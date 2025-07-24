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

	// Compute the hash of the prompt (messages, model, model parameters)
	promptHash, err := ComputePromptHash(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to compute prompt hash: %w", err)
	}

	context := &PromptPexContext{
		// Unique identifier for the run
		RunID: runID,
		// The prompt content and metadata
		Prompt: prompt,
		// Hash of the prompt messages, model, and parameters
		PromptHash: promptHash,
		// Infered intent of the prompt
		Intent:       "",
		Rules:        []string{},
		InverseRules: "",
		InputSpec:    "",
		Tests:        "",
		TestData:     "",
		TestOutputs:  "",
		TestEvals:    "",
		Options:      h.options,
	}

	return context, nil
}
