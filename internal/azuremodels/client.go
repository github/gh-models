package azuremodels

import (
	"context"
	"os"
)

// httpLogFileKey is the context key for the HTTP log filename
type httpLogFileKey struct{}

// WithHTTPLogFile returns a new context with the HTTP log filename attached
func WithHTTPLogFile(ctx context.Context, httpLogFile string) context.Context {
	// reset http-log file
	if httpLogFile != "" {
		_ = os.Remove(httpLogFile)
	}
	return context.WithValue(ctx, httpLogFileKey{}, httpLogFile)
}

// HTTPLogFileFromContext returns the HTTP log filename from the context, if any
func HTTPLogFileFromContext(ctx context.Context) string {
	if httpLogFile, ok := ctx.Value(httpLogFileKey{}).(string); ok {
		return httpLogFile
	}
	return ""
}

// Client represents a client for interacting with an API about models.
type Client interface {
	// GetChatCompletionStream returns a stream of chat completions using the given options.
	// HTTP logging configuration is extracted from the context if present.
	GetChatCompletionStream(ctx context.Context, req ChatCompletionOptions, org string) (*ChatCompletionResponse, error)
	// GetModelDetails returns the details of the specified model in a particular registry.
	GetModelDetails(ctx context.Context, registry, modelName, version string) (*ModelDetails, error)
	// ListModels returns a list of available models.
	ListModels(context.Context) ([]*ModelSummary, error)
}
