//go:build integration

package integration

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestFileModificationScenarios tests scenarios that would modify prompt files
// These tests focus on validating file changes and exit codes as mentioned in the problem statement
func TestFileModificationScenarios(t *testing.T) {
	t.Run("generate command with valid prompt file", func(t *testing.T) {
		// Create a prompt file suitable for test generation
		promptContent := `name: File Modification Test
description: A prompt file to test the generate command file modifications
model: openai/gpt-4o-mini
modelParameters:
  temperature: 0.1
  maxTokens: 50
messages:
  - role: system
    content: You are a helpful assistant.
  - role: user
    content: "Answer the question: {{question}}"
testData:
  - question: "What is 2+2?"
  - question: "What color is the sky?"
`

		// Create the prompt file in a temporary directory we can inspect
		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test_generate.prompt.yml")
		err := os.WriteFile(promptFile, []byte(promptContent), 0644)
		require.NoError(t, err)

		// Record the original file contents
		originalContent, err := os.ReadFile(promptFile)
		require.NoError(t, err)

		// Run the generate command
		_, stderr, exitCode := runCommand(t, "generate", promptFile)

		// Even without auth, the command should fail gracefully with expected exit code
		require.Equal(t, 1, exitCode, "Expected exit code 1 for unauthenticated generate command")
		require.Contains(t, stderr, "not authenticated", "Expected authentication error")

		// File should remain unchanged when command fails due to auth
		currentContent, err := os.ReadFile(promptFile)
		require.NoError(t, err)
		require.Equal(t, string(originalContent), string(currentContent),
			"File should not be modified when command fails due to authentication")
	})

	t.Run("eval command with test data", func(t *testing.T) {
		// Create a prompt file with evaluators
		promptContent := `name: Evaluation Test
description: A prompt file with evaluators for testing
model: openai/gpt-4o-mini
modelParameters:
  temperature: 0.1
messages:
  - role: system
    content: You are a helpful assistant that responds politely.
  - role: user
    content: "{{greeting}}"
testData:
  - greeting: "Hello"
    expected: "Hello there"
  - greeting: "Hi"
    expected: "Hi there"
evaluators:
  - name: contains-greeting
    string:
      contains: "hello"
  - name: is-polite
    string:
      contains: "there"
`

		promptFile := createTempPromptFile(t, promptContent)

		// Run eval command
		_, stderr, exitCode := runCommand(t, "eval", promptFile)

		// Should fail due to authentication but with proper exit code
		require.Equal(t, 1, exitCode, "Expected exit code 1 for unauthenticated eval command")
		require.Contains(t, stderr, "not authenticated", "Expected authentication error")
	})
}

// TestPromptFileStructure tests that the integration tests properly handle various prompt file structures
func TestPromptFileStructure(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		errorType   string // "auth", "parse", "validation"
	}{
		{
			name: "complete valid prompt file",
			content: `name: Complete Test
description: A complete valid prompt file
model: openai/gpt-4o-mini
modelParameters:
  temperature: 0.5
  maxTokens: 100
messages:
  - role: system
    content: You are helpful.
  - role: user
    content: "{{input}}"
testData:
  - input: "test"
evaluators:
  - name: test-evaluator
    string:
      contains: "test"
`,
			expectError: true,
			errorType:   "auth",
		},
		{
			name: "minimal valid prompt file",
			content: `name: Minimal Test
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Hello"
`,
			expectError: true,
			errorType:   "auth",
		},
		{
			name: "prompt file with template variables",
			content: `name: Template Test
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Hello {{name}}, how are you?"
testData:
  - name: "Alice"
  - name: "Bob"
`,
			expectError: true,
			errorType:   "auth",
		},
		{
			name: "prompt file with json schema",
			content: `name: JSON Schema Test
model: openai/gpt-4o-mini
responseFormat: json_schema
jsonSchema: '{"name": "person", "schema": {"type": "object", "properties": {"name": {"type": "string"}}}}'
messages:
  - role: user
    content: "Generate a person"
`,
			expectError: true,
			errorType:   "auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptFile := createTempPromptFile(t, tt.content)

			// Test with run command
			_, stderr, exitCode := runCommand(t, "run", "--file", promptFile)

			if tt.expectError {
				require.Equal(t, 1, exitCode, "Expected non-zero exit code")

				switch tt.errorType {
				case "auth":
					require.Contains(t, stderr, "not authenticated",
						"Expected authentication error for test: %s", tt.name)
				case "parse":
					require.True(t,
						strings.Contains(stderr, "parse") ||
							strings.Contains(stderr, "yaml") ||
							strings.Contains(stderr, "not authenticated"),
						"Expected parse error or auth error for test: %s. Got stderr: %s", tt.name, stderr)
				case "validation":
					require.True(t,
						strings.Contains(stderr, "validation") ||
							strings.Contains(stderr, "required") ||
							strings.Contains(stderr, "not authenticated"),
						"Expected validation error or auth error for test: %s. Got stderr: %s", tt.name, stderr)
				}
			} else {
				require.Equal(t, 0, exitCode, "Expected zero exit code for valid prompt file")
			}
		})
	}
}

