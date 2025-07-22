// Package generate provides a gh command to generate tests.
package generate

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/spf13/cobra"
)

type generateCommandHandler struct {
	ctx     context.Context
	cfg     *command.Config
	client  azuremodels.Client
	options PromptPexOptions
	org     string
}

// NewGenerateCommand returns a new command to generate tests using PromptPex.
func NewGenerateCommand(cfg *command.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [prompt-file]",
		Short: "Generate tests using PromptPex",
		Long: heredoc.Docf(`
			Augment prompt.yml file with generated test cases using the PromptPex methodology.
			
			This command analyzes a prompt file and generates comprehensive test cases to evaluate
			the prompt's behavior across different scenarios and edge cases.
		`, "`"),
		Example: heredoc.Doc(`
			gh models generate prompt.yml
			gh models generate --effort medium --models-under-test "openai/gpt-4o-mini,openai/gpt-4o" prompt.yml
			gh models generate --org my-org --groundtruth-model "openai/gpt-4o" prompt.yml
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			promptFile := args[0]

			// Parse command-line options
			options := GetDefaultOptions()

			// Parse flags and apply to options
			if err := parseFlags(cmd, &options); err != nil {
				return fmt.Errorf("failed to parse flags: %w", err)
			}

			// Get organization
			org, _ := cmd.Flags().GetString("org")

			// Create the command handler
			handler := &generateCommandHandler{
				ctx:     cmd.Context(),
				cfg:     cfg,
				client:  cfg.Client,
				options: options,
				org:     org,
			}

			// Create PromptPex context
			context, err := handler.CreateContext(promptFile)
			if err != nil {
				return fmt.Errorf("failed to create context: %w", err)
			}

			// Run the PromptPex pipeline
			if err := handler.runPipeline(context); err != nil {
				return fmt.Errorf("pipeline failed: %w", err)
			}

			return nil
		},
	}

	// Add command-line flags
	cmd.Flags().String("org", "", "Organization to attribute usage to")
	cmd.Flags().String("effort", "", "Effort level (min, low, medium, high)")
	cmd.Flags().StringSlice("models-under-test", []string{}, "Models to test (can be used multiple times)")
	cmd.Flags().String("groundtruth-model", "", "Model to use for generating groundtruth outputs")
	cmd.Flags().Int("tests-per-rule", 0, "Number of tests to generate per rule")
	cmd.Flags().Int("runs-per-test", 0, "Number of times to run each test")
	cmd.Flags().Int("test-expansions", 0, "Number of test expansion phases")
	cmd.Flags().Bool("rate-tests", false, "Enable test rating")
	cmd.Flags().Bool("evals", false, "Enable evaluations")
	cmd.Flags().StringSlice("eval-models", []string{}, "Models to use for evaluation")
	cmd.Flags().String("custom-metric", "", "Custom evaluation metric")
	cmd.Flags().Float64("temperature", 0.0, "Temperature for model inference")

	return cmd
}

// parseFlags parses command-line flags and applies them to the options
func parseFlags(cmd *cobra.Command, options *PromptPexOptions) error {
	// Parse effort first so it can set defaults
	if effort, _ := cmd.Flags().GetString("effort"); effort != "" {
		options.Effort = &effort
	}

	// Apply effort configuration
	if options.Effort != nil {
		ApplyEffortConfiguration(options, *options.Effort)
	}

	// Parse other flags (these override effort defaults)
	if modelsUnderTest, _ := cmd.Flags().GetStringSlice("models-under-test"); len(modelsUnderTest) > 0 {
		options.ModelsUnderTest = modelsUnderTest
	}

	if groundtruthModel, _ := cmd.Flags().GetString("groundtruth-model"); groundtruthModel != "" {
		options.GroundtruthModel = &groundtruthModel
	}

	if cmd.Flags().Changed("tests-per-rule") {
		testsPerRule, _ := cmd.Flags().GetInt("tests-per-rule")
		options.TestsPerRule = &testsPerRule
	}

	if cmd.Flags().Changed("runs-per-test") {
		runsPerTest, _ := cmd.Flags().GetInt("runs-per-test")
		options.RunsPerTest = &runsPerTest
	}

	if cmd.Flags().Changed("test-expansions") {
		testExpansions, _ := cmd.Flags().GetInt("test-expansions")
		options.TestExpansions = &testExpansions
	}

	if cmd.Flags().Changed("rate-tests") {
		rateTests, _ := cmd.Flags().GetBool("rate-tests")
		options.RateTests = &rateTests
	}

	if cmd.Flags().Changed("evals") {
		evals, _ := cmd.Flags().GetBool("evals")
		options.Evals = &evals
	}

	if evalModels, _ := cmd.Flags().GetStringSlice("eval-models"); len(evalModels) > 0 {
		options.EvalModels = evalModels
	}

	if customMetric, _ := cmd.Flags().GetString("custom-metric"); customMetric != "" {
		options.CustomMetric = &customMetric
	}

	if cmd.Flags().Changed("temperature") {
		temperature, _ := cmd.Flags().GetFloat64("temperature")
		options.Temperature = &temperature
	}

	return nil
}
