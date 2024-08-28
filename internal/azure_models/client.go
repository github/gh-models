package azure_models

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/github/gh-models/internal/sse"
)

type Client struct {
	client *http.Client
	token  string
}

const (
	prodInferenceURL = "https://models.inference.ai.azure.com/chat/completions"
	prodModelsURL    = "https://models.inference.ai.azure.com/models"
)

func NewClient(authToken string) *Client {
	httpClient, _ := api.DefaultHTTPClient()
	return &Client{
		client: httpClient,
		token:  authToken,
	}
}

func (c *Client) GetChatCompletionStream(req ChatCompletionOptions) (*ChatCompletionResponse, error) {
	req.Stream = true

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
	chatCompletionResponse.Reader = sse.NewEventReader[ChatCompletion](resp.Body)

	return &chatCompletionResponse, nil
}

func (c *Client) ListModels() ([]*ModelSummary, error) {
	httpReq, err := http.NewRequest("GET", prodModelsURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp)
	}

	var models []*ModelSummary
	err = json.NewDecoder(resp.Body).Decode(&models)
	if err != nil {
		return nil, err
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
