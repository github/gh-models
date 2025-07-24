package generate

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/github/gh-models/pkg/prompt"
)

// computePromptHash computes a SHA256 hash of the prompt's messages, model, and model parameters
func computePromptHash(p *prompt.File) (string, error) {
	// Create a hashable structure containing only the fields we want to hash
	hashData := struct {
		Messages        []prompt.Message       `json:"messages"`
		Model           string                 `json:"model"`
		ModelParameters prompt.ModelParameters `json:"modelParameters"`
	}{
		Messages:        p.Messages,
		Model:           p.Model,
		ModelParameters: p.ModelParameters,
	}

	// Convert to JSON for consistent hashing
	jsonData, err := json.Marshal(hashData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal prompt data for hashing: %w", err)
	}

	// Compute SHA256 hash
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash), nil
}

// createContext creates a new PromptPexContext from a prompt file
func (h *generateCommandHandler) CreateContextFromPrompt(promptFile string) (*PromptPexContext, error) {
	runID := fmt.Sprintf("run_%d", time.Now().Unix())

	prompt, err := prompt.LoadFromFile(promptFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt file: %w", err)
	}

	// Compute the hash of the prompt (messages, model, model parameters)
	promptHash, err := computePromptHash(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to compute prompt hash: %w", err)
	}

	context := &PromptPexContext{
		RunID:         runID,
		Prompt:        prompt,
		PromptHash:    promptHash,
		Intent:        "",
		Rules:         "",
		InverseRules:  "",
		InputSpec:     "",
		Tests:         "",
		TestData:      "",
		TestOutputs:   "",
		TestEvals:     "",
		RuleEvals:     "",
		RuleCoverages: "",
		Options:       h.options,
	}

	return context, nil
}
