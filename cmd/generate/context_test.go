package generate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/github/gh-models/pkg/prompt"
	"github.com/github/gh-models/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestGenerateDefaultSessionFileName(t *testing.T) {
	tests := []struct {
		name       string
		promptFile string
		expected   string
	}{
		{
			name:       "prompt.yml file",
			promptFile: "test.prompt.yml",
			expected:   "test.generate.json",
		},
		{
			name:       "prompt.yml with path",
			promptFile: "/path/to/test.prompt.yml",
			expected:   "/path/to/test.generate.json",
		},
		{
			name:       "non-prompt.yml file",
			promptFile: "test.yml",
			expected:   "test.yml.generate.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateDefaultSessionFileName(tt.promptFile)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadContextFromFile(t *testing.T) {
	// Create a temporary context file
	tmpDir := t.TempDir()
	contextFile := filepath.Join(tmpDir, "test.generate.json")

	// Create test context
	testContext := &PromptPexContext{
		RunID:      util.Ptr("test_run_123"),
		PromptHash: util.Ptr("testhash123"),
		Intent:     util.Ptr("Test intent"),
		Rules:      []string{"rule1", "rule2"},
		Tests: []PromptPexTest{
			{
				TestID:    util.Ptr(1),
				TestInput: "test input",
			},
		},
	}

	// Write context to file
	data, err := json.Marshal(testContext)
	require.NoError(t, err)
	err = os.WriteFile(contextFile, data, 0644)
	require.NoError(t, err)

	// Load context from file
	loaded, err := loadContextFromFile(contextFile)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	require.Equal(t, *testContext.RunID, *loaded.RunID)
	require.Equal(t, *testContext.PromptHash, *loaded.PromptHash)
	require.Equal(t, *testContext.Intent, *loaded.Intent)
	require.Equal(t, testContext.Rules, loaded.Rules)
	require.Len(t, loaded.Tests, 1)
	require.Equal(t, *testContext.Tests[0].TestID, *loaded.Tests[0].TestID)
}

func TestLoadContextFromFileNotExists(t *testing.T) {
	_, err := loadContextFromFile("/nonexistent/file.json")
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestMergeContexts(t *testing.T) {
	existing := &PromptPexContext{
		RunID:        util.Ptr("old_run"),
		PromptHash:   util.Ptr("oldhash"),
		Intent:       util.Ptr("Existing intent"),
		Rules:        []string{"existing_rule1", "existing_rule2"},
		InverseRules: []string{"inverse_rule1"},
		InputSpec:    util.Ptr("Existing input spec"),
		Tests: []PromptPexTest{
			{
				TestID:    util.Ptr(1),
				TestInput: "existing test",
			},
		},
	}

	new := &PromptPexContext{
		RunID:      util.Ptr("new_run"),
		PromptHash: util.Ptr("newhash"),
		Prompt: &prompt.File{
			Name: "New prompt",
		},
		Options: &PromptPexOptions{
			Temperature: util.Ptr(0.7),
		},
	}

	merged := mergeContexts(existing, new)

	// New context values should take precedence
	require.Equal(t, *new.RunID, *merged.RunID)
	require.Equal(t, *new.PromptHash, *merged.PromptHash)
	require.Equal(t, new.Prompt, merged.Prompt)
	require.Equal(t, new.Options, merged.Options)

	// Existing context values should be preserved
	require.Equal(t, *existing.Intent, *merged.Intent)
	require.Equal(t, existing.Rules, merged.Rules)
	require.Equal(t, existing.InverseRules, merged.InverseRules)
	require.Equal(t, *existing.InputSpec, *merged.InputSpec)
	require.Equal(t, existing.Tests, merged.Tests)
}

func TestCreateContextFromPromptWithSessionFile(t *testing.T) {
	// Create temporary files
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "test.prompt.yml")
	sessionFile := filepath.Join(tmpDir, "test.generate.json")

	// Create a minimal prompt file
	promptContent := `name: "Test Prompt"
model: "openai/gpt-4o-mini"
messages:
  - role: user
    content: "Hello"
`
	err := os.WriteFile(promptFile, []byte(promptContent), 0644)
	require.NoError(t, err)

	// Create handler
	handler := &generateCommandHandler{
		options: GetDefaultOptions(),
	}

	// Test 1: No existing session file
	context, err := handler.CreateContextFromPrompt(promptFile, sessionFile)
	require.NoError(t, err)
	require.NotNil(t, context)
	require.NotNil(t, context.RunID)
	require.NotNil(t, context.Prompt)
	require.NotNil(t, context.PromptHash)

	// Save the context to session file for next test
	data, err := json.Marshal(context)
	require.NoError(t, err)
	err = os.WriteFile(sessionFile, data, 0644)
	require.NoError(t, err)

	// Add some additional data to simulate existing pipeline results
	context.Intent = util.Ptr("Test intent from pipeline")
	context.Rules = []string{"rule1", "rule2"}
	data, err = json.Marshal(context)
	require.NoError(t, err)
	err = os.WriteFile(sessionFile, data, 0644)
	require.NoError(t, err)

	// Test 2: Load existing session file with same prompt hash
	context2, err := handler.CreateContextFromPrompt(promptFile, sessionFile)
	require.NoError(t, err)
	require.NotNil(t, context2)
	require.Equal(t, "Test intent from pipeline", *context2.Intent)
	require.Equal(t, []string{"rule1", "rule2"}, context2.Rules)
}

func TestCreateContextFromPromptWithDefaultSessionFile(t *testing.T) {
	// Create temporary files
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "test.prompt.yml")

	// Create a minimal prompt file
	promptContent := `name: "Test Prompt"
model: "openai/gpt-4o-mini"
messages:
  - role: user
    content: "Hello"
`
	err := os.WriteFile(promptFile, []byte(promptContent), 0644)
	require.NoError(t, err)

	// Create handler
	handler := &generateCommandHandler{
		options: GetDefaultOptions(),
	}

	// Test with empty session file (should use default)
	context, err := handler.CreateContextFromPrompt(promptFile, "")
	require.NoError(t, err)
	require.NotNil(t, context)
	require.NotNil(t, context.RunID)
	require.NotNil(t, context.Prompt)
	require.NotNil(t, context.PromptHash)
}

func TestCreateContextFromPromptHashMismatch(t *testing.T) {
	// Create temporary files
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "test.prompt.yml")
	sessionFile := filepath.Join(tmpDir, "test.generate.json")

	// Create a minimal prompt file
	promptContent := `name: "Test Prompt"
model: "openai/gpt-4o-mini"
messages:
  - role: user
    content: "Hello"
`
	err := os.WriteFile(promptFile, []byte(promptContent), 0644)
	require.NoError(t, err)

	// Create handler
	handler := &generateCommandHandler{
		options: GetDefaultOptions(),
	}

	// Create context with different hash
	existingContext := &PromptPexContext{
		RunID:      util.Ptr("existing_run"),
		PromptHash: util.Ptr("different_hash"),
		Intent:     util.Ptr("Existing intent"),
	}

	// Write existing context to session file
	data, err := json.Marshal(existingContext)
	require.NoError(t, err)
	err = os.WriteFile(sessionFile, data, 0644)
	require.NoError(t, err)

	// Try to create context - should fail due to hash mismatch
	_, err = handler.CreateContextFromPrompt(promptFile, sessionFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "prompt hash mismatch")
}
