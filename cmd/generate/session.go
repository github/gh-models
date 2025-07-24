package generate

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// SessionFile represents the session file structure with metadata
type SessionFile struct {
	Version      string            `json:"version"`
	Created      time.Time         `json:"created"`
	LastModified time.Time         `json:"lastModified"`
	PromptFile   string            `json:"promptFile"`
	Context      *PromptPexContext `json:"context"`
}

const SessionFileVersion = "1.0"

// LoadOrCreateSession loads an existing session file or creates a new one
func (h *generateCommandHandler) LoadOrCreateSession(promptFile string) (*PromptPexContext, error) {
	// If no session file is provided, create a context without session persistence
	if h.sessionFile == "" {
		context, err := h.CreateContextFromPrompt(promptFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create context: %w", err)
		}
		return context, nil
	}

	// Calculate prompt file hash for consistency checking
	promptHash, err := calculateFileHash(promptFile)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate prompt file hash: %w", err)
	}

	// Check if session file exists
	if _, err := os.Stat(h.sessionFile); err == nil {
		// Session file exists, load it
		context, err := h.loadExistingSession(promptFile, promptHash)
		if err != nil {
			return nil, fmt.Errorf("failed to load existing session: %w", err)
		}
		return context, nil
	} else if os.IsNotExist(err) {
		// Session file doesn't exist, create new session
		context, err := h.createNewSession(promptFile, promptHash)
		if err != nil {
			return nil, fmt.Errorf("failed to create new session: %w", err)
		}
		return context, nil
	} else {
		return nil, fmt.Errorf("failed to check session file: %w", err)
	}
}

// loadExistingSession loads and validates an existing session file
func (h *generateCommandHandler) loadExistingSession(promptFile, promptHash string) (*PromptPexContext, error) {
	h.cfg.WriteToOut(fmt.Sprintf("Loading existing session from: %s\n", h.sessionFile))

	// Read session file
	data, err := os.ReadFile(h.sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	// Parse session file
	var session SessionFile
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session file: %w", err)
	}

	// Validate prompt file consistency
	if session.PromptFile != promptFile {
		return nil, fmt.Errorf("prompt file mismatch: session expects '%s' but got '%s'", session.PromptFile, promptFile)
	}

	if session.Context.PromptHash != promptHash {
		return nil, fmt.Errorf("prompt file has been modified since session was created (hash mismatch)")
	}

	h.cfg.WriteToOut("Session loaded successfully. Checking for completed steps...\n")
	h.logSessionProgress(session.Context)

	return session.Context, nil
}

// createNewSession creates a new session with the given prompt file
func (h *generateCommandHandler) createNewSession(promptFile, promptHash string) (*PromptPexContext, error) {
	h.cfg.WriteToOut(fmt.Sprintf("Creating new session: %s\n", h.sessionFile))

	// Create context from prompt file
	context, err := h.CreateContextFromPrompt(promptFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create context: %w", err)
	}

	// Set the prompt hash in the context
	context.PromptHash = promptHash

	// Save initial session
	if err := h.SaveSession(context, promptFile); err != nil {
		return nil, fmt.Errorf("failed to save initial session: %w", err)
	}

	return context, nil
}

// SaveSession saves the current context to the session file
func (h *generateCommandHandler) SaveSession(context *PromptPexContext, promptFile string) error {
	// If no session file is provided, skip saving
	if h.sessionFile == "" {
		return nil
	}

	// Create session structure
	session := SessionFile{
		Version:      SessionFileVersion,
		Created:      time.Now(),
		LastModified: time.Now(),
		PromptFile:   promptFile,
		Context:      context,
	}

	// If session file already exists, preserve the created time
	if existingData, err := os.ReadFile(h.sessionFile); err == nil {
		var existingSession SessionFile
		if json.Unmarshal(existingData, &existingSession) == nil {
			session.Created = existingSession.Created
		}
	}

	// Ensure session file directory exists
	if err := os.MkdirAll(filepath.Dir(h.sessionFile), 0755); err != nil {
		return fmt.Errorf("failed to create session file directory: %w", err)
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Write to file atomically using a temporary file
	tempFile := h.sessionFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary session file: %w", err)
	}

	if err := os.Rename(tempFile, h.sessionFile); err != nil {
		os.Remove(tempFile) // Clean up temp file on failure
		return fmt.Errorf("failed to replace session file: %w", err)
	}

	return nil
}

// calculateFileHash calculates SHA256 hash of a file
func calculateFileHash(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// logSessionProgress logs what steps have been completed in the session
func (h *generateCommandHandler) logSessionProgress(context *PromptPexContext) {
	completed := []string{}
	
	if context.Intent != "" {
		completed = append(completed, "Intent")
	}
	if context.InputSpec != "" {
		completed = append(completed, "Input Specification")
	}
	if context.Rules != "" {
		completed = append(completed, "Output Rules")
	}
	if context.InverseRules != "" {
		completed = append(completed, "Inverse Rules")
	}
	if len(context.PromptPexTests) > 0 {
		completed = append(completed, fmt.Sprintf("Tests (%d)", len(context.PromptPexTests)))
		
		// Check if groundtruth exists
		hasGroundtruth := false
		for _, test := range context.PromptPexTests {
			if test.Groundtruth != nil && *test.Groundtruth != "" {
				hasGroundtruth = true
				break
			}
		}
		if hasGroundtruth {
			completed = append(completed, "Groundtruth")
		}
	}
	if context.TestOutputs != "" {
		completed = append(completed, "Test Outputs")
	}
	if context.TestEvals != "" {
		completed = append(completed, "Test Evaluations")
	}

	if len(completed) > 0 {
		h.cfg.WriteToOut(fmt.Sprintf("Completed steps: %v\n", completed))
	} else {
		h.cfg.WriteToOut("No steps completed yet.\n")
	}
}

// IsStepCompleted checks if a specific step has been completed based on the context
func IsStepCompleted(context *PromptPexContext, step string) bool {
	switch step {
	case "intent":
		return context.Intent != ""
	case "inputSpec":
		return context.InputSpec != ""
	case "rules":
		return context.Rules != ""
	case "inverseRules":
		return context.InverseRules != ""
	case "tests":
		return len(context.PromptPexTests) > 0
	case "testExpansions":
		// Test expansions are considered complete if any tests have generation > 0
		for _, test := range context.PromptPexTests {
			if test.Generation != nil && *test.Generation > 0 {
				return true
			}
		}
		return false
	case "groundtruth":
		// Check if any tests have groundtruth data
		for _, test := range context.PromptPexTests {
			if test.Groundtruth != nil && *test.Groundtruth != "" {
				return true
			}
		}
		return false
	case "testOutputs":
		return context.TestOutputs != ""
	case "testEvals":
		return context.TestEvals != ""
	default:
		return false
	}
}