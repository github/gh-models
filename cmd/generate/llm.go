package generate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/github/gh-models/internal/azuremodels"
)

// callModelWithRetry makes an API call with automatic retry on rate limiting
func (h *generateCommandHandler) callModelWithRetry(step string, req azuremodels.ChatCompletionOptions) (string, error) {
	const maxRetries = 3
	ctx := h.ctx

	h.logLLMRequest(step, req)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := h.client.GetChatCompletionStream(ctx, req, h.org)
		if err != nil {
			var rateLimitErr *azuremodels.RateLimitError
			if errors.As(err, &rateLimitErr) {
				if attempt < maxRetries {
					h.cfg.WriteToOut(fmt.Sprintf("    Rate limited, waiting %v before retry (attempt %d/%d)...\n",
						rateLimitErr.RetryAfter, attempt+1, maxRetries+1))

					// Wait for the specified duration
					select {
					case <-ctx.Done():
						return "", ctx.Err()
					case <-time.After(rateLimitErr.RetryAfter):
						continue
					}
				}
				return "", fmt.Errorf("rate limit exceeded after %d attempts: %w", attempt+1, err)
			}
			// For non-rate-limit errors, return immediately
			return "", err
		}

		var content strings.Builder
		for {
			completion, err := resp.Reader.Read()
			if err != nil {
				if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "EOF") {
					break
				}
				return "", err
			}
			if len(completion.Choices) == 0 {
				return "", fmt.Errorf("no completion choices returned from model")
			}
			for _, choice := range completion.Choices {
				if choice.Delta != nil && choice.Delta.Content != nil {
					content.WriteString(*choice.Delta.Content)
				}
				if choice.Message != nil && choice.Message.Content != nil {
					content.WriteString(*choice.Message.Content)
				}
			}
		}

		res := strings.TrimSpace(content.String())
		h.logLLMResponse(res)
		return res, nil
	}

	// This should never be reached, but just in case
	return "", errors.New("unexpected error calling model")
}

// logLLMPayload logs the LLM request and response if verbose mode is enabled
func (h *generateCommandHandler) logLLMResponse(response string) {
	if h.options.Verbose != nil && *h.options.Verbose {
		h.cfg.WriteToOut(fmt.Sprintf("╭─assistant\n%s\n╰─🏁\n", response))
	}
}

func (h *generateCommandHandler) logLLMRequest(step string, options azuremodels.ChatCompletionOptions) {
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