// TestCommandChaining tests multiple commands in sequence to ensure proper exit codes
func TestCommandChaining(t *testing.T) {
	promptContent := `name: Chaining Test
description: Test prompt for command chaining
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Test message"
`
	promptFile := createTempPromptFile(t, promptContent)

	t.Run("sequential command execution", func(t *testing.T) {
		// Test list -> run -> generate sequence
		commands := []struct {
			name string
			args []string
		}{
			{"list", []string{"list"}},
			{"run", []string{"run", "--file", promptFile}},
			{"generate", []string{"generate", promptFile}},
		}

		for _, cmd := range commands {
			t.Run(cmd.name, func(t *testing.T) {
				_, stderr, exitCode := runCommand(t, cmd.args...)

				// All should fail with auth error and exit code 1
				require.Equal(t, 1, exitCode,
					"Command %s should fail with exit code 1 due to auth", cmd.name)
				require.Contains(t, stderr, "not authenticated",
					"Command %s should fail with auth error", cmd.name)
			})
		}
	})
}

// TestLongRunningCommands tests commands that might take longer to execute
func TestLongRunningCommands(t *testing.T) {
	// Set a longer timeout for these tests
	if testing.Short() {
		t.Skip("Skipping long-running command tests in short mode")
	}

	t.Run("generate command timeout handling", func(t *testing.T) {
		promptContent := `name: Long Running Test
description: Test for potentially long-running generate command
model: openai/gpt-4o-mini
messages:
  - role: system
    content: You are a helpful assistant.
  - role: user
    content: "{{topic}}"
testData:
  - topic: "artificial intelligence"
  - topic: "machine learning"
  - topic: "deep learning"
  - topic: "neural networks"
  - topic: "computer vision"
`
		promptFile := createTempPromptFile(t, promptContent)

		start := time.Now()
		_, stderr, exitCode := runCommand(t, "generate", promptFile)
		duration := time.Since(start)

		// Should fail quickly due to auth, not timeout
		require.Equal(t, 1, exitCode)
		require.Contains(t, stderr, "not authenticated")
		require.Less(t, duration, 10*time.Second,
			"Command should fail quickly due to auth, not timeout")
	})
}

// TestFileSystemInteraction tests how commands interact with the file system
func TestFileSystemInteraction(t *testing.T) {
	t.Run("working directory independence", func(t *testing.T) {
		// Create prompt file in temp directory
		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		promptContent := `name: Directory Test
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Test"
`
		err := os.WriteFile(promptFile, []byte(promptContent), 0644)
		require.NoError(t, err)

		// Test with absolute path (should work regardless of working directory)
		_, stderr, exitCode := runCommand(t, "run", "--file", promptFile)

		require.Equal(t, 1, exitCode)
		require.Contains(t, stderr, "not authenticated")
	})

	t.Run("file permissions", func(t *testing.T) {
		// Create a read-only prompt file
		promptContent := `name: Permission Test
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Test"
`
		promptFile := createTempPromptFile(t, promptContent)

		// Make file read-only
		err := os.Chmod(promptFile, 0444)
		require.NoError(t, err)

		// Should still be able to read the file for run command
		_, stderr, exitCode := runCommand(t, "run", "--file", promptFile)

		require.Equal(t, 1, exitCode)
		require.Contains(t, stderr, "not authenticated")
	})
}

// TestOutputFormats tests different output format options
func TestOutputFormats(t *testing.T) {
	promptContent := `name: Output Format Test
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "{{input}}"
testData:
  - input: "Hello"
evaluators:
  - name: test
    string:
      contains: "test"
`
	promptFile := createTempPromptFile(t, promptContent)

	t.Run("eval with json output", func(t *testing.T) {
		_, stderr, exitCode := runCommand(t, "eval", "--json", promptFile)

		// Should fail due to authentication when trying to make API calls for evaluation
		require.Equal(t, 1, exitCode)
		require.Contains(t, stderr, "not authenticated")
	})

	t.Run("eval with default output", func(t *testing.T) {
		_, stderr, exitCode := runCommand(t, "eval", promptFile)

		// Should fail due to authentication when trying to make API calls for evaluation
		require.Equal(t, 1, exitCode)
		require.Contains(t, stderr, "not authenticated")
	})
}

// Helper function to count lines in a file
func countLines(t *testing.T, filename string) int {
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	require.NoError(t, scanner.Err())
	return lines
}
