package azuremodels

import (
	"context"
	"errors"
)

// MockClient provides a client for interacting with the Azure models API in tests.
type MockClient struct {
	MockGetChatCompletionStream func(context.Context, ChatCompletionOptions, string, string) (*ChatCompletionResponse, error)
	MockGetModelDetails         func(context.Context, string, string, string) (*ModelDetails, error)
	MockListModels              func(context.Context) ([]*ModelSummary, error)
}

// NewMockClient returns a new mock client for stubbing out interactions with the models API.
func NewMockClient() *MockClient {
	return &MockClient{
		MockGetChatCompletionStream: func(context.Context, ChatCompletionOptions, string, string) (*ChatCompletionResponse, error) {
			return nil, errors.New("GetChatCompletionStream not implemented")
		},
		MockGetModelDetails: func(context.Context, string, string, string) (*ModelDetails, error) {
			return nil, errors.New("GetModelDetails not implemented")
		},
		MockListModels: func(context.Context) ([]*ModelSummary, error) {
			return nil, errors.New("ListModels not implemented")
		},
	}
}

// GetChatCompletionStream calls the mocked function for getting a stream of chat completions for the given request.
func (c *MockClient) GetChatCompletionStream(ctx context.Context, opt ChatCompletionOptions, org, httpLogFile string) (*ChatCompletionResponse, error) {
	return c.MockGetChatCompletionStream(ctx, opt, org, httpLogFile)
}

// GetModelDetails calls the mocked function for getting the details of the specified model in a particular registry.
func (c *MockClient) GetModelDetails(ctx context.Context, registry, modelName, version string) (*ModelDetails, error) {
	return c.MockGetModelDetails(ctx, registry, modelName, version)
}

// ListModels calls the mocked function for getting a list of available models.
func (c *MockClient) ListModels(ctx context.Context) ([]*ModelSummary, error) {
	return c.MockListModels(ctx)
}
