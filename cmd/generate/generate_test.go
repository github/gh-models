package generate

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/internal/sse"
	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestNewGenerateCommand(t *testing.T) {
	t.Run("creates command with correct structure", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(new(bytes.Buffer), new(bytes.Buffer), client, true, 80)

		cmd := NewGenerateCommand(cfg)

		require.Equal(t, "generate [prompt-file]", cmd.Use)
		require.Equal(t, "Generate tests using PromptPex", cmd.Short)
		require.Contains(t, cmd.Long, "PromptPex methodology")
		require.True(t, cmd.Args != nil) // Should have ExactArgs(1)

		// Check that flags are added
		flags := cmd.Flags()
		require.True(t, flags.Lookup("org") != nil)
		require.True(t, flags.Lookup("effort") != nil)
		require.True(t, flags.Lookup("models-under-test") != nil)
		require.True(t, flags.Lookup("groundtruth-model") != nil)
		require.True(t, flags.Lookup("tests-per-rule") != nil)
		require.True(t, flags.Lookup("runs-per-test") != nil)
		require.True(t, flags.Lookup("test-expansions") != nil)
		require.True(t, flags.Lookup("rate-tests") != nil)
		require.True(t, flags.Lookup("evals") != nil)
		require.True(t, flags.Lookup("eval-models") != nil)
		require.True(t, flags.Lookup("custom-metric") != nil)
		require.True(t, flags.Lookup("temperature") != nil)
	})

	t.Run("--help prints usage info", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		cmd := NewGenerateCommand(nil)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--help"})

		err := cmd.Help()

		require.NoError(t, err)
		output := outBuf.String()
		require.Contains(t, output, "Augment prompt.yml file with generated test cases")
		require.Contains(t, output, "PromptPex methodology")
		require.Regexp(t, regexp.MustCompile(`--effort string\s+Effort level`), output)
		require.Regexp(t, regexp.MustCompile(`--models-under-test strings\s+Models to test`), output)
		require.Regexp(t, regexp.MustCompile(`--groundtruth-model string\s+Model to use for generating groundtruth`), output)
		require.Regexp(t, regexp.MustCompile(`--temperature float\s+Temperature for model inference`), output)
		require.Empty(t, errBuf.String())
	})
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(*testing.T, PromptPexOptions)
	}{
		{
			name: "default options preserve initial state",
			args: []string{},
			validate: func(t *testing.T, opts PromptPexOptions) {
				require.Equal(t, 3, *opts.TestsPerRule)
				require.Equal(t, 2, *opts.RunsPerTest)
				require.Equal(t, 0, *opts.TestExpansions)
				require.Equal(t, false, *opts.RateTests)
				require.Equal(t, false, *opts.Evals)
			},
		},
		{
			name: "effort flag is set",
			args: []string{"--effort", "medium"},
			validate: func(t *testing.T, opts PromptPexOptions) {
				require.NotNil(t, opts.Effort)
				require.Equal(t, "medium", *opts.Effort)
			},
		},
		{
			name: "models under test flag",
			args: []string{"--models-under-test", "openai/gpt-4o", "--models-under-test", "openai/gpt-4o-mini"},
			validate: func(t *testing.T, opts PromptPexOptions) {
				require.Equal(t, []string{"openai/gpt-4o", "openai/gpt-4o-mini"}, opts.ModelsUnderTest)
			},
		},
		{
			name: "groundtruth model flag",
			args: []string{"--groundtruth-model", "openai/gpt-4o"},
			validate: func(t *testing.T, opts PromptPexOptions) {
				require.NotNil(t, opts.GroundtruthModel)
				require.Equal(t, "openai/gpt-4o", *opts.GroundtruthModel)
			},
		},
		{
			name: "numeric flags",
			args: []string{"--tests-per-rule", "10", "--runs-per-test", "3", "--test-expansions", "2"},
			validate: func(t *testing.T, opts PromptPexOptions) {
				require.NotNil(t, opts.TestsPerRule)
				require.Equal(t, 10, *opts.TestsPerRule)
				require.NotNil(t, opts.RunsPerTest)
				require.Equal(t, 3, *opts.RunsPerTest)
				require.NotNil(t, opts.TestExpansions)
				require.Equal(t, 2, *opts.TestExpansions)
			},
		},
		{
			name: "boolean flags",
			args: []string{"--rate-tests", "--evals"},
			validate: func(t *testing.T, opts PromptPexOptions) {
				require.NotNil(t, opts.RateTests)
				require.Equal(t, true, *opts.RateTests)
				require.NotNil(t, opts.Evals)
				require.Equal(t, true, *opts.Evals)
			},
		},
		{
			name: "temperature flag",
			args: []string{"--temperature", "0.7"},
			validate: func(t *testing.T, opts PromptPexOptions) {
				require.NotNil(t, opts.Temperature)
				require.Equal(t, 0.7, *opts.Temperature)
			},
		},
		{
			name: "custom metric flag",
			args: []string{"--custom-metric", "Rate the quality of response from 1-10"},
			validate: func(t *testing.T, opts PromptPexOptions) {
				require.NotNil(t, opts.CustomMetric)
				require.Equal(t, "Rate the quality of response from 1-10", *opts.CustomMetric)
			},
		},
		{
			name: "eval models flag",
			args: []string{"--eval-models", "openai/gpt-4o", "--eval-models", "openai/gpt-4o-mini"},
			validate: func(t *testing.T, opts PromptPexOptions) {
				require.Equal(t, []string{"openai/gpt-4o", "openai/gpt-4o-mini"}, opts.EvalModels)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary command to parse flags
			cmd := NewGenerateCommand(nil)
			cmd.SetArgs(append(tt.args, "dummy.yml")) // Add required positional arg

			// Parse flags but don't execute
			err := cmd.ParseFlags(tt.args)
			require.NoError(t, err)

			// Parse options from the flags
			options := GetDefaultOptions()
			err = ParseFlags(cmd, &options)
			require.NoError(t, err)

			// Validate using the test-specific validation function
			tt.validate(t, options)
		})
	}
}

