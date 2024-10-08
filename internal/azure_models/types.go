package azure_models

import (
	"encoding/json"

	"github.com/github/gh-models/internal/sse"
)

type ChatMessageRole string

const (
	ChatMessageRoleAssistant ChatMessageRole = "assistant"
	ChatMessageRoleSystem    ChatMessageRole = "system"
	ChatMessageRoleUser      ChatMessageRole = "user"
)

type ChatMessage struct {
	Content *string         `json:"content,omitempty"`
	Role    ChatMessageRole `json:"role"`
}

type ChatCompletionOptions struct {
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Messages    []ChatMessage `json:"messages"`
	Model       string        `json:"model"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
}

type ChatChoiceMessage struct {
	Content *string `json:"content,omitempty"`
	Role    *string `json:"role,omitempty"`
}

type ChatChoiceDelta struct {
	Content *string `json:"content,omitempty"`
	Role    *string `json:"role,omitempty"`
}

type ChatChoice struct {
	Delta        *ChatChoiceDelta   `json:"delta,omitempty"`
	FinishReason string             `json:"finish_reason"`
	Index        int32              `json:"index"`
	Message      *ChatChoiceMessage `json:"message,omitempty"`
}

type ChatCompletion struct {
	Choices []ChatChoice `json:"choices"`
}

type ChatCompletionResponse struct {
	Reader sse.Reader[ChatCompletion]
}

type modelCatalogSearchResponse struct {
	Summaries []modelCatalogSearchSummary `json:"summaries"`
}

type modelCatalogSearchSummary struct {
	AssetID        string      `json:"assetId"`
	DisplayName    string      `json:"displayName"`
	InferenceTasks []string    `json:"inferenceTasks"`
	Name           string      `json:"name"`
	Popularity     json.Number `json:"popularity"`
	Publisher      string      `json:"publisher"`
	RegistryName   string      `json:"registryName"`
	Version        string      `json:"version"`
	Summary        string      `json:"summary"`
}

type ModelSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	FriendlyName string `json:"friendly_name"`
	Task         string `json:"task"`
	Publisher    string `json:"publisher"`
	Summary      string `json:"summary"`
	Version      string `json:"version"`
	RegistryName string `json:"registry_name"`
}

type modelCatalogDetailsResponse struct {
	AssetID            string   `json:"assetId"`
	Name               string   `json:"name"`
	DisplayName        string   `json:"displayName"`
	Publisher          string   `json:"publisher"`
	Version            string   `json:"version"`
	RegistryName       string   `json:"registryName"`
	Evaluation         string   `json:"evaluation"`
	Summary            string   `json:"summary"`
	Description        string   `json:"description"`
	License            string   `json:"license"`
	LicenseDescription string   `json:"licenseDescription"`
	Notes              string   `json:"notes"`
	Keywords           []string `json:"keywords"`
	InferenceTasks     []string `json:"inferenceTasks"`
	FineTuningTasks    []string `json:"fineTuningTasks"`
	Labels             []string `json:"labels"`
	TradeRestricted    bool     `json:"tradeRestricted"`
	CreatedTime        string   `json:"createdTime"`
	ModelLimits        struct {
		SupportedLanguages []string `json:"supportedLanguages"`
		TextLimits         struct {
			MaxOutputTokens    int `json:"maxOutputTokens"`
			InputContextWindow int `json:"inputContextWindow"`
		} `json:"textLimits"`
		SupportedInputModalities  []string `json:"supportedInputModalities"`
		SupportedOutputModalities []string `json:"supportedOutputModalities"`
	} `json:"modelLimits"`
}

type ModelDetails struct {
	Description        string `json:"description"`
	License            string `json:"license"`
	LicenseDescription string `json:"license_description"`
}

func Ptr[T any](value T) *T {
	return &value
}
