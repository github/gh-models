package run

import (
	"bytes"
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/internal/sse"
	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Run("NewRunCommand happy path", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		modelSummary := &azuremodels.ModelSummary{
			ID:           "openai/test-model-1",
			Name:         "test-model-1",
			FriendlyName: "Test Model 1",
			Task:         "chat-completion",
			Publisher:    "OpenAI",
			Summary:      "This is a test model",
			Version:      "1.0",
		}
		listModelsCallCount := 0
		client.MockListModels = func(ctx context.Context) ([]*azuremodels.ModelSummary, error) {
			listModelsCallCount++
			return []*azuremodels.ModelSummary{modelSummary}, nil
		}
		fakeMessageFromModel := "yes hello this is dog"
		chatChoice := azuremodels.ChatChoice{
			Message: &azuremodels.ChatChoiceMessage{
				Content: util.Ptr(fakeMessageFromModel),
				Role:    util.Ptr(string(azuremodels.ChatMessageRoleAssistant)),
			},
		}
		chatCompletion := azuremodels.ChatCompletion{Choices: []azuremodels.ChatChoice{chatChoice}}
		chatResp := &azuremodels.ChatCompletionResponse{
			Reader: sse.NewMockEventReader([]azuremodels.ChatCompletion{chatCompletion}),
		}
		getChatCompletionCallCount := 0
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			getChatCompletionCallCount++
			return chatResp, nil
		}
		buf := new(bytes.Buffer)
		cfg := command.NewConfig(buf, buf, client, true, 80)
		runCmd := NewRunCommand(cfg)
		runCmd.SetArgs([]string{modelSummary.ID, "this is my prompt"})

		_, err := runCmd.ExecuteC()

		require.NoError(t, err)
		require.Equal(t, 1, listModelsCallCount)
		require.Equal(t, 1, getChatCompletionCallCount)
		output := buf.String()
		require.Contains(t, output, fakeMessageFromModel)
	})

	t.Run("--help prints usage info", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		runCmd := NewRunCommand(nil)
		runCmd.SetOut(outBuf)
		runCmd.SetErr(errBuf)
		runCmd.SetArgs([]string{"--help"})

		err := runCmd.Help()

		require.NoError(t, err)
		output := outBuf.String()
		require.Contains(t, output, "Use `gh models run` to run in interactive mode. It will provide a list of the current\nmodels and allow you to select the one you want to run an inference with.")
		require.Regexp(t, regexp.MustCompile(`--max-tokens string\s+Limit the maximum tokens for the model response\.`), output)
		require.Regexp(t, regexp.MustCompile(`--system-prompt string\s+Prompt the system\.`), output)
		require.Regexp(t, regexp.MustCompile(`--temperature string\s+Controls randomness in the response, use lower to be more deterministic\.`), output)
		require.Regexp(t, regexp.MustCompile(`--top-p string\s+Controls text diversity by selecting the most probable words until a set probability is reached\.`), output)
		require.Empty(t, errBuf.String())
	})

	t.Run("--file pre-loads YAML from file", func(t *testing.T) {
		const yamlBody = `
name: Text Summarizer
description: Summarizes input text concisely
model: openai/test-model
modelParameters:
  temperature: 0.5
messages:
  - role: system
    content: You are a text summarizer.
  - role: user
    content: Hello there!
`
		tmp, err := os.CreateTemp(t.TempDir(), "*.prompt.yml")
		require.NoError(t, err)
		_, err = tmp.WriteString(yamlBody)
		require.NoError(t, err)
		require.NoError(t, tmp.Close())

		client := azuremodels.NewMockClient()
		modelSummary := &azuremodels.ModelSummary{
			ID:        "openai/test-model",
			Name:      "test-model",
			Publisher: "openai",
			Task:      "chat-completion",
		}
		client.MockListModels = func(ctx context.Context) ([]*azuremodels.ModelSummary, error) {
			return []*azuremodels.ModelSummary{modelSummary}, nil
		}

		var capturedReq azuremodels.ChatCompletionOptions
		reply := "Summary - foo"
		chatCompletion := azuremodels.ChatCompletion{
			Choices: []azuremodels.ChatChoice{{
				Message: &azuremodels.ChatChoiceMessage{
					Content: util.Ptr(reply),
					Role:    util.Ptr(string(azuremodels.ChatMessageRoleAssistant)),
				},
			}},
		}
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			capturedReq = opt
			return &azuremodels.ChatCompletionResponse{
				Reader: sse.NewMockEventReader([]azuremodels.ChatCompletion{chatCompletion}),
			}, nil
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)
		runCmd := NewRunCommand(cfg)
		runCmd.SetArgs([]string{
			"--file", tmp.Name(),
			"openai/test-model",
		})

		_, err = runCmd.ExecuteC()
		require.NoError(t, err)

		require.Equal(t, 2, len(capturedReq.Messages))
		require.Equal(t, "You are a text summarizer.", *capturedReq.Messages[0].Content)
		require.Equal(t, "Hello there!", *capturedReq.Messages[1].Content)

		require.NotNil(t, capturedReq.Temperature)
		require.Equal(t, 0.5, *capturedReq.Temperature)

		require.Contains(t, out.String(), reply) // response streamed to output
	})

	t.Run("--file with {{input}} placeholder is substituted with initial prompt and stdin", func(t *testing.T) {
		const yamlBody = `
name: Summarizer
description: Summarizes input text
model: openai/test-model
messages:
  - role: system
    content: You are a text summarizer.
  - role: user
    content: "{{input}}"
`

		tmp, err := os.CreateTemp(t.TempDir(), "*.prompt.yml")
		require.NoError(t, err)
		_, err = tmp.WriteString(yamlBody)
		require.NoError(t, err)
		require.NoError(t, tmp.Close())

		client := azuremodels.NewMockClient()
		modelSummary := &azuremodels.ModelSummary{
			ID:        "openai/test-model",
			Name:      "test-model",
			Publisher: "openai",
			Task:      "chat-completion",
		}
		client.MockListModels = func(ctx context.Context) ([]*azuremodels.ModelSummary, error) {
			return []*azuremodels.ModelSummary{modelSummary}, nil
		}

		var capturedReq azuremodels.ChatCompletionOptions
		reply := "Summary - bar"
		chatCompletion := azuremodels.ChatCompletion{
			Choices: []azuremodels.ChatChoice{{
				Message: &azuremodels.ChatChoiceMessage{
					Content: util.Ptr(reply),
					Role:    util.Ptr(string(azuremodels.ChatMessageRoleAssistant)),
				},
			}},
		}
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			capturedReq = opt
			return &azuremodels.ChatCompletionResponse{
				Reader: sse.NewMockEventReader([]azuremodels.ChatCompletion{chatCompletion}),
			}, nil
		}

		// create a pipe to fake stdin so that isPipe(os.Stdin)==true
		r, w, err := os.Pipe()
		require.NoError(t, err)
		oldStdin := os.Stdin
		os.Stdin = r
		defer func() { os.Stdin = oldStdin }()
		piped := "Hello there!"
		go func() {
			_, _ = w.Write([]byte(piped))
			_ = w.Close()
		}()

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		initialPrompt := "Please summarize the provided text."
		runCmd := NewRunCommand(cfg)
		runCmd.SetArgs([]string{
			"--file", tmp.Name(),
			"openai/test-model",
			initialPrompt,
		})

		_, err = runCmd.ExecuteC()
		require.NoError(t, err)

		require.Len(t, capturedReq.Messages, 2)
		require.Equal(t, "You are a text summarizer.", *capturedReq.Messages[0].Content)
		require.Equal(t, initialPrompt+"\n"+piped, *capturedReq.Messages[1].Content) // {{input}} -> "Please summarize the provided text.\nHello there!"

		require.Contains(t, out.String(), reply)
	})

	t.Run("cli flags override params set in the prompt.yaml file", func(t *testing.T) {
		// Begin setup:
		const yamlBody = `
    name: Example Prompt
    description: Example description
    model: openai/example-model
    modelParameters:
      maxTokens: 300
      temperature: 0.8
      topP: 0.9
    messages:
      - role: system
        content: System message
      - role: user
        content: User message
    `
		tmp, err := os.CreateTemp(t.TempDir(), "*.prompt.yaml")
		require.NoError(t, err)
		_, err = tmp.WriteString(yamlBody)
		require.NoError(t, err)
		require.NoError(t, tmp.Close())

		client := azuremodels.NewMockClient()
		modelSummary := &azuremodels.ModelSummary{
			ID:        "openai/example-model",
			Name:      "example-model",
			Publisher: "openai",
			Task:      "chat-completion",
		}
		modelSummary2 := &azuremodels.ModelSummary{
			ID:        "openai/example-model-4o-mini-plus",
			Name:      "example-model-4o-mini-plus",
			Publisher: "openai",
			Task:      "chat-completion",
		}

		client.MockListModels = func(ctx context.Context) ([]*azuremodels.
			ModelSummary, error) {
			return []*azuremodels.ModelSummary{modelSummary, modelSummary2}, nil
		}

		var capturedReq azuremodels.ChatCompletionOptions
		reply := "hello"
		chatCompletion := azuremodels.ChatCompletion{
			Choices: []azuremodels.ChatChoice{{
				Message: &azuremodels.ChatChoiceMessage{
					Content: util.Ptr(reply),
					Role:    util.Ptr(string(azuremodels.ChatMessageRoleAssistant)),
				},
			}},
		}

		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			capturedReq = opt
			return &azuremodels.ChatCompletionResponse{
				Reader: sse.NewMockEventReader([]azuremodels.ChatCompletion{chatCompletion}),
			}, nil
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)
		runCmd := NewRunCommand(cfg)

		// End setup.
		// ---
		// We're finally ready to start making assertions.

		// Test case 1: with no flags, the model params come from the YAML file
		runCmd.SetArgs([]string{
			"--file", tmp.Name(),
		})

		_, err = runCmd.ExecuteC()
		require.NoError(t, err)

		require.Equal(t, "openai/example-model", capturedReq.Model)
		require.Equal(t, 300, *capturedReq.MaxTokens)
		require.Equal(t, 0.8, *capturedReq.Temperature)
		require.Equal(t, 0.9, *capturedReq.TopP)

		require.Equal(t, "System message", *capturedReq.Messages[0].Content)
		require.Equal(t, "User message", *capturedReq.Messages[1].Content)

		// Hooray!
		// Test case 2: values from flags override the params from the YAML file
		runCmd = NewRunCommand(cfg)
		runCmd.SetArgs([]string{
			"openai/example-model-4o-mini-plus",
			"--file", tmp.Name(),
			"--max-tokens", "150",
			"--temperature", "0.1",
			"--top-p", "0.3",
		})

		_, err = runCmd.ExecuteC()
		require.NoError(t, err)

		require.Equal(t, "openai/example-model-4o-mini-plus", capturedReq.Model)
		require.Equal(t, 150, *capturedReq.MaxTokens)
		require.Equal(t, 0.1, *capturedReq.Temperature)
		require.Equal(t, 0.3, *capturedReq.TopP)

		require.Equal(t, "System message", *capturedReq.Messages[0].Content)
		require.Equal(t, "User message", *capturedReq.Messages[1].Content)
	})
}