func TestGenerateCommandExecution(t *testing.T) {
	t.Run("successful execution with mock prompt file", func(t *testing.T) {
		// Create test prompt file
		const yamlBody = `
name: Test Sentiment Analysis
description: Analyze sentiment of user input
model: openai/gpt-4o-mini
messages:
  - role: system
    content: You are a sentiment analysis expert.
  - role: user
    content: "Classify sentiment: {{text}}"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		// Setup mock client
		client := azuremodels.NewMockClient()
		callCount := 0
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			callCount++
			var response string

			// Mock different responses for different pipeline stages
			switch callCount {
			case 1: // Intent generation
				response = "This prompt analyzes the sentiment of text input to classify it as positive, negative, or neutral."
			case 2: // Input spec generation
				response = "Input: text (string) - Any text to analyze for sentiment"
			case 3: // Output rules generation
				response = "1. Output must be one of: positive, negative, neutral\n2. Response should be lowercase\n3. No additional text or explanation"
			case 4: // Inverse rules generation
				response = "1. Output should not contain multiple sentiment words\n2. Output should not be uppercase\n3. Output should not contain explanations"
			case 5: // Test generation
				response = `[
					{
						"scenario": "Positive sentiment detection",
						"testinput": "I love this amazing product!",
						"reasoning": "Tests ability to detect clear positive sentiment"
					},
					{
						"scenario": "Negative sentiment detection", 
						"testinput": "This is terrible and disappointing",
						"reasoning": "Tests ability to detect clear negative sentiment"
					},
					{
						"scenario": "Neutral sentiment detection",
						"testinput": "The weather is cloudy today",
						"reasoning": "Tests ability to detect neutral sentiment"
					}
				]`
			default:
				response = "Test response"
			}

			chatCompletion := azuremodels.ChatCompletion{
				Choices: []azuremodels.ChatChoice{
					{
						Message: &azuremodels.ChatChoiceMessage{
							Content: util.Ptr(response),
							Role:    util.Ptr(string(azuremodels.ChatMessageRoleAssistant)),
						},
					},
				},
			}

			return &azuremodels.ChatCompletionResponse{
				Reader: sse.NewMockEventReader([]azuremodels.ChatCompletion{chatCompletion}),
			}, nil
		}

		// Create config and run command
		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewGenerateCommand(cfg)
		cmd.SetArgs([]string{
			"--effort", "low",
			promptFile,
		})

		err = cmd.Execute()
		require.NoError(t, err)

		// Check that pipeline stages were called
		require.Greater(t, callCount, 3, "Should have called multiple pipeline stages")

		output := out.String()
		require.Contains(t, output, "Running pipeline for prompt")
		require.Contains(t, output, "Generating intent")
		require.Contains(t, output, "Generating input specification")
		require.Contains(t, output, "Generating output rules")
		require.Contains(t, output, "Generating tests")
		require.Contains(t, output, "Pipeline completed successfully")
	})

	t.Run("fails with invalid prompt file", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewGenerateCommand(cfg)
		cmd.SetArgs([]string{"nonexistent.yml"})

		err := cmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to create context")
	})

	t.Run("handles LLM errors gracefully", func(t *testing.T) {
		// Create test prompt file
		const yamlBody = `
name: Test Prompt
description: Test description
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Test prompt"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		// Setup mock client to return error
		client := azuremodels.NewMockClient()
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			return nil, errors.New("Mock API error")
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewGenerateCommand(cfg)
		cmd.SetArgs([]string{promptFile})

		err = cmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "pipeline failed")
	})

	t.Run("executes with models under test", func(t *testing.T) {
		// Create test prompt file
		const yamlBody = `
name: Simple Test
description: Simple test prompt
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Say hello"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		// Setup mock client
		client := azuremodels.NewMockClient()
		callCount := 0
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			callCount++
			var response string

			if len(opt.Messages) > 0 && opt.Messages[0].Content != nil {
				content := *opt.Messages[0].Content
				// Generate different responses for different pipeline stages
				if contains(content, "intent") && !contains(content, "test") {
					response = "This prompt generates a greeting."
				} else if contains(content, "input") && !contains(content, "test") {
					response = "No specific input required."
				} else if contains(content, "rules") && !contains(content, "test") {
					response = "1. Must contain greeting\n2. Should be friendly"
				} else {
					// For any other prompt, especially test generation, return valid JSON
					response = `[{"scenario": "Basic greeting", "testinput": "Hello", "reasoning": "Tests basic functionality"}]`
				}
			} else {
				response = `[{"scenario": "Default test", "testinput": "test", "reasoning": "Default test case"}]`
			}

			chatCompletion := azuremodels.ChatCompletion{
				Choices: []azuremodels.ChatChoice{
					{
						Message: &azuremodels.ChatChoiceMessage{
							Content: util.Ptr(response),
							Role:    util.Ptr(string(azuremodels.ChatMessageRoleAssistant)),
						},
					},
				},
			}

			return &azuremodels.ChatCompletionResponse{
				Reader: sse.NewMockEventReader([]azuremodels.ChatCompletion{chatCompletion}),
			}, nil
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewGenerateCommand(cfg)
		cmd.SetArgs([]string{
			"--models-under-test", "openai/gpt-4o-mini",
			"--runs-per-test", "1",
			promptFile,
		})

		err = cmd.Execute()
		require.NoError(t, err)

		output := out.String()
		require.Contains(t, output, "Running tests against models")
		require.Contains(t, output, "openai/gpt-4o-mini")
	})

	t.Run("executes with groundtruth model", func(t *testing.T) {
		// Create test prompt file
		const yamlBody = `
name: Groundtruth Test
description: Test with groundtruth generation
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Generate response"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		// Setup mock client
		client := azuremodels.NewMockClient()
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			var response string
			if len(opt.Messages) > 0 && opt.Messages[0].Content != nil {
				content := *opt.Messages[0].Content
				if contains(content, "intent") && !contains(content, "test") {
					response = "This prompt generates responses."
				} else if contains(content, "input") && !contains(content, "test") {
					response = "Input: Any text input"
				} else if contains(content, "rules") && !contains(content, "test") {
					response = "1. Response should be relevant\n2. Response should be helpful"
				} else {
					response = `[{"scenario": "Response generation", "testinput": "Input", "reasoning": "Tests generation"}]`
				}
			} else {
				response = `[{"scenario": "Default test", "testinput": "test", "reasoning": "Default test case"}]`
			}

			chatCompletion := azuremodels.ChatCompletion{
				Choices: []azuremodels.ChatChoice{
					{
						Message: &azuremodels.ChatChoiceMessage{
							Content: util.Ptr(response),
							Role:    util.Ptr(string(azuremodels.ChatMessageRoleAssistant)),
						},
					},
				},
			}

			return &azuremodels.ChatCompletionResponse{
				Reader: sse.NewMockEventReader([]azuremodels.ChatCompletion{chatCompletion}),
			}, nil
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewGenerateCommand(cfg)
		cmd.SetArgs([]string{
			"--groundtruth-model", "openai/gpt-4o",
			promptFile,
		})

		err = cmd.Execute()
		require.NoError(t, err)

		output := out.String()
		require.Contains(t, output, "Generating groundtruth with model")
		require.Contains(t, output, "openai/gpt-4o")
	})

	t.Run("executes with test expansions", func(t *testing.T) {
		// Create test prompt file
		const yamlBody = `
name: Expansion Test
description: Test with test expansion
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Test input"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		// Setup mock client
		client := azuremodels.NewMockClient()
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			var response string
			if len(opt.Messages) > 0 && opt.Messages[0].Content != nil {
				content := *opt.Messages[0].Content
				if contains(content, "intent") && !contains(content, "test") {
					response = "This prompt processes test input."
				} else if contains(content, "input") && !contains(content, "test") {
					response = "Input: Test input data"
				} else if contains(content, "rules") && !contains(content, "test") {
					response = "1. Output should be processed\n2. Output should be valid"
				} else if contains(content, "variations") {
					response = `[{"scenario": "Variation 1", "testinput": "Input variant 1", "reasoning": "Test variation"}]`
				} else {
					response = `[{"scenario": "Basic test", "testinput": "Input", "reasoning": "Basic test"}]`
				}
			} else {
				response = `[{"scenario": "Default test", "testinput": "test", "reasoning": "Default test case"}]`
			}

			chatCompletion := azuremodels.ChatCompletion{
				Choices: []azuremodels.ChatChoice{
					{
						Message: &azuremodels.ChatChoiceMessage{
							Content: util.Ptr(response),
							Role:    util.Ptr(string(azuremodels.ChatMessageRoleAssistant)),
						},
					},
				},
			}

			return &azuremodels.ChatCompletionResponse{
				Reader: sse.NewMockEventReader([]azuremodels.ChatCompletion{chatCompletion}),
			}, nil
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewGenerateCommand(cfg)
		cmd.SetArgs([]string{
			"--test-expansions", "1",
			promptFile,
		})

		err = cmd.Execute()
		require.NoError(t, err)

		output := out.String()
		require.Contains(t, output, "Expanding tests with 1 expansion phases")
	})

	t.Run("executes with evaluations", func(t *testing.T) {
		// Create test prompt file
		const yamlBody = `
name: Eval Test
description: Test with evaluations
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Test prompt"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		// Setup mock client
		client := azuremodels.NewMockClient()
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			var response string
			if len(opt.Messages) > 0 && opt.Messages[0].Content != nil {
				content := *opt.Messages[0].Content
				if contains(content, "intent") && !contains(content, "test") {
					response = "This prompt tests functionality."
				} else if contains(content, "input") && !contains(content, "test") {
					response = "Input: Test data"
				} else if contains(content, "rules") && !contains(content, "test") {
					response = "1. Output should be valid\n2. Output should be accurate"
				} else if contains(content, "Evaluate") && contains(content, "compliance") {
					response = "ok"
				} else {
					response = `[{"scenario": "Test scenario", "testinput": "Test input", "reasoning": "Test reasoning"}]`
				}
			} else {
				response = `[{"scenario": "Default test", "testinput": "test", "reasoning": "Default test case"}]`
			}

			chatCompletion := azuremodels.ChatCompletion{
				Choices: []azuremodels.ChatChoice{
					{
						Message: &azuremodels.ChatChoiceMessage{
							Content: util.Ptr(response),
							Role:    util.Ptr(string(azuremodels.ChatMessageRoleAssistant)),
						},
					},
				},
			}

			return &azuremodels.ChatCompletionResponse{
				Reader: sse.NewMockEventReader([]azuremodels.ChatCompletion{chatCompletion}),
			}, nil
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewGenerateCommand(cfg)
		cmd.SetArgs([]string{
			"--evals",
			"--eval-models", "openai/gpt-4o-mini",
			"--models-under-test", "openai/gpt-4o-mini",
			promptFile,
		})

		err = cmd.Execute()
		require.NoError(t, err)

		output := out.String()
		require.Contains(t, output, "Evaluating test results")
	})
}

