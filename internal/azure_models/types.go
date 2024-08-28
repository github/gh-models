package azure_models

import "github.com/github/gh-models/internal/sse"

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
	Reader *sse.EventReader[ChatCompletion]
}

type ModelSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	FriendlyName string `json:"friendly_name"`
	Task         string `json:"task"`
	Publisher    string `json:"publisher"`
	Summary      string `json:"summary"`
}

func Ptr[T any](value T) *T {
	return &value
}
