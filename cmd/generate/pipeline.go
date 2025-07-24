package generate

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/prompt"
	"github.com/github/gh-models/pkg/util"
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

	// Step 8: Generate Groundtruth (if model specified)
	if h.options.Models.Groundtruth != nil {
		if err := h.generateGroundtruth(context); err != nil {
			return fmt.Errorf("failed to generate groundtruth: %w", err)
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
		Model:       *h.options.Models.Rules, // GitHub Models compatible model
		Messages:    messages,
		Temperature: util.Ptr(0.0),
		Stream:      false,
	}
	intent, err := h.callModelWithRetry("intent", options)
	if err != nil {
		return err
	}
	context.Intent = intent

	return nil
}

// generateInputSpec generates the input specification
func (h *generateCommandHandler) generateInputSpec(context *PromptPexContext) error {
	h.cfg.WriteToOut("Generating input specification...\n")

	system := `Analyze the following prompt and generate a specification for its inputs.
List the expected input parameters, their types, constraints, and examples.`
	prompt := fmt.Sprintf(`<prompt>
%s
</prompt>

Input Specification:`, RenderMessagesToString(context.Prompt.Messages))

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleSystem, Content: &system},
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       *h.options.Models.Rules,
		Messages:    messages,
		Temperature: util.Ptr(0.0),
	}

	inputSpec, err := h.callModelWithRetry("input spec", options)
	if err != nil {
		return err
	}
	context.InputSpec = inputSpec

	return nil
}

// generateOutputRules generates output rules for the prompt
func (h *generateCommandHandler) generateOutputRules(context *PromptPexContext) error {
	h.cfg.WriteToOut("Generating output rules...\n")

	system := `Analyze the following prompt and generate a list of output rules.
These rules should describe what makes a valid output from this prompt.
List each rule on a separate line starting with a number.`
	prompt := fmt.Sprintf(`<prompt>
%s
</prompt>

Output Rules:`, RenderMessagesToString(context.Prompt.Messages))

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleSystem, Content: &system},
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       *h.options.Models.Rules, // GitHub Models compatible model
		Messages:    messages,
		Temperature: util.Ptr(0.0),
	}

	rules, err := h.callModelWithRetry("output rules", options)
	if err != nil {
		return err
	}

	parsed := ParseRules(rules)
	if parsed == nil {
		return fmt.Errorf("failed to parse output rules: %s", rules)
	}

	context.Rules = parsed

	return nil
}

// generateInverseRules generates inverse rules (what makes an invalid output)
func (h *generateCommandHandler) generateInverseRules(context *PromptPexContext) error {
	h.cfg.WriteToOut("Generating inverse rules...\n")

	system := `Based on the following <output_rules>, generate inverse rules that describe what would make an INVALID output.
These should be the opposite or negation of the original rules.`
	prompt := fmt.Sprintf(`

<output_rules>
%s
</output_rules>

Inverse Rules:`, context.Rules)

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleSystem, Content: &system},
		{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       *h.options.Models.Rules, // GitHub Models compatible model
		Messages:    messages,
		Temperature: util.Ptr(0.0),
	}

	inverseRules, err := h.callModelWithRetry("inverse output rules", options)
	if err != nil {
		return err
	}

	parsed := ParseRules(inverseRules)
	if parsed == nil {
		return fmt.Errorf("failed to parse inverse output rules: %s", inverseRules)
	}
	context.InverseRules = parsed

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

<intent>
%s
</intent>

<input_specification>
%s
</input_specification>

<output_rules>
%s
</output_rules>

<prompt>
%s
</prompt>

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
		Model:       *h.options.Models.Tests, // GitHub Models compatible model
		Messages:    messages,
		Temperature: util.Ptr(0.3),
	}

	content, err := h.callModelWithRetry("tests", options)
	if err != nil {
		return fmt.Errorf("failed to generate tests: %w", err)
	}
	tests, err := h.ParseTestsFromLLMResponse(content)
	if err != nil {
		return fmt.Errorf("failed to parse test JSON: %w", err)
	}
	context.Tests = tests

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
		Temperature: util.Ptr(0.0),
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

// generateGroundtruth generates groundtruth outputs using the specified model
func (h *generateCommandHandler) generateGroundtruth(context *PromptPexContext) error {
	groundtruthModel := h.options.Models.Groundtruth
	h.cfg.WriteToOut(fmt.Sprintf("Generating groundtruth with model: %s", *groundtruthModel))

	for i := range context.Tests {
		test := &context.Tests[i]

		// Generate groundtruth output
		output, err := h.runSingleTestWithContext(test.TestInput, *groundtruthModel, context)
		if err != nil {
			h.cfg.WriteToOut(fmt.Sprintf("Failed to generate groundtruth for test %d: %v", i, err))
			continue
		}

		test.Groundtruth = &output
		test.GroundtruthModel = groundtruthModel
	}

	return nil
}

// expandTests implements test expansion functionality
func (h *generateCommandHandler) expandTests(context *PromptPexContext) error {
	h.cfg.WriteToOut(fmt.Sprintf("Expanding tests with %d expansion phases", *h.options.TestExpansions))

	originalTestCount := len(context.Tests)

	for phase := 0; phase < *h.options.TestExpansions; phase++ {
		h.cfg.WriteToOut(fmt.Sprintf("Test expansion phase %d/%d", phase+1, *h.options.TestExpansions))

		var newTests []PromptPexTest

		for _, test := range context.Tests {
			// Generate expanded versions of each test
			expandedTests, err := h.expandSingleTest(test, context)
			if err != nil {
				h.cfg.WriteToOut(fmt.Sprintf("Failed to expand test: %v", err))
				continue
			}

			newTests = append(newTests, expandedTests...)
		}

		// Add new tests to the collection
		context.Tests = append(context.Tests, newTests...)
	}

	h.cfg.WriteToOut(fmt.Sprintf("Expanded from %d to %d tests", originalTestCount, len(context.Tests)))

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
		Temperature: util.Ptr(0.5),
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
			expandedTests[i].Generation = util.Ptr(*test.Generation + 1)
		} else {
			expandedTests[i].Generation = util.Ptr(1)
		}
	}

	return expandedTests, nil
}
