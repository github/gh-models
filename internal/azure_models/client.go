// Package azure_models provides a client for interacting with the Azure models API.
package azure_models

import (
	"bytes"
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

type Client struct {
	client *http.Client
	token  string
}

const (
	prodInferenceURL = "https://models.inference.ai.azure.com/chat/completions"
	azureAiStudioURL = "https://api.catalog.azureml.ms"
	prodModelsURL    = azureAiStudioURL + "/asset-gallery/v1.0/models"
)

// NewClient returns a new client using the given auth token.
func NewClient(authToken string) *Client {
	httpClient, _ := api.DefaultHTTPClient()
	return &Client{
		client: httpClient,
		token:  authToken,
	}
}

// GetChatCompletionStream returns a stream of chat completions for the given request.
func (c *Client) GetChatCompletionStream(req ChatCompletionOptions) (*ChatCompletionResponse, error) {
	// Check if the model name is `o1-mini` or `o1-preview`
	if req.Model == "o1-mini" || req.Model == "o1-preview" {
		req.Stream = false
	} else {
		req.Stream = true
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(bodyBytes)

	httpReq, err := http.NewRequest("POST", prodInferenceURL, body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")

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
		chatCompletionResponse.Reader = sse.NewEventReader[chatCompletion](resp.Body)
	} else {
		var completion chatCompletion
		if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
			return nil, err
		}

		// Create a mock reader that returns the decoded completion
		mockReader := sse.NewMockEventReader([]chatCompletion{completion})
		chatCompletionResponse.Reader = mockReader
	}

	return &chatCompletionResponse, nil
}

// GetModelDetails returns the details of the specified model in a prticular registry.
func (c *Client) GetModelDetails(registry string, modelName string, version string) (*ModelDetails, error) {
	url := fmt.Sprintf("%s/asset-gallery/v1.0/%s/models/%s/version/%s", azureAiStudioURL, registry, modelName, version)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

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
func (c *Client) ListModels() ([]*ModelSummary, error) {
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

	httpReq, err := http.NewRequest("POST", prodModelsURL, body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

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

func (c *Client) handleHTTPError(resp *http.Response) error {
	sb := strings.Builder{}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		sb.WriteString("unauthorized")

	case http.StatusBadRequest:
		sb.WriteString("bad request")

	default:
		sb.WriteString("unexpected response from the server: " + resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) > 0 {
		sb.WriteString("\n")
		sb.Write(body)
		sb.WriteString("\n")
	}

	return errors.New(sb.String())
}
