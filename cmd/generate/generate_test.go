package generate

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/stretchr/testify/require"
)

func TestNewGenerateCommand(t *testing.T) {
	t.Run("creates command with correct structure", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(new(bytes.Buffer), new(bytes.Buffer), client, true, 80)

		cmd := NewGenerateCommand(cfg)

		require.Equal(t, "generate [prompt-file]", cmd.Use)
		require.Equal(t, "Generate tests and evaluations for prompts", cmd.Short)
		require.Contains(t, cmd.Long, "PromptPex methodology")
		require.True(t, cmd.Args != nil) // Should have ExactArgs(1)

		// Check that flags are added
		flags := cmd.Flags()
		require.True(t, flags.Lookup("org") != nil)
		require.True(t, flags.Lookup("effort") != nil)
		require.True(t, flags.Lookup("groundtruth-model") != nil)
	})

	t.Run("--help prints usage info", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		cmd := NewGenerateCommand(nil)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--help"})

		err := cmd.Help()

		require.NoError(t, err)
		output := outBuf.String()
		require.Contains(t, output, "Augment prompt.yml file with generated test cases")
		require.Contains(t, output, "PromptPex methodology")
		require.Regexp(t, regexp.MustCompile(`--effort string\s+Effort level`), output)
		require.Regexp(t, regexp.MustCompile(`--groundtruth-model string\s+Model to use for generating groundtruth`), output)
		require.Empty(t, errBuf.String())
	})
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(*testing.T, *PromptPexOptions)
	}{
		{
			name: "default options preserve initial state",
			args: []string{},
			validate: func(t *testing.T, opts *PromptPexOptions) {
				require.Equal(t, 3, *opts.TestsPerRule)
				require.Equal(t, 2, *opts.RunsPerTest)
			},
		},
		{
			name: "effort flag is set",
			args: []string{"--effort", "medium"},
			validate: func(t *testing.T, opts *PromptPexOptions) {
				require.NotNil(t, opts.Effort)
				require.Equal(t, "medium", *opts.Effort)
			},
		},
		{
			name: "groundtruth model flag",
			args: []string{"--groundtruth-model", "openai/gpt-4o"},
			validate: func(t *testing.T, opts *PromptPexOptions) {
				require.NotNil(t, opts.Models.Groundtruth)
				require.Equal(t, "openai/gpt-4o", *opts.Models.Groundtruth)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary command to parse flags
			cmd := NewGenerateCommand(nil)
			cmd.SetArgs(append(tt.args, "dummy.yml")) // Add required positional arg

			// Parse flags but don't execute
			err := cmd.ParseFlags(tt.args)
			require.NoError(t, err)

			// Parse options from the flags
			options := GetDefaultOptions()
			err = ParseFlags(cmd, options)
			require.NoError(t, err)

			// Validate using the test-specific validation function
			tt.validate(t, options)
		})
	}
}

func TestGenerateCommandExecution(t *testing.T) {

	t.Run("fails with invalid prompt file", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewGenerateCommand(cfg)
		cmd.SetArgs([]string{"nonexistent.yml"})

		err := cmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to create context")
	})

	t.Run("handles LLM errors gracefully", func(t *testing.T) {
		// Create test prompt file
		const yamlBody = `
name: Test Prompt
description: Test description
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Test prompt"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		// Setup mock client to return error
		client := azuremodels.NewMockClient()
		client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
			return nil, errors.New("Mock API error")
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewGenerateCommand(cfg)
		cmd.SetArgs([]string{promptFile})

		err = cmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "pipeline failed")
	})
}

func TestGenerateCommandHandlerContext(t *testing.T) {
	t.Run("creates context with valid prompt file", func(t *testing.T) {
		// Create test prompt file
		const yamlBody = `
name: Test Context Creation
description: Test description for context
model: openai/gpt-4o-mini
messages:
  - role: user
    content: "Test content"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		// Create handler
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(new(bytes.Buffer), new(bytes.Buffer), client, true, 100)
		options := GetDefaultOptions()

		handler := &generateCommandHandler{
			ctx:        context.Background(),
			cfg:        cfg,
			client:     client,
			options:    options,
			promptFile: promptFile,
			org:        "",
		}

		// Test context creation
		ctx, err := handler.CreateContextFromPrompt()
		require.NoError(t, err)
		require.NotNil(t, ctx)
		require.NotEmpty(t, ctx.RunID)
		require.True(t, ctx.RunID != nil)
		require.Equal(t, "Test Context Creation", ctx.Prompt.Name)
		require.Equal(t, "Test description for context", ctx.Prompt.Description)
		require.Equal(t, options, ctx.Options)
	})

	t.Run("fails with invalid prompt file", func(t *testing.T) {
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(new(bytes.Buffer), new(bytes.Buffer), client, true, 100)
		options := GetDefaultOptions()

		handler := &generateCommandHandler{
			ctx:        context.Background(),
			cfg:        cfg,
			client:     client,
			options:    options,
			promptFile: "nonexistent.yml",
			org:        "",
		}

		// Test with nonexistent file
		_, err := handler.CreateContextFromPrompt()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to load prompt file")
	})
}
