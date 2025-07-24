package generate

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/github/gh-models/pkg/prompt"
)

// ComputePromptHash computes a SHA256 hash of the prompt's messages, model, and model parameters
func ComputePromptHash(p *prompt.File) (string, error) {
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
