package generate

import (
	"fmt"

	"github.com/github/gh-models/internal/azuremodels"
)

// logLLMPayload logs the LLM request and response if verbose mode is enabled
func (h *generateCommandHandler) logLLMResponse(response string) {
	if h.options.Verbose != nil && *h.options.Verbose {
		h.cfg.WriteToOut(fmt.Sprintf("╭─assistant\n%s\n╰─🏁\n", response))
	}
}

func (h *generateCommandHandler) logLLMRequest(step string, options azuremodels.ChatCompletionOptions, messages []azuremodels.ChatMessage) {
	if h.options.Verbose != nil && *h.options.Verbose {
		h.cfg.WriteToOut(fmt.Sprintf("\n╭─💬 %s %s\n", step, options.Model))
		for _, msg := range messages {
			content := ""
			if msg.Content != nil {
				content = *msg.Content
			}
			h.cfg.WriteToOut(fmt.Sprintf("╭─%s\n%s\n", msg.Role, content))
		}
		h.cfg.WriteToOut("╰─\n")
	}
}
