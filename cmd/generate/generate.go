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
			if err := ParseFlags(cmd, &options); err != nil {
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
	AddCommandLineFlags(cmd)

	return cmd
}

func AddCommandLineFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.String("org", "", "Organization to attribute usage to")
	flags.String("effort", "", "Effort level (min, low, medium, high)")
	flags.StringSlice("models-under-test", []string{}, "Models to test (can be used multiple times)")
	flags.String("groundtruth-model", "", "Model to use for generating groundtruth outputs")
	flags.Int("tests-per-rule", 0, "Number of tests to generate per rule")
	flags.Int("runs-per-test", 0, "Number of times to run each test")
	flags.Int("test-expansions", 0, "Number of test expansion phases")
	flags.Bool("rate-tests", false, "Enable test rating")
	flags.Bool("evals", false, "Enable evaluations")
	flags.StringSlice("eval-models", []string{}, "Models to use for evaluation")
	flags.String("custom-metric", "", "Custom evaluation metric")
	flags.Float64("temperature", 0.0, "Temperature for model inference")
}

// parseFlags parses command-line flags and applies them to the options
func ParseFlags(cmd *cobra.Command, options *PromptPexOptions) error {
	flags := cmd.Flags()
	// Parse effort first so it can set defaults
	if effort, _ := flags.GetString("effort"); effort != "" {
		options.Effort = &effort
	}

	// Apply effort configuration
	if options.Effort != nil {
		ApplyEffortConfiguration(options, *options.Effort)
	}

	// Parse other flags (these override effort defaults)
	if modelsUnderTest, _ := flags.GetStringSlice("models-under-test"); len(modelsUnderTest) > 0 {
		options.ModelsUnderTest = modelsUnderTest
	}

	if groundtruthModel, _ := flags.GetString("groundtruth-model"); groundtruthModel != "" {
		options.GroundtruthModel = &groundtruthModel
	}

	if flags.Changed("tests-per-rule") {
		testsPerRule, _ := flags.GetInt("tests-per-rule")
		options.TestsPerRule = &testsPerRule
	}

	if flags.Changed("runs-per-test") {
		runsPerTest, _ := flags.GetInt("runs-per-test")
		options.RunsPerTest = &runsPerTest
	}

	if flags.Changed("test-expansions") {
		testExpansions, _ := flags.GetInt("test-expansions")
		options.TestExpansions = &testExpansions
	}

	if flags.Changed("rate-tests") {
		rateTests, _ := flags.GetBool("rate-tests")
		options.RateTests = &rateTests
	}

	if flags.Changed("evals") {
		evals, _ := flags.GetBool("evals")
		options.Evals = &evals
	}

	if evalModels, _ := flags.GetStringSlice("eval-models"); len(evalModels) > 0 {
		options.EvalModels = evalModels
	}

	if customMetric, _ := flags.GetString("custom-metric"); customMetric != "" {
		options.CustomMetric = &customMetric
	}

	if flags.Changed("temperature") {
		temperature, _ := flags.GetFloat64("temperature")
		options.Temperature = &temperature
	}

	return nil
}
