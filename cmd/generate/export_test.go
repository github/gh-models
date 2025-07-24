package generate

import (
	"bytes"
	"testing"

	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/prompt"
	"github.com/github/gh-models/pkg/util"
)

func TestToGitHubModelsPrompt(t *testing.T) {
	tests := []struct {
		name        string
		modelID     string
		context     *PromptPexContext
		options     *PromptPexOptions
		expected    func(*prompt.File) bool // validation function
		expectError bool
	}{
		{
			name:    "basic conversion with default model",
			modelID: "evals",
			context: &PromptPexContext{
				Prompt: &prompt.File{
					Name:        "test-prompt",
					Description: "Test description",
					Messages: []prompt.Message{
						{
							Role:    "system",
							Content: "You are a helpful assistant.",
						},
						{
							Role:    "user",
							Content: "Hello {{input}}!",
						},
					},
				},
				Rules: []string{"1. Be helpful", "2. Be accurate"},
				Tests: []PromptPexTest{
					{
						TestInput:   `{"input": "world"}`,
						Groundtruth: util.Ptr("Hello world!"),
						Reasoning:   util.Ptr("Basic greeting test"),
					},
				},
			},
			options: util.Ptr(PromptPexOptions{
				Temperature: util.Ptr(0.7),
			}),
			expected: func(pf *prompt.File) bool {
				return pf.Model == "gpt-4o" &&
					pf.Name == "test-prompt" &&
					pf.Description == "Test description" &&
					len(pf.Messages) == 2 &&
					len(pf.TestData) == 1 &&
					len(pf.Evaluators) == 1 &&
					*pf.ModelParameters.Temperature == 0.7
			},
			expectError: false,
		},
		{
			name:    "custom model",
			modelID: "gpt-3.5-turbo",
			context: &PromptPexContext{
				Prompt: &prompt.File{
					Name:        "custom-model-test",
					Description: "Custom model test",
					Messages: []prompt.Message{
						{
							Role:    "user",
							Content: "Test message",
						},
					},
				},
				Rules: []string{"Test rules"},
				Tests: []PromptPexTest{
					{
						TestInput: "simple test",
					},
				},
			},
			options: util.Ptr(PromptPexOptions{
				Temperature: util.Ptr(0.5),
			}),
			expected: func(pf *prompt.File) bool {
				return pf.Model == "gpt-3.5-turbo" &&
					pf.Name == "custom-model-test" &&
					len(pf.Messages) == 1 &&
					len(pf.TestData) == 1 &&
					*pf.ModelParameters.Temperature == 0.5
			},
			expectError: false,
		},
		{
			name:    "JSON test input parsing",
			modelID: "gpt-4",
			context: &PromptPexContext{
				Prompt: &prompt.File{
					Name:        "json-test",
					Description: "JSON parsing test",
					Messages: []prompt.Message{
						{
							Role:    "user",
							Content: "Process {{data}} and {{type}}",
						},
					},
				},
				Rules: []string{"Process data correctly"},
				Tests: []PromptPexTest{
					{
						TestInput:   `{"data": "test data", "type": "analysis"}`,
						Groundtruth: util.Ptr("Analysis result"),
					},
				},
			},
			options: util.Ptr(PromptPexOptions{}),
			expected: func(pf *prompt.File) bool {
				if len(pf.TestData) != 1 {
					return false
				}
				testData := pf.TestData[0]
				return testData["data"] == "test data" &&
					testData["type"] == "analysis" &&
					testData["expected"] == "Analysis result"
			},
			expectError: false,
		},
		{
			name:    "empty test input should be skipped",
			modelID: "gpt-4",
			context: &PromptPexContext{
				Prompt: &prompt.File{
					Name:        "empty-test",
					Description: "Empty test case",
					Messages: []prompt.Message{
						{
							Role:    "user",
							Content: "Test {{input}}",
						},
					},
				},
				Rules: []string{"Test rules"},
				Tests: []PromptPexTest{
					{
						TestInput: "",
					},
					{
						TestInput: "   ",
					},
					{
						TestInput: "valid input",
					},
				},
			},
			options: util.Ptr(PromptPexOptions{}),
			expected: func(pf *prompt.File) bool {
				// Only the valid input should remain
				return len(pf.TestData) == 1 &&
					pf.TestData[0]["input"] == "valid input"
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler with proper config
			var outputBuffer bytes.Buffer
			cfg := &command.Config{
				Out: &outputBuffer,
			}
			handler := &generateCommandHandler{
				cfg:     cfg,
				options: tt.options,
			}

			result, err := handler.toGitHubModelsPrompt(tt.modelID, tt.context)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Result should not be nil")
				return
			}

			if !tt.expected(result) {
				t.Errorf("Result validation failed")
			}
		})
	}
}

