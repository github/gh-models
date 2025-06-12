package azuremodels

import "context"

// Client represents a client for interacting with an API about models.
type Client interface {
	// GetChatCompletionStream returns a stream of chat completions using the given options.
	GetChatCompletionStream(context.Context, ChatCompletionOptions, string) (*ChatCompletionResponse, error)
	// GetModelDetails returns the details of the specified model in a particular registry.
	GetModelDetails(ctx context.Context, registry, modelName, version string) (*ModelDetails, error)
	// ListModels returns a list of available models.
	ListModels(context.Context) ([]*ModelSummary, error)
}
