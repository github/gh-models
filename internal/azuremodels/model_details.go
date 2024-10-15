package azuremodels

import "fmt"

// ModelDetails includes detailed information about a model.
type ModelDetails struct {
	Description               string   `json:"description"`
	Evaluation                string   `json:"evaluation"`
	License                   string   `json:"license"`
	LicenseDescription        string   `json:"license_description"`
	Notes                     string   `json:"notes"`
	Tags                      []string `json:"tags"`
	SupportedInputModalities  []string `json:"supported_input_modalities"`
	SupportedOutputModalities []string `json:"supported_output_modalities"`
	SupportedLanguages        []string `json:"supported_languages"`
	MaxOutputTokens           int      `json:"max_output_tokens"`
	MaxInputTokens            int      `json:"max_input_tokens"`
	RateLimitTier             string   `json:"rateLimitTier"`
}

// ContextLimits returns a summary of the context limits for the model.
func (m *ModelDetails) ContextLimits() string {
	return fmt.Sprintf("up to %d input tokens and %d output tokens", m.MaxInputTokens, m.MaxOutputTokens)
}