func TestExtractTemplateVariables(t *testing.T) {
	tests := []struct {
		name     string
		context  *PromptPexContext
		expected map[string]bool
	}{
		{
			name: "basic template variables",
			context: &PromptPexContext{
				Prompt: &prompt.File{
					Messages: []prompt.Message{
						{
							Role:    "user",
							Content: "Hello {{name}}, how are you?",
						},
						{
							Role:    "system",
							Content: "Process {{data}} with {{method}}",
						},
					},
				},
			},
			expected: map[string]bool{
				"name":      true,
				"data":      true,
				"method":    true,
				"expected":  true,
				"reasoning": true,
			},
		},
		{
			name: "no template variables",
			context: &PromptPexContext{
				Prompt: &prompt.File{
					Messages: []prompt.Message{
						{
							Role:    "user",
							Content: "Hello world",
						},
					},
				},
			},
			expected: map[string]bool{
				"expected":  true,
				"reasoning": true,
			},
		},
		{
			name: "duplicate variables",
			context: &PromptPexContext{
				Prompt: &prompt.File{
					Messages: []prompt.Message{
						{
							Role:    "user",
							Content: "{{input}} processing {{input}}",
						},
						{
							Role:    "assistant",
							Content: "Result for {{input}}",
						},
					},
				},
			},
			expected: map[string]bool{
				"input":     true,
				"expected":  true,
				"reasoning": true,
			},
		},
		{
			name: "nil prompt",
			context: &PromptPexContext{
				Prompt: nil,
			},
			expected: map[string]bool{
				"expected":  true,
				"reasoning": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outputBuffer bytes.Buffer
			cfg := &command.Config{
				Out: &outputBuffer,
			}
			handler := &generateCommandHandler{
				cfg:     cfg,
				options: util.Ptr(PromptPexOptions{}),
			}

			result := handler.extractTemplateVariables(tt.context)

			for expectedKey, expectedValue := range tt.expected {
				if result[expectedKey] != expectedValue {
					t.Errorf("Expected key '%s' to be %v, got %v", expectedKey, expectedValue, result[expectedKey])
				}
			}

			for actualKey := range result {
				if _, exists := tt.expected[actualKey]; !exists {
					t.Errorf("Unexpected key '%s' in result", actualKey)
				}
			}
		})
	}
}

func TestExtractVariablesFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "single variable",
			text:     "Hello {{name}}!",
			expected: []string{"name"},
		},
		{
			name:     "multiple variables",
			text:     "Process {{data}} with {{method}} for {{user}}",
			expected: []string{"data", "method", "user"},
		},
		{
			name:     "no variables",
			text:     "Hello world!",
			expected: []string{},
		},
		{
			name:     "variables with spaces",
			text:     "{{ name }} and {{ age }}",
			expected: []string{"name", "age"},
		},
		{
			name:     "nested braces",
			text:     "{{outer}} and {{{inner}}}",
			expected: []string{"outer", "{inner"},
		},
		{
			name:     "empty text",
			text:     "",
			expected: []string{},
		},
		{
			name:     "malformed variables",
			text:     "{{incomplete and {not_valid}}",
			expected: []string{"incomplete and {not_valid"}, // This is what the regex actually captures
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVariablesFromText(tt.text)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d variables, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected variable %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestGetMapKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]bool
		expected int
	}{
		{
			name: "non-empty map",
			input: map[string]bool{
				"key1": true,
				"key2": false,
				"key3": true,
			},
			expected: 3,
		},
		{
			name:     "empty map",
			input:    map[string]bool{},
			expected: 0,
		},
		{
			name:     "nil map",
			input:    nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMapKeys(tt.input)

			if len(result) != tt.expected {
				t.Errorf("Expected %d keys, got %d", tt.expected, len(result))
			}

			// Verify all keys are present
			for key := range tt.input {
				found := false
				for _, resultKey := range result {
					if resultKey == key {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected key '%s' not found in result", key)
				}
			}
		})
	}
}

func TestGetTestScenario(t *testing.T) {
	tests := []struct {
		name     string
		test     PromptPexTest
		expected string
	}{
		{
			name: "with reasoning",
			test: PromptPexTest{
				TestInput: "test input",
				Reasoning: util.Ptr("Test reasoning"),
			},
			expected: "Test reasoning",
		},
		{
			name: "with groundtruth (short)",
			test: PromptPexTest{
				TestInput:   "test input",
				Groundtruth: util.Ptr("Short groundtruth"),
			},
			expected: "Short groundtruth",
		},
		{
			name: "with groundtruth (long)",
			test: PromptPexTest{
				TestInput:   "test input",
				Groundtruth: util.Ptr("This is a very long groundtruth that should be truncated"),
			},
			expected: "This is a very long groundtruth that should be tru...", // First 50 chars + "..."
		},
		{
			name: "with both reasoning and groundtruth (reasoning takes precedence)",
			test: PromptPexTest{
				TestInput:   "test input",
				Reasoning:   util.Ptr("Test reasoning"),
				Groundtruth: util.Ptr("Test groundtruth"),
			},
			expected: "Test reasoning",
		},
		{
			name: "with empty reasoning",
			test: PromptPexTest{
				TestInput:   "test input",
				Reasoning:   util.Ptr(""),
				Groundtruth: util.Ptr("Test groundtruth"),
			},
			expected: "Test groundtruth",
		},
		{
			name: "no reasoning or groundtruth",
			test: PromptPexTest{
				TestInput: "test input",
			},
			expected: "unknown scenario",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTestScenario(tt.test)

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
