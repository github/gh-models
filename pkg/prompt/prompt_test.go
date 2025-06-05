package prompt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromptFile(t *testing.T) {
	t.Run("loads and parses prompt file", func(t *testing.T) {
		const yamlBody = `
name: Test Prompt
description: A test prompt file
model: openai/gpt-4o
modelParameters:
  temperature: 0.5
  maxTokens: 100
messages:
  - role: system
    content: You are a helpful assistant.
  - role: user
    content: "Hello {{name}}"
testData:
  - name: "Alice"
  - name: "Bob"
evaluators:
  - name: contains-greeting
    string:
      contains: "hello"
`

		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFilePath, []byte(yamlBody), 0644)
		require.NoError(t, err)

		promptFile, err := LoadFromFile(promptFilePath)
		require.NoError(t, err)
		require.Equal(t, "Test Prompt", promptFile.Name)
		require.Equal(t, "A test prompt file", promptFile.Description)
		require.Equal(t, "openai/gpt-4o", promptFile.Model)
		require.Equal(t, 0.5, *promptFile.ModelParameters.Temperature)
		require.Equal(t, 100, *promptFile.ModelParameters.MaxTokens)
		require.Len(t, promptFile.Messages, 2)
		require.Equal(t, "system", promptFile.Messages[0].Role)
		require.Equal(t, "You are a helpful assistant.", promptFile.Messages[0].Content)
		require.Equal(t, "user", promptFile.Messages[1].Role)
		require.Equal(t, "Hello {{name}}", promptFile.Messages[1].Content)
		require.Len(t, promptFile.TestData, 2)
		require.Equal(t, "Alice", promptFile.TestData[0]["name"])
		require.Equal(t, "Bob", promptFile.TestData[1]["name"])
		require.Len(t, promptFile.Evaluators, 1)
		require.Equal(t, "contains-greeting", promptFile.Evaluators[0].Name)
		require.Equal(t, "hello", promptFile.Evaluators[0].String.Contains)
	})

	t.Run("templates messages correctly", func(t *testing.T) {
		testData := map[string]interface{}{
			"name": "World",
			"age":  25,
		}

		result, err := TemplateString("Hello {{name}}, you are {{age}} years old", testData)
		require.NoError(t, err)
		require.Equal(t, "Hello World, you are 25 years old", result)
	})

	t.Run("handles missing template variables", func(t *testing.T) {
		testData := map[string]interface{}{
			"name": "World",
		}

		result, err := TemplateString("Hello {{name}}, you are {{missing}} years old", testData)
		require.NoError(t, err)
		require.Equal(t, "Hello World, you are {{missing}} years old", result)
	})

	t.Run("handles file not found", func(t *testing.T) {
		_, err := LoadFromFile("/nonexistent/file.yml")
		require.Error(t, err)
	})

	t.Run("handles invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "invalid.prompt.yml")
		err := os.WriteFile(promptFilePath, []byte("invalid: yaml: content: ["), 0644)
		require.NoError(t, err)

		_, err = LoadFromFile(promptFilePath)
		require.Error(t, err)
	})
}
