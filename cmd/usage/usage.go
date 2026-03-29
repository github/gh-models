// Package usage provides a gh command to show GitHub Models and Copilot usage information.
package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/github/gh-models/pkg/command"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
)

const defaultGitHubAPIBase = "https://api.github.com"

var (
	headerColor = ansi.ColorFunc("white+du")
	greenColor  = ansi.ColorFunc("green")
	yellowColor = ansi.ColorFunc("yellow")
)

// usageOptions holds configuration for the usage command, allowing dependency injection for testing.
type usageOptions struct {
	apiBase    string
	httpClient *http.Client
}

// premiumRequestUsageResponse represents the API response for premium request usage.
type premiumRequestUsageResponse struct {
	TimePeriod timePeriod  `json:"timePeriod"`
	User       string      `json:"user"`
	UsageItems []usageItem `json:"usageItems"`
}

type timePeriod struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day,omitempty"`
}

type usageItem struct {
	Product          string  `json:"product"`
	SKU              string  `json:"sku"`
	Model            string  `json:"model"`
	UnitType         string  `json:"unitType"`
	PricePerUnit     float64 `json:"pricePerUnit"`
	GrossQuantity    float64 `json:"grossQuantity"`
	GrossAmount      float64 `json:"grossAmount"`
	DiscountQuantity float64 `json:"discountQuantity"`
	DiscountAmount   float64 `json:"discountAmount"`
	NetQuantity      float64 `json:"netQuantity"`
	NetAmount        float64 `json:"netAmount"`
}

