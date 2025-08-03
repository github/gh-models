package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	binaryName      = "gh-models-test"
	timeoutDuration = 30 * time.Second
)

// getBinaryPath returns the path to the compiled gh-models binary
func getBinaryPath(t *testing.T) string {
	wd, err := os.Getwd()
	require.NoError(t, err)

	// Binary should be in the parent directory
	binaryPath := filepath.Join(filepath.Dir(wd), binaryName)

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("Binary %s not found. Run 'script/build' first.", binaryPath)
	}

	return binaryPath
}

// hasAuthToken checks if GitHub authentication is available
func hasAuthToken() bool {
	// Check if gh CLI is available and authenticated
	cmd := exec.Command("gh", "auth", "status")
	return cmd.Run() == nil
}

// runCommand executes the gh-models binary with given arguments
func runCommand(t *testing.T, args ...string) (stdout, stderr string, err error) {
	binaryPath := getBinaryPath(t)

	cmd := exec.Command(binaryPath, args...)
	cmd.Env = os.Environ()

	// Set timeout
	done := make(chan error, 1)
	var stdoutBytes, stderrBytes []byte

	go func() {
		stdoutBytes, err = cmd.Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				stderrBytes = exitError.Stderr
			}
		}
		done <- err
	}()

	select {
	case err = <-done:
		return string(stdoutBytes), string(stderrBytes), err
	case <-time.After(timeoutDuration):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("Command timed out after %v", timeoutDuration)
		return "", "", nil
	}
}

func TestIntegrationHelp(t *testing.T) {
	stdout, stderr, err := runCommand(t, "--help")

	// Help should always work, even without auth
	require.NoError(t, err, "stderr: %s", stderr)
	require.Contains(t, stdout, "GitHub Models CLI extension")
	require.Contains(t, stdout, "Available Commands:")
	require.Contains(t, stdout, "list")
	require.Contains(t, stdout, "run")
	require.Contains(t, stdout, "view")
	require.Contains(t, stdout, "eval")
}

func TestIntegrationList(t *testing.T) {
	if !hasAuthToken() {
		t.Skip("Skipping integration test - no GitHub authentication available")
	}

	stdout, stderr, err := runCommand(t, "list")

	if err != nil {
		t.Logf("List command failed. stdout: %s, stderr: %s", stdout, stderr)
		// If the command fails due to auth issues, skip the test
		if strings.Contains(stderr, "authentication") || strings.Contains(stderr, "token") {
			t.Skip("Skipping - authentication issue")
		}
		require.NoError(t, err, "List command should succeed with valid auth")
	}

	// Basic verification that list command produces expected output format
	require.NotEmpty(t, stdout, "List should produce output")
	// Should contain some indication of models or table headers
	lowerOut := strings.ToLower(stdout)
	hasExpectedContent := strings.Contains(lowerOut, "model") ||
		strings.Contains(lowerOut, "name") ||
		strings.Contains(lowerOut, "id") ||
		strings.Contains(lowerOut, "display")
	require.True(t, hasExpectedContent, "List output should contain model information")
}

func TestIntegrationListHelp(t *testing.T) {
	stdout, stderr, err := runCommand(t, "list", "--help")

	require.NoError(t, err, "stderr: %s", stderr)
	require.Contains(t, stdout, "Returns a list of models")
	require.Contains(t, stdout, "Usage:")
}

