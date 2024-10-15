package azuremodels

const (
	defaultInferenceURL     = "https://models.inference.ai.azure.com/chat/completions"
	defaultAzureAiStudioURL = "https://api.catalog.azureml.ms"
	defaultModelsURL        = defaultAzureAiStudioURL + "/asset-gallery/v1.0/models"
)

// AzureClientConfig represents configurable settings for the Azure client.
type AzureClientConfig struct {
	InferenceURL     string
	AzureAiStudioURL string
	ModelsURL        string
}

// NewDefaultAzureClientConfig returns a new AzureClientConfig with default values for API URLs.
func NewDefaultAzureClientConfig() *AzureClientConfig {
	return &AzureClientConfig{
		InferenceURL:     defaultInferenceURL,
		AzureAiStudioURL: defaultAzureAiStudioURL,
		ModelsURL:        defaultModelsURL,
	}
}
