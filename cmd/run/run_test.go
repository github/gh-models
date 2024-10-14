package run

import (
	"bytes"
	"context"
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
		runCmd.SetArgs([]string{modelSummary.Name, "this is my prompt"})

		_, err := runCmd.ExecuteC()

		require.NoError(t, err)
		require.Equal(t, 1, listModelsCallCount)
		require.Equal(t, 1, getChatCompletionCallCount)
		output := buf.String()
		require.Contains(t, output, fakeMessageFromModel)
	})
}