func TestIntegrationView(t *testing.T) {
	if !hasAuthToken() {
		t.Skip("Skipping integration test - no GitHub authentication available")
	}

	// First get a model to view
	listOut, _, listErr := runCommand(t, "list")
	if listErr != nil {
		t.Skip("Cannot run view test - list command failed")
	}

	// Extract a model name from list output (this is basic parsing)
	lines := strings.Split(listOut, "\n")
	var modelName string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines that might contain model IDs (containing forward slash)
		if strings.Contains(line, "/") && !strings.HasPrefix(line, "Usage:") &&
			!strings.HasPrefix(line, "gh models") && line != "" {
			// Try to extract what looks like a model ID
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.Contains(field, "/") {
					modelName = field
					break
				}
			}
			if modelName != "" {
				break
			}
		}
	}

	if modelName == "" {
		t.Skip("Could not extract model name from list output")
	}

	stdout, stderr, err := runCommand(t, "view", modelName)

	if err != nil {
		t.Logf("View command failed for model %s. stdout: %s, stderr: %s", modelName, stdout, stderr)
		// If the command fails due to auth issues, skip the test
		if strings.Contains(stderr, "authentication") || strings.Contains(stderr, "token") {
			t.Skip("Skipping - authentication issue")
		}
		require.NoError(t, err, "View command should succeed with valid model")
	}

	// Basic verification that view command produces expected output
	require.NotEmpty(t, stdout, "View should produce output")
	lowerOut := strings.ToLower(stdout)
	hasExpectedContent := strings.Contains(lowerOut, "model") ||
		strings.Contains(lowerOut, "name") ||
		strings.Contains(lowerOut, "description") ||
		strings.Contains(lowerOut, "publisher")
	require.True(t, hasExpectedContent, "View output should contain model details")
}

func TestIntegrationViewHelp(t *testing.T) {
	stdout, stderr, err := runCommand(t, "view", "--help")

	require.NoError(t, err, "stderr: %s", stderr)
	require.Contains(t, stdout, "Returns details about the specified model")
	require.Contains(t, stdout, "Usage:")
}

func TestIntegrationRunHelp(t *testing.T) {
	stdout, stderr, err := runCommand(t, "run", "--help")

	require.NoError(t, err, "stderr: %s", stderr)
	require.Contains(t, stdout, "Prompts the specified model")
	require.Contains(t, stdout, "Usage:")
}

func TestIntegrationEvalHelp(t *testing.T) {
	stdout, stderr, err := runCommand(t, "eval", "--help")

	require.NoError(t, err, "stderr: %s", stderr)
	require.Contains(t, stdout, "Runs evaluation tests against a model")
	require.Contains(t, stdout, "Usage:")
}

// TestIntegrationRun tests the run command with a simple prompt
// This test is more limited since it requires actual model inference
func TestIntegrationRun(t *testing.T) {
	if !hasAuthToken() {
		t.Skip("Skipping integration test - no GitHub authentication available")
	}

	// We'll test with a very simple prompt to minimize cost and time
	// Using a basic model and short prompt
	stdout, _, err := runCommand(t, "run", "--help")
	require.NoError(t, err, "Run help should work")

	// For now, just verify the help works.
	// A full test would require setting up a prompt and model,
	// which might be expensive for CI
	require.Contains(t, stdout, "Prompts the specified model")
}

// TestIntegrationRunWithOrg tests the run command with --org flag
func TestIntegrationRunWithOrg(t *testing.T) {
	if !hasAuthToken() {
		t.Skip("Skipping integration test - no GitHub authentication available")
	}

	// Test run command with --org flag (using help to avoid expensive API calls)
	stdout, _, err := runCommand(t, "run", "--org", "test-org", "--help")
	require.NoError(t, err, "Run help with --org should work")
	require.Contains(t, stdout, "Prompts the specified model")
	require.Contains(t, stdout, "--org string")
}

// TestIntegrationEvalWithOrg tests the eval command with --org flag
func TestIntegrationEvalWithOrg(t *testing.T) {
	if !hasAuthToken() {
		t.Skip("Skipping integration test - no GitHub authentication available")
	}

	// Test eval command with --org flag (using help to avoid expensive API calls)
	stdout, _, err := runCommand(t, "eval", "--org", "test-org", "--help")
	require.NoError(t, err, "Eval help with --org should work")
	require.Contains(t, stdout, "Runs evaluation tests against a model")
	require.Contains(t, stdout, "--org string")
}
