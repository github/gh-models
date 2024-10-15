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
