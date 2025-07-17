package azuremodels

const (
	defaultInferenceRoot    = "https://models.github.ai"
	defaultInferencePath    = "inference/chat/completions"
	defaultAzureAiStudioURL = "https://api.catalog.azureml.ms"
	defaultModelsURL        = "https://models.github.ai/catalog/models"
)

// AzureClientConfig represents configurable settings for the Azure client.
type AzureClientConfig struct {
	InferenceRoot    string
	InferencePath    string
	AzureAiStudioURL string
	ModelsURL        string
}

// NewDefaultAzureClientConfig returns a new AzureClientConfig with default values for API URLs.
func NewDefaultAzureClientConfig() *AzureClientConfig {
	return &AzureClientConfig{
		InferenceRoot:    defaultInferenceRoot,
		InferencePath:    defaultInferencePath,
		AzureAiStudioURL: defaultAzureAiStudioURL,
		ModelsURL:        defaultModelsURL,
	}
}
