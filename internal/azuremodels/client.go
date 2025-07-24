package azuremodels

import "context"

// Client represents a client for interacting with an API about models.
type Client interface {
	// GetChatCompletionStream returns a stream of chat completions using the given options.
	// The httpLogFile parameter, if non-empty, specifies the file to log HTTP requests to.
	GetChatCompletionStream(ctx context.Context, req ChatCompletionOptions, org, httpLogFile string) (*ChatCompletionResponse, error)
	// GetModelDetails returns the details of the specified model in a particular registry.
	GetModelDetails(ctx context.Context, registry, modelName, version string) (*ModelDetails, error)
	// ListModels returns a list of available models.
	ListModels(context.Context) ([]*ModelSummary, error)
}
