package list

import (
	"bytes"
	"context"
	"testing"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("NewListCommand happy path", func(t *testing.T) {
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
		buf := new(bytes.Buffer)
		cfg := command.NewConfig(buf, buf, client, true, 80)
		listCmd := NewListCommand(cfg)

		_, err := listCmd.ExecuteC()

		require.NoError(t, err)
		require.Equal(t, 1, listModelsCallCount)
		output := buf.String()
		require.Contains(t, output, "Showing 1 available chat models")
		require.Contains(t, output, "DISPLAY NAME")
		require.Contains(t, output, "MODEL NAME")
		require.Contains(t, output, modelSummary.FriendlyName)
		require.Contains(t, output, modelSummary.Name)
	})

	t.Run("--help prints usage info", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		listCmd := NewListCommand(nil)
		listCmd.SetOut(outBuf)
		listCmd.SetErr(errBuf)
		listCmd.SetArgs([]string{"--help"})

		err := listCmd.Help()

		require.NoError(t, err)
		require.Contains(t, outBuf.String(), "Returns a list of models that are available to use via the CLI.\n\nValues from the \"MODEL NAME\" column can be used as the `[model]`\nargument in other commands.")
		require.Empty(t, errBuf.String())
	})
}
