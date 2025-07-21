// Package prompt provides shared types and utilities for working with .prompt.yml files
package prompt

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/github/gh-models/internal/azuremodels"
	"gopkg.in/yaml.v3"
)

// File represents the structure of a .prompt.yml file
type File struct {
	Name            string          `yaml:"name"`
	Description     string          `yaml:"description"`
	Model           string          `yaml:"model"`
	ModelParameters ModelParameters `yaml:"modelParameters"`
	ResponseFormat  *string         `yaml:"responseFormat,omitempty"`
	JsonSchema      *JsonSchema     `yaml:"jsonSchema,omitempty"`
	Messages        []Message       `yaml:"messages"`
	// TestData and Evaluators are only used by eval command
	TestData   []TestDataItem `yaml:"testData,omitempty"`
	Evaluators []Evaluator    `yaml:"evaluators,omitempty"`
}

// ModelParameters represents model configuration parameters
type ModelParameters struct {
	MaxTokens   *int     `yaml:"maxTokens"`
	Temperature *float64 `yaml:"temperature"`
	TopP        *float64 `yaml:"topP"`
}

// Message represents a conversation message
type Message struct {
	Role    string `yaml:"role"`
	Content string `yaml:"content"`
}

// TestDataItem represents a single test data item for evaluation
type TestDataItem map[string]interface{}

// Evaluator represents an evaluation method (only used by eval command)
type Evaluator struct {
	Name   string           `yaml:"name"`
	String *StringEvaluator `yaml:"string,omitempty"`
	LLM    *LLMEvaluator    `yaml:"llm,omitempty"`
	Uses   string           `yaml:"uses,omitempty"`
}

// StringEvaluator represents string-based evaluation
type StringEvaluator struct {
	EndsWith   string `yaml:"endsWith,omitempty"`
	StartsWith string `yaml:"startsWith,omitempty"`
	Contains   string `yaml:"contains,omitempty"`
	Equals     string `yaml:"equals,omitempty"`
}

// LLMEvaluator represents LLM-based evaluation
type LLMEvaluator struct {
	ModelID      string   `yaml:"modelId"`
	Prompt       string   `yaml:"prompt"`
	Choices      []Choice `yaml:"choices"`
	SystemPrompt string   `yaml:"systemPrompt,omitempty"`
}

// Choice represents a scoring choice for LLM evaluation
type Choice struct {
	Choice string  `yaml:"choice"`
	Score  float64 `yaml:"score"`
}

// JsonSchema represents a JSON schema for structured responses
type JsonSchema struct {
	Raw    string
	Parsed map[string]interface{}
}

// UnmarshalYAML implements custom YAML unmarshaling for JsonSchema
// Only supports JSON string format
func (js *JsonSchema) UnmarshalYAML(node *yaml.Node) error {
	// Only support string nodes (JSON format)
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("jsonSchema must be a JSON string")
	}

	var jsonStr string
	if err := node.Decode(&jsonStr); err != nil {
		return err
	}

	// Parse and validate the JSON schema
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return fmt.Errorf("invalid JSON in jsonSchema: %w", err)
	}

	js.Raw = jsonStr
	js.Parsed = parsed
	return nil
}

// LoadFromFile loads and parses a prompt file from the given path
func LoadFromFile(filePath string) (*File, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var promptFile File
	if err := yaml.Unmarshal(data, &promptFile); err != nil {
		return nil, err
	}

	if err := promptFile.validateResponseFormat(); err != nil {
		return nil, err
	}

	return &promptFile, nil
}

// validateResponseFormat validates the responseFormat field
func (f *File) validateResponseFormat() error {
	if f.ResponseFormat == nil {
		return nil
	}

	switch *f.ResponseFormat {
	case "text", "json_object", "json_schema":
	default:
		return fmt.Errorf("invalid responseFormat: %s. Must be 'text', 'json_object', or 'json_schema'", *f.ResponseFormat)
	}

	// If responseFormat is "json_schema", jsonSchema must be provided
	if *f.ResponseFormat == "json_schema" {
		if f.JsonSchema == nil {
			return fmt.Errorf("jsonSchema is required when responseFormat is 'json_schema'")
		}

		// Check for required fields in the already parsed schema
		if _, ok := f.JsonSchema.Parsed["name"]; !ok {
			return fmt.Errorf("jsonSchema must contain 'name' field")
		}
		if _, ok := f.JsonSchema.Parsed["schema"]; !ok {
			return fmt.Errorf("jsonSchema must contain 'schema' field")
		}
	}

	return nil
}

// TemplateString templates a string with the given data using simple {{variable}} replacement
func TemplateString(templateStr string, data interface{}) (string, error) {
	result := templateStr

	// Convert data to map[string]interface{} if it's not already
	var dataMap map[string]interface{}
	switch d := data.(type) {
	case map[string]interface{}:
		dataMap = d
	case map[string]string:
		dataMap = make(map[string]interface{})
		for k, v := range d {
			dataMap[k] = v
		}
	default:
		// If it's not a map, we can't template it
		return result, nil
	}

	// Replace all {{variable}} patterns with values from the data map
	for key, value := range dataMap {
		placeholder := "{{" + key + "}}"
		if valueStr, ok := value.(string); ok {
			result = strings.ReplaceAll(result, placeholder, valueStr)
		} else {
			// Convert non-string values to string
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}

	return result, nil
}

// GetAzureChatMessageRole converts a role string to azuremodels.ChatMessageRole
func GetAzureChatMessageRole(role string) (azuremodels.ChatMessageRole, error) {
	switch strings.ToLower(role) {
	case "system":
		return azuremodels.ChatMessageRoleSystem, nil
	case "user":
		return azuremodels.ChatMessageRoleUser, nil
	case "assistant":
		return azuremodels.ChatMessageRoleAssistant, nil
	default:
		return "", fmt.Errorf("unknown message role: %s", role)
	}
}

// BuildChatCompletionOptions creates a ChatCompletionOptions with the file's model and parameters
func (f *File) BuildChatCompletionOptions(messages []azuremodels.ChatMessage) azuremodels.ChatCompletionOptions {
	req := azuremodels.ChatCompletionOptions{
		Messages: messages,
		Model:    f.Model,
		Stream:   false,
	}

	if f.ModelParameters.MaxTokens != nil {
		req.MaxTokens = f.ModelParameters.MaxTokens
	}
	if f.ModelParameters.Temperature != nil {
		req.Temperature = f.ModelParameters.Temperature
	}
	if f.ModelParameters.TopP != nil {
		req.TopP = f.ModelParameters.TopP
	}

	if f.ResponseFormat != nil {
		responseFormat := &azuremodels.ResponseFormat{
			Type: *f.ResponseFormat,
		}
		if f.JsonSchema != nil {
			responseFormat.JsonSchema = &f.JsonSchema.Parsed
		}
		req.ResponseFormat = responseFormat
	}

	return req
}
