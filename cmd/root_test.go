package cmd

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoot(t *testing.T) {
	t.Run("usage info describes sub-commands", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd := NewRootCommand()
		rootCmd.SetOut(buf)

		err := rootCmd.Help()

		require.NoError(t, err)
		output := buf.String()
		require.Regexp(t, regexp.MustCompile(`Usage:\n\s+gh models \[command\]`), output)
		require.Regexp(t, regexp.MustCompile(`eval\s+Evaluate prompts using test data and evaluators`), output)
		require.Regexp(t, regexp.MustCompile(`list\s+List available models`), output)
		require.Regexp(t, regexp.MustCompile(`run\s+Run inference with the specified model`), output)
		require.Regexp(t, regexp.MustCompile(`view\s+View details about a model`), output)
		require.Regexp(t, regexp.MustCompile(`generate\s+Generate tests and evaluations for prompts`), output)
	})
}
