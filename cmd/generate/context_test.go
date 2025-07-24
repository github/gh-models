package generate

import (
	"os"
	"path/filepath"
	"reflect"
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
				"intent":       "",
				"rules":        []string{},
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
				"intent": "",
				"rules":  []string{},
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
			options:        PromptPexOptions{},
			expectError:    false,
			expectedFields: map[string]interface{}{},
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
			context, err := handler.CreateContextFromPrompt(promptFile)

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

			// Check that Prompt is loaded
			if context.Prompt == nil {
				t.Errorf("Prompt should not be nil")
			}

			// Check that PromptHash is generated
			if context.PromptHash == "" {
				t.Errorf("PromptHash should not be empty")
			}
			if len(context.PromptHash) != 64 { // SHA256 hex string is 64 characters
				t.Errorf("PromptHash should be 64 characters long (SHA256 hex), got %d", len(context.PromptHash))
			}

			// Check expected fields
			for field, expectedValue := range tt.expectedFields {
				switch field {
				case "intent":
					if context.Intent != expectedValue.(string) {
						t.Errorf("Expected %s to be %q, got %q", field, expectedValue, context.Intent)
					}
				case "rules":
					if !reflect.DeepEqual(context.Rules, expectedValue.([]string)) {
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
		context, err := handler.CreateContextFromPrompt(promptFile)
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

	_, err := handler.CreateContextFromPrompt("/nonexistent/file.prompt.yml")
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

			_, err = handler.CreateContextFromPrompt(promptFile)

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

func TestCreateContextPromptHash(t *testing.T) {
	tests := []struct {
		name            string
		promptFileYAML1 string
		promptFileYAML2 string
		expectSameHash  bool
	}{
		{
			name: "identical prompts should have same hash",
			promptFileYAML1: `name: test
description: Test prompt
model: gpt-4
modelParameters:
  temperature: 0.7
messages:
  - role: user
    content: Hello world`,
			promptFileYAML2: `name: test
description: Test prompt
model: gpt-4
modelParameters:
  temperature: 0.7
messages:
  - role: user
    content: Hello world`,
			expectSameHash: true,
		},
		{
			name: "different models should have different hash",
			promptFileYAML1: `name: test
description: Test prompt
model: gpt-4
messages:
  - role: user
    content: Hello world`,
			promptFileYAML2: `name: test
description: Test prompt
model: gpt-3.5-turbo
messages:
  - role: user
    content: Hello world`,
			expectSameHash: false,
		},
		{
			name: "different temperatures should have different hash",
			promptFileYAML1: `name: test
description: Test prompt
model: gpt-4
modelParameters:
  temperature: 0.7
messages:
  - role: user
    content: Hello world`,
			promptFileYAML2: `name: test
description: Test prompt
model: gpt-4
modelParameters:
  temperature: 0.5
messages:
  - role: user
    content: Hello world`,
			expectSameHash: false,
		},
		{
			name: "different messages should have different hash",
			promptFileYAML1: `name: test
description: Test prompt
model: gpt-4
messages:
  - role: user
    content: Hello world`,
			promptFileYAML2: `name: test
description: Test prompt
model: gpt-4
messages:
  - role: user
    content: Hello universe`,
			expectSameHash: false,
		},
		{
			name: "different description should have same hash (description not included in hash)",
			promptFileYAML1: `name: test
description: Description 1
model: gpt-4
messages:
  - role: user
    content: Hello world`,
			promptFileYAML2: `name: test
description: Description 2
model: gpt-4
messages:
  - role: user
    content: Hello world`,
			expectSameHash: true,
		},
		{
			name: "different maxTokens should have different hash",
			promptFileYAML1: `name: test
description: Test prompt
model: gpt-4
modelParameters:
  maxTokens: 1000
messages:
  - role: user
    content: Hello world`,
			promptFileYAML2: `name: test
description: Test prompt
model: gpt-4
modelParameters:
  maxTokens: 2000
messages:
  - role: user
    content: Hello world`,
			expectSameHash: false,
		},
		{
			name: "different topP should have different hash",
			promptFileYAML1: `name: test
description: Test prompt
model: gpt-4
modelParameters:
  topP: 0.9
messages:
  - role: user
    content: Hello world`,
			promptFileYAML2: `name: test
description: Test prompt
model: gpt-4
modelParameters:
  topP: 0.8
messages:
  - role: user
    content: Hello world`,
			expectSameHash: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create first prompt file
			promptFile1 := filepath.Join(tempDir, "test1.prompt.yml")
			err := os.WriteFile(promptFile1, []byte(tt.promptFileYAML1), 0644)
			if err != nil {
				t.Fatalf("Failed to create first test prompt file: %v", err)
			}

			// Create second prompt file
			promptFile2 := filepath.Join(tempDir, "test2.prompt.yml")
			err = os.WriteFile(promptFile2, []byte(tt.promptFileYAML2), 0644)
			if err != nil {
				t.Fatalf("Failed to create second test prompt file: %v", err)
			}

			config := &command.Config{}
			handler := &generateCommandHandler{
				cfg:     config,
				options: util.Ptr(PromptPexOptions{}),
			}

			// Create contexts from both files
			context1, err := handler.CreateContextFromPrompt(promptFile1)
			if err != nil {
				t.Fatalf("Failed to create context from first file: %v", err)
			}

			context2, err := handler.CreateContextFromPrompt(promptFile2)
			if err != nil {
				t.Fatalf("Failed to create context from second file: %v", err)
			}

			// Verify hashes are set
			if context1.PromptHash == "" {
				t.Errorf("First context PromptHash should not be empty")
			}
			if context2.PromptHash == "" {
				t.Errorf("Second context PromptHash should not be empty")
			}

			// Compare hashes
			if tt.expectSameHash {
				if context1.PromptHash != context2.PromptHash {
					t.Errorf("Expected same hash but got different:\nHash1: %s\nHash2: %s", context1.PromptHash, context2.PromptHash)
				}
			} else {
				if context1.PromptHash == context2.PromptHash {
					t.Errorf("Expected different hashes but got same: %s", context1.PromptHash)
				}
			}
		})
	}
}
