package generate

import (
	"fmt"
)

// generateSummary generates a summary report
func (h *generateCommandHandler) generateSummary(context *PromptPexContext) error {

	h.WriteBox(fmt.Sprintf(`🚀 Done! Saved %d tests in %s`, len(context.Tests), h.promptFile), fmt.Sprintf(`
To run the tests and evaluations, use the following command:

    gh models eval %s
	
`, h.promptFile))
	return nil
}
