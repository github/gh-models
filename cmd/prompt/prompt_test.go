package prompt

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/internal/sse"
	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestPrompt(t *testing.T) {
	t.Run("NewPromptCommand happy path with model in prompt file", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		fakeMessageFromModel := "Here's a summary of the text in bullet points:\n\n- Point 1\n- Point 2"
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

			// Verify that the request contains the expected data
			require.Equal(t, "gpt-4o-mini", opt.Model)

			// Find the user message and verify it contains our test content
			var userMsgFound bool
			var systemMsgFound bool
			for _, msg := range opt.Messages {
				if msg.Role == azuremodels.ChatMessageRoleUser && *msg.Content == "Summarize the given text:\n\n<text>\nTest content to summarize\n</text>" {
					userMsgFound = true
				}
				if msg.Role == azuremodels.ChatMessageRoleSystem && *msg.Content == "You are a text summarizer. Your only job is to summarize a given text to you. I want you to summarize in bullet points." {
					systemMsgFound = true
				}
			}
			require.True(t, userMsgFound, "User message with correct content not found")
			require.True(t, systemMsgFound, "System message with correct content not found")

			return chatResp, nil
		}

		// Create a temp prompt file
		promptContent := `---
name: Summarizer
description: Summarizes a given text
model: gpt-4o-mini
model_parameters:
  temperature: 0.5
---

system:
You are a text summarizer. Your only job is to summarize a given text to you. I want you to summarize in bullet points.

user:
Summarize the given text:

<text>
{{text}}
</text>
`
		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test-prompt.md")
		err := os.WriteFile(promptFilePath, []byte(promptContent), 0644)
		require.NoError(t, err)

		buf := new(bytes.Buffer)
		cfg := command.NewConfig(buf, buf, client, true, 80)
		promptCmd := NewPromptCommand(cfg)
		promptCmd.SetArgs([]string{promptFilePath, "text=Test content to summarize"})

		_, err = promptCmd.ExecuteC()

		require.NoError(t, err)
		require.Equal(t, 1, getChatCompletionCallCount)
		output := buf.String()
		require.Contains(t, output, fakeMessageFromModel)
	})

	t.Run("NewPromptCommand error on missing prompt file", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		buf := new(bytes.Buffer)
		cfg := command.NewConfig(buf, buf, client, true, 80)
		promptCmd := NewPromptCommand(cfg)

		// Don't provide any args
		promptCmd.SetArgs([]string{})

		_, err := promptCmd.ExecuteC()
		require.Error(t, err)
		require.Contains(t, err.Error(), "prompt file path is required")
	})

	t.Run("NewPromptCommand prompts for model when not in prompt file", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		modelSummary := &azuremodels.ModelSummary{
			ID:           "test-id-1",
			Name:         "gpt-4o-mini",
			FriendlyName: "OpenAI GPT-4o mini",
			Task:         "chat-completion",
			Publisher:    "OpenAI",
			Summary:      "This is a test model",
			Version:      "1.0",
			RegistryName: "azure-openai",
		}
		client.MockListModels = func(ctx context.Context) ([]*azuremodels.ModelSummary, error) {
			return []*azuremodels.ModelSummary{modelSummary}, nil
		}

		// This test can't easily simulate interactive selection, so we'll just check that it
		// responds with an error when we don't provide a model and it tries to prompt for one

		// Create a temp prompt file without a model
		promptContent := `---
name: Summarizer
description: Summarizes a given text
---

system:
You are a text summarizer. Your only job is to summarize a given text to you.

user:
Summarize the given text:

<text>
{{text}}
</text>
`
		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test-prompt-no-model.md")
		err := os.WriteFile(promptFilePath, []byte(promptContent), 0644)
		require.NoError(t, err)

		buf := new(bytes.Buffer)
		cfg := command.NewConfig(buf, buf, client, true, 80)
		promptCmd := NewPromptCommand(cfg)
		promptCmd.SetArgs([]string{promptFilePath, "text=Test content to summarize"})

		// This will error because in test we can't respond to the interactive prompt
		_, err = promptCmd.ExecuteC()
		require.Error(t, err)
	})

	t.Run("NewPromptCommand error on invalid key-value argument", func(t *testing.T) {
		client := azuremodels.NewMockClient()

		// Create a temp prompt file
		promptContent := `---
name: Summarizer
description: Summarizes a given text
model: gpt-4o-mini
---

system:
You are a text summarizer.

user:
Summarize: {{text}}
`
		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test-prompt.md")
		err := os.WriteFile(promptFilePath, []byte(promptContent), 0644)
		require.NoError(t, err)

		buf := new(bytes.Buffer)
		cfg := command.NewConfig(buf, buf, client, true, 80)
		promptCmd := NewPromptCommand(cfg)
		promptCmd.SetArgs([]string{promptFilePath, "invalid-format"}) // Invalid format, should be key=value

		_, err = promptCmd.ExecuteC()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid argument: invalid-format, expected key=value")
	})

	t.Run("error handling from GetChatCompletionStream", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		client.MockListModels = func(ctx context.Context) ([]*azuremodels.ModelSummary, error) {
			return []*azuremodels.ModelSummary{{
				Name:         "gpt-4o-mini",
				FriendlyName: "GPT-4o mini",
				Task:         "chat-completion",
			}}, nil
		}
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions) (*azuremodels.ChatCompletionResponse, error) {
			return nil, errors.New("API error")
		}

		// Create a temp prompt file
		promptContent := `---
name: Summarizer
model: gpt-4o-mini
---

user:
Hello
`
		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test-prompt.md")
		err := os.WriteFile(promptFilePath, []byte(promptContent), 0644)
		require.NoError(t, err)

		buf := new(bytes.Buffer)
		cfg := command.NewConfig(buf, buf, client, true, 80)
		promptCmd := NewPromptCommand(cfg)
		promptCmd.SetArgs([]string{promptFilePath})

		_, err = promptCmd.ExecuteC()
		require.Error(t, err)
		require.Contains(t, err.Error(), "API error")
	})

	t.Run("--help prints usage info", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		promptCmd := NewPromptCommand(nil)
		promptCmd.SetOut(outBuf)
		promptCmd.SetErr(errBuf)
		promptCmd.SetArgs([]string{"--help"})

		err := promptCmd.Help()

		require.NoError(t, err)
		output := outBuf.String()
		require.Contains(t, output, "Prompts the specified model with the given prompt. Replace any {{placeholders}} in the prompts with \"key=value\" arguments.")
		require.Contains(t, output, "gh models prompt my-prompt.prompt.md")
		require.Empty(t, errBuf.String())
	})
}
