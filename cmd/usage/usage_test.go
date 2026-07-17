package usage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, userResp map[string]string, usageResp *premiumRequestUsageResponse, usageStatus int, assertUsageRequest ...func(*http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Match exact paths so tests fail loudly if the command starts
		// hitting an unexpected endpoint, instead of silently returning
		// canned responses for any matching prefix/suffix.
		switch {
		case r.URL.Path == "/user":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(userResp)
		case r.URL.Path == "/users/testuser/settings/billing/premium_request/usage":
			if len(assertUsageRequest) > 0 {
				assertUsageRequest[0](r)
			}
			w.Header().Set("Content-Type", "application/json")
			if usageStatus != 0 && usageStatus != http.StatusOK {
				w.WriteHeader(usageStatus)
				_, _ = w.Write([]byte(`{"message":"Not Found"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(usageResp)
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"unknown path: ` + r.URL.Path + `"}`))
		}
	}))
}

func newTestOpts(serverURL string) *usageOptions {
	return &usageOptions{
		apiBase:    serverURL,
		httpClient: http.DefaultClient,
		now:        time.Now,
	}
}

func TestUsageCommand(t *testing.T) {
	t.Run("shows usage table", func(t *testing.T) {
		usageResp := &premiumRequestUsageResponse{
			TimePeriod: timePeriod{Year: 2026, Month: 3},
			User:       "testuser",
			UsageItems: []usageItem{
				{
					Product:       "Copilot",
					Model:         "Claude Opus 4.6",
					GrossQuantity: 100,
					GrossAmount:   4.0,
					NetAmount:     0.0,
				},
				{
					Product:       "Copilot",
					Model:         "GPT-5.2",
					GrossQuantity: 50,
					GrossAmount:   2.0,
					NetAmount:     0.0,
				},
			},
		}

		server := newTestServer(
			t,
			map[string]string{"login": "testuser"},
			usageResp,
			0,
			func(r *http.Request) {
				query := r.URL.Query()
				require.Equal(t, "2026", query.Get("year"))
				require.Equal(t, "3", query.Get("month"))
				require.Empty(t, query.Get("day"))
			},
		)
		defer server.Close()

		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, newTestOpts(server.URL))
		cmd.SetArgs([]string{"--year", "2026", "--month", "3"})

		_, err := cmd.ExecuteC()
		require.NoError(t, err)

		output := buf.String()
		require.Contains(t, output, "testuser")
		require.Contains(t, output, "2026-03")
		require.Contains(t, output, "Claude Opus 4.6")
		require.Contains(t, output, "GPT-5.2")
		require.Contains(t, output, "100")
		require.Contains(t, output, "$4.00")
		require.Contains(t, output, "Total: 150 requests")
		require.Contains(t, output, "100% discount")
	})

	t.Run("shows net cost when applicable", func(t *testing.T) {
		usageResp := &premiumRequestUsageResponse{
			TimePeriod: timePeriod{Year: 2026, Month: 3},
			User:       "testuser",
			UsageItems: []usageItem{
				{
					Product:       "Copilot",
					Model:         "Claude Opus 4.6",
					GrossQuantity: 100,
					GrossAmount:   4.0,
					NetAmount:     2.0,
				},
			},
		}

		server := newTestServer(t, map[string]string{"login": "testuser"}, usageResp, 0)
		defer server.Close()

		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, newTestOpts(server.URL))
		cmd.SetArgs([]string{"--year", "2026", "--month", "3"})

		_, err := cmd.ExecuteC()
		require.NoError(t, err)

		output := buf.String()
		require.Contains(t, output, "Net cost: $2.00")
	})

	t.Run("preserves fractional request quantities", func(t *testing.T) {
		usageResp := &premiumRequestUsageResponse{
			TimePeriod: timePeriod{Year: 2026, Month: 3},
			User:       "testuser",
			UsageItems: []usageItem{
				{
					Product:       "Copilot",
					Model:         "Fractional model",
					GrossQuantity: 0.25,
					GrossAmount:   0.01,
				},
			},
		}

		server := newTestServer(t, map[string]string{"login": "testuser"}, usageResp, 0)
		defer server.Close()

		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, newTestOpts(server.URL))
		cmd.SetArgs([]string{"--year", "2026", "--month", "3"})

		_, err := cmd.ExecuteC()
		require.NoError(t, err)

		output := buf.String()
		require.Contains(t, output, "0.25")
		require.Contains(t, output, "Total: 0.25 requests")
	})

	t.Run("shows no usage message when empty", func(t *testing.T) {
		usageResp := &premiumRequestUsageResponse{
			TimePeriod: timePeriod{Year: 2026, Month: 1},
			User:       "testuser",
			UsageItems: []usageItem{},
		}

		server := newTestServer(t, map[string]string{"login": "testuser"}, usageResp, 0)
		defer server.Close()

		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, newTestOpts(server.URL))
		cmd.SetArgs([]string{"--year", "2026", "--month", "1"})

		_, err := cmd.ExecuteC()
		require.NoError(t, err)

		output := buf.String()
		require.Contains(t, output, "No usage found")
	})

	t.Run("errors without token", func(t *testing.T) {
		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = ""

		cmd := newUsageCommand(cfg, nil)
		cmd.SetArgs([]string{})

		_, err := cmd.ExecuteC()
		require.Error(t, err)
		require.Contains(t, err.Error(), "no GitHub token found")
	})

	t.Run("404 returns helpful scope error", func(t *testing.T) {
		server := newTestServer(t, map[string]string{"login": "testuser"}, nil, http.StatusNotFound)
		defer server.Close()

		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, newTestOpts(server.URL))
		cmd.SetArgs([]string{"--year", "2026", "--month", "3"})

		_, err := cmd.ExecuteC()
		require.Error(t, err)
		require.Contains(t, err.Error(), "HTTP 404")
		require.Contains(t, err.Error(), "gh auth refresh")
	})

	t.Run("--today flag works", func(t *testing.T) {
		fixedNow := time.Date(2026, time.March, 29, 23, 59, 59, 0, time.UTC)
		usageResp := &premiumRequestUsageResponse{
			TimePeriod: timePeriod{Year: 2026, Month: 3, Day: 29},
			User:       "testuser",
			UsageItems: []usageItem{
				{
					Product:       "Copilot",
					Model:         "Claude Opus 4.6",
					GrossQuantity: 5,
					GrossAmount:   0.20,
					NetAmount:     0.0,
				},
			},
		}

		server := newTestServer(
			t,
			map[string]string{"login": "testuser"},
			usageResp,
			0,
			func(r *http.Request) {
				query := r.URL.Query()
				require.Equal(t, fmt.Sprintf("%d", fixedNow.Year()), query.Get("year"))
				require.Equal(t, fmt.Sprintf("%d", fixedNow.Month()), query.Get("month"))
				require.Equal(t, fmt.Sprintf("%d", fixedNow.Day()), query.Get("day"))
			},
		)
		defer server.Close()

		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		opts := newTestOpts(server.URL)
		opts.now = func() time.Time { return fixedNow }
		cmd := newUsageCommand(cfg, opts)
		cmd.SetArgs([]string{"--today"})

		_, err := cmd.ExecuteC()
		require.NoError(t, err)

		output := buf.String()
		require.Contains(t, output, "today")
	})

	t.Run("sends explicit day query", func(t *testing.T) {
		usageResp := &premiumRequestUsageResponse{
			TimePeriod: timePeriod{Year: 2026, Month: 3, Day: 15},
			User:       "testuser",
			UsageItems: []usageItem{},
		}

		server := newTestServer(
			t,
			map[string]string{"login": "testuser"},
			usageResp,
			0,
			func(r *http.Request) {
				query := r.URL.Query()
				require.Equal(t, "2026", query.Get("year"))
				require.Equal(t, "3", query.Get("month"))
				require.Equal(t, "15", query.Get("day"))
			},
		)
		defer server.Close()

		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, newTestOpts(server.URL))
		cmd.SetArgs([]string{"--year", "2026", "--month", "3", "--day", "15"})

		_, err := cmd.ExecuteC()
		require.NoError(t, err)
	})

	t.Run("--help prints usage info", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		cmd := NewUsageCommand(nil)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--help"})

		err := cmd.Help()

		require.NoError(t, err)
		require.Contains(t, outBuf.String(), "Display premium request usage statistics")
		require.Empty(t, errBuf.String())
	})

	t.Run("sorts by gross amount descending", func(t *testing.T) {
		usageResp := &premiumRequestUsageResponse{
			TimePeriod: timePeriod{Year: 2026, Month: 3},
			User:       "testuser",
			UsageItems: []usageItem{
				{
					Product:       "Copilot",
					Model:         "GPT-5.2",
					GrossQuantity: 10,
					GrossAmount:   0.40,
					NetAmount:     0.0,
				},
				{
					Product:       "Copilot",
					Model:         "Claude Opus 4.6",
					GrossQuantity: 100,
					GrossAmount:   4.0,
					NetAmount:     0.0,
				},
			},
		}

		server := newTestServer(t, map[string]string{"login": "testuser"}, usageResp, 0)
		defer server.Close()

		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, newTestOpts(server.URL))
		cmd.SetArgs([]string{"--year", "2026", "--month", "3"})

		_, err := cmd.ExecuteC()
		require.NoError(t, err)

		output := buf.String()
		opusIdx := bytes.Index([]byte(output), []byte("Claude Opus"))
		gptIdx := bytes.Index([]byte(output), []byte("GPT-5.2"))
		require.GreaterOrEqual(t, opusIdx, 0, "Claude Opus should be present in the output")
		require.GreaterOrEqual(t, gptIdx, 0, "GPT-5.2 should be present in the output")
		require.True(t, opusIdx < gptIdx, "Claude Opus 4.6 should appear before GPT-5.2 (sorted by gross amount)")
	})

	t.Run("net cost with zero gross does not produce NaN/Inf", func(t *testing.T) {
		usageResp := &premiumRequestUsageResponse{
			TimePeriod: timePeriod{Year: 2026, Month: 3},
			User:       "testuser",
			UsageItems: []usageItem{
				{
					Product:       "Copilot",
					Model:         "Edge case model",
					GrossQuantity: 1,
					GrossAmount:   0.0,
					NetAmount:     1.50,
				},
			},
		}

		server := newTestServer(t, map[string]string{"login": "testuser"}, usageResp, 0)
		defer server.Close()

		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, newTestOpts(server.URL))
		cmd.SetArgs([]string{"--year", "2026", "--month", "3"})

		_, err := cmd.ExecuteC()
		require.NoError(t, err)

		output := buf.String()
		require.NotContains(t, output, "NaN")
		require.NotContains(t, output, "+Inf")
		require.NotContains(t, output, "-Inf")
		require.Contains(t, output, "gross amount unavailable")
	})

	t.Run("--today is mutually exclusive with date flags", func(t *testing.T) {
		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, nil)
		cmd.SetArgs([]string{"--today", "--year", "2026"})

		_, err := cmd.ExecuteC()
		require.Error(t, err)
		require.Contains(t, err.Error(), "none of the others can be")
	})

	t.Run("rejects invalid month", func(t *testing.T) {
		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, nil)
		cmd.SetArgs([]string{"--year", "2026", "--month", "13"})

		_, err := cmd.ExecuteC()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid value for --month")
	})

	t.Run("rejects non-four-digit year", func(t *testing.T) {
		for _, year := range []string{"1", "10000"} {
			buf := new(bytes.Buffer)
			client := azuremodels.NewMockClient()
			cfg := command.NewConfig(buf, buf, client, true, 120)
			cfg.Token = "test-token"

			cmd := newUsageCommand(cfg, nil)
			cmd.SetArgs([]string{"--year", year, "--month", "3"})

			_, err := cmd.ExecuteC()
			require.Error(t, err)
			require.Contains(t, err.Error(), "must be a four-digit year")
		}
	})

	t.Run("rejects invalid day", func(t *testing.T) {
		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, nil)
		cmd.SetArgs([]string{"--year", "2026", "--month", "3", "--day", "99"})

		_, err := cmd.ExecuteC()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid value for --day")
	})
}
