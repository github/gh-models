package generate

import (
	"fmt"
	"strings"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/prompt"
)

// RenderMessagesToString converts a slice of Messages to a human-readable string representation
func RenderMessagesToString(messages []prompt.Message) string {
	if len(messages) == 0 {
		return ""
	}

	var builder strings.Builder

	for i, msg := range messages {
		// Add role header
		roleLower := strings.ToLower(msg.Role)
		builder.WriteString(fmt.Sprintf("%s:\n", roleLower))

		// Add content with proper indentation
		content := strings.TrimSpace(msg.Content)
		if content != "" {
			// Split content into lines and indent each line
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				builder.WriteString(fmt.Sprintf("%s\n", line))
			}
		}

		// Add separator between messages (except for the last one)
		if i < len(messages)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// logLLMPayload logs the LLM request and response if verbose mode is enabled
func (h *generateCommandHandler) LogLLMResponse(response string) {
	if h.options.Verbose != nil && *h.options.Verbose {
		h.cfg.WriteToOut(fmt.Sprintf("╭─assistant\n%s\n╰─🏁\n", response))
	}
}

func (h *generateCommandHandler) LogLLMRequest(step string, options azuremodels.ChatCompletionOptions) {
	if h.options.Verbose != nil && *h.options.Verbose {
		h.cfg.WriteToOut(fmt.Sprintf("\n╭─💬 %s %s\n", step, options.Model))
		for _, msg := range options.Messages {
			content := ""
			if msg.Content != nil {
				content = *msg.Content
			}
			h.cfg.WriteToOut(fmt.Sprintf("╭─%s\n%s\n", msg.Role, content))
		}
		h.cfg.WriteToOut("╰─\n")
	}
}
