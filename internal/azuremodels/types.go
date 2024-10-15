package azuremodels

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/github/gh-models/internal/sse"
)

// ChatMessageRole represents the role of a chat message.
type ChatMessageRole string

const (
	// ChatMessageRoleAssistant represents a message from the model.
	ChatMessageRoleAssistant ChatMessageRole = "assistant"
	// ChatMessageRoleSystem represents a system message.
	ChatMessageRoleSystem ChatMessageRole = "system"
	// ChatMessageRoleUser represents a message from the user.
	ChatMessageRoleUser ChatMessageRole = "user"
)

// ChatMessage represents a message from a chat thread with a model.
type ChatMessage struct {
	Content *string         `json:"content,omitempty"`
	Role    ChatMessageRole `json:"role"`
}

// ChatCompletionOptions represents available options for a chat completion request.
type ChatCompletionOptions struct {
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Messages    []ChatMessage `json:"messages"`
	Model       string        `json:"model"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
}

// ChatChoiceMessage is a message from a choice in a chat conversation.
type ChatChoiceMessage struct {
	Content *string `json:"content,omitempty"`
	Role    *string `json:"role,omitempty"`
}

type chatChoiceDelta struct {
	Content *string `json:"content,omitempty"`
	Role    *string `json:"role,omitempty"`
}

// ChatChoice represents a choice in a chat completion.
type ChatChoice struct {
	Delta        *chatChoiceDelta   `json:"delta,omitempty"`
	FinishReason string             `json:"finish_reason"`
	Index        int32              `json:"index"`
	Message      *ChatChoiceMessage `json:"message,omitempty"`
}

// ChatCompletion represents a chat completion.
type ChatCompletion struct {
	Choices []ChatChoice `json:"choices"`
}

// ChatCompletionResponse represents a response to a chat completion request.
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

// ModelSummary includes basic information about a model.
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

// IsChatModel returns true if the model is for chat completions.
func (m *ModelSummary) IsChatModel() bool {
	return m.Task == "chat-completion"
}

// HasName checks if the model has the given name.
func (m *ModelSummary) HasName(name string) bool {
	return strings.EqualFold(m.FriendlyName, name) || strings.EqualFold(m.Name, name)
}

type modelCatalogTextLimits struct {
	MaxOutputTokens    int `json:"maxOutputTokens"`
	InputContextWindow int `json:"inputContextWindow"`
}

type modelCatalogLimits struct {
	SupportedLanguages        []string                `json:"supportedLanguages"`
	TextLimits                *modelCatalogTextLimits `json:"textLimits"`
	SupportedInputModalities  []string                `json:"supportedInputModalities"`
	SupportedOutputModalities []string                `json:"supportedOutputModalities"`
}

type modelCatalogPlaygroundLimits struct {
	RateLimitTier string `json:"rateLimitTier"`
}

type modelCatalogDetailsResponse struct {
	AssetID            string                        `json:"assetId"`
	Name               string                        `json:"name"`
	DisplayName        string                        `json:"displayName"`
	Publisher          string                        `json:"publisher"`
	Version            string                        `json:"version"`
	RegistryName       string                        `json:"registryName"`
	Evaluation         string                        `json:"evaluation"`
	Summary            string                        `json:"summary"`
	Description        string                        `json:"description"`
	License            string                        `json:"license"`
	LicenseDescription string                        `json:"licenseDescription"`
	Notes              string                        `json:"notes"`
	Keywords           []string                      `json:"keywords"`
	InferenceTasks     []string                      `json:"inferenceTasks"`
	FineTuningTasks    []string                      `json:"fineTuningTasks"`
	Labels             []string                      `json:"labels"`
	TradeRestricted    bool                          `json:"tradeRestricted"`
	CreatedTime        string                        `json:"createdTime"`
	PlaygroundLimits   *modelCatalogPlaygroundLimits `json:"playgroundLimits"`
	ModelLimits        *modelCatalogLimits           `json:"modelLimits"`
}

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