func TestParseTemplateVariables(t *testing.T) {
	tests := []struct {
		name      string
		varFlags  []string
		expected  map[string]string
		expectErr bool
	}{
		{
			name:     "empty vars",
			varFlags: []string{},
			expected: map[string]string{},
		},
		{
			name:     "single var",
			varFlags: []string{"name=John"},
			expected: map[string]string{"name": "John"},
		},
		{
			name:     "multiple vars",
			varFlags: []string{"name=John", "age=25", "city=New York"},
			expected: map[string]string{"name": "John", "age": "25", "city": "New York"},
		},
		{
			name:     "multi-word values",
			varFlags: []string{"full_name=John Smith", "description=A senior developer"},
			expected: map[string]string{"full_name": "John Smith", "description": "A senior developer"},
		},
		{
			name:     "value with equals sign",
			varFlags: []string{"equation=x = y + 2"},
			expected: map[string]string{"equation": "x = y + 2"},
		},
		{
			name:     "empty strings are skipped",
			varFlags: []string{"", "name=John", "  "},
			expected: map[string]string{"name": "John"},
		},
		{
			name:      "invalid format - no equals",
			varFlags:  []string{"invalid"},
			expectErr: true,
		},
		{
			name:      "invalid format - empty key",
			varFlags:  []string{"=value"},
			expectErr: true,
		},
		{
			name:      "duplicate keys",
			varFlags:  []string{"name=John", "name=Jane"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.StringSlice("var", tt.varFlags, "test flag")

			result, err := parseTemplateVariables(flags)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidateModelName(t *testing.T) {
	tests := []struct {
		name          string
		modelName     string
		expectedModel string
		expectError   bool
	}{
		{
			name:          "custom provider skips validation",
			modelName:     "custom/mycompany/custom-model",
			expectedModel: "custom/mycompany/custom-model",
			expectError:   false,
		},
		{
			name:          "azureml provider requires validation",
			modelName:     "openai/gpt-4",
			expectedModel: "openai/gpt-4",
			expectError:   false,
		},
		{
			name:        "invalid model format",
			modelName:   "invalid-format",
			expectError: true,
		},
		{
			name:        "nonexistent azureml model",
			modelName:   "nonexistent/model",
			expectError: true,
		},
	}

	// Create a mock model for testing
	mockModel := &azuremodels.ModelSummary{
		ID:        "openai/gpt-4",
		Name:      "gpt-4",
		Publisher: "openai",
		Task:      "chat-completion",
	}
	models := []*azuremodels.ModelSummary{mockModel}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateModelName(tt.modelName, models)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedModel, result)
			}
		})
	}
}
