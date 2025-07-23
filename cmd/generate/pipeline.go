package generate

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/prompt"
)

// RunTestGenerationPipeline executes the main PromptPex pipeline
func (h *generateCommandHandler) RunTestGenerationPipeline(context *PromptPexContext) error {
	h.cfg.WriteToOut(fmt.Sprintf("Generating tests for '%s'\n", context.Prompt.Name))

	// Step 1: Generate Intent
	if err := h.generateIntent(context); err != nil {
		return fmt.Errorf("failed to generate intent: %w", err)
	}

	// Step 2: Generate Input Specification
	if err := h.generateInputSpec(context); err != nil {
		return fmt.Errorf("failed to generate input specification: %w", err)
	}

	// Step 3: Generate Output Rules
	if err := h.generateOutputRules(context); err != nil {
		return fmt.Errorf("failed to generate output rules: %w", err)
	}

	// Step 4: Generate Inverse Output Rules
	if err := h.generateInverseRules(context); err != nil {
		return fmt.Errorf("failed to generate inverse rules: %w", err)
	}

	// Step 5: Generate Tests
	if err := h.generateTests(context); err != nil {
		return fmt.Errorf("failed to generate tests: %w", err)
	}

	// Step 6: Test Expansions (if enabled)
	if h.options.TestExpansions != nil && *h.options.TestExpansions > 0 {
		if err := h.expandTests(context); err != nil {
			return fmt.Errorf("failed to expand tests: %w", err)
		}
	}

	// Step 7: Rate Tests (if enabled)
	if h.options.RateTests != nil && *h.options.RateTests {
		if err := h.rateTests(context); err != nil {
			return fmt.Errorf("failed to rate tests: %w", err)
		}
	}

	// Step 8: Generate Groundtruth (if model specified)
	if h.options.GroundtruthModel != nil {
		if err := h.generateGroundtruth(context); err != nil {
			return fmt.Errorf("failed to generate groundtruth: %w", err)
		}
	}

	// Step 9: Run Tests (if models specified)
	if len(h.options.ModelsUnderTest) > 0 {
		if err := h.runTests(context); err != nil {
			return fmt.Errorf("failed to run tests: %w", err)
		}
	}

	// Step 10: Evaluate Results (if enabled)
	if h.options.Evals != nil && *h.options.Evals && len(h.options.EvalModels) > 0 {
		if err := h.evaluateResults(context); err != nil {
			return fmt.Errorf("failed to evaluate results: %w", err)
		}
	}

	// Step 11: Generate GitHub Models Evals
	// TODO
	//if err := h.githubModelsEvalsGenerate(context); err != nil {
	//	return fmt.Errorf("failed to generate GitHub Models evals: %w", err)
	//}

	// Generate summary report
	if err := h.GenerateSummary(context); err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	h.cfg.WriteToOut("Pipeline completed successfully.")
	return nil
}

// extractContentFromCompletion safely extracts content from a completion response
func (h *generateCommandHandler) extractContentFromCompletion(completion azuremodels.ChatCompletion) (string, error) {
	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned from model")
	}
	if completion.Choices[0].Message == nil || completion.Choices[0].Message.Content == nil {
		return "", fmt.Errorf("no content in completion response")
	}
	return *completion.Choices[0].Message.Content, nil
}

// generateIntent generates the intent of the prompt
func (h *generateCommandHandler) generateIntent(context *PromptPexContext) error {
	h.cfg.WriteToOut("Generating intent...\n")

	system := `Analyze the following prompt and describe its intent in 2-3 sentences.`
	prompt := fmt.Sprintf(`<prompt>
%s
</prompt>

Intent:`, RenderMessagesToString(context.Prompt.Messages))

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleSystem, Content: &system},
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       "openai/gpt-4o", // GitHub Models compatible model
		Messages:    messages,
		Temperature: Float64Ptr(0.0),
	}

	h.logLLMRequest("intent", options, messages)

	response, err := h.client.GetChatCompletionStream(h.ctx, options, h.org)
	if err != nil {
		return err
	}
	completion, err := response.Reader.Read()
	if err != nil {
		return err
	}
	intent, err := h.extractContentFromCompletion(completion)
	if err != nil {
		return err
	}

	h.logLLMResponse(intent)

	context.Intent = intent

	return nil
}

