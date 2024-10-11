package azuremodels

import (
	"context"
)

// MockClient provides a client for interacting with the Azure models API in tests.
type MockClient struct {
	MockGetChatCompletionStream func(context.Context, ChatCompletionOptions) (*ChatCompletionResponse, error)
	MockGetModelDetails         func(context.Context, string, string, string) (*ModelDetails, error)
	MockListModels              func(context.Context) ([]*ModelSummary, error)
}

// NewMockClient returns a new mock client for stubbing out interactions with the models API.
func NewMockClient() *MockClient {
	return &MockClient{
		MockGetChatCompletionStream: func(context.Context, ChatCompletionOptions) (*ChatCompletionResponse, error) {
			return nil, nil
		},
		MockGetModelDetails: func(context.Context, string, string, string) (*ModelDetails, error) {
			return nil, nil
		},
		MockListModels: func(context.Context) ([]*ModelSummary, error) {
			return nil, nil
		},
	}
}

// GetChatCompletionStream calls the mocked function for getting a stream of chat completions for the given request.
func (c *MockClient) GetChatCompletionStream(ctx context.Context, opt ChatCompletionOptions) (*ChatCompletionResponse, error) {
	return c.MockGetChatCompletionStream(ctx, opt)
}

// GetModelDetails calls the mocked function for getting the details of the specified model in a particular registry.
func (c *MockClient) GetModelDetails(ctx context.Context, registry, modelName, version string) (*ModelDetails, error) {
	return c.MockGetModelDetails(ctx, registry, modelName, version)
}

// ListModels calls the mocked function for getting a list of available models.
func (c *MockClient) ListModels(ctx context.Context) ([]*ModelSummary, error) {
	return c.MockListModels(ctx)
}
