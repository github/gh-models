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
			name: "standard format with azureml provider",
			modelKey: &ModelKey{
				Provider:  "azureml",
				Publisher: "openai",
				ModelName: "gpt-4",
			},
			expected: "azureml/openai/gpt-4",
		},
		{
			name: "custom provider",
			modelKey: &ModelKey{
				Provider:  "custom",
				Publisher: "microsoft",
				ModelName: "phi-3",
			},
			expected: "custom/microsoft/phi-3",
		},
		{
			name: "model name with hyphens",
			modelKey: &ModelKey{
				Provider:  "azureml",
				Publisher: "cohere",
				ModelName: "command-r-plus",
			},
			expected: "azureml/cohere/command-r-plus",
		},
		{
			name: "model name with underscores",
			modelKey: &ModelKey{
				Provider:  "azureml",
				Publisher: "ai21",
				ModelName: "jamba_instruct",
			},
			expected: "azureml/ai21/jamba_instruct",
		},
		{
			name: "long provider name",
			modelKey: &ModelKey{
				Provider:  "custom-provider",
				Publisher: "test-publisher",
				ModelName: "test-model",
			},
			expected: "custom-provider/test-publisher/test-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.modelKey.String()
			require.Equal(t, tt.expected, result)
		})
	}
}