// generateInputSpec generates the input specification
func (h *generateCommandHandler) generateInputSpec(context *PromptPexContext) error {
	h.cfg.WriteToOut("Generating input specification...\n")

	prompt := fmt.Sprintf(`Analyze the following prompt and generate a specification for its inputs.
List the expected input parameters, their types, constraints, and examples.

Prompt:
%s

Input Specification:`, RenderMessagesToString(context.Prompt.Messages))

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       "openai/gpt-4o-mini", // GitHub Models compatible model
		Messages:    messages,
		Temperature: Float64Ptr(0.0),
	}

	h.logLLMRequest("input spec", options, messages)

	response, err := h.client.GetChatCompletionStream(h.ctx, options, h.org)
	if err != nil {
		return err
	}
	completion, err := response.Reader.Read()
	if err != nil {
		return err
	}
	inputSpec, err := h.extractContentFromCompletion(completion)
	if err != nil {
		return err
	}

	h.logLLMResponse(inputSpec)

	context.InputSpec = inputSpec

	return nil
}

// generateOutputRules generates output rules for the prompt
func (h *generateCommandHandler) generateOutputRules(context *PromptPexContext) error {
	h.cfg.WriteToOut("Generating output rules...\n")

	prompt := fmt.Sprintf(`Analyze the following prompt and generate a list of output rules.
These rules should describe what makes a valid output from this prompt.
List each rule on a separate line starting with a number.

Prompt:
%s

Output Rules:`, RenderMessagesToString(context.Prompt.Messages))

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       "openai/gpt-4o-mini", // GitHub Models compatible model
		Messages:    messages,
		Temperature: Float64Ptr(0.0),
	}

	h.logLLMRequest("output rules", options, messages)

	response, err := h.client.GetChatCompletionStream(h.ctx, options, h.org)
	if err != nil {
		return err
	}
	completion, err := response.Reader.Read()
	if err != nil {
		return err
	}
	rules, err := h.extractContentFromCompletion(completion)
	if err != nil {
		return err
	}

	h.logLLMResponse(rules)

	context.Rules = rules

	return nil
}

// generateInverseRules generates inverse rules (what makes an invalid output)
func (h *generateCommandHandler) generateInverseRules(context *PromptPexContext) error {
	h.cfg.WriteToOut("Generating inverse rules...\n")

	prompt := fmt.Sprintf(`Based on the following output rules, generate inverse rules that describe what would make an INVALID output.
These should be the opposite or negation of the original rules.

Original Rules:
%s

Inverse Rules:`, context.Rules)

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       "openai/gpt-4o-mini", // GitHub Models compatible model
		Messages:    messages,
		Temperature: Float64Ptr(0.0),
	}

	h.logLLMRequest("inverse rules", options, messages)

	response, err := h.client.GetChatCompletionStream(h.ctx, options, h.org)

	if err != nil {
		return err
	}
	completion, err := response.Reader.Read()
	if err != nil {
		return err
	}
	inverseRules, err := h.extractContentFromCompletion(completion)
	if err != nil {
		return err
	}

	h.logLLMResponse(inverseRules)

	context.InverseRules = inverseRules

	return nil
}

// generateTests generates test cases for the prompt
func (h *generateCommandHandler) generateTests(context *PromptPexContext) error {
	h.cfg.WriteToOut("Generating tests...\n")

	testsPerRule := 3
	if h.options.TestsPerRule != nil {
		testsPerRule = *h.options.TestsPerRule
	}

	// Build dynamic prompt based on the actual content (like TypeScript reference)
	prompt := fmt.Sprintf(`Generate %d test cases for the following prompt based on the intent, input specification, and output rules.

INTENT:
%s

INPUT SPECIFICATION:
%s

OUTPUT RULES:
%s

PROMPT:
%s

Generate test cases that:
1. Test the core functionality described in the intent
2. Cover edge cases and boundary conditions
3. Validate that outputs follow the specified rules
4. Use realistic inputs that match the input specification

Return only a JSON array with this exact format:
[
  {
    "scenario": "Description of what this test validates",
    "testinput": "The actual input text or data",
    "reasoning": "Why this test is important and what it validates"
  }
]

Generate exactly %d diverse test cases:`, testsPerRule*3,
		context.Intent,
		context.InputSpec,
		context.Rules,
		RenderMessagesToString(context.Prompt.Messages),
		testsPerRule*3)

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       "openai/gpt-4o-mini", // GitHub Models compatible model
		Messages:    messages,
		Temperature: Float64Ptr(0.3),
	}

	h.logLLMRequest("tests", options, messages)

	response, err := h.client.GetChatCompletionStream(h.ctx, options, h.org)

	if err != nil {
		return err
	}

	// Parse the JSON response
	completion, err := response.Reader.Read()
	if err != nil {
		return err
	}
	content := *completion.Choices[0].Message.Content

	h.logLLMResponse(content)

	h.cfg.WriteToOut(fmt.Sprintf("LLM Response for tests: %s", content))

	tests, err := h.ParseTestsFromLLMResponse(content)
	if err != nil {
		return fmt.Errorf("failed to parse test JSON: %w", err)
	}

	context.PromptPexTests = tests

	// Serialize tests to JSON
	testsJSON, err := json.MarshalIndent(tests, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tests: %w", err)
	}
	context.Tests = string(testsJSON)

	// Create test data file
	context.TestData = string(testsJSON)

	return nil
}

