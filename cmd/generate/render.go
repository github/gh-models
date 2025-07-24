package generate

import (
	"fmt"
	"strings"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/prompt"
)

var BOX_START = "╭─"
var BOX_END = "╰─"
var BOX_LINE = "─"
var PREVIEW_TEST_COUNT = 16

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

func (h *generateCommandHandler) WriteStartBox(title string) {
	h.cfg.WriteToOut(fmt.Sprintf("%s %s\n", BOX_START, title))
}

func (h *generateCommandHandler) WriteEndBox(suffix string) {
	if suffix == "" {
		suffix = BOX_LINE
	}
	h.cfg.WriteToOut(fmt.Sprintf("%s%s\n", BOX_END, suffix))
}

func (h *generateCommandHandler) WriteBox(title, content string) {
	h.WriteStartBox(title)
	if content != "" {
		h.cfg.WriteToOut(content)
		if !strings.HasSuffix(content, "\n") {
			h.cfg.WriteToOut("\n")
		}
	}
	h.WriteEndBox("")
}

func (h *generateCommandHandler) WriteToLine(item string) {
	if len(item) > h.cfg.TerminalWidth-2 {
		item = item[:h.cfg.TerminalWidth-2] + "…"
	}
	if strings.HasSuffix(item, "\n") {
		h.cfg.WriteToOut(item)
	} else {
		h.cfg.WriteToOut(fmt.Sprintf("%s\n", item))
	}
}

func (h *generateCommandHandler) WriteEndListBox(items []string, maxItems int) {
	renderedItems := items
	if len(renderedItems) > maxItems {
		renderedItems = renderedItems[:maxItems]
	}
	for _, item := range renderedItems {
		h.WriteToLine(item)
	}
	if len(items) != len(renderedItems) {
		h.cfg.WriteToOut("…\n")
	}
	h.WriteEndBox(fmt.Sprintf("%d items", len(items)))
}

// logLLMPayload logs the LLM request and response if verbose mode is enabled
func (h *generateCommandHandler) LogLLMResponse(response string) {
	if h.options.Verbose != nil && *h.options.Verbose {
		h.WriteStartBox("🏁")
		h.cfg.WriteToOut(response)
		if !strings.HasSuffix(response, "\n") {
			h.cfg.WriteToOut("\n")
		}
		h.WriteEndBox("")
	}
}

func (h *generateCommandHandler) LogLLMRequest(step string, options azuremodels.ChatCompletionOptions) {
	if h.options.Verbose != nil && *h.options.Verbose {
		h.WriteStartBox(fmt.Sprintf("💬 %s %s", step, options.Model))
		for _, msg := range options.Messages {
			content := ""
			if msg.Content != nil {
				content = *msg.Content
			}
			h.cfg.WriteToOut(fmt.Sprintf("%s%s\n%s\n", BOX_START, msg.Role, content))
		}
		h.WriteEndBox("")
	}
}