func defaultOptions() *usageOptions {
	return &usageOptions{
		apiBase: defaultGitHubAPIBase,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewUsageCommand returns a new command to show model usage information.
func NewUsageCommand(cfg *command.Config) *cobra.Command {
	return newUsageCommand(cfg, nil)
}

// newUsageCommand is the internal constructor that accepts options for testing.
func newUsageCommand(cfg *command.Config, opts *usageOptions) *cobra.Command {
	if opts == nil {
		opts = defaultOptions()
	}

	var (
		flagMonth int
		flagYear  int
		flagDay   int
		flagToday bool
	)

	cmd := &cobra.Command{
		Use:   "usage",
		Short: "Show premium request usage and costs",
		Long: heredoc.Docf(`
			Display premium request usage statistics for GitHub Models and Copilot.

			Shows a breakdown of requests by model, with gross and net costs.
			By default, shows usage for the current billing period (month).

			Use %[1]s--today%[1]s to see only today's usage, or %[1]s--month%[1]s and
			%[1]s--year%[1]s to query a specific billing period.

			Requires the %[1]suser%[1]s scope on your GitHub token. If you get a 404 error,
			run: %[1]sgh auth refresh -h github.com -s user%[1]s
		`, "`"),
		Example: heredoc.Doc(`
			# Show current month's usage
			$ gh models usage

			# Show today's usage
			$ gh models usage --today

			# Show usage for a specific month
			$ gh models usage --year 2026 --month 2

			# Show usage for a specific day
			$ gh models usage --year 2026 --month 3 --day 15
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			token := cfg.Token
			if token == "" {
				return fmt.Errorf("no GitHub token found. Please run 'gh auth login' to authenticate")
			}

			ctx := cmd.Context()

			// Resolve time period
			now := time.Now().UTC()
			year := flagYear
			month := flagMonth
			day := flagDay

			if flagToday {
				year = now.Year()
				month = int(now.Month())
				day = now.Day()
			}

			if year == 0 {
				year = now.Year()
			}
			if month == 0 {
				month = int(now.Month())
			}

			// Get username
			username, err := getUsername(ctx, opts, token)
			if err != nil {
				return fmt.Errorf("failed to get username: %w", err)
			}

			// Build query params
			query := fmt.Sprintf("?year=%d&month=%d", year, month)
			if day > 0 {
				query += fmt.Sprintf("&day=%d", day)
			}

			// Fetch usage
			data, err := fetchPremiumRequestUsage(ctx, opts, token, username, query)
			if err != nil {
				return err
			}

			// Format period string
			periodStr := fmt.Sprintf("%d-%02d", data.TimePeriod.Year, data.TimePeriod.Month)
			if data.TimePeriod.Day > 0 {
				periodStr += fmt.Sprintf("-%02d", data.TimePeriod.Day)
			}

			if len(data.UsageItems) == 0 {
				cfg.WriteToOut(fmt.Sprintf("\nNo usage found for %s\n\n", periodStr))
				return nil
			}

			// Sort by gross amount descending
			items := data.UsageItems
			sort.Slice(items, func(i, j int) bool {
				return items[i].GrossAmount > items[j].GrossAmount
			})

			// Calculate totals
			var totalReqs, totalGross, totalNet float64
			for _, item := range items {
				totalReqs += item.GrossQuantity
				totalGross += item.GrossAmount
				totalNet += item.NetAmount
			}

			// Print header
			if cfg.IsTerminalOutput {
				cfg.WriteToOut("\n")
				if flagToday {
					cfg.WriteToOut(fmt.Sprintf("Premium request usage for %s (%s, today)\n", username, periodStr))
				} else {
					cfg.WriteToOut(fmt.Sprintf("Premium request usage for %s (%s)\n", username, periodStr))
				}
				cfg.WriteToOut("\n")
			}

			// Print table
			printer := cfg.NewTablePrinter()

			printer.AddHeader([]string{"PRODUCT", "MODEL", "REQUESTS", "GROSS", "NET"}, tableprinter.WithColor(headerColor))
			printer.EndRow()

			for _, item := range items {
				if item.GrossQuantity == 0 {
					continue
				}
				printer.AddField(item.Product)
				printer.AddField(item.Model)
				printer.AddField(fmt.Sprintf("%.1f", item.GrossQuantity))
				printer.AddField(fmt.Sprintf("$%.2f", item.GrossAmount))
				printer.AddField(fmt.Sprintf("$%.2f", item.NetAmount))
				printer.EndRow()
			}

			if err := printer.Render(); err != nil {
				return err
			}

			// Print summary
			if cfg.IsTerminalOutput {
				cfg.WriteToOut("\n")
				cfg.WriteToOut(fmt.Sprintf("Total: %.0f requests, $%.2f gross, $%.2f net\n", totalReqs, totalGross, totalNet))

				if totalGross > 0 && totalNet == 0 {
					cfg.WriteToOut(greenColor("All usage included in your plan (100% discount)") + "\n")
				} else if totalNet > 0 {
					pct := (totalNet / totalGross) * 100
					cfg.WriteToOut(yellowColor(fmt.Sprintf("Net cost: $%.2f (%.0f%% of gross)", totalNet, pct)) + "\n")
				}
				cfg.WriteToOut("\n")
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&flagYear, "year", 0, "Filter by year (default: current year)")
	cmd.Flags().IntVar(&flagMonth, "month", 0, "Filter by month (default: current month)")
	cmd.Flags().IntVar(&flagDay, "day", 0, "Filter by specific day")
	cmd.Flags().BoolVar(&flagToday, "today", false, "Show only today's usage")

	return cmd
}

func getUsername(ctx context.Context, opts *usageOptions, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, opts.apiBase+"/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := opts.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get user info: HTTP %d", resp.StatusCode)
	}

	var user struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}
	return user.Login, nil
}

func fetchPremiumRequestUsage(ctx context.Context, opts *usageOptions, token, username, query string) (*premiumRequestUsageResponse, error) {
	url := fmt.Sprintf("%s/users/%s/settings/billing/premium_request/usage%s", opts.apiBase, username, query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := opts.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("usage data not available (HTTP 404). You may need the 'user' scope.\nRun: gh auth refresh -h github.com -s user\n\nResponse: %s", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch usage data: HTTP %d\n%s", resp.StatusCode, string(body))
	}

	var data premiumRequestUsageResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse usage response: %w", err)
	}

	return &data, nil
}
