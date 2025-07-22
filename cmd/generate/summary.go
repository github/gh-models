package generate

import (
	"fmt"
)

// generateSummary generates a summary report
func (h *generateCommandHandler) GenerateSummary(context *PromptPexContext) error {
	h.cfg.WriteToOut(fmt.Sprintf("Summary: Generated %d tests for prompt '%s'", len(context.PromptPexTests), context.Prompt.Name))

	return nil
}
