// Package azuremodels provides a client for interacting with the Azure models API.
package azuremodels

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/github/gh-models/internal/sse"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// AzureClient provides a client for interacting with the Azure models API.
type AzureClient struct {
	client *http.Client
	token  string
	cfg    *AzureClientConfig
}

// NewDefaultAzureClient returns a new Azure client using the given auth token using default API URLs.
func NewDefaultAzureClient(authToken string) (*AzureClient, error) {
	httpClient, err := api.DefaultHTTPClient()
	if err != nil {
		return nil, err
	}
	cfg := NewDefaultAzureClientConfig()
	return &AzureClient{client: httpClient, token: authToken, cfg: cfg}, nil
}

// NewAzureClient returns a new Azure client using the given HTTP client, configuration, and auth token.
func NewAzureClient(httpClient *http.Client, authToken string, cfg *AzureClientConfig) *AzureClient {
	return &AzureClient{client: httpClient, token: authToken, cfg: cfg}
}

// GetChatCompletionStream returns a stream of chat completions using the given options.
func (c *AzureClient) GetChatCompletionStream(ctx context.Context, req ChatCompletionOptions, org string) (*ChatCompletionResponse, error) {
	// Check for o1 models, which don't support streaming
	if req.Model == "o1-mini" || req.Model == "o1-preview" || req.Model == "o1" {
		req.Stream = false
	} else {
		req.Stream = true
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(bodyBytes)

	var inferenceURL string
	if org != "" {
		inferenceURL = fmt.Sprintf("%s/orgs/%s/%s", c.cfg.InferenceRoot, org, c.cfg.InferencePath)
	} else {
		inferenceURL = c.cfg.InferenceRoot + "/" + c.cfg.InferencePath
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, inferenceURL, body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")

	// Azure would like us to send specific user agents to help distinguish
	// traffic from known sources and other web requests
	httpReq.Header.Set("x-ms-useragent", "github-cli-models")
	httpReq.Header.Set("x-ms-user-agent", "github-cli-models") // send both to accommodate various Azure consumers

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		// If we aren't going to return an SSE stream, then ensure the response body is closed.
		defer resp.Body.Close()
		return nil, c.handleHTTPError(resp)
	}

	var chatCompletionResponse ChatCompletionResponse

	if req.Stream {
		// Handle streamed response
		chatCompletionResponse.Reader = sse.NewEventReader[ChatCompletion](resp.Body)
	} else {
		var completion ChatCompletion
		if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
			return nil, err
		}

		// Create a mock reader that returns the decoded completion
		mockReader := sse.NewMockEventReader([]ChatCompletion{completion})
		chatCompletionResponse.Reader = mockReader
	}

	return &chatCompletionResponse, nil
}

// GetModelDetails returns the details of the specified model in a particular registry.
func (c *AzureClient) GetModelDetails(ctx context.Context, registry, modelName, version string) (*ModelDetails, error) {
	url := fmt.Sprintf("%s/asset-gallery/v1.0/%s/models/%s/version/%s", c.cfg.AzureAiStudioURL, registry, modelName, version)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp)
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	var detailsResponse modelCatalogDetailsResponse
	err = decoder.Decode(&detailsResponse)
	if err != nil {
		return nil, err
	}

	modelDetails := &ModelDetails{
		Description:        detailsResponse.Description,
		License:            detailsResponse.License,
		LicenseDescription: detailsResponse.LicenseDescription,
		Notes:              detailsResponse.Notes,
		Tags:               lowercaseStrings(detailsResponse.Keywords),
		Evaluation:         detailsResponse.Evaluation,
	}

	modelLimits := detailsResponse.ModelLimits
	if modelLimits != nil {
		modelDetails.SupportedInputModalities = modelLimits.SupportedInputModalities
		modelDetails.SupportedOutputModalities = modelLimits.SupportedOutputModalities
		modelDetails.SupportedLanguages = convertLanguageCodesToNames(modelLimits.SupportedLanguages)

		textLimits := modelLimits.TextLimits
		if textLimits != nil {
			modelDetails.MaxOutputTokens = textLimits.MaxOutputTokens
			modelDetails.MaxInputTokens = textLimits.InputContextWindow
		}
	}

	playgroundLimits := detailsResponse.PlaygroundLimits
	if playgroundLimits != nil {
		modelDetails.RateLimitTier = playgroundLimits.RateLimitTier
	}

	return modelDetails, nil
}

func convertLanguageCodesToNames(input []string) []string {
	output := make([]string, len(input))
	english := display.English.Languages()
	for i, code := range input {
		tag := language.MustParse(code)
		output[i] = english.Name(tag)
	}
	return output
}

func lowercaseStrings(input []string) []string {
	output := make([]string, len(input))
	for i, s := range input {
		output[i] = strings.ToLower(s)
	}
	return output
}

// ListModels returns a list of available models.
func (c *AzureClient) ListModels(ctx context.Context) ([]*ModelSummary, error) {
	body := bytes.NewReader([]byte(`
		{
			"filters": [
				{ "field": "freePlayground", "values": ["true"], "operator": "eq"},
				{ "field": "labels", "values": ["latest"], "operator": "eq"}
			],
			"order": [
				{ "field": "displayName", "direction": "asc" }
			]
		}
	`))

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.ModelsURL, body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp)
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	var searchResponse modelCatalogSearchResponse
	err = decoder.Decode(&searchResponse)
	if err != nil {
		return nil, err
	}

	models := make([]*ModelSummary, 0, len(searchResponse.Summaries))
	for _, summary := range searchResponse.Summaries {
		inferenceTask := ""
		if len(summary.InferenceTasks) > 0 {
			inferenceTask = summary.InferenceTasks[0]
		}

		models = append(models, &ModelSummary{
			ID:           summary.AssetID,
			Name:         summary.Name,
			FriendlyName: summary.DisplayName,
			Task:         inferenceTask,
			Publisher:    summary.Publisher,
			Summary:      summary.Summary,
			Version:      summary.Version,
			RegistryName: summary.RegistryName,
		})
	}

	return models, nil
}

func (c *AzureClient) handleHTTPError(resp *http.Response) error {
	sb := strings.Builder{}
	var err error

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		_, err = sb.WriteString("unauthorized")
		if err != nil {
			return err
		}

	case http.StatusBadRequest:
		_, err = sb.WriteString("bad request")
		if err != nil {
			return err
		}

	default:
		_, err = sb.WriteString("unexpected response from the server: " + resp.Status)
		if err != nil {
			return err
		}
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) > 0 {
		_, err = sb.WriteString("\n")
		if err != nil {
			return err
		}

		_, err = sb.Write(body)
		if err != nil {
			return err
		}

		_, err = sb.WriteString("\n")
		if err != nil {
			return err
		}
	}

	return errors.New(sb.String())
}
