package generate

import (
	"fmt"
	"time"

	"github.com/github/gh-models/pkg/prompt"
	"github.com/github/gh-models/pkg/util"
)

// createContext creates a new PromptPexContext from a prompt file
func (h *generateCommandHandler) CreateContext(inputFile string) (*PromptPexContext, error) {
	runID := fmt.Sprintf("run_%d", time.Now().Unix())

	prompt, err := prompt.LoadFromFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt file: %w", err)
	}

	context := &PromptPexContext{
		RunID:             runID,
		WriteResults:      util.Ptr(true),
		Prompt:            prompt,
		Intent:            "",
		Rules:             "",
		InverseRules:      "",
		InputSpec:         "",
		BaselineTests:     "",
		Tests:             "",
		TestData:          "",
		RateTests:         "",
		TestOutputs:       "",
		TestEvals:         "",
		RuleEvals:         "",
		RuleCoverages:     "",
		BaselineTestEvals: "",
		Options:           h.options,
	}

	return context, nil
}
