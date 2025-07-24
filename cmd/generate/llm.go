package generate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/github/gh-models/internal/azuremodels"
)

// callModelWithRetry makes an API call with automatic retry on rate limiting
func (h *generateCommandHandler) callModelWithRetry(step string, req azuremodels.ChatCompletionOptions) (string, error) {
	const maxRetries = 3
	ctx := h.ctx

	h.LogLLMRequest(step, req)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(h.cfg.ErrOut))
		sp.Start()
		//nolint:gocritic,revive // TODO
		defer sp.Stop()

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
		reader := resp.Reader
		//nolint:gocritic,revive // TODO
		defer reader.Close()

		var content strings.Builder
		for {
			completion, err := reader.Read()
			if err != nil {
				if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "EOF") {
					break
				}
				return "", err
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
		h.LogLLMResponse(res)
		return res, nil
	}

	// This should never be reached, but just in case
	return "", errors.New("unexpected error calling model")
}
