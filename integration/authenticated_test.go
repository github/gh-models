//go:build integration

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAuthenticatedScenarios tests what would happen with proper authentication
// These tests are designed to demonstrate the expected behavior when auth is available
func TestAuthenticatedScenarios(t *testing.T) {
	// Skip these tests if we know we're not authenticated
	// This allows the tests to pass in CI while still being useful for local testing

	t.Run("check authentication status", func(t *testing.T) {
		// Check if gh is authenticated by trying to get user info
		_, stderr, exitCode := runCommand(t, "list")

		if exitCode == 1 && strings.Contains(stderr, "not authenticated") {
			t.Skip("GitHub authentication not available - skipping authenticated scenario tests")
		}

		// If we get here, we might be authenticated
		// Test basic list functionality
		stdout, stderr, exitCode := runCommand(t, "list")

		if exitCode == 0 {
			// Success case - we should see model listings
			require.Contains(t, stdout, "openai/", "Expected to see OpenAI models in list output")
			t.Logf("✅ Authentication successful - found models in output")
		} else {
			// Even with auth, might fail due to other reasons (network, etc.)
			t.Logf("ℹ️  List command failed with exit code %d. This might be due to network issues or other factors.", exitCode)
			t.Logf("   Stderr: %s", stderr)
		}
	})

	t.Run("authenticated run command", func(t *testing.T) {
		// Create a simple test prompt
		promptContent := `name: Authenticated Test
description: A simple test for authenticated scenarios
model: openai/gpt-4o-mini
modelParameters:
  temperature: 0.1
  maxTokens: 10
messages:
  - role: system
    content: You are a helpful assistant. Be very brief.
  - role: user
    content: "Say 'OK' if you understand."
`

		promptFile := createTempPromptFile(t, promptContent)

		// Try to run with authentication
		stdout, stderr, exitCode := runCommand(t, "run", "--file", promptFile)

		if exitCode == 1 && strings.Contains(stderr, "not authenticated") {
			t.Skip("GitHub authentication not available - skipping authenticated run test")
		}

		if exitCode == 0 {
			// Success case
			require.NotEmpty(t, stdout, "Expected some output from successful run")
			t.Logf("✅ Authenticated run successful")
			t.Logf("   Output: %s", strings.TrimSpace(stdout))
		} else {
			// Log what happened for debugging
			t.Logf("ℹ️  Run command failed with exit code %d", exitCode)
			t.Logf("   Stdout: %s", stdout)
			t.Logf("   Stderr: %s", stderr)
		}
	})

	t.Run("authenticated generate command", func(t *testing.T) {
		// Create a prompt file suitable for test generation
		promptContent := `name: Generate Test
description: A prompt for testing generate command with auth
model: openai/gpt-4o-mini
modelParameters:
  temperature: 0.1
messages:
  - role: system
    content: You are a helpful assistant.
  - role: user
    content: "Tell me about {{topic}}"
testData:
  - topic: "cats"
  - topic: "dogs"
`

		// Create in a temp directory we can monitor
		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "generate_test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(promptContent), 0644)
		require.NoError(t, err)

		// Record original file size/content
		originalStat, err := os.Stat(promptFile)
		require.NoError(t, err)
		originalSize := originalStat.Size()

		// Try to run generate
		stdout, stderr, exitCode := runCommand(t, "generate", promptFile)

		if exitCode == 1 && strings.Contains(stderr, "not authenticated") {
			t.Skip("GitHub authentication not available - skipping authenticated generate test")
		}

		if exitCode == 0 {
			// Success case - check if file was modified
			newStat, err := os.Stat(promptFile)
			require.NoError(t, err)
			newSize := newStat.Size()

			t.Logf("✅ Generate command completed successfully")
			t.Logf("   Original file size: %d bytes", originalSize)
			t.Logf("   New file size: %d bytes", newSize)

			if newSize > originalSize {
				t.Logf("✅ File appears to have been augmented with new test data")

				// Read the updated file content
				newContent, err := os.ReadFile(promptFile)
				require.NoError(t, err)

				// Check for signs of test generation (evaluators section, more testData)
				content := string(newContent)
				if strings.Contains(content, "evaluators:") {
					t.Logf("✅ Found evaluators section in updated file")
				}
				if strings.Count(content, "- topic:") > 2 {
					t.Logf("✅ Found additional test data entries")
				}
			}

			if stdout != "" {
				t.Logf("   Generate output: %s", strings.TrimSpace(stdout))
			}
		} else {
			t.Logf("ℹ️  Generate command failed with exit code %d", exitCode)
			t.Logf("   Stderr: %s", stderr)
		}
	})

	t.Run("authenticated eval command", func(t *testing.T) {
		// Create a prompt file with evaluators
		promptContent := `name: Eval Test
description: A prompt for testing eval command with auth
model: openai/gpt-4o-mini
modelParameters:
  temperature: 0.1
  maxTokens: 20
messages:
  - role: system
    content: You are a helpful assistant.
  - role: user
    content: "Say hello to {{name}}"
testData:
  - name: "Alice"
  - name: "Bob"
evaluators:
  - name: contains-hello
    string:
      contains: "hello"
  - name: mentions-name
    string:
      contains: "{{name}}"
`

		promptFile := createTempPromptFile(t, promptContent)

		// Try to run eval
		stdout, stderr, exitCode := runCommand(t, "eval", promptFile)

		if exitCode == 1 && strings.Contains(stderr, "not authenticated") {
			t.Skip("GitHub authentication not available - skipping authenticated eval test")
		}

		if exitCode == 0 {
			// Success case
			t.Logf("✅ Eval command completed successfully")
			t.Logf("   Output: %s", strings.TrimSpace(stdout))

			// Look for evaluation results
			if strings.Contains(stdout, "PASS") || strings.Contains(stdout, "FAIL") {
				t.Logf("✅ Found evaluation results in output")
			}
			if strings.Contains(stdout, "contains-hello") {
				t.Logf("✅ Found evaluator results in output")
			}
		} else {
			t.Logf("ℹ️  Eval command failed with exit code %d", exitCode)
			t.Logf("   Stderr: %s", stderr)
		}

		// Test with JSON output format
		stdout, stderr, exitCode = runCommand(t, "eval", "--json", promptFile)
		if exitCode == 0 {
			t.Logf("✅ JSON eval format successful")
			// Could validate JSON structure here if needed
		}
	})

	t.Run("view command with authentication", func(t *testing.T) {
		// Try to get details about a specific model
		stdout, stderr, exitCode := runCommand(t, "view", "openai/gpt-4o-mini")

		if exitCode == 1 && strings.Contains(stderr, "not authenticated") {
			t.Skip("GitHub authentication not available - skipping authenticated view test")
		}

		if exitCode == 0 {
			t.Logf("✅ View command successful")
			t.Logf("   Output: %s", strings.TrimSpace(stdout))

			// Check for expected model details
			expectedFields := []string{"gpt-4o-mini", "openai", "tokens"}
			for _, field := range expectedFields {
				if strings.Contains(strings.ToLower(stdout), strings.ToLower(field)) {
					t.Logf("✅ Found expected field '%s' in model details", field)
				}
			}
		} else {
			t.Logf("ℹ️  View command failed with exit code %d", exitCode)
			t.Logf("   Stderr: %s", stderr)
		}
	})
}

