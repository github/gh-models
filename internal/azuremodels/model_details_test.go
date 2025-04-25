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

	t.Run("FormatIdentifier", func(t *testing.T) {
		publisher := "Open AI"
		name := "GPT 3"
		expected := "open-ai/gpt-3"
		result := FormatIdentifier(publisher, name)
		require.Equal(t, expected, result)
	})
}