// runTests executes tests against the specified models
func (h *generateCommandHandler) runTests(context *PromptPexContext) error {
	h.cfg.WriteToOut("Running tests against models...\n")

	var results []PromptPexTestResult
	runsPerTest := 1
	if h.options.RunsPerTest != nil {
		runsPerTest = *h.options.RunsPerTest
	}

	for _, modelName := range h.options.ModelsUnderTest {
		h.cfg.WriteToOut(fmt.Sprintf("Running tests with model: %s", modelName))

		for i, test := range context.PromptPexTests {
			for run := 0; run < runsPerTest; run++ {
				result := PromptPexTestResult{
					ID:        fmt.Sprintf("test_%d_run_%d_%s", i, run, modelName),
					PromptID:  context.RunID,
					RuleID:    i,
					Rule:      fmt.Sprintf("Rule %d", i),
					Scenario:  *test.Scenario,
					TestInput: test.TestInput,
					Model:     modelName,
					Input:     test.TestInput,
					Metrics:   make(map[string]PromptPexEvaluation),
				}

				// Run the test by sending the input to the model
				output, err := h.runSingleTestWithContext(test.TestInput, modelName, context)
				if err != nil {
					errStr := err.Error()
					result.Error = &errStr
					result.Output = ""
				} else {
					result.Output = output
				}

				results = append(results, result)
			}
		}
	}

	// Save results
	resultsJSON, _ := json.MarshalIndent(results, "", "  ")
	context.TestOutputs = string(resultsJSON)

	return nil
}

// runSingleTestWithContext runs a single test against a model with context
func (h *generateCommandHandler) runSingleTestWithContext(input, modelName string, context *PromptPexContext) (string, error) {
	// Use the context if provided, otherwise use the stored context
	var messages []prompt.Message
	if context != nil {
		messages = context.Prompt.Messages
	} else {
		// Fallback to basic sentiment analysis prompt
		systemContent := "You are a sentiment analysis expert. Classify the sentiment of the given text."
		userContent := "Classify the sentiment of this text as positive, negative, or neutral: {{text}}\n\nRespond with only the sentiment word."
		messages = []prompt.Message{
			{Role: "system", Content: systemContent},
			{Role: "user", Content: userContent},
		}
	}

	// Build OpenAI messages from our messages format
	var openaiMessages []azuremodels.ChatMessage
	for _, msg := range messages {
		// Replace template variables in content
		var content string
		if msg.Content != "" {
			content = strings.ReplaceAll(msg.Content, "{{text}}", input)
		}

		// Convert role format
		var role azuremodels.ChatMessageRole
		if msg.Role == "A" || msg.Role == "assistant" {
			role = azuremodels.ChatMessageRoleAssistant
		} else if msg.Role == "system" {
			role = azuremodels.ChatMessageRoleSystem
		} else {
			role = azuremodels.ChatMessageRoleUser
		}

		openaiMessages = append(openaiMessages, azuremodels.ChatMessage{
			Role:    role,
			Content: &content,
		})
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       "openai/gpt-4o-mini", // GitHub Models compatible model
		Messages:    openaiMessages,
		Temperature: Float64Ptr(0.0),
	}

	response, err := h.client.GetChatCompletionStream(h.ctx, options, h.org)
	if err != nil {
		return "", err
	}
	completion, err := response.Reader.Read()
	if err != nil {
		return "", err
	}
	result := *completion.Choices[0].Message.Content

	return result, nil
}

