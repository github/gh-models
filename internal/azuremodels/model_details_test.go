package azuremodels

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelDetails(t *testing.T) {
	t.Run("ContextLimits", func(t *testing.T) {
		details := &ModelDetails{MaxInputTokens: 123, MaxOutputTokens: 456}
		result := details.ContextLimits()
		require.Equal(t, "up to 123 input tokens and 456 output tokens", result)
	})
}
