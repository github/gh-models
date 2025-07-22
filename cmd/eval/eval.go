// Package eval provides a gh command to evaluate prompts against GitHub models.
package eval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/prompt"
	"github.com/github/gh-models/pkg/util"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
)

var (
	lightGrayUnderline = ansi.ColorFunc("white+du")
	red                = ansi.ColorFunc("red")
	green              = ansi.ColorFunc("green")
)

// EvaluationSummary represents the overall evaluation summary
type EvaluationSummary struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Model       string       `json:"model"`
	TestResults []TestResult `json:"testResults"`
	Summary     Summary      `json:"summary"`
}

// Summary represents the evaluation summary statistics
type Summary struct {
	TotalTests  int     `json:"totalTests"`
	PassedTests int     `json:"passedTests"`
	FailedTests int     `json:"failedTests"`
	PassRate    float64 `json:"passRate"`
}

// TestResult represents the result of running a test case
type TestResult struct {
	TestCase          map[string]interface{} `json:"testCase"`
	ModelResponse     string                 `json:"modelResponse"`
	EvaluationResults []EvaluationResult     `json:"evaluationResults"`
}

// EvaluationResult represents the result of a single evaluator
type EvaluationResult struct {
	EvaluatorName string  `json:"evaluatorName"`
	Score         float64 `json:"score"`
	Passed        bool    `json:"passed"`
	Details       string  `json:"details,omitempty"`
}

var FailedTests = errors.New("❌ Some tests failed.")