// evaluateResults evaluates test results using the specified evaluation models
func (h *generateCommandHandler) evaluateResults(context *PromptPexContext) error {
	h.cfg.WriteToOut("Evaluating test results...\n")

	// Parse existing test results
	var results []PromptPexTestResult
	if err := json.Unmarshal([]byte(context.TestOutputs), &results); err != nil {
		return fmt.Errorf("failed to parse test results: %w", err)
	}

	// Evaluate each result
	for i := range results {
		if results[i].Error != nil {
			continue // Skip failed tests
		}

		// Evaluate against output rules
		compliance, err := h.evaluateCompliance(results[i].Output, context.Rules)
		if err != nil {
			h.cfg.WriteToOut(fmt.Sprintf("Failed to evaluate compliance for test %s: %v", results[i].ID, err))
		} else {
			results[i].Compliance = &compliance
		}

		// Add custom metrics evaluation
		if h.options.CustomMetric != nil {
			score, err := h.evaluateCustomMetric(results[i].Output, *h.options.CustomMetric)
			if err != nil {
				h.cfg.WriteToOut(fmt.Sprintf("Failed to evaluate custom metric for test %s: %v", results[i].ID, err))
			} else {
				results[i].Metrics["custom"] = PromptPexEvaluation{
					Content: "Custom metric evaluation",
					Score:   &score,
				}
			}
		}
	}

	// Save updated results
	resultsJSON, _ := json.MarshalIndent(results, "", "  ")
	context.TestOutputs = string(resultsJSON)

	return nil
}

// evaluateCompliance evaluates if an output complies with the given rules
func (h *generateCommandHandler) evaluateCompliance(output, rules string) (PromptPexEvalResultType, error) {
	prompt := fmt.Sprintf(`Evaluate if the following output complies with the given rules.
Respond with only one word: "ok" if it complies, "err" if it doesn't, or "unknown" if uncertain.

Rules:
%s

Output to evaluate:
%s

Compliance:`, rules, output)

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       "openai/gpt-4o-mini", // GitHub Models compatible model
		Messages:    messages,
		Temperature: Float64Ptr(0.0),
	}

	response, err := h.client.GetChatCompletionStream(h.ctx, options, h.org)

	if err != nil {
		return EvalResultUnknown, err
	}

	completion, err := response.Reader.Read()
	if err != nil {
		return EvalResultUnknown, err
	}
	result := strings.ToLower(strings.TrimSpace(*completion.Choices[0].Message.Content))

	switch result {
	case "ok":
		return EvalResultOK, nil
	case "err":
		return EvalResultError, nil
	default:
		return EvalResultUnknown, nil
	}
}

// evaluateCustomMetric evaluates output using a custom metric
func (h *generateCommandHandler) evaluateCustomMetric(output, metric string) (float64, error) {
	prompt := fmt.Sprintf(`%s

Output to evaluate:
%s

Score (0-1):`, metric, output)

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       "openai/gpt-4o-mini", // GitHub Models compatible model
		Messages:    messages,
		Temperature: Float64Ptr(0.0),
	}

	response, err := h.client.GetChatCompletionStream(h.ctx, options, h.org)

	if err != nil {
		return 0.0, err
	}

	completion, err := response.Reader.Read()
	if err != nil {
		return 0.0, err
	}

	// Parse the score from the response
	scoreStr := strings.TrimSpace(*completion.Choices[0].Message.Content)

	var score float64
	if _, err := fmt.Sscanf(scoreStr, "%f", &score); err != nil {
		return 0.0, fmt.Errorf("failed to parse score: %w", err)
	}

	return score, nil
}

// generateGroundtruth generates groundtruth outputs using the specified model
func (h *generateCommandHandler) generateGroundtruth(context *PromptPexContext) error {
	h.cfg.WriteToOut(fmt.Sprintf("Generating groundtruth with model: %s", *h.options.GroundtruthModel))

	for i := range context.PromptPexTests {
		test := &context.PromptPexTests[i]

		// Generate groundtruth output
		output, err := h.runSingleTestWithContext(test.TestInput, *h.options.GroundtruthModel, context)
		if err != nil {
			h.cfg.WriteToOut(fmt.Sprintf("Failed to generate groundtruth for test %d: %v", i, err))
			continue
		}

		test.Groundtruth = &output
		test.GroundtruthModel = h.options.GroundtruthModel
	}

	// Update test data
	testData, _ := json.MarshalIndent(context.PromptPexTests, "", "  ")
	context.TestData = string(testData)

	return nil
}

