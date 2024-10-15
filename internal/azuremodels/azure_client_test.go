package azuremodels

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/github/gh-models/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestAzureClient(t *testing.T) {
	ctx := context.Background()

	t.Run("GetChatCompletionStream", func(t *testing.T) {
		newTestServerForChatCompletion := func(handlerFn http.HandlerFunc) *httptest.Server {
			return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "application/json", r.Header.Get("Content-Type"))
				require.Equal(t, "/", r.URL.Path)
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "github-cli-models", r.Header.Get("x-ms-useragent"))
				require.Equal(t, "github-cli-models", r.Header.Get("x-ms-user-agent"))

				handlerFn(w, r)
			}))
		}

		t.Run("non-streaming happy path", func(t *testing.T) {
			message := &ChatChoiceMessage{
				Role:    util.Ptr("assistant"),
				Content: util.Ptr("This is my test story in response to your test prompt."),
			}
			choice := ChatChoice{Index: 1, FinishReason: "stop", Message: message}
			authToken := "fake-token-123abc"
			testServer := newTestServerForChatCompletion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "Bearer "+authToken, r.Header.Get("Authorization"))

				data := new(bytes.Buffer)
				err := json.NewEncoder(data).Encode(&ChatCompletion{Choices: []ChatChoice{choice}})
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("data: " + data.String() + "\n\ndata: [DONE]\n"))
			}))
			defer testServer.Close()
			cfg := &AzureClientConfig{InferenceURL: testServer.URL}
			httpClient := testServer.Client()
			client := NewAzureClient(httpClient, authToken, cfg)
			opts := ChatCompletionOptions{
				Model:  "some-test-model",
				Stream: false,
				Messages: []ChatMessage{
					{
						Role:    "user",
						Content: util.Ptr("Tell me a story, test model."),
					},
				},
			}

			chatCompletionStreamResp, err := client.GetChatCompletionStream(ctx, opts)

			require.NoError(t, err)
			require.NotNil(t, chatCompletionStreamResp)
			reader := chatCompletionStreamResp.Reader
			defer reader.Close()
			choicesReceived := []ChatChoice{}
			for {
				chatCompletionResp, err := reader.Read()
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)
				choicesReceived = append(choicesReceived, chatCompletionResp.Choices...)
			}
			require.Equal(t, 1, len(choicesReceived))
			require.Equal(t, choice.FinishReason, choicesReceived[0].FinishReason)
			require.Equal(t, choice.Index, choicesReceived[0].Index)
			require.Equal(t, message.Role, choicesReceived[0].Message.Role)
			require.Equal(t, message.Content, choicesReceived[0].Message.Content)
		})

		t.Run("streaming happy path", func(t *testing.T) {
			message1 := &ChatChoiceMessage{
				Role:    util.Ptr("assistant"),
				Content: util.Ptr("This is the first part of my test story in response to your test prompt."),
			}
			message2 := &ChatChoiceMessage{
				Role:    util.Ptr("assistant"),
				Content: util.Ptr("This is the second part of my test story in response to your test prompt."),
			}
			choice1 := ChatChoice{Index: 1, Message: message1}
			choice2 := ChatChoice{Index: 2, FinishReason: "stop", Message: message2}
			authToken := "fake-token-123abc"
			testServer := newTestServerForChatCompletion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "Bearer "+authToken, r.Header.Get("Authorization"))

				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Connection", "keep-alive")
				w.(http.Flusher).Flush()

				data1 := new(bytes.Buffer)
				err := json.NewEncoder(data1).Encode(&ChatCompletion{Choices: []ChatChoice{choice1}})
				require.NoError(t, err)
				w.Write([]byte("data: " + data1.String() + "\n\n"))
				w.(http.Flusher).Flush()
				time.Sleep(1 * time.Millisecond)

				data2 := new(bytes.Buffer)
				err = json.NewEncoder(data2).Encode(&ChatCompletion{Choices: []ChatChoice{choice2}})
				require.NoError(t, err)
				w.Write([]byte("data: " + data2.String() + "\n\n"))
				w.(http.Flusher).Flush()
				time.Sleep(1 * time.Millisecond)

				w.Write([]byte("data: [DONE]\n"))
			}))
			defer testServer.Close()
			cfg := &AzureClientConfig{InferenceURL: testServer.URL}
			httpClient := testServer.Client()
			client := NewAzureClient(httpClient, authToken, cfg)
			opts := ChatCompletionOptions{
				Model:  "some-test-model",
				Stream: true,
				Messages: []ChatMessage{
					{
						Role:    "user",
						Content: util.Ptr("Tell me a story, test model."),
					},
				},
			}

			chatCompletionStreamResp, err := client.GetChatCompletionStream(ctx, opts)

			require.NoError(t, err)
			require.NotNil(t, chatCompletionStreamResp)
			reader := chatCompletionStreamResp.Reader
			defer reader.Close()
			choicesReceived := []ChatChoice{}
			for {
				chatCompletionResp, err := reader.Read()
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)
				choicesReceived = append(choicesReceived, chatCompletionResp.Choices...)
			}
			require.Equal(t, 2, len(choicesReceived))
			require.Equal(t, choice1.FinishReason, choicesReceived[0].FinishReason)
			require.Equal(t, choice1.Index, choicesReceived[0].Index)
			require.Equal(t, message1.Role, choicesReceived[0].Message.Role)
			require.Equal(t, message1.Content, choicesReceived[0].Message.Content)
			require.Equal(t, choice2.FinishReason, choicesReceived[1].FinishReason)
			require.Equal(t, choice2.Index, choicesReceived[1].Index)
			require.Equal(t, message2.Role, choicesReceived[1].Message.Role)
			require.Equal(t, message2.Content, choicesReceived[1].Message.Content)
		})

		t.Run("handles non-OK status", func(t *testing.T) {
			errRespBody := `{"error": "o noes"}`
			testServer := newTestServerForChatCompletion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(errRespBody))
			}))
			defer testServer.Close()
			cfg := &AzureClientConfig{InferenceURL: testServer.URL}
			httpClient := testServer.Client()
			client := NewAzureClient(httpClient, "fake-token-123abc", cfg)
			opts := ChatCompletionOptions{
				Model:    "some-test-model",
				Messages: []ChatMessage{{Role: "user", Content: util.Ptr("Tell me a story, test model.")}},
			}

			chatCompletionResp, err := client.GetChatCompletionStream(ctx, opts)

			require.Error(t, err)
			require.Nil(t, chatCompletionResp)
			require.Equal(t, "unexpected response from the server: 500 Internal Server Error\n"+errRespBody+"\n", err.Error())
		})
	})

	t.Run("ListModels", func(t *testing.T) {
		newTestServerForListModels := func(handlerFn http.HandlerFunc) *httptest.Server {
			return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "application/json", r.Header.Get("Content-Type"))
				require.Equal(t, "/", r.URL.Path)
				require.Equal(t, http.MethodPost, r.Method)

				handlerFn(w, r)
			}))
		}

		t.Run("happy path", func(t *testing.T) {
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
			testServer := newTestServerForListModels(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		t.Run("handles non-OK status", func(t *testing.T) {
			errRespBody := `{"error": "o noes"}`
			testServer := newTestServerForListModels(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(errRespBody))
			}))
			defer testServer.Close()
			cfg := &AzureClientConfig{ModelsURL: testServer.URL}
			httpClient := testServer.Client()
			client := NewAzureClient(httpClient, "fake-token-123abc", cfg)

			models, err := client.ListModels(ctx)

			require.Error(t, err)
			require.Nil(t, models)
			require.Equal(t, "unauthorized\n"+errRespBody+"\n", err.Error())
		})
	})

	t.Run("GetModelDetails", func(t *testing.T) {
		newTestServerForModelDetails := func(handlerFn http.HandlerFunc) *httptest.Server {
			return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "application/json", r.Header.Get("Content-Type"))
				require.Equal(t, http.MethodGet, r.Method)

				handlerFn(w, r)
			}))
		}

		t.Run("happy path", func(t *testing.T) {
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
			testServer := newTestServerForModelDetails(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		t.Run("handles non-OK status", func(t *testing.T) {
			errRespBody := `{"error": "o noes"}`
			testServer := newTestServerForModelDetails(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(errRespBody))
			}))
			defer testServer.Close()
			cfg := &AzureClientConfig{AzureAiStudioURL: testServer.URL}
			httpClient := testServer.Client()
			client := NewAzureClient(httpClient, "fake-token-123abc", cfg)

			details, err := client.GetModelDetails(ctx, "someRegistry", "someModel", "someVersion")

			require.Error(t, err)
			require.Nil(t, details)
			require.Equal(t, "bad request\n"+errRespBody+"\n", err.Error())
		})
	})
}
