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
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Run("NewRunCommand happy path", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		modelSummary := &azuremodels.ModelSummary{
			ID:           "test-id-1",
			Name:         "test-model-1",
			FriendlyName: "Test Model 1",
			Task:         "chat-completion",
			Publisher:    "OpenAI",
			Summary:      "This is a test model",
			Version:      "1.0",
			RegistryName: "azure-openai",
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
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions) (*azuremodels.ChatCompletionResponse, error) {
			getChatCompletionCallCount++
			return chatResp, nil
		}
		buf := new(bytes.Buffer)
		cfg := command.NewConfig(buf, buf, client, true, 80)
		runCmd := NewRunCommand(cfg)
		runCmd.SetArgs([]string{azuremodels.FormatIdentifier(modelSummary.Publisher, modelSummary.Name), "this is my prompt"})

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
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions) (*azuremodels.ChatCompletionResponse, error) {
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
			azuremodels.FormatIdentifier("openai", "test-model"),
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
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions) (*azuremodels.ChatCompletionResponse, error) {
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
			azuremodels.FormatIdentifier("openai", "test-model"),
			initialPrompt,
		})

		_, err = runCmd.ExecuteC()
		require.NoError(t, err)

		require.Len(t, capturedReq.Messages, 2)
		require.Equal(t, "You are a text summarizer.", *capturedReq.Messages[0].Content)
		require.Equal(t, initialPrompt+"\n"+piped, *capturedReq.Messages[1].Content) // {{input}} -> "Please summarize the provided text.\nHello there!"

		require.Contains(t, out.String(), reply)
	})
}
