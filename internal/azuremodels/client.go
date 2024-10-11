package azuremodels

import "context"

type Client interface {
	GetChatCompletionStream(context.Context, ChatCompletionOptions) (*ChatCompletionResponse, error)
	GetModelDetails(context.Context, string, string, string) (*ModelDetails, error)
	ListModels(context.Context) ([]*ModelSummary, error)
}