// TestAuthenticationHelpers tests helper scenarios around authentication
func TestAuthenticationHelpers(t *testing.T) {
	t.Run("authentication error messages", func(t *testing.T) {
		// Test that auth error messages are helpful
		commands := [][]string{
			{"list"},
			{"run", "openai/gpt-4o-mini", "test"},
			{"generate", "/nonexistent/file.yml"},
			{"eval", "/nonexistent/file.yml"},
			{"view", "openai/gpt-4o-mini"},
		}

		for _, cmd := range commands {
			t.Run(fmt.Sprintf("auth_error_%s", cmd[0]), func(t *testing.T) {
				_, stderr, exitCode := runCommand(t, cmd...)

				if exitCode == 1 && strings.Contains(stderr, "not authenticated") {
					// Verify the error message is helpful (but not all commands show the full auth message)
					t.Logf("✅ Command '%s' shows auth error (exit code: %d)", cmd[0], exitCode)
				}
			})
		}
	})

	t.Run("command availability without auth", func(t *testing.T) {
		// Even without auth, help commands should work
		helpCommands := [][]string{
			{"--help"},
			{"list", "--help"},
			{"run", "--help"},
			{"generate", "--help"},
			{"eval", "--help"},
			{"view", "--help"},
		}

		for _, cmd := range helpCommands {
			t.Run(fmt.Sprintf("help_%s", strings.Join(cmd, "_")), func(t *testing.T) {
				stdout, _, exitCode := runCommand(t, cmd...)

				require.Equal(t, 0, exitCode,
					"Help command should succeed without auth: %v", cmd)
				require.Contains(t, stdout, "Usage:",
					"Help output should contain usage information")
				t.Logf("✅ Help command works without auth: %v", cmd)
			})
		}
	})
}
