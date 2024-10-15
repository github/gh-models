package azuremodels

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAzureClient(t *testing.T) {
	ctx := context.Background()
	t.Run("GetModelDetails", func(t *testing.T) {
		token := "fake-token-123abc"
		cfg := &AzureClientConfig{}
		client := NewAzureClient(token, cfg)
		registry := "foo"
		modelName := "bar"
		version := "baz"

		details, err := client.GetModelDetails(ctx, registry, modelName, version)

		require.NoError(t, err)
		require.NotNil(t, details)
	})
}
