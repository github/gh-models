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

func TestList(t *testing.T) {
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
	hasExpectedContent := strings.Contains(lowerOut, "openai/gpt-4.1")
	require.True(t, hasExpectedContent, "List output should contain model information")
}

// TestRun tests the run command with a simple prompt
// This test is more limited since it requires actual model inference
func TestRun(t *testing.T) {
	stdout, _, err := runCommand(t, "run", "openai/gpt-4.1-nano", "say 'pain' in french")
	require.NoError(t, err, "Run should work")
	require.Contains(t, strings.ToLower(stdout), "pain")
}

// TestIntegrationRunWithOrg tests the run command with --org flag
func TestRunWithOrg(t *testing.T) {
	// Test run command with --org flag (using help to avoid expensive API calls)
	stdout, _, err := runCommand(t, "run", "openai/gpt-4.1-nano", "say 'pain' in french", "--org", "github")
	require.NoError(t, err, "Run should work")
	require.Contains(t, strings.ToLower(stdout), "pain")
}
