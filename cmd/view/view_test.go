package view

import (
	"bytes"
	"context"
	"testing"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/stretchr/testify/require"
)

func TestView(t *testing.T) {
	t.Run("NewViewCommand happy path", func(t *testing.T) {
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
		getModelDetailsCallCount := 0
		modelDetails := &azuremodels.ModelDetails{
			Description:               "Fake description",
			Evaluation:                "Fake evaluation",
			License:                   "MIT",
			LicenseDescription:        "This is a test license",
			Tags:                      []string{"tag1", "tag2"},
			SupportedInputModalities:  []string{"text", "carrier-pigeon"},
			SupportedOutputModalities: []string{"underwater-signals"},
			SupportedLanguages:        []string{"English", "Spanish"},
			MaxOutputTokens:           123,
			MaxInputTokens:            456,
			RateLimitTier:             "mediumish",
		}
		client.MockGetModelDetails = func(ctx context.Context, registryName, modelName, version string) (*azuremodels.ModelDetails, error) {
			getModelDetailsCallCount++
			return modelDetails, nil
		}
		buf := new(bytes.Buffer)
		cfg := command.NewConfig(buf, buf, client, true, 80)
		viewCmd := NewViewCommand(cfg)
		viewCmd.SetArgs([]string{azuremodels.FormatIdentifier(modelSummary.Publisher, modelSummary.Name)})

		_, err := viewCmd.ExecuteC()

		require.NoError(t, err)
		require.Equal(t, 1, listModelsCallCount)
		require.Equal(t, 1, getModelDetailsCallCount)
		output := buf.String()
		require.Contains(t, output, "Display name:")
		require.Contains(t, output, modelSummary.FriendlyName)
		require.Contains(t, output, "Model name:")
		require.Contains(t, output, modelSummary.Name)
		require.Contains(t, output, "Publisher:")
		require.Contains(t, output, modelSummary.Publisher)
		require.Contains(t, output, "Summary:")
		require.Contains(t, output, modelSummary.Summary)
		require.Contains(t, output, "Context:")
		require.Contains(t, output, "up to 456 input tokens and 123 output tokens")
		require.Contains(t, output, "Rate limit tier:")
		require.Contains(t, output, "mediumish")
		require.Contains(t, output, "Tags:")
		require.Contains(t, output, "tag1, tag2")
		require.Contains(t, output, "Supported input types:")
		require.Contains(t, output, "text, carrier-pigeon")
		require.Contains(t, output, "Supported output types:")
		require.Contains(t, output, "underwater-signals")
		require.Contains(t, output, "Supported languages:")
		require.Contains(t, output, "English, Spanish")
		require.Contains(t, output, "License:")
		require.Contains(t, output, modelDetails.License)
		require.Contains(t, output, "License description:")
		require.Contains(t, output, modelDetails.LicenseDescription)
		require.Contains(t, output, "Description:")
		require.Contains(t, output, modelDetails.Description)
		require.Contains(t, output, "Evaluation:")
		require.Contains(t, output, modelDetails.Evaluation)
	})

	t.Run("--help prints usage info", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		viewCmd := NewViewCommand(nil)
		viewCmd.SetOut(outBuf)
		viewCmd.SetErr(errBuf)
		viewCmd.SetArgs([]string{"--help"})

		err := viewCmd.Help()

		require.NoError(t, err)
		require.Contains(t, outBuf.String(), "Use `gh models view` to run in interactive mode. It will provide a list of the current\nmodels and allow you to select the one you want information about.")
		require.Empty(t, errBuf.String())
	})
}
