// Package generate provides a gh command to generate tests.
package generate

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/cobra"
)

type generateCommandHandler struct {
	ctx         context.Context
	cfg         *command.Config
	client      azuremodels.Client
	options     *PromptPexOptions
	promptFile  string
	org         string
	sessionFile *string
}

// NewGenerateCommand returns a new command to generate tests using PromptPex.
func NewGenerateCommand(cfg *command.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [prompt-file]",
		Short: "Generate tests and evaluations for prompts",
		Long: heredoc.Docf(`
			Augment prompt.yml file with generated test cases.
			
			This command analyzes a prompt file and generates comprehensive test cases to evaluate
			the prompt's behavior across different scenarios and edge cases using the PromptPex methodology.
		`, "`"),
		Example: heredoc.Doc(`
			gh models generate prompt.yml
			gh models generate --org my-org --groundtruth-model "openai/gpt-4.1" prompt.yml
			gh models generate --session-file prompt.session.json prompt.yml
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			promptFile := args[0]

			// Parse command-line options
			options := GetDefaultOptions()

			// Parse flags and apply to options
			if err := ParseFlags(cmd, options); err != nil {
				return fmt.Errorf("failed to parse flags: %w", err)
			}

			// Get organization
			org, _ := cmd.Flags().GetString("org")

			// Get session-file flag
			sessionFile, _ := cmd.Flags().GetString("session-file")

			// Get http-log flag
			httpLog, _ := cmd.Flags().GetString("http-log")

			ctx := cmd.Context()
			// Add HTTP log filename to context if provided
			if httpLog != "" {
				ctx = azuremodels.WithHTTPLogFile(ctx, httpLog)
			}

			// Create the command handler
			handler := &generateCommandHandler{
				ctx:         ctx,
				cfg:         cfg,
				client:      cfg.Client,
				options:     options,
				promptFile:  promptFile,
				org:         org,
				sessionFile: util.Ptr(sessionFile),
			}

			// Create PromptPex context
			context, err := handler.CreateContextFromPrompt()
			if err != nil {
				return fmt.Errorf("failed to create context: %w", err)
			}

			// Run the PromptPex pipeline
			if err := handler.RunTestGenerationPipeline(context); err != nil {
				// Disable usage help for pipeline failures
				cmd.SilenceUsage = true
				return fmt.Errorf("pipeline failed: %w", err)
			}

			return nil
		},
	}

	// Add command-line flags
	AddCommandLineFlags(cmd)

	return cmd
}

func AddCommandLineFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.String("org", "", "Organization to attribute usage to")
	flags.String("effort", "", "Effort level (low, medium, high)")
	flags.String("groundtruth-model", "", "Model to use for generating groundtruth outputs. Defaults to openai/gpt-4o. Use 'none' to disable groundtruth generation.")
	flags.String("session-file", "", "Session file to load existing context from (defaults to <prompt-file>.generate.json)")
}

// parseFlags parses command-line flags and applies them to the options
func ParseFlags(cmd *cobra.Command, options *PromptPexOptions) error {
	flags := cmd.Flags()
	// Parse effort first so it can set defaults
	if effort, _ := flags.GetString("effort"); effort != "" {
		options.Effort = effort
	}

	// Apply effort configuration
	if options.Effort != "" {
		ApplyEffortConfiguration(options, options.Effort)
	}

	if groundtruthModel, _ := flags.GetString("groundtruth-model"); groundtruthModel != "" {
		options.Models.Groundtruth = groundtruthModel
	}

	return nil
}
