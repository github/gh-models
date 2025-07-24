package generate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/github/gh-models/pkg/command"
	"github.com/stretchr/testify/require"
)

func TestSessionFile(t *testing.T) {
	t.Run("create new session", func(t *testing.T) {
		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test.prompt.yml")
		sessionFilePath := filepath.Join(tmpDir, "session.json")

		// Create a test prompt file
		yamlBody := `name: Test Prompt
description: A test prompt file
model: openai/gpt-4o
messages:
  - role: user
    content: "Hello world"`

		err := os.WriteFile(promptFilePath, []byte(yamlBody), 0644)
		require.NoError(t, err)

		// Create handler with proper config
		cfg := &command.Config{
			Out: os.Stdout,
		}
		options := GetDefaultOptions()
		handler := &generateCommandHandler{
			cfg:         cfg,
			options:     options,
			sessionFile: sessionFilePath,
		}

		// Load or create session (should create new)
		context, err := handler.LoadOrCreateSession(promptFilePath)
		require.NoError(t, err)
		require.NotNil(t, context)
		require.Equal(t, "Test Prompt", context.Prompt.Name)

		// Check session file was created
		require.FileExists(t, sessionFilePath)

		// Verify session file contents
		data, err := os.ReadFile(sessionFilePath)
		require.NoError(t, err)

		var session SessionFile
		err = json.Unmarshal(data, &session)
		require.NoError(t, err)
		require.Equal(t, SessionFileVersion, session.Version)
		require.Equal(t, promptFilePath, session.PromptFile)
		require.NotEmpty(t, session.PromptHash)
		require.NotNil(t, session.Context)
	})

	t.Run("load existing session", func(t *testing.T) {
		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test.prompt.yml")
		sessionFilePath := filepath.Join(tmpDir, "session.json")

		// Create a test prompt file
		yamlBody := `name: Test Prompt
description: A test prompt file
model: openai/gpt-4o
messages:
  - role: user
    content: "Hello world"`

		err := os.WriteFile(promptFilePath, []byte(yamlBody), 0644)
		require.NoError(t, err)

		// Create handler with proper config
		cfg := &command.Config{
			Out: os.Stdout,
		}
		options := GetDefaultOptions()
		handler := &generateCommandHandler{
			cfg:         cfg,
			options:     options,
			sessionFile: sessionFilePath,
		}

		// Create initial session
		context1, err := handler.LoadOrCreateSession(promptFilePath)
		require.NoError(t, err)

		// Modify context to simulate progress
		context1.Intent = "Test intent"
		promptHash, _ := calculateFileHash(promptFilePath)
		err = handler.SaveSession(context1, promptFilePath, promptHash)
		require.NoError(t, err)

		// Load session again (should load existing)
		context2, err := handler.LoadOrCreateSession(promptFilePath)
		require.NoError(t, err)
		require.Equal(t, "Test intent", context2.Intent)
	})

	t.Run("prompt file mismatch", func(t *testing.T) {
		tmpDir := t.TempDir()
		promptFilePath1 := filepath.Join(tmpDir, "test1.prompt.yml")
		promptFilePath2 := filepath.Join(tmpDir, "test2.prompt.yml")
		sessionFilePath := filepath.Join(tmpDir, "session.json")

		// Create test prompt files
		yamlBody1 := `name: Test Prompt 1
model: openai/gpt-4o
messages:
  - role: user
    content: "Hello world 1"`

		yamlBody2 := `name: Test Prompt 2
model: openai/gpt-4o
messages:
  - role: user
    content: "Hello world 2"`

		err := os.WriteFile(promptFilePath1, []byte(yamlBody1), 0644)
		require.NoError(t, err)
		err = os.WriteFile(promptFilePath2, []byte(yamlBody2), 0644)
		require.NoError(t, err)

		// Create handler with proper config
		cfg := &command.Config{
			Out: os.Stdout,
		}
		options := GetDefaultOptions()
		handler := &generateCommandHandler{
			cfg:         cfg,
			options:     options,
			sessionFile: sessionFilePath,
		}

		// Create session with first prompt file
		_, err = handler.LoadOrCreateSession(promptFilePath1)
		require.NoError(t, err)

		// Try to load session with different prompt file (should fail)
		_, err = handler.LoadOrCreateSession(promptFilePath2)
		require.Error(t, err)
		require.Contains(t, err.Error(), "prompt file mismatch")
	})

	t.Run("prompt file modified", func(t *testing.T) {
		tmpDir := t.TempDir()
		promptFilePath := filepath.Join(tmpDir, "test.prompt.yml")
		sessionFilePath := filepath.Join(tmpDir, "session.json")

		// Create initial prompt file
		yamlBody1 := `name: Test Prompt
model: openai/gpt-4o
messages:
  - role: user
    content: "Hello world"`

		err := os.WriteFile(promptFilePath, []byte(yamlBody1), 0644)
		require.NoError(t, err)

		// Create handler with proper config
		cfg := &command.Config{
			Out: os.Stdout,
		}
		options := GetDefaultOptions()
		handler := &generateCommandHandler{
			cfg:         cfg,
			options:     options,
			sessionFile: sessionFilePath,
		}

		// Create session
		_, err = handler.LoadOrCreateSession(promptFilePath)
		require.NoError(t, err)

		// Modify prompt file
		yamlBody2 := `name: Test Prompt Modified
model: openai/gpt-4o
messages:
  - role: user
    content: "Hello world modified"`

		err = os.WriteFile(promptFilePath, []byte(yamlBody2), 0644)
		require.NoError(t, err)

		// Try to load session with modified prompt file (should fail)
		_, err = handler.LoadOrCreateSession(promptFilePath)
		require.Error(t, err)
		require.Contains(t, err.Error(), "hash mismatch")
	})
}

