package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/github/gh-models/pkg/prompt"
	"github.com/github/gh-models/pkg/util"
)

// createContext creates a new PromptPexContext from a prompt file
func (h *generateCommandHandler) CreateContextFromPrompt() (*PromptPexContext, error) {
	prompt, err := prompt.LoadFromFile(h.promptFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt file: %w", err)
	}

	// Compute the hash of the prompt (messages, model, model parameters)
	promptHash, err := ComputePromptHash(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to compute prompt hash: %w", err)
	}

	runID := fmt.Sprintf("run_%d", time.Now().Unix())
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

	if h.sessionFile != nil {
		// Try to load existing context from session file
		existingContext, err := loadContextFromFile(*h.sessionFile)
		if err != nil {
			h.cfg.WriteToOut(fmt.Sprintf("Creating session file at %s\n", *h.sessionFile))
			// If file doesn't exist, that's okay - we'll start fresh
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load existing context from %s: %w", *h.sessionFile, err)
			}
		} else {
			h.cfg.WriteToOut(fmt.Sprintf("Reloading session file at %s\n", *h.sessionFile))
			// Check if prompt hashes match
			if existingContext.PromptHash != nil && context.PromptHash != nil &&
				*existingContext.PromptHash != *context.PromptHash {
				return nil, fmt.Errorf("prompt changed unable to reuse session file")
			}

			// Merge existing context data
			if existingContext != nil {
				context = mergeContexts(existingContext, context)
			}
		}
	}

	return context, nil
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

// saveContext saves the context to the session file
func (h *generateCommandHandler) SaveContext(context *PromptPexContext) error {
	if h.sessionFile == nil {
		return nil // No session file specified, skip saving
	}
	data, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context to JSON: %w", err)
	}

	if err := os.WriteFile(*h.sessionFile, data, 0644); err != nil {
		h.cfg.WriteToOut(fmt.Sprintf("Failed to write context to session file %s: %v", *h.sessionFile, err))
	}

	return nil
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
		if existing.InputSpec != nil {
			merged.InputSpec = existing.InputSpec
			if existing.Rules != nil {
				merged.Rules = existing.Rules
				if existing.InverseRules != nil {
					merged.InverseRules = existing.InverseRules
					if existing.Tests != nil {
						merged.Tests = existing.Tests
					}
				}
			}
		}
	}

	return merged
}