// expandTests implements test expansion functionality
func (h *generateCommandHandler) expandTests(context *PromptPexContext) error {
	h.cfg.WriteToOut(fmt.Sprintf("Expanding tests with %d expansion phases", *h.options.TestExpansions))

	originalTestCount := len(context.PromptPexTests)

	for phase := 0; phase < *h.options.TestExpansions; phase++ {
		h.cfg.WriteToOut(fmt.Sprintf("Test expansion phase %d/%d", phase+1, *h.options.TestExpansions))

		var newTests []PromptPexTest

		for _, test := range context.PromptPexTests {
			// Generate expanded versions of each test
			expandedTests, err := h.expandSingleTest(test, context)
			if err != nil {
				h.cfg.WriteToOut(fmt.Sprintf("Failed to expand test: %v", err))
				continue
			}

			newTests = append(newTests, expandedTests...)
		}

		// Add new tests to the collection
		context.PromptPexTests = append(context.PromptPexTests, newTests...)
	}

	h.cfg.WriteToOut(fmt.Sprintf("Expanded from %d to %d tests", originalTestCount, len(context.PromptPexTests)))

	// Update test data
	testData, _ := json.MarshalIndent(context.PromptPexTests, "", "  ")
	context.TestData = string(testData)

	return nil
}

// expandSingleTest expands a single test into multiple variations
func (h *generateCommandHandler) expandSingleTest(test PromptPexTest, context *PromptPexContext) ([]PromptPexTest, error) {
	prompt := fmt.Sprintf(`Given this test case, generate 2-3 variations that test similar scenarios but with different inputs.
Keep the same scenario type but vary the specific details.

Original test:
Scenario: %s
Input: %s
Reasoning: %s

Generate variations in JSON format as an array of objects with "scenario", "testinput", and "reasoning" fields.`,
		*test.Scenario, test.TestInput, *test.Reasoning)

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       "openai/gpt-4o-mini", // GitHub Models compatible model
		Messages:    messages,
		Temperature: Float64Ptr(0.5),
	}

	response, err := h.client.GetChatCompletionStream(h.ctx, options, h.org)

	if err != nil {
		return nil, err
	}

	completion, err := response.Reader.Read()
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var expandedTests []PromptPexTest
	content := *completion.Choices[0].Message.Content
	jsonStr := ExtractJSON(content)

	if err := json.Unmarshal([]byte(jsonStr), &expandedTests); err != nil {
		return nil, fmt.Errorf("failed to parse expanded tests JSON: %w", err)
	}

	// Set the original test input for tracking
	for i := range expandedTests {
		expandedTests[i].TestInputOriginal = &test.TestInput
		if test.Generation != nil {
			expandedTests[i].Generation = IntPtr(*test.Generation + 1)
		} else {
			expandedTests[i].Generation = IntPtr(1)
		}
	}

	return expandedTests, nil
}

// rateTests generates a quality assessment of the test collection
func (h *generateCommandHandler) rateTests(context *PromptPexContext) error {
	h.cfg.WriteToOut("Rating test collection quality...\n")

	testSummary := make([]string, len(context.PromptPexTests))
	for i, test := range context.PromptPexTests {
		testSummary[i] = fmt.Sprintf("Test %d: %s - %s", i+1, *test.Scenario, test.TestInput)
	}

	prompt := fmt.Sprintf(`Analyze the following collection of test cases and provide a quality assessment.
Rate the overall test coverage, diversity, and effectiveness on a scale of 1-10.
Identify any gaps or areas for improvement.

Test Collection:
%s

Analysis:`, strings.Join(testSummary, "\n"))

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       "openai/gpt-4o-mini", // GitHub Models compatible model
		Messages:    messages,
		Temperature: Float64Ptr(0.2),
	}

	response, err := h.client.GetChatCompletionStream(h.ctx, options, h.org)

	if err != nil {
		return err
	}

	completion, err := response.Reader.Read()
	if err != nil {
		return err
	}

	rating := *completion.Choices[0].Message.Content

	context.RateTests = rating

	return nil
}