func TestIsStepCompleted(t *testing.T) {
	context := &PromptPexContext{}

	t.Run("intent step", func(t *testing.T) {
		require.False(t, IsStepCompleted(context, "intent"))
		context.Intent = "Test intent"
		require.True(t, IsStepCompleted(context, "intent"))
	})

	t.Run("rules step", func(t *testing.T) {
		require.False(t, IsStepCompleted(context, "rules"))
		context.Rules = "Test rules"
		require.True(t, IsStepCompleted(context, "rules"))
	})

	t.Run("tests step", func(t *testing.T) {
		require.False(t, IsStepCompleted(context, "tests"))
		context.PromptPexTests = []PromptPexTest{{TestInput: "test"}}
		require.True(t, IsStepCompleted(context, "tests"))
	})

	t.Run("groundtruth step", func(t *testing.T) {
		require.False(t, IsStepCompleted(context, "groundtruth"))
		groundtruth := "Test groundtruth"
		context.PromptPexTests = []PromptPexTest{{
			TestInput:   "test",
			Groundtruth: &groundtruth,
		}}
		require.True(t, IsStepCompleted(context, "groundtruth"))
	})

	t.Run("unknown step", func(t *testing.T) {
		require.False(t, IsStepCompleted(context, "unknown"))
	})
}

func TestCalculateFileHash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	content := "Hello world"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Calculate hash
	hash1, err := calculateFileHash(testFile)
	require.NoError(t, err)
	require.NotEmpty(t, hash1)

	// Calculate hash again (should be same)
	hash2, err := calculateFileHash(testFile)
	require.NoError(t, err)
	require.Equal(t, hash1, hash2)

	// Modify file
	err = os.WriteFile(testFile, []byte("Hello world modified"), 0644)
	require.NoError(t, err)

	// Calculate hash again (should be different)
	hash3, err := calculateFileHash(testFile)
	require.NoError(t, err)
	require.NotEqual(t, hash1, hash3)
}

func TestSessionFileSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	sessionFilePath := filepath.Join(tmpDir, "session.json")

	// Create test context
	context := &PromptPexContext{
		RunID:  "test-run-123",
		Intent: "Test intent",
	}

	// Create handler with proper config
	cfg := &command.Config{
		Out: os.Stdout,
	}
	options := GetDefaultOptions()
	handler := &generateCommandHandler{
		cfg:         cfg,
		options:     options,
		sessionFile: sessionFilePath,
	}

	// Save session
	err := handler.SaveSession(context, "test.yml", "testhash")
	require.NoError(t, err)

	// Verify file exists
	require.FileExists(t, sessionFilePath)

	// Load and verify
	data, err := os.ReadFile(sessionFilePath)
	require.NoError(t, err)

	var session SessionFile
	err = json.Unmarshal(data, &session)
	require.NoError(t, err)

	require.Equal(t, SessionFileVersion, session.Version)
	require.Equal(t, "test.yml", session.PromptFile)
	require.Equal(t, "testhash", session.PromptHash)
	require.Equal(t, "test-run-123", session.Context.RunID)
	require.Equal(t, "Test intent", session.Context.Intent)
	require.True(t, time.Since(session.Created) < time.Minute)
	require.True(t, time.Since(session.LastModified) < time.Minute)
}