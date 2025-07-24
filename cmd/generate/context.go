package generate

import (
	"fmt"
	"time"

	"github.com/github/gh-models/pkg/prompt"
)

// createContext creates a new PromptPexContext from a prompt file
func (h *generateCommandHandler) CreateContextFromPrompt(promptFile string) (*PromptPexContext, error) {
	runID := fmt.Sprintf("run_%d", time.Now().Unix())

	prompt, err := prompt.LoadFromFile(promptFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt file: %w", err)
	}

	context := &PromptPexContext{
		RunID:             runID,
		Prompt:            prompt,
		Intent:            "",
		Rules:             "",
		InverseRules:      "",
		InputSpec:         "",
		BaselineTests:     "",
		Tests:             "",
		TestData:          "",
		TestOutputs:       "",
		TestEvals:         "",
		RuleEvals:         "",
		RuleCoverages:     "",
		BaselineTestEvals: "",
		Options:           h.options,
	}

	return context, nil
}
