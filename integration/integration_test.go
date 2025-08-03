//go:build integration

package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	binaryName = "gh-models"
	timeout    = 30 * time.Second
)

// getBinaryPath returns the path to the compiled gh-models binary
func getBinaryPath(t *testing.T) string {
	// Look for binary in project root
	wd, err := os.Getwd()
	require.NoError(t, err)

	// Go up one level from integration/ to project root
	projectRoot := filepath.Dir(wd)
	binaryPath := filepath.Join(projectRoot, binaryName)

	// Verify binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("Binary %s not found. Run 'make build' first.", binaryPath)
	}

	return binaryPath
}

// runCommand executes the gh-models binary with given args and returns stdout, stderr, and exit code
func runCommand(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	binaryPath := getBinaryPath(t)

	cmd := exec.Command(binaryPath, args...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()

	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("Failed to run command: %v", err)
		}
	} else {
		exitCode = 0
	}

	return stdout, stderr, exitCode
}

// createTempPromptFile creates a temporary prompt file for testing
func createTempPromptFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "test.prompt.yml")
	err := os.WriteFile(promptFile, []byte(content), 0644)
	require.NoError(t, err)
	return promptFile
}

// TestBasicCommands tests basic command functionality and exit codes
func TestBasicCommands(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectExitCode int
		expectStdout   []string // strings that should be present in stdout
		expectStderr   []string // strings that should be present in stderr
	}{
		{
			name:           "help command",
			args:           []string{"--help"},
			expectExitCode: 0,
			expectStdout:   []string{"GitHub Models CLI extension", "Available Commands:", "Usage:"},
		},
		{
			name:           "list command without auth",
			args:           []string{"list"},
			expectExitCode: 1,
			expectStderr:   []string{"not authenticated"},
		},
		{
			name:           "run command help",
			args:           []string{"run", "--help"},
			expectExitCode: 0,
			expectStdout:   []string{"Prompts the specified model", "Usage:", "Examples:"},
		},
		{
			name:           "generate command help",
			args:           []string{"generate", "--help"},
			expectExitCode: 0,
			expectStdout:   []string{"Augment prompt.yml file", "Usage:", "Examples:"},
		},
		{
			name:           "eval command help",
			args:           []string{"eval", "--help"},
			expectExitCode: 0,
			expectStdout:   []string{"Runs evaluation tests", "Usage:", "Examples:"},
		},
		{
			name:           "view command help",
			args:           []string{"view", "--help"},
			expectExitCode: 0,
			expectStdout:   []string{"Returns details about the specified model", "Usage:", "Examples:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runCommand(t, tt.args...)

			require.Equal(t, tt.expectExitCode, exitCode,
				"Expected exit code %d, got %d. Stdout: %s, Stderr: %s",
				tt.expectExitCode, exitCode, stdout, stderr)

			for _, expected := range tt.expectStdout {
				require.Contains(t, stdout, expected,
					"Expected stdout to contain '%s'. Full stdout: %s", expected, stdout)
			}

			for _, expected := range tt.expectStderr {
				require.Contains(t, stderr, expected,
					"Expected stderr to contain '%s'. Full stderr: %s", expected, stderr)
			}
		})
	}
}

// TestRunCommandWithPromptFile tests the run command with a prompt file
func TestRunCommandWithPromptFile(t *testing.T) {
	// Create a simple test prompt file
	promptContent := `name: Integration Test Prompt
description: A simple test prompt for integration testing
model: openai/gpt-4o-mini
modelParameters:
  temperature: 0.1
  maxTokens: 10
messages:
  - role: system
    content: You are a helpful assistant. Be very brief.
  - role: user
    content: Say "test successful" in exactly 2 words.
`

	promptFile := createTempPromptFile(t, promptContent)

	t.Run("run with prompt file without auth", func(t *testing.T) {
		_, stderr, exitCode := runCommand(t, "run", "--file", promptFile)

		// Should fail due to authentication
		require.Equal(t, 1, exitCode)
		require.Contains(t, stderr, "not authenticated")
	})

	t.Run("run with invalid model", func(t *testing.T) {
		_, stderr, exitCode := runCommand(t, "run", "invalid/model", "test prompt")

		// Should fail due to authentication first
		require.Equal(t, 1, exitCode)
		require.Contains(t, stderr, "not authenticated")
	})
}

