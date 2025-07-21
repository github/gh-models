package prompt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/github/gh-models/internal/azuremodels"
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

	t.Run("loads prompt file with responseFormat text", func(t *testing.T) {
		const yamlBody = `
name: Text Response Format Test
description: Test with text response format
model: openai/gpt-4o
responseFormat: text
messages:
  - role: user
    content: "Hello"
`

		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFilePath, []byte(yamlBody), 0644)
		require.NoError(t, err)

		promptFile, err := LoadFromFile(promptFilePath)
		require.NoError(t, err)
		require.NotNil(t, promptFile.ResponseFormat)
		require.Equal(t, "text", *promptFile.ResponseFormat)
		require.Nil(t, promptFile.JsonSchema)
	})

	t.Run("loads prompt file with responseFormat json_object", func(t *testing.T) {
		const yamlBody = `
name: JSON Object Response Format Test
description: Test with JSON object response format
model: openai/gpt-4o
responseFormat: json_object
messages:
  - role: user
    content: "Hello"
`

		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFilePath, []byte(yamlBody), 0644)
		require.NoError(t, err)

		promptFile, err := LoadFromFile(promptFilePath)
		require.NoError(t, err)
		require.NotNil(t, promptFile.ResponseFormat)
		require.Equal(t, "json_object", *promptFile.ResponseFormat)
		require.Nil(t, promptFile.JsonSchema)
	})

	t.Run("loads prompt file with responseFormat json_schema and jsonSchema as JSON string", func(t *testing.T) {
		const yamlBody = `
name: JSON Schema String Format Test
description: Test with JSON schema as JSON string
model: openai/gpt-4o
responseFormat: json_schema
jsonSchema: |-
  {
    "name": "describe_animal",
    "strict": true,
    "schema": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "description": "The name of the animal"
        },
        "habitat": {
          "type": "string",
          "description": "The habitat the animal lives in"
        }
      },
      "additionalProperties": false,
      "required": [
        "name",
        "habitat"
      ]
    }
  }
messages:
  - role: user
    content: "Hello"
`

		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFilePath, []byte(yamlBody), 0644)
		require.NoError(t, err)

		promptFile, err := LoadFromFile(promptFilePath)
		require.NoError(t, err)
		require.NotNil(t, promptFile.ResponseFormat)
		require.Equal(t, "json_schema", *promptFile.ResponseFormat)
		require.NotNil(t, promptFile.JsonSchema)

		// Parse the JSON schema string to verify its contents
		var schema map[string]interface{}
		err = json.Unmarshal([]byte(*promptFile.JsonSchema), &schema)
		require.NoError(t, err)

		require.Equal(t, "describe_animal", schema["name"])
		require.Equal(t, true, schema["strict"])
		require.Contains(t, schema, "schema")

		// Verify the nested schema structure
		nestedSchema := schema["schema"].(map[string]interface{})
		require.Equal(t, "object", nestedSchema["type"])
		require.Contains(t, nestedSchema, "properties")
		require.Contains(t, nestedSchema, "required")

		properties := nestedSchema["properties"].(map[string]interface{})
		require.Contains(t, properties, "name")
		require.Contains(t, properties, "habitat")

		required := nestedSchema["required"].([]interface{})
		require.Contains(t, required, "name")
		require.Contains(t, required, "habitat")
	})

	t.Run("validates invalid responseFormat", func(t *testing.T) {
		const yamlBody = `
name: Invalid Response Format Test
description: Test with invalid response format
model: openai/gpt-4o
responseFormat: invalid_format
messages:
  - role: user
    content: "Hello"
`

		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFilePath, []byte(yamlBody), 0644)
		require.NoError(t, err)

		_, err = LoadFromFile(promptFilePath)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid responseFormat: invalid_format")
	})

	t.Run("validates json_schema requires jsonSchema", func(t *testing.T) {
		const yamlBody = `
name: JSON Schema Missing Test
description: Test json_schema without jsonSchema
model: openai/gpt-4o
responseFormat: json_schema
messages:
  - role: user
    content: "Hello"
`

		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFilePath, []byte(yamlBody), 0644)
		require.NoError(t, err)

		_, err = LoadFromFile(promptFilePath)
		require.Error(t, err)
		require.Contains(t, err.Error(), "jsonSchema is required when responseFormat is 'json_schema'")
	})

	t.Run("BuildChatCompletionOptions includes responseFormat and jsonSchema", func(t *testing.T) {
		jsonSchemaStr := `{
			"name": "test_schema",
			"strict": true,
			"schema": {
				"type": "object",
				"properties": {
					"name": {
						"type": "string",
						"description": "The name"
					}
				},
				"required": ["name"]
			}
		}`

		promptFile := &File{
			Model:          "openai/gpt-4o",
			ResponseFormat: func() *string { s := "json_schema"; return &s }(),
			JsonSchema:     func() *JsonSchema { js := JsonSchema(jsonSchemaStr); return &js }(),
		}

		messages := []azuremodels.ChatMessage{
			{
				Role:    azuremodels.ChatMessageRoleUser,
				Content: func() *string { s := "Hello"; return &s }(),
			},
		}
		options := promptFile.BuildChatCompletionOptions(messages)
		require.NotNil(t, options.ResponseFormat)
		require.Equal(t, "json_schema", options.ResponseFormat.Type)
		require.NotNil(t, options.ResponseFormat.JsonSchema)

		schema := *options.ResponseFormat.JsonSchema
		require.Equal(t, "test_schema", schema["name"])
		require.Equal(t, true, schema["strict"])
		require.Contains(t, schema, "schema")

		schemaContent := schema["schema"].(map[string]interface{})
		require.Equal(t, "object", schemaContent["type"])
		require.Contains(t, schemaContent, "properties")
	})
}
