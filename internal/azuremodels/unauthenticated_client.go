package azuremodels

import (
	"context"
	"errors"
)

// UnauthenticatedClient is for use by anonymous viewers to talk to the models API.
type UnauthenticatedClient struct {
}

// NewUnauthenticatedClient contructs a new models API client for an anonymous viewer.
func NewUnauthenticatedClient() *UnauthenticatedClient {
	return &UnauthenticatedClient{}
}

// GetChatCompletionStream returns an error because this functionality requires authentication.
func (c *UnauthenticatedClient) GetChatCompletionStream(ctx context.Context, opt ChatCompletionOptions) (*ChatCompletionResponse, error) {
	return nil, errors.New("not authenticated")
}

// GetModelDetails returns an error because this functionality requires authentication.
func (c *UnauthenticatedClient) GetModelDetails(ctx context.Context, registry, modelName, version string) (*ModelDetails, error) {
	return nil, errors.New("not authenticated")
}

// ListModels returns an error because this functionality requires authentication.
func (c *UnauthenticatedClient) ListModels(ctx context.Context) ([]*ModelSummary, error) {
	return nil, errors.New("not authenticated")
}