// TestGenerateCommand tests the generate command for creating test data
func TestGenerateCommand(t *testing.T) {
	// Create a prompt file suitable for test generation
	promptContent := `name: Test Generation Example
description: A prompt for testing the generate command
model: openai/gpt-4o-mini
messages:
  - role: system
    content: You are a helpful assistant.
  - role: user
    content: "Tell me about {{topic}}"
testData:
  - topic: "cats"
  - topic: "dogs"
`

	promptFile := createTempPromptFile(t, promptContent)

	t.Run("generate without auth", func(t *testing.T) {
		_, stderr, exitCode := runCommand(t, "generate", promptFile)

		// Should fail due to authentication
		require.Equal(t, 1, exitCode)
		require.Contains(t, stderr, "not authenticated")
	})

	t.Run("generate with invalid file", func(t *testing.T) {
		_, stderr, exitCode := runCommand(t, "generate", "/nonexistent/file.yml")

		// Should fail due to file not found
		require.Equal(t, 1, exitCode)
		// Error could be about file not found or authentication, both are acceptable
		require.True(t, strings.Contains(stderr, "not authenticated") ||
			strings.Contains(stderr, "no such file") ||
			strings.Contains(stderr, "cannot find"))
	})
}

// TestEvalCommand tests the eval command for prompt evaluation
func TestEvalCommand(t *testing.T) {
	// Create a prompt file with evaluators
	promptContent := `name: Evaluation Test
description: A prompt with evaluators for testing eval command
model: openai/gpt-4o-mini
messages:
  - role: system
    content: You are a helpful assistant.
  - role: user
    content: "Say hello"
testData:
  - input: "hello"
evaluators:
  - name: contains-greeting
    string:
      contains: "hello"
`

	promptFile := createTempPromptFile(t, promptContent)

	t.Run("eval without auth", func(t *testing.T) {
		_, stderr, exitCode := runCommand(t, "eval", promptFile)

		// Should fail due to authentication
		require.Equal(t, 1, exitCode)
		require.Contains(t, stderr, "not authenticated")
	})
}

// TestInvalidCommands tests error handling for invalid commands and arguments
func TestInvalidCommands(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectExitCode int
	}{
		{
			name:           "invalid command",
			args:           []string{"invalid-command"},
			expectExitCode: 1,
		},
		{
			name:           "run without arguments",
			args:           []string{"run"},
			expectExitCode: 1,
		},
		{
			name:           "run with too few arguments",
			args:           []string{"run", "model-name"},
			expectExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, exitCode := runCommand(t, tt.args...)
			require.Equal(t, tt.expectExitCode, exitCode)
		})
	}
}

// TestPromptFileValidation tests validation of prompt files
func TestPromptFileValidation(t *testing.T) {
	t.Run("invalid yaml", func(t *testing.T) {
		invalidYaml := `name: Test
invalid: yaml: content:
messages
  - invalid
`
		promptFile := createTempPromptFile(t, invalidYaml)

		_, stderr, exitCode := runCommand(t, "run", "--file", promptFile)

		// Should fail due to invalid YAML (or auth error first)
		require.Equal(t, 1, exitCode)
		// Could fail on YAML parsing or authentication
		require.True(t, strings.Contains(stderr, "not authenticated") ||
			strings.Contains(stderr, "yaml") ||
			strings.Contains(stderr, "parse"))
	})

	t.Run("missing required fields", func(t *testing.T) {
		incompleteYaml := `name: Test
# missing model and messages
description: Incomplete prompt file
`
		promptFile := createTempPromptFile(t, incompleteYaml)

		_, stderr, exitCode := runCommand(t, "run", "--file", promptFile)

		// Should fail due to missing fields (or auth error first)
		require.Equal(t, 1, exitCode)
		require.True(t, strings.Contains(stderr, "not authenticated") ||
			strings.Contains(stderr, "model") ||
			strings.Contains(stderr, "messages"))
	})
}
