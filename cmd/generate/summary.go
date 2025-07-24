package generate

import (
	"fmt"
)

// generateSummary generates a summary report
func (h *generateCommandHandler) GenerateSummary(context *PromptPexContext) error {
	h.cfg.WriteToOut(fmt.Sprintf("\n---\nGenerated %d tests for prompt '%s'\n", len(context.Tests), context.Prompt.Name))

	return nil
}
