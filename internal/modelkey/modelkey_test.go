package modelkey

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseModelKey(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *ModelKey
		expectError bool
	}{
		{
			name:  "valid format with provider",
			input: "custom/openai/gpt-4",
			expected: &ModelKey{
				Provider:  "custom",
				Publisher: "openai",
				ModelName: "gpt-4",
			},
			expectError: false,
		},
		{
			name:  "valid format without provider (defaults to azureml)",
			input: "openai/gpt-4",
			expected: &ModelKey{
				Provider:  "azureml",
				Publisher: "openai",
				ModelName: "gpt-4",
			},
			expectError: false,
		},
		{
			name:  "valid format with azureml provider explicitly",
			input: "azureml/microsoft/phi-3",
			expected: &ModelKey{
				Provider:  "azureml",
				Publisher: "microsoft",
				ModelName: "phi-3",
			},
			expectError: false,
		},
		{
			name:  "valid format with hyphens in model name",
			input: "cohere/command-r-plus",
			expected: &ModelKey{
				Provider:  "azureml",
				Publisher: "cohere",
				ModelName: "command-r-plus",
			},
			expectError: false,
		},
		{
			name:  "valid format with underscores in model name",
			input: "ai21/jamba_instruct",
			expected: &ModelKey{
				Provider:  "azureml",
				Publisher: "ai21",
				ModelName: "jamba_instruct",
			},
			expectError: false,
		},
		{
			name:        "invalid format with only one part",
			input:       "gpt-4",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid format with four parts",
			input:       "provider/publisher/model/extra",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid format with empty string",
			input:       "",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid format with only slashes",
			input:       "//",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid format with empty parts",
			input:       "provider//model",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseModelKey(tt.input)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.expected.Provider, result.Provider)
				require.Equal(t, tt.expected.Publisher, result.Publisher)
				require.Equal(t, tt.expected.ModelName, result.ModelName)
			}
		})
	}
}

func TestModelKey_String(t *testing.T) {
	tests := []struct {
		name     string
		modelKey *ModelKey
		expected string
	}{
		{
			name: "standard format with azureml provider - should omit provider",
			modelKey: &ModelKey{
				Provider:  "azureml",
				Publisher: "openai",
				ModelName: "gpt-4",
			},
			expected: "openai/gpt-4",
		},
		{
			name: "custom provider - should include provider",
			modelKey: &ModelKey{
				Provider:  "custom",
				Publisher: "microsoft",
				ModelName: "phi-3",
			},
			expected: "custom/microsoft/phi-3",
		},
		{
			name: "azureml provider with hyphens - should omit provider",
			modelKey: &ModelKey{
				Provider:  "azureml",
				Publisher: "cohere",
				ModelName: "command-r-plus",
			},
			expected: "cohere/command-r-plus",
		},
		{
			name: "azureml provider with underscores - should omit provider",
			modelKey: &ModelKey{
				Provider:  "azureml",
				Publisher: "ai21",
				ModelName: "jamba_instruct",
			},
			expected: "ai21/jamba_instruct",
		},
		{
			name: "non-azureml provider - should include provider",
			modelKey: &ModelKey{
				Provider:  "custom-provider",
				Publisher: "test-publisher",
				ModelName: "test-model",
			},
			expected: "custom-provider/test-publisher/test-model",
		},
		{
			name: "azureml provider with uppercase and spaces - should format and omit provider",
			modelKey: &ModelKey{
				Provider:  "azureml",
				Publisher: "Open AI",
				ModelName: "GPT 4",
			},
			expected: "open-ai/gpt-4",
		},
		{
			name: "non-azureml provider with uppercase and spaces - should format and include provider",
			modelKey: &ModelKey{
				Provider:  "Custom Provider",
				Publisher: "Test Publisher",
				ModelName: "Test Model Name",
			},
			expected: "custom-provider/test-publisher/test-model-name",
		},
		{
			name: "mixed case with multiple spaces",
			modelKey: &ModelKey{
				Provider:  "azureml",
				Publisher: "Microsoft Corporation",
				ModelName: "Phi 3 Mini Instruct",
			},
			expected: "microsoft-corporation/phi-3-mini-instruct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.modelKey.String()
			require.Equal(t, tt.expected, result)
		})
	}
}
