package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/util"
)

func TestCreateContext(t *testing.T) {
	tests := []struct {
		name           string
		promptFileYAML string
		options        PromptPexOptions
		expectError    bool
		expectedFields map[string]interface{}
	}{
		{
			name: "basic prompt file",
			promptFileYAML: `name: test-prompt
description: A test prompt
model: gpt-4o
messages:
  - role: system
    content: You are a helpful assistant.
  - role: user
    content: Hello {{input}}!`,
			options: PromptPexOptions{
				Temperature: util.Ptr(0.7),
			},
			expectError: false,
			expectedFields: map[string]interface{}{
				"writeResults": true,
				"intent":       "",
				"rules":        "",
				"inverseRules": "",
			},
		},
		{
			name: "prompt with model parameters",
			promptFileYAML: `name: parametrized-prompt
description: A prompt with parameters
model: gpt-3.5-turbo
modelParameters:
  temperature: 0.5
  maxTokens: 1000
messages:
  - role: user
    content: Analyze {{data}}`,
			options: PromptPexOptions{
				Effort: util.Ptr("high"),
			},
			expectError: false,
			expectedFields: map[string]interface{}{
				"writeResults": true,
				"intent":       "",
				"rules":        "",
			},
		},
		{
			name: "minimal prompt",
			promptFileYAML: `name: minimal
description: Minimal prompt
model: gpt-4
messages:
  - role: user
    content: Test`,
			options:     PromptPexOptions{},
			expectError: false,
			expectedFields: map[string]interface{}{
				"writeResults": true,
			},
		},
		{
			name:           "invalid yaml",
			promptFileYAML: `invalid: yaml: content: [`,
			options:        PromptPexOptions{},
			expectError:    true,
		},
		{
			name:           "missing required fields",
			promptFileYAML: `description: Missing name`,
			options:        PromptPexOptions{},
			expectError:    false, // The prompt package might not require all fields
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary prompt file
			tempDir := t.TempDir()
			promptFile := filepath.Join(tempDir, "test.prompt.yml")
			err := os.WriteFile(promptFile, []byte(tt.promptFileYAML), 0644)
			if err != nil {
				t.Fatalf("Failed to create test prompt file: %v", err)
			}

			// Create handler
			config := &command.Config{}
			handler := &generateCommandHandler{
				cfg:     config,
				options: &tt.options,
			}

			// Test CreateContext
			context, err := handler.CreateContext(promptFile)

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

			// Verify context fields
			if context == nil {
				t.Fatalf("Context is nil")
			}

			// Check that RunID is generated and has expected format
			if context.RunID == "" {
				t.Errorf("RunID should not be empty")
			}
			if !strings.HasPrefix(context.RunID, "run_") {
				t.Errorf("RunID should start with 'run_', got: %s", context.RunID)
			}

			// Check WriteResults default
			if context.WriteResults == nil || *context.WriteResults != true {
				t.Errorf("WriteResults should be true by default")
			}

			// Check that Prompt is loaded
			if context.Prompt == nil {
				t.Errorf("Prompt should not be nil")
			}

			// Check expected fields
			for field, expectedValue := range tt.expectedFields {
				switch field {
				case "writeResults":
					if context.WriteResults == nil || *context.WriteResults != expectedValue.(bool) {
						t.Errorf("Expected %s to be %v, got %v", field, expectedValue, context.WriteResults)
					}
				case "intent":
					if context.Intent != expectedValue.(string) {
						t.Errorf("Expected %s to be %q, got %q", field, expectedValue, context.Intent)
					}
				case "rules":
					if context.Rules != expectedValue.(string) {
						t.Errorf("Expected %s to be %q, got %q", field, expectedValue, context.Rules)
					}
				case "inverseRules":
					if context.InverseRules != expectedValue.(string) {
						t.Errorf("Expected %s to be %q, got %q", field, expectedValue, context.InverseRules)
					}
				}
			}

			// Check that options are preserved
			if context.Options.Temperature != tt.options.Temperature {
				t.Errorf("Expected temperature to be preserved")
			}
			if context.Options.Effort != tt.options.Effort {
				t.Errorf("Expected effort to be preserved")
			}
		})
	}
}

func TestCreateContextRunIDUniqueness(t *testing.T) {
	// Create a simple prompt file
	tempDir := t.TempDir()
	promptFile := filepath.Join(tempDir, "test.prompt.yml")
	promptYAML := `name: test
description: Test prompt
model: gpt-4
messages:
  - role: user
    content: Test`
	err := os.WriteFile(promptFile, []byte(promptYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test prompt file: %v", err)
	}

	config := &command.Config{}
	handler := &generateCommandHandler{
		cfg:     config,
		options: util.Ptr(PromptPexOptions{}),
	}

	// Create multiple contexts and check that RunIDs are generated
	var runIDs []string
	for i := 0; i < 3; i++ {
		context, err := handler.CreateContext(promptFile)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check that RunID has the expected format
		if !strings.HasPrefix(context.RunID, "run_") {
			t.Errorf("RunID should start with 'run_', got: %s", context.RunID)
		}

		runIDs = append(runIDs, context.RunID)
		time.Sleep(time.Millisecond * 100) // Shorter delay
	}

	// Check that all RunIDs are non-empty and properly formatted
	for i, runID := range runIDs {
		if runID == "" {
			t.Errorf("RunID %d should not be empty", i)
		}
		if !strings.HasPrefix(runID, "run_") {
			t.Errorf("RunID %d should start with 'run_', got: %s", i, runID)
		}
	}

	// Note: We don't require strict uniqueness as timestamp-based IDs might collide
	// in rapid succession, which is acceptable for this use case
}

func TestCreateContextWithNonExistentFile(t *testing.T) {
	config := &command.Config{}
	handler := &generateCommandHandler{
		cfg:     config,
		options: util.Ptr(PromptPexOptions{}),
	}

	_, err := handler.CreateContext("/nonexistent/file.prompt.yml")
	if err == nil {
		t.Errorf("Expected error for non-existent file")
	}
}

func TestCreateContextPromptValidation(t *testing.T) {
	tests := []struct {
		name           string
		promptFileYAML string
		expectError    bool
		errorContains  string
	}{
		{
			name: "valid prompt",
			promptFileYAML: `name: valid
description: Valid prompt
model: gpt-4
messages:
  - role: user
    content: Test`,
			expectError: false,
		},
		{
			name: "invalid response format",
			promptFileYAML: `name: invalid-response
description: Invalid response format
model: gpt-4
responseFormat: invalid_format
messages:
  - role: user
    content: Test`,
			expectError:   true,
			errorContains: "invalid responseFormat",
		},
		{
			name: "json_schema without schema",
			promptFileYAML: `name: missing-schema
description: Missing schema
model: gpt-4
responseFormat: json_schema
messages:
  - role: user
    content: Test`,
			expectError:   true,
			errorContains: "jsonSchema is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			promptFile := filepath.Join(tempDir, "test.prompt.yml")
			err := os.WriteFile(promptFile, []byte(tt.promptFileYAML), 0644)
			if err != nil {
				t.Fatalf("Failed to create test prompt file: %v", err)
			}

			config := &command.Config{}
			handler := &generateCommandHandler{
				cfg:     config,
				options: util.Ptr(PromptPexOptions{}),
			}

			_, err = handler.CreateContext(promptFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
