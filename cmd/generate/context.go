package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/github/gh-models/pkg/prompt"
	"github.com/github/gh-models/pkg/util"
)

// createContext creates a new PromptPexContext from a prompt file
func (h *generateCommandHandler) CreateContextFromPrompt(promptFile string, contextFile string) (*PromptPexContext, error) {
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
		RunID: util.Ptr(runID),
		// The prompt content and metadata
		Prompt: prompt,
		// Hash of the prompt messages, model, and parameters
		PromptHash: util.Ptr(promptHash),
		// The options used to generate the prompt
		Options: h.options,
	}

	// Determine session file path
	sessionFile := contextFile
	if sessionFile == "" {
		// Generate default session file name by replacing 'prompt.yml' with '.generate.json'
		sessionFile = generateDefaultSessionFileName(promptFile)
	}

	// Try to load existing context from session file
	if sessionFile != "" {
		existingContext, err := loadContextFromFile(sessionFile)
		if err != nil {
			// If file doesn't exist, that's okay - we'll start fresh
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load existing context from %s: %w", sessionFile, err)
			}
		} else {
			// Check if prompt hashes match
			if existingContext.PromptHash != nil && context.PromptHash != nil &&
				*existingContext.PromptHash != *context.PromptHash {
				return nil, fmt.Errorf("prompt hash mismatch: existing context has different prompt than current file")
			}

			// Merge existing context data
			context = mergeContexts(existingContext, context)
		}
	}

	return context, nil
}

// generateDefaultSessionFileName generates the default session file name
func generateDefaultSessionFileName(promptFile string) string {
	// Replace .prompt.yml with .generate.json
	if strings.HasSuffix(promptFile, ".prompt.yml") {
		return strings.TrimSuffix(promptFile, ".prompt.yml") + ".generate.json"
	}
	// If it doesn't end with .prompt.yml, just append .generate.json
	return promptFile + ".generate.json"
}

// loadContextFromFile loads a PromptPexContext from a JSON file
func loadContextFromFile(filePath string) (*PromptPexContext, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var context PromptPexContext
	if err := json.Unmarshal(data, &context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context JSON: %w", err)
	}

	return &context, nil
}

// mergeContexts merges an existing context with a new context
// The new context takes precedence for prompt, options, and hash
// Other data from existing context is preserved
func mergeContexts(existing *PromptPexContext, new *PromptPexContext) *PromptPexContext {
	merged := &PromptPexContext{
		// Use new context's core data
		RunID:      new.RunID,
		Prompt:     new.Prompt,
		PromptHash: new.PromptHash,
		Options:    new.Options,
	}

	// Preserve existing pipeline data if it exists
	if existing.Intent != nil {
		merged.Intent = existing.Intent
	}
	if existing.Rules != nil {
		merged.Rules = existing.Rules
	}
	if existing.InverseRules != nil {
		merged.InverseRules = existing.InverseRules
	}
	if existing.InputSpec != nil {
		merged.InputSpec = existing.InputSpec
	}
	if existing.Tests != nil {
		merged.Tests = existing.Tests
	}

	return merged
}
