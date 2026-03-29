package usage

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, userResp map[string]string, usageResp *premiumRequestUsageResponse, usageStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/user":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(userResp)
		case strings.HasPrefix(r.URL.Path, "/users/") && strings.HasSuffix(r.URL.Path, "/settings/billing/premium_request/usage"):
			w.Header().Set("Content-Type", "application/json")
			if usageStatus != 0 && usageStatus != http.StatusOK {
				w.WriteHeader(usageStatus)
				_, _ = w.Write([]byte(`{"message":"Not Found"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(usageResp)
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"unknown path"}`))
		}
	}))
}

func newTestOpts(serverURL string) *usageOptions {
	return &usageOptions{
		apiBase:    serverURL,
		httpClient: http.DefaultClient,
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
		require.Contains(t, output, "testuser")
		require.Contains(t, output, "2026-03")
		require.Contains(t, output, "Claude Opus 4.6")
		require.Contains(t, output, "GPT-5.2")
		require.Contains(t, output, "100.0")
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

		server := newTestServer(t, map[string]string{"login": "testuser"}, usageResp, 0)
		defer server.Close()

		buf := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(buf, buf, client, true, 120)
		cfg.Token = "test-token"

		cmd := newUsageCommand(cfg, newTestOpts(server.URL))
		cmd.SetArgs([]string{"--today"})

		_, err := cmd.ExecuteC()
		require.NoError(t, err)

		output := buf.String()
		require.Contains(t, output, "today")
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
		require.True(t, opusIdx < gptIdx, "Claude Opus 4.6 should appear before GPT-5.2 (sorted by gross amount)")
	})
}
