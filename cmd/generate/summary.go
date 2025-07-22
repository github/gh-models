package generate

import (
	"encoding/json"
	"fmt"
)

// generateSummary generates a summary report
func (h *generateCommandHandler) GenerateSummary(context *PromptPexContext) (string, error) {
	h.cfg.WriteToOut(fmt.Sprintf("Summary: Generated %d tests for prompt '%s'", len(context.PromptPexTests), context.Prompt.Name))

	summary := map[string]interface{}{
		"name":  context.Prompt.Name,
		"tests": len(context.PromptPexTests),
		"runId": context.RunID,
	}

	data, _ := json.MarshalIndent(summary, "", "  ")

	return string(data), nil
}
