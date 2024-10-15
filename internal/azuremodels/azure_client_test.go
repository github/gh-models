package azuremodels

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAzureClient(t *testing.T) {
	ctx := context.Background()

	t.Run("ListModels happy path", func(t *testing.T) {
		summary1 := modelCatalogSearchSummary{
			AssetID:        "test-id-1",
			Name:           "test-model-1",
			DisplayName:    "I Can't Believe It's Not a Real Model",
			InferenceTasks: []string{"this model has an inference task but the other model will not"},
			Publisher:      "OpenAI",
			Summary:        "This is a test model",
			Version:        "1.0",
			RegistryName:   "azure-openai",
		}
		summary2 := modelCatalogSearchSummary{
			AssetID:      "test-id-2",
			Name:         "test-model-2",
			DisplayName:  "Down the Rabbit-Hole",
			Publisher:    "Project Gutenberg",
			Summary:      "The first chapter of Alice's Adventures in Wonderland by Lewis Carroll.",
			Version:      "THE MILLENNIUM FULCRUM EDITION 3.0",
			RegistryName: "proj-gutenberg-website",
		}
		searchResponse := &modelCatalogSearchResponse{
			Summaries: []modelCatalogSearchSummary{summary1, summary2},
		}
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			require.Equal(t, "/", r.URL.Path)

			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(searchResponse)
			require.NoError(t, err)
		}))
		defer testServer.Close()
		cfg := &AzureClientConfig{ModelsURL: testServer.URL}
		httpClient := testServer.Client()
		client := NewAzureClient(httpClient, "fake-token-123abc", cfg)

		models, err := client.ListModels(ctx)

		require.NoError(t, err)
		require.NotNil(t, models)
		require.Equal(t, 2, len(models))
		require.Equal(t, summary1.AssetID, models[0].ID)
		require.Equal(t, summary2.AssetID, models[1].ID)
		require.Equal(t, summary1.Name, models[0].Name)
		require.Equal(t, summary2.Name, models[1].Name)
		require.Equal(t, summary1.DisplayName, models[0].FriendlyName)
		require.Equal(t, summary2.DisplayName, models[1].FriendlyName)
		require.Equal(t, summary1.InferenceTasks[0], models[0].Task)
		require.Empty(t, models[1].Task)
		require.Equal(t, summary1.Publisher, models[0].Publisher)
		require.Equal(t, summary2.Publisher, models[1].Publisher)
		require.Equal(t, summary1.Summary, models[0].Summary)
		require.Equal(t, summary2.Summary, models[1].Summary)
		require.Equal(t, summary1.Version, models[0].Version)
		require.Equal(t, summary2.Version, models[1].Version)
		require.Equal(t, summary1.RegistryName, models[0].RegistryName)
		require.Equal(t, summary2.RegistryName, models[1].RegistryName)
	})

	t.Run("GetModelDetails happy path", func(t *testing.T) {
		registry := "foo"
		modelName := "bar"
		version := "baz"
		textLimits := &modelCatalogTextLimits{MaxOutputTokens: 8675309, InputContextWindow: 3}
		modelLimits := &modelCatalogLimits{
			SupportedInputModalities:  []string{"books", "VHS"},
			SupportedOutputModalities: []string{"watercolor paintings"},
			SupportedLanguages:        []string{"fr", "zh"},
			TextLimits:                textLimits,
		}
		playgroundLimits := &modelCatalogPlaygroundLimits{RateLimitTier: "big-ish"}
		catalogDetails := &modelCatalogDetailsResponse{
			Description:        "some model description",
			License:            "My Favorite License",
			LicenseDescription: "This is a test license",
			Notes:              "You aren't gonna believe these notes.",
			Keywords:           []string{"Tag1", "TAG2"},
			Evaluation:         "This model is the best.",
			ModelLimits:        modelLimits,
			PlaygroundLimits:   playgroundLimits,
		}
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			require.Equal(t, "/asset-gallery/v1.0/"+registry+"/models/"+modelName+"/version/"+version, r.URL.Path)

			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(catalogDetails)
			require.NoError(t, err)
		}))
		defer testServer.Close()
		cfg := &AzureClientConfig{AzureAiStudioURL: testServer.URL}
		httpClient := testServer.Client()
		client := NewAzureClient(httpClient, "fake-token-123abc", cfg)

		details, err := client.GetModelDetails(ctx, registry, modelName, version)

		require.NoError(t, err)
		require.NotNil(t, details)
		require.Equal(t, catalogDetails.Description, details.Description)
		require.Equal(t, catalogDetails.License, details.License)
		require.Equal(t, catalogDetails.LicenseDescription, details.LicenseDescription)
		require.Equal(t, catalogDetails.Notes, details.Notes)
		require.Equal(t, []string{"tag1", "tag2"}, details.Tags)
		require.Equal(t, catalogDetails.Evaluation, details.Evaluation)
		require.Equal(t, modelLimits.SupportedInputModalities, details.SupportedInputModalities)
		require.Equal(t, modelLimits.SupportedOutputModalities, details.SupportedOutputModalities)
		require.Equal(t, []string{"French", "Chinese"}, details.SupportedLanguages)
		require.Equal(t, textLimits.MaxOutputTokens, details.MaxOutputTokens)
		require.Equal(t, textLimits.InputContextWindow, details.MaxInputTokens)
		require.Equal(t, playgroundLimits.RateLimitTier, details.RateLimitTier)
	})
}
