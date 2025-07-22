package generate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/prompt"
)

func TestGithubModelsEvalsGenerate(t *testing.T) {
	tests := []struct {
		name            string
		context         *PromptPexContext
		options         PromptPexOptions
		expectedFiles   []string
		expectedContent []string
		expectError     bool
		expectedOutput  string
	}{
		{
			name: "basic generation with default model",
			context: &PromptPexContext{
				WriteResults: BoolPtr(true),
				Prompt: &prompt.File{
					Name:        "test-prompt",
					Description: "Test description",
					Model:       "gpt-4o",
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
				Rules: "1. Be helpful\n2. Be accurate",
				PromptPexTests: []PromptPexTest{
					{
						TestInput:   `{"input": "world"}`,
						Groundtruth: StringPtr("Hello world!"),
						Reasoning:   StringPtr("Basic greeting test"),
					},
				},
			},
			options: PromptPexOptions{
				Temperature:     Float64Ptr(0.7),
				ModelsUnderTest: []string{},
				Out:             StringPtr(t.TempDir()),
			},
			expectedFiles: []string{"gpt-4o.prompt.yml"},
			expectedContent: []string{
				"name: test-prompt",
				"description: Test description",
				"model: gpt-4o",
				"temperature: 0.7",
				"input: world",
				"expected: Hello world!",
			},
			expectError:    false,
			expectedOutput: "Generating GitHub Models Evals...\nGenerating GitHub Models eval for model: evals\nGenerated GitHub Models eval file:",
		},
		{
			name: "multiple custom models",
			context: &PromptPexContext{
				WriteResults: BoolPtr(true),
				Prompt: &prompt.File{
					Name:        "multi-model-test",
					Description: "Multi-model test",
					Model:       "gpt-4",
					Messages: []prompt.Message{
						{
							Role:    "user",
							Content: "Test message",
						},
					},
				},
				Rules: "Test rules",
				PromptPexTests: []PromptPexTest{
					{
						TestInput: "simple test",
					},
				},
			},
			options: PromptPexOptions{
				Temperature:     Float64Ptr(0.5),
				ModelsUnderTest: []string{"gpt-3.5-turbo", "gpt-4"},
				Out:             StringPtr(t.TempDir()),
			},
			expectedFiles: []string{
				"gpt-4o.prompt.yml", // default "evals" model
				"gpt-3.5-turbo.prompt.yml",
				"gpt-4.prompt.yml",
			},
			expectedContent: []string{
				"temperature: 0.5",
				"name: multi-model-test",
				"description: Multi-model test",
			},
			expectError:    false,
			expectedOutput: "Generating GitHub Models Evals...\nGenerating GitHub Models eval for model: evals\nGenerated GitHub Models eval file:\nGenerating GitHub Models eval for model: gpt-3.5-turbo\nGenerated GitHub Models eval file:\nGenerating GitHub Models eval for model: gpt-4\nGenerated GitHub Models eval file:",
		},
		{
			name: "no tests - should skip generation",
			context: &PromptPexContext{
				WriteResults: BoolPtr(true),
				Prompt: &prompt.File{
					Name:        "no-tests",
					Description: "No tests case",
					Model:       "gpt-4",
					Messages: []prompt.Message{
						{
							Role:    "user",
							Content: "Test",
						},
					},
				},
				Rules:          "Test rules",
				PromptPexTests: []PromptPexTest{},
			},
			options: PromptPexOptions{
				Temperature: Float64Ptr(0.8),
				Out:         StringPtr(t.TempDir()),
			},
			expectedFiles:   []string{},
			expectedContent: []string{},
			expectError:     false,
			expectedOutput:  "Generating GitHub Models Evals...\nNo tests found. Skipping GitHub Models Evals generation.",
		},
		{
			name: "write results disabled",
			context: &PromptPexContext{
				WriteResults: BoolPtr(false),
				Prompt: &prompt.File{
					Name:        "no-write",
					Description: "No write test",
					Model:       "gpt-4",
					Messages: []prompt.Message{
						{
							Role:    "user",
							Content: "Test",
						},
					},
				},
				Rules: "Test rules",
				PromptPexTests: []PromptPexTest{
					{
						TestInput: "test",
					},
				},
			},
			options: PromptPexOptions{
				Temperature: Float64Ptr(0.3),
				Out:         StringPtr(t.TempDir()),
			},
			expectedFiles:   []string{}, // No files should be written
			expectedContent: []string{},
			expectError:     false,
			expectedOutput:  "Generating GitHub Models Evals...\nGenerating GitHub Models eval for model: evals\nGenerated GitHub Models eval file:",
		},
		{
			name: "model with slash in name",
			context: &PromptPexContext{
				WriteResults: BoolPtr(true),
				Prompt: &prompt.File{
					Name:        "slash-model-test",
					Description: "Slash model test",
					Model:       "gpt-4",
					Messages: []prompt.Message{
						{
							Role:    "user",
							Content: "Test",
						},
					},
				},
				Rules: "Test rules",
				PromptPexTests: []PromptPexTest{
					{
						TestInput: "test",
					},
				},
			},
			options: PromptPexOptions{
				Temperature:     Float64Ptr(0.9),
				ModelsUnderTest: []string{"openai/gpt-4o-mini"},
				Out:             StringPtr(t.TempDir()),
			},
			expectedFiles: []string{
				"gpt-4o.prompt.yml",             // default "evals" model
				"openai_gpt-4o-mini.prompt.yml", // slash replaced with underscore
			},
			expectedContent: []string{
				"temperature: 0.9",
				"name: slash-model-test",
				"description: Slash model test",
			},
			expectError:    false,
			expectedOutput: "Generating GitHub Models Evals...\nGenerating GitHub Models eval for model: evals\nGenerated GitHub Models eval file:\nGenerating GitHub Models eval for model: openai/gpt-4o-mini\nGenerated GitHub Models eval file:",
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

			err := handler.githubModelsEvalsGenerate(tt.context)

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

			// Check output
			output := outputBuffer.String()
			if len(tt.expectedOutput) > 0 {
				outputLines := strings.Split(tt.expectedOutput, "\n")
				for _, expectedLine := range outputLines {
					if strings.TrimSpace(expectedLine) == "" {
						continue
					}
					found := strings.Contains(output, expectedLine)
					if !found {
						t.Errorf("Expected output to contain '%s', but got: %s", expectedLine, output)
					}
				}
			}

			// Check file creation only if WriteResults is true
			if tt.context.WriteResults != nil && *tt.context.WriteResults {
				outputDir := "."
				if tt.options.Out != nil {
					outputDir = *tt.options.Out
				}

				// Check that expected files were created
				for _, expectedFile := range tt.expectedFiles {
					filePath := filepath.Join(outputDir, expectedFile)
					if _, err := os.Stat(filePath); os.IsNotExist(err) {
						t.Errorf("Expected file %s was not created", expectedFile)
					} else if err != nil {
						t.Errorf("Error checking file %s: %v", expectedFile, err)
					}
				}

				// Check file contents if files were expected
				if len(tt.expectedFiles) > 0 {
					for _, expectedFile := range tt.expectedFiles {
						filePath := filepath.Join(outputDir, expectedFile)
						content, err := os.ReadFile(filePath)
						if err != nil {
							t.Errorf("Error reading file %s: %v", expectedFile, err)
							continue
						}

						contentStr := string(content)

						// Check for specific content in each file based on the file name
						if strings.Contains(expectedFile, "gpt-4o.prompt.yml") {
							if !strings.Contains(contentStr, "model: gpt-4o") {
								t.Errorf("File %s should contain 'model: gpt-4o', but doesn't", expectedFile)
							}
						} else if strings.Contains(expectedFile, "gpt-3.5-turbo.prompt.yml") {
							if !strings.Contains(contentStr, "model: gpt-3.5-turbo") {
								t.Errorf("File %s should contain 'model: gpt-3.5-turbo', but doesn't", expectedFile)
							}
						} else if strings.Contains(expectedFile, "gpt-4.prompt.yml") {
							if !strings.Contains(contentStr, "model: gpt-4") {
								t.Errorf("File %s should contain 'model: gpt-4', but doesn't", expectedFile)
							}
						} else if strings.Contains(expectedFile, "openai_gpt-4o-mini.prompt.yml") {
							if !strings.Contains(contentStr, "model: openai/gpt-4o-mini") {
								t.Errorf("File %s should contain 'model: openai/gpt-4o-mini', but doesn't", expectedFile)
							}
						}

						// Check for common content that should be in all files
						for _, expectedContent := range tt.expectedContent {
							// Skip model-specific content checks here since we handle them above
							if !strings.HasPrefix(expectedContent, "model: ") {
								if !strings.Contains(contentStr, expectedContent) {
									t.Errorf("File %s should contain '%s', but content is: %s", expectedFile, expectedContent, contentStr)
								}
							}
						}
					}
				}
			} else {
				// If WriteResults is false, no files should be created
				outputDir := "."
				if tt.options.Out != nil {
					outputDir = *tt.options.Out
				}
				files, err := os.ReadDir(outputDir)
				if err == nil {
					// Count only .prompt.yml files
					promptFiles := 0
					for _, file := range files {
						if strings.HasSuffix(file.Name(), ".prompt.yml") {
							promptFiles++
						}
					}
					if promptFiles > 0 {
						t.Errorf("No .prompt.yml files should be written when WriteResults is false, but found %d", promptFiles)
					}
				}
			}
		})
	}
}

