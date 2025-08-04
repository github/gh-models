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
	ctx          context.Context
	cfg          *command.Config
	client       azuremodels.Client
	options      *PromptPexOptions
	promptFile   string
	org          string
	sessionFile  *string
	templateVars map[string]string
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
			gh models generate --var name=Alice --var topic="machine learning" prompt.yml
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

			// Parse template variables from flags
			templateVars, err := util.ParseTemplateVariables(cmd.Flags())
			if err != nil {
				return err
			}

			// Check for reserved keys specific to generate command
			if _, exists := templateVars["input"]; exists {
				return fmt.Errorf("'input' is a reserved variable name and cannot be used with --var")
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
				ctx:          ctx,
				cfg:          cfg,
				client:       cfg.Client,
				options:      options,
				promptFile:   promptFile,
				org:          org,
				sessionFile:  util.Ptr(sessionFile),
				templateVars: templateVars,
			}

			// Create prompt context
			promptContext, err := handler.CreateContextFromPrompt()
			if err != nil {
				return fmt.Errorf("failed to create context: %w", err)
			}

			// Run the PromptPex pipeline
			if err := handler.RunTestGenerationPipeline(promptContext); err != nil {
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
	flags.String("effort", "", "Effort level (min, low, medium, high)")
	flags.String("groundtruth-model", "", "Model to use for generating groundtruth outputs. Defaults to openai/gpt-4o. Use 'none' to disable groundtruth generation.")
	flags.String("session-file", "", "Session file to load existing context from")
	flags.StringArray("var", []string{}, "Template variables for prompt files (can be used multiple times: --var name=value)")

	// Custom instruction flags for each phase
	flags.String("instruction-intent", "", "Custom system instruction for intent generation phase")
	flags.String("instruction-inputspec", "", "Custom system instruction for input specification generation phase")
	flags.String("instruction-outputrules", "", "Custom system instruction for output rules generation phase")
	flags.String("instruction-inverseoutputrules", "", "Custom system instruction for inverse output rules generation phase")
	flags.String("instruction-tests", "", "Custom system instruction for tests generation phase")
}

// ParseFlags parses command-line flags and applies them to the options
func ParseFlags(cmd *cobra.Command, options *PromptPexOptions) error {
	flags := cmd.Flags()
	// Parse effort first so it can set defaults
	if effort, _ := flags.GetString("effort"); effort != "" {
		// Validate effort value
		if effort != EffortMin && effort != EffortLow && effort != EffortMedium && effort != EffortHigh {
			return fmt.Errorf("invalid effort level '%s': must be one of %s, %s, %s, or %s", effort, EffortMin, EffortLow, EffortMedium, EffortHigh)
		}
		options.Effort = effort
	}

	// Apply effort configuration
	if options.Effort != "" {
		ApplyEffortConfiguration(options, options.Effort)
	}

	if groundtruthModel, _ := flags.GetString("groundtruth-model"); groundtruthModel != "" {
		options.Models.Groundtruth = groundtruthModel
	}

	// Parse custom instruction flags
	if options.Instructions == nil {
		options.Instructions = &PromptPexPrompts{}
	}

	if intentInstruction, _ := flags.GetString("instruction-intent"); intentInstruction != "" {
		options.Instructions.Intent = intentInstruction
	}

	if inputSpecInstruction, _ := flags.GetString("instruction-inputspec"); inputSpecInstruction != "" {
		options.Instructions.InputSpec = inputSpecInstruction
	}

	if outputRulesInstruction, _ := flags.GetString("instruction-outputrules"); outputRulesInstruction != "" {
		options.Instructions.OutputRules = outputRulesInstruction
	}

	if inverseOutputRulesInstruction, _ := flags.GetString("instruction-inverseoutputrules"); inverseOutputRulesInstruction != "" {
		options.Instructions.InverseOutputRules = inverseOutputRulesInstruction
	}

	if testsInstruction, _ := flags.GetString("instruction-tests"); testsInstruction != "" {
		options.Instructions.Tests = testsInstruction
	}

	return nil
}
