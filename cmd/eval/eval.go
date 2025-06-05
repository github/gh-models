// Package eval provides a gh command to evaluate prompts against GitHub models.
package eval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/prompt"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/cobra"
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
			  model: gpt-4o
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

			See https://docs.github.com/github-models/use-github-models/storing-prompts-in-github-repositories#supported-file-format for more information.
		`),
		Example: "gh models eval my_prompt.prompt.yml",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			promptFilePath := args[0]
			
			// Get the json flag
			jsonOutput, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

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
			}

			return handler.runEvaluation(cmd.Context())
		},
	}

	cmd.Flags().Bool("json", false, "Output results in JSON format")
	return cmd
}

type evalCommandHandler struct {
	cfg        *command.Config
	client     azuremodels.Client
	evalFile   *prompt.File
	jsonOutput bool
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
	passRate := 0.0
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

	return nil
}

func (h *evalCommandHandler) printTestResult(result TestResult, testPassed bool) {
	if testPassed {
		h.cfg.WriteToOut("  ✓ PASSED\n")
	} else {
		h.cfg.WriteToOut("  ✗ FAILED\n")
		// Show the first 100 characters of the model response when test fails
		preview := result.ModelResponse
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		h.cfg.WriteToOut(fmt.Sprintf("    Model Response: %s\n", preview))
	}

	// Show evaluation details
	for _, evalResult := range result.EvaluationResults {
		status := "✓"
		if !evalResult.Passed {
			status = "✗"
		}
		h.cfg.WriteToOut(fmt.Sprintf("    %s %s (score: %.2f)\n",
			status, evalResult.EvaluatorName, evalResult.Score))
		if evalResult.Details != "" {
			h.cfg.WriteToOut(fmt.Sprintf("      %s\n", evalResult.Details))
		}
	}
	h.cfg.WriteToOut("\n")
}

func (h *evalCommandHandler) printSummary(passedTests, totalTests int, passRate float64) {
	// Summary
	h.cfg.WriteToOut("Evaluation Summary:\n")
	if totalTests == 0 {
		h.cfg.WriteToOut("Passed: 0/0 (0.0%)\n")
	} else {
		h.cfg.WriteToOut(fmt.Sprintf("Passed: %d/%d (%.1f%%)\n",
			passedTests, totalTests, passRate))
	}

	if passedTests == totalTests {
		h.cfg.WriteToOut("🎉 All tests passed!\n")
	} else {
		h.cfg.WriteToOut("❌ Some tests failed.\n")
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

		var role azuremodels.ChatMessageRole
		switch strings.ToLower(msg.Role) {
		case "system":
			role = azuremodels.ChatMessageRoleSystem
		case "user":
			role = azuremodels.ChatMessageRoleUser
		case "assistant":
			role = azuremodels.ChatMessageRoleAssistant
		default:
			return nil, fmt.Errorf("unknown message role: %s", msg.Role)
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

func (h *evalCommandHandler) callModel(ctx context.Context, messages []azuremodels.ChatMessage) (string, error) {
	req := azuremodels.ChatCompletionOptions{
		Messages: messages,
		Model:    h.evalFile.Model,
		Stream:   false,
	}

	// Apply model parameters
	if h.evalFile.ModelParameters.MaxTokens != nil {
		req.MaxTokens = h.evalFile.ModelParameters.MaxTokens
	}
	if h.evalFile.ModelParameters.Temperature != nil {
		req.Temperature = h.evalFile.ModelParameters.Temperature
	}
	if h.evalFile.ModelParameters.TopP != nil {
		req.TopP = h.evalFile.ModelParameters.TopP
	}

	resp, err := h.client.GetChatCompletionStream(ctx, req)
	if err != nil {
		return "", err
	}

	// For non-streaming requests, we should get a single response
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
		return h.runStringEvaluator(evaluator.Name, *evaluator.String, response)
	case evaluator.LLM != nil:
		return h.runLLMEvaluator(ctx, evaluator.Name, *evaluator.LLM, testCase, response)
	case evaluator.Uses != "":
		return h.runPluginEvaluator(ctx, evaluator.Name, evaluator.Uses, testCase, response)
	default:
		return EvaluationResult{}, fmt.Errorf("no evaluation method specified for evaluator %s", evaluator.Name)
	}
}

func (h *evalCommandHandler) runStringEvaluator(name string, eval prompt.StringEvaluator, response string) (EvaluationResult, error) {
	var passed bool
	var details string

	switch {
	case eval.Equals != "":
		passed = response == eval.Equals
		details = fmt.Sprintf("Expected exact match: '%s'", eval.Equals)
	case eval.Contains != "":
		passed = strings.Contains(strings.ToLower(response), strings.ToLower(eval.Contains))
		details = fmt.Sprintf("Expected to contain: '%s'", eval.Contains)
	case eval.StartsWith != "":
		passed = strings.HasPrefix(strings.ToLower(response), strings.ToLower(eval.StartsWith))
		details = fmt.Sprintf("Expected to start with: '%s'", eval.StartsWith)
	case eval.EndsWith != "":
		passed = strings.HasSuffix(strings.ToLower(response), strings.ToLower(eval.EndsWith))
		details = fmt.Sprintf("Expected to end with: '%s'", eval.EndsWith)
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
	// Template the evaluation prompt
	evalData := make(map[string]interface{})
	for k, v := range testCase {
		evalData[k] = v
	}
	evalData["completion"] = response

	promptContent, err := h.templateString(eval.Prompt, evalData)
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("failed to template evaluation prompt: %w", err)
	}

	// Prepare messages for evaluation
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

	// Call the evaluation model
	req := azuremodels.ChatCompletionOptions{
		Messages: messages,
		Model:    eval.ModelID,
		Stream:   false,
	}

	resp, err := h.client.GetChatCompletionStream(ctx, req)
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("failed to call evaluation model: %w", err)
	}

	var evalResponse strings.Builder
	for {
		completion, err := resp.Reader.Read()
		if err != nil {
			if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "EOF") {
				break
			}
			return EvaluationResult{}, err
		}

		for _, choice := range completion.Choices {
			if choice.Delta != nil && choice.Delta.Content != nil {
				evalResponse.WriteString(*choice.Delta.Content)
			}
			if choice.Message != nil && choice.Message.Content != nil {
				evalResponse.WriteString(*choice.Message.Content)
			}
		}
	}

	// Match response to choices
	evalResponseText := strings.TrimSpace(strings.ToLower(evalResponse.String()))
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