// NewEvalCommand returns a new command to evaluate prompts against models
func NewEvalCommand(cfg *command.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eval",
		Short: "Evaluate prompts using test data and evaluators",
		Long: heredoc.Docf(`
			Runs evaluation tests against a model using a prompt.yml file.

			The prompt.yml file should contain:
			- Model configuration and parameters
			- Test data with input variables
			- Messages with templated content
			- Evaluators to assess model responses

			Example prompt.yml structure:
			  name: My Evaluation
			  model: openai/gpt-4o
			  testData:
			    - input: "Hello world"
			      expected: "Hello there"
			  messages:
			    - role: user
			      content: "Respond to: {{input}}"
			  evaluators:
			    - name: contains-hello
			      string:
			        contains: "hello"

			By default, results are displayed in a human-readable format. Use the --json flag
			to output structured JSON data for programmatic use or integration with CI/CD pipelines.
			This command will automatically retry on rate limiting errors, waiting for the specified
			duration before retrying the request.

			See https://docs.github.com/github-models/use-github-models/storing-prompts-in-github-repositories#supported-file-format for more information.
		`),
		Example: heredoc.Doc(`
			gh models eval my_prompt.prompt.yml
			gh models eval --org my-org my_prompt.prompt.yml
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			promptFilePath := args[0]

			// Get the json flag
			jsonOutput, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			// Get the org flag
			org, _ := cmd.Flags().GetString("org")

			// Load the evaluation prompt file
			evalFile, err := loadEvaluationPromptFile(promptFilePath)
			if err != nil {
				return fmt.Errorf("failed to load prompt file: %w", err)
			}

			// Run evaluation
			handler := &evalCommandHandler{
				cfg:        cfg,
				client:     cfg.Client,
				evalFile:   evalFile,
				jsonOutput: jsonOutput,
				org:        org,
			}

			err = handler.runEvaluation(cmd.Context())
			if err == FailedTests {
				// Cobra by default will show the help message when an error occurs,
				// which is not what we want for failed evaluations.
				// Instead, we just want to exit with a non-zero code.
				cmd.SilenceUsage = true
			}
			return err
		},
	}

	cmd.Flags().Bool("json", false, "Output results in JSON format")
	cmd.Flags().String("org", "", "Organization to attribute usage to (omitting will attribute usage to the current actor")
	return cmd
}

type evalCommandHandler struct {
	cfg        *command.Config
	client     azuremodels.Client
	evalFile   *prompt.File
	jsonOutput bool
	org        string
}

func loadEvaluationPromptFile(filePath string) (*prompt.File, error) {
	evalFile, err := prompt.LoadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt file: %w", err)
	}

	return evalFile, nil
}

func (h *evalCommandHandler) runEvaluation(ctx context.Context) error {
	// Print header info only for human-readable output
	if !h.jsonOutput {
		h.cfg.WriteToOut(fmt.Sprintf("Running evaluation: %s\n", h.evalFile.Name))
		h.cfg.WriteToOut(fmt.Sprintf("Description: %s\n", h.evalFile.Description))
		h.cfg.WriteToOut(fmt.Sprintf("Model: %s\n", h.evalFile.Model))
		h.cfg.WriteToOut(fmt.Sprintf("Test cases: %d\n", len(h.evalFile.TestData)))
		h.cfg.WriteToOut("\n")
	}

	var testResults []TestResult
	passedTests := 0
	totalTests := len(h.evalFile.TestData)

	for i, testCase := range h.evalFile.TestData {
		if !h.jsonOutput {
			h.cfg.WriteToOut("-------------------------\n")
			h.cfg.WriteToOut(fmt.Sprintf("Running test case %d/%d...\n", i+1, totalTests))
		}

		result, err := h.runTestCase(ctx, testCase)
		if err != nil {
			return fmt.Errorf("test case %d failed: %w", i+1, err)
		}

		testResults = append(testResults, result)

		// Check if all evaluators passed
		testPassed := true
		for _, evalResult := range result.EvaluationResults {
			if !evalResult.Passed {
				testPassed = false
				break
			}
		}

		if testPassed {
			passedTests++
		}

		if !h.jsonOutput {
			h.printTestResult(result, testPassed)
		}
	}

	// Calculate pass rate
	passRate := 100.0
	if totalTests > 0 {
		passRate = float64(passedTests) / float64(totalTests) * 100
	}

	if h.jsonOutput {
		// Output JSON format
		summary := EvaluationSummary{
			Name:        h.evalFile.Name,
			Description: h.evalFile.Description,
			Model:       h.evalFile.Model,
			TestResults: testResults,
			Summary: Summary{
				TotalTests:  totalTests,
				PassedTests: passedTests,
				FailedTests: totalTests - passedTests,
				PassRate:    passRate,
			},
		}

		jsonData, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		h.cfg.WriteToOut(string(jsonData) + "\n")
	} else {
		// Output human-readable format summary
		h.printSummary(passedTests, totalTests, passRate)
	}

	if totalTests-passedTests > 0 {
		return FailedTests
	}

	return nil
}

func (h *evalCommandHandler) printTestResult(result TestResult, testPassed bool) {
	printer := h.cfg.NewTablePrinter()
	if testPassed {
		printer.AddField("Result", tableprinter.WithColor(lightGrayUnderline))
		printer.AddField("✓ PASSED", tableprinter.WithColor(green))
		printer.EndRow()
	} else {
		printer.AddField("Result", tableprinter.WithColor(lightGrayUnderline))
		printer.AddField("✗ FAILED", tableprinter.WithColor(red))
		printer.EndRow()
		// Show the first 100 characters of the model response when test fails
		preview := result.ModelResponse
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}

		printer.AddField("Model Response", tableprinter.WithColor(lightGrayUnderline))
		printer.AddField(preview)
		printer.EndRow()
	}

	err := printer.Render()
	if err != nil {
		return
	}

	h.cfg.WriteToOut("\n")

	table := h.cfg.NewTablePrinter()
	table.AddHeader([]string{"EVALUATION", "RESULT", "SCORE", "CRITERIA"}, tableprinter.WithColor(lightGrayUnderline))
	// Show evaluation details
	for _, evalResult := range result.EvaluationResults {
		status, color := "✓", green
		if !evalResult.Passed {
			status, color = "✗", red
		}
		table.AddField(evalResult.EvaluatorName)
		table.AddField(status, tableprinter.WithColor(color))
		table.AddField(fmt.Sprintf("%.2f", evalResult.Score), tableprinter.WithColor(color))

		if evalResult.Details != "" {
			table.AddField(evalResult.Details)
		}
		table.EndRow()
	}

	err = table.Render()
	if err != nil {
		return
	}

	h.cfg.WriteToOut("\n")
}

func (h *evalCommandHandler) printSummary(passedTests, totalTests int, passRate float64) {
	// Summary
	h.cfg.WriteToOut("Evaluation Summary:\n")
	if totalTests == 0 {
		h.cfg.WriteToOut("Passed: 0/0 (0.00%)\n")
	} else {
		h.cfg.WriteToOut(fmt.Sprintf("Passed: %d/%d (%.2f%%)\n",
			passedTests, totalTests, passRate))
	}

	if passedTests == totalTests {
		h.cfg.WriteToOut("🎉 All tests passed!\n")
	}
}

func (h *evalCommandHandler) runTestCase(ctx context.Context, testCase map[string]interface{}) (TestResult, error) {
	// Template the messages with test case data
	messages, err := h.templateMessages(testCase)
	if err != nil {
		return TestResult{}, fmt.Errorf("failed to template messages: %w", err)
	}

	// Call the model
	response, err := h.callModel(ctx, messages)
	if err != nil {
		return TestResult{}, fmt.Errorf("failed to call model: %w", err)
	}

	// Run evaluators
	evalResults, err := h.runEvaluators(ctx, testCase, response)
	if err != nil {
		return TestResult{}, fmt.Errorf("failed to run evaluators: %w", err)
	}

	return TestResult{
		TestCase:          testCase,
		ModelResponse:     response,
		EvaluationResults: evalResults,
	}, nil
}

func (h *evalCommandHandler) templateMessages(testCase map[string]interface{}) ([]azuremodels.ChatMessage, error) {
	var messages []azuremodels.ChatMessage

	for _, msg := range h.evalFile.Messages {
		content, err := h.templateString(msg.Content, testCase)
		if err != nil {
			return nil, fmt.Errorf("failed to template message content: %w", err)
		}

		role, err := prompt.GetAzureChatMessageRole(msg.Role)
		if err != nil {
			return nil, err
		}

		messages = append(messages, azuremodels.ChatMessage{
			Role:    role,
			Content: util.Ptr(content),
		})
	}

	return messages, nil
}

func (h *evalCommandHandler) templateString(templateStr string, data map[string]interface{}) (string, error) {
	return prompt.TemplateString(templateStr, data)
}

// callModelWithRetry makes an API call with automatic retry on rate limiting
func (h *evalCommandHandler) callModelWithRetry(ctx context.Context, req azuremodels.ChatCompletionOptions) (string, error) {
	const maxRetries = 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := h.client.GetChatCompletionStream(ctx, req, h.org)
		if err != nil {
			var rateLimitErr *azuremodels.RateLimitError
			if errors.As(err, &rateLimitErr) {
				if attempt < maxRetries {
					if !h.jsonOutput {
						h.cfg.WriteToOut(fmt.Sprintf("    Rate limited, waiting %v before retry (attempt %d/%d)...\n",
							rateLimitErr.RetryAfter, attempt+1, maxRetries+1))
					}

					// Wait for the specified duration
					select {
					case <-ctx.Done():
						return "", ctx.Err()
					case <-time.After(rateLimitErr.RetryAfter):
						continue
					}
				}
				return "", fmt.Errorf("rate limit exceeded after %d attempts: %w", attempt+1, err)
			}
			// For non-rate-limit errors, return immediately
			return "", err
		}

		var content strings.Builder
		for {
			completion, err := resp.Reader.Read()
			if err != nil {
				if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "EOF") {
					break
				}
				return "", err
			}

			for _, choice := range completion.Choices {
				if choice.Delta != nil && choice.Delta.Content != nil {
					content.WriteString(*choice.Delta.Content)
				}
				if choice.Message != nil && choice.Message.Content != nil {
					content.WriteString(*choice.Message.Content)
				}
			}
		}

		return strings.TrimSpace(content.String()), nil
	}

	// This should never be reached, but just in case
	return "", errors.New("unexpected error calling model")
}

func (h *evalCommandHandler) callModel(ctx context.Context, messages []azuremodels.ChatMessage) (string, error) {
	req := h.evalFile.BuildChatCompletionOptions(messages)
	return h.callModelWithRetry(ctx, req)
}

func (h *evalCommandHandler) runEvaluators(ctx context.Context, testCase map[string]interface{}, response string) ([]EvaluationResult, error) {
	var results []EvaluationResult

	for _, evaluator := range h.evalFile.Evaluators {
		result, err := h.runSingleEvaluator(ctx, evaluator, testCase, response)
		if err != nil {
			return nil, fmt.Errorf("evaluator %s failed: %w", evaluator.Name, err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (h *evalCommandHandler) runSingleEvaluator(ctx context.Context, evaluator prompt.Evaluator, testCase map[string]interface{}, response string) (EvaluationResult, error) {
	switch {
	case evaluator.String != nil:
		return h.runStringEvaluator(evaluator.Name, *evaluator.String, testCase, response)
	case evaluator.LLM != nil:
		return h.runLLMEvaluator(ctx, evaluator.Name, *evaluator.LLM, testCase, response)
	case evaluator.Uses != "":
		return h.runPluginEvaluator(ctx, evaluator.Name, evaluator.Uses, testCase, response)
	default:
		return EvaluationResult{}, fmt.Errorf("no evaluation method specified for evaluator %s", evaluator.Name)
	}
}

func (h *evalCommandHandler) runStringEvaluator(name string, eval prompt.StringEvaluator, testCase map[string]interface{}, response string) (EvaluationResult, error) {
	var passed bool
	var details string

	switch {
	case eval.Equals != "":
		equals, err := h.templateString(eval.Equals, testCase)
		if err != nil {
			return EvaluationResult{}, fmt.Errorf("failed to template message content: %w", err)
		}
		passed = response == equals
		details = fmt.Sprintf("Expected exact match: '%s'", equals)
	case eval.Contains != "":
		contains, err := h.templateString(eval.Contains, testCase)
		if err != nil {
			return EvaluationResult{}, fmt.Errorf("failed to template message content: %w", err)
		}
		passed = strings.Contains(strings.ToLower(response), strings.ToLower(contains))
		details = fmt.Sprintf("Expected to contain: '%s'", contains)
	case eval.StartsWith != "":
		startsWith, err := h.templateString(eval.StartsWith, testCase)
		if err != nil {
			return EvaluationResult{}, fmt.Errorf("failed to template message content: %w", err)
		}
		passed = strings.HasPrefix(strings.ToLower(response), strings.ToLower(startsWith))
		details = fmt.Sprintf("Expected to start with: '%s'", startsWith)
	case eval.EndsWith != "":
		endsWith, err := h.templateString(eval.EndsWith, testCase)
		if err != nil {
			return EvaluationResult{}, fmt.Errorf("failed to template message content: %w", err)
		}
		passed = strings.HasSuffix(strings.ToLower(response), strings.ToLower(endsWith))
		details = fmt.Sprintf("Expected to end with: '%s'", endsWith)
	default:
		return EvaluationResult{}, errors.New("no string evaluation criteria specified")
	}

	score := 0.0
	if passed {
		score = 1.0
	}

	return EvaluationResult{
		EvaluatorName: name,
		Score:         score,
		Passed:        passed,
		Details:       details,
	}, nil
}

func (h *evalCommandHandler) runLLMEvaluator(ctx context.Context, name string, eval prompt.LLMEvaluator, testCase map[string]interface{}, response string) (EvaluationResult, error) {
	evalData := make(map[string]interface{})
	for k, v := range testCase {
		evalData[k] = v
	}
	evalData["completion"] = response

	promptContent, err := h.templateString(eval.Prompt, evalData)
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("failed to template evaluation prompt: %w", err)
	}

	var messages []azuremodels.ChatMessage
	if eval.SystemPrompt != "" {
		messages = append(messages, azuremodels.ChatMessage{
			Role:    azuremodels.ChatMessageRoleSystem,
			Content: util.Ptr(eval.SystemPrompt),
		})
	}
	messages = append(messages, azuremodels.ChatMessage{
		Role:    azuremodels.ChatMessageRoleUser,
		Content: util.Ptr(promptContent),
	})

	req := azuremodels.ChatCompletionOptions{
		Messages: messages,
		Model:    eval.ModelID,
		Stream:   false,
	}

	evalResponseText, err := h.callModelWithRetry(ctx, req)
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("failed to call evaluation model: %w", err)
	}

	// Match response to choices
	evalResponseText = strings.TrimSpace(strings.ToLower(evalResponseText))
	for _, choice := range eval.Choices {
		if strings.Contains(evalResponseText, strings.ToLower(choice.Choice)) {
			return EvaluationResult{
				EvaluatorName: name,
				Score:         choice.Score,
				Passed:        choice.Score > 0,
				Details:       fmt.Sprintf("LLM evaluation matched choice: '%s'", choice.Choice),
			}, nil
		}
	}

	// No match found
	return EvaluationResult{
		EvaluatorName: name,
		Score:         0.0,
		Passed:        false,
		Details:       fmt.Sprintf("LLM evaluation response '%s' did not match any defined choices", evalResponseText),
	}, nil
}

func (h *evalCommandHandler) runPluginEvaluator(ctx context.Context, name, plugin string, testCase map[string]interface{}, response string) (EvaluationResult, error) {
	// Handle built-in evaluators like github/similarity, github/coherence, etc.
	if strings.HasPrefix(plugin, "github/") {
		evaluatorName := strings.TrimPrefix(plugin, "github/")
		if builtinEvaluator, exists := BuiltInEvaluators[evaluatorName]; exists {
			return h.runLLMEvaluator(ctx, name, builtinEvaluator, testCase, response)
		}
	}

	return EvaluationResult{
		EvaluatorName: name,
		Score:         0.0,
		Passed:        false,
		Details:       fmt.Sprintf("Plugin evaluator '%s' not found", plugin),
	}, nil
}