func TestToGitHubModelsPrompt(t *testing.T) {
	tests := []struct {
		name        string
		modelID     string
		context     *PromptPexContext
		options     PromptPexOptions
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
				Rules: "1. Be helpful\n2. Be accurate",
				PromptPexTests: []PromptPexTest{
					{
						TestInput:   `{"input": "world"}`,
						Groundtruth: StringPtr("Hello world!"),
						Reasoning:   StringPtr("Basic greeting test"),
					},
				},
			},
			options: PromptPexOptions{
				Temperature: Float64Ptr(0.7),
			},
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
				Rules: "Test rules",
				PromptPexTests: []PromptPexTest{
					{
						TestInput: "simple test",
					},
				},
			},
			options: PromptPexOptions{
				Temperature: Float64Ptr(0.5),
			},
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
				Rules: "Process data correctly",
				PromptPexTests: []PromptPexTest{
					{
						TestInput:   `{"data": "test data", "type": "analysis"}`,
						Groundtruth: StringPtr("Analysis result"),
					},
				},
			},
			options: PromptPexOptions{},
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
				Rules: "Test rules",
				PromptPexTests: []PromptPexTest{
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
			options: PromptPexOptions{},
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
				options: PromptPexOptions{},
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
			expected: []string{"not_valid"},
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
				Reasoning: StringPtr("Test reasoning"),
			},
			expected: "Test reasoning",
		},
		{
			name: "with groundtruth (short)",
			test: PromptPexTest{
				TestInput:   "test input",
				Groundtruth: StringPtr("Short groundtruth"),
			},
			expected: "Short groundtruth",
		},
		{
			name: "with groundtruth (long)",
			test: PromptPexTest{
				TestInput:   "test input",
				Groundtruth: StringPtr("This is a very long groundtruth that should be truncated"),
			},
			expected: "This is a very long groundtruth that should be t...",
		},
		{
			name: "with both reasoning and groundtruth (reasoning takes precedence)",
			test: PromptPexTest{
				TestInput:   "test input",
				Reasoning:   StringPtr("Test reasoning"),
				Groundtruth: StringPtr("Test groundtruth"),
			},
			expected: "Test reasoning",
		},
		{
			name: "with empty reasoning",
			test: PromptPexTest{
				TestInput:   "test input",
				Reasoning:   StringPtr(""),
				Groundtruth: StringPtr("Test groundtruth"),
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
