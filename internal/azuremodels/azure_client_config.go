package azuremodels

const (
	defaultInferenceUrl     = "https://models.inference.ai.azure.com/chat/completions"
	defaultAzureAiStudioUrl = "https://api.catalog.azureml.ms"
	defaultModelsUrl        = defaultAzureAiStudioUrl + "/asset-gallery/v1.0/models"
)

// AzureClientConfig represents configurable settings for the Azure client.
type AzureClientConfig struct {
	InferenceUrl     string
	AzureAiStudioUrl string
	ModelsUrl        string
}

// NewDefaultAzureClientConfig returns a new AzureClientConfig with default values for API URLs.
func NewDefaultAzureClientConfig() *AzureClientConfig {
	return &AzureClientConfig{
		InferenceUrl:     defaultInferenceUrl,
		AzureAiStudioUrl: defaultAzureAiStudioUrl,
		ModelsUrl:        defaultModelsUrl,
	}
}