func TestGenerateCommandWithValidPromptFile(t *testing.T) {
	t.Run("loads existing prompt file correctly", func(t *testing.T) {
		// Use the existing test prompt file
		promptFile := filepath.Join("..", "..", "examples", "test_generate.yml")

		// Setup mock client
		client := azuremodels.NewMockClient()
		callCount := 0
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			callCount++
			var response string

			// Mock different responses for different pipeline stages
			switch callCount {
			case 1: // Intent generation
				response = "This prompt analyzes sentiment of text input to classify it as positive, negative, or neutral."
			case 2: // Input spec generation
				response = "Input: text (string) - Any text to analyze for sentiment"
			case 3: // Output rules generation
				response = "1. Output must be one of: positive, negative, neutral\n2. Response should be lowercase\n3. No additional text or explanation"
			case 4: // Inverse rules generation
				response = "1. Output should not contain multiple sentiment words\n2. Output should not be uppercase\n3. Output should not contain explanations"
			case 5: // Test generation
				response = `[
					{
						"scenario": "Positive sentiment detection",
						"testinput": "I love this amazing product!",
						"reasoning": "Tests ability to detect clear positive sentiment"
					},
					{
						"scenario": "Negative sentiment detection", 
						"testinput": "This is terrible and disappointing",
						"reasoning": "Tests ability to detect clear negative sentiment"
					},
					{
						"scenario": "Neutral sentiment detection",
						"testinput": "The weather is cloudy today",
						"reasoning": "Tests ability to detect neutral sentiment"
					}
				]`
			default:
				response = "Test response"
			}

			chatCompletion := azuremodels.ChatCompletion{
				Choices: []azuremodels.ChatChoice{
					{
						Message: &azuremodels.ChatChoiceMessage{
							Content: util.Ptr(response),
							Role:    util.Ptr(string(azuremodels.ChatMessageRoleAssistant)),
						},
					},
				},
			}

			return &azuremodels.ChatCompletionResponse{
				Reader: sse.NewMockEventReader([]azuremodels.ChatCompletion{chatCompletion}),
			}, nil
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewGenerateCommand(cfg)
		cmd.SetArgs([]string{promptFile})

		err := cmd.Execute()
		require.NoError(t, err)

		output := out.String()
		require.Contains(t, output, "Running pipeline for prompt")
	})
}

func TestGenerateCommandHandlerContext(t *testing.T) {
	t.Run("creates context with valid prompt file", func(t *testing.T) {
		// Create test prompt file
		const yamlBody = `
name: Test Context Creation
description: Test description for context
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Test content"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		// Create handler
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(new(bytes.Buffer), new(bytes.Buffer), client, true, 100)
		options := GetDefaultOptions()

		handler := &generateCommandHandler{
			ctx:     context.Background(),
			cfg:     cfg,
			client:  client,
			options: options,
			org:     "",
		}

		// Test context creation
		ctx, err := handler.CreateContext(promptFile)
		require.NoError(t, err)
		require.NotNil(t, ctx)
		require.NotEmpty(t, ctx.RunID)
		require.True(t, ctx.RunID != "")
		require.Equal(t, "Test Context Creation", ctx.Prompt.Name)
		require.Equal(t, "Test description for context", ctx.Prompt.Description)
		require.Equal(t, options, ctx.Options)
	})

	t.Run("fails with invalid prompt file", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(new(bytes.Buffer), new(bytes.Buffer), client, true, 100)
		options := GetDefaultOptions()

		handler := &generateCommandHandler{
			ctx:     context.Background(),
			cfg:     cfg,
			client:  client,
			options: options,
			org:     "",
		}

		// Test with nonexistent file
		_, err := handler.CreateContext("nonexistent.yml")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to load prompt file")
	})
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return regexp.MustCompile("(?i)" + regexp.QuoteMeta(substr)).MatchString(s)
}
