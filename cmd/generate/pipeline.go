package generate

import (
	"fmt"
	"strings"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/prompt"
	"github.com/github/gh-models/pkg/util"
)

// RunTestGenerationPipeline executes the main PromptPex pipeline
func (h *generateCommandHandler) RunTestGenerationPipeline(context *PromptPexContext) error {
	// Step 1: Generate Intent
	if err := h.generateIntent(context); err != nil {
		return fmt.Errorf("failed to generate intent: %w", err)
	}
	if err := h.SaveContext(context); err != nil {
		return err
	}

	// Step 2: Generate Input Specification
	if err := h.generateInputSpec(context); err != nil {
		return fmt.Errorf("failed to generate input specification: %w", err)
	}
	if err := h.SaveContext(context); err != nil {
		return err
	}

	// Step 3: Generate Output Rules
	if err := h.generateOutputRules(context); err != nil {
		return fmt.Errorf("failed to generate output rules: %w", err)
	}
	if err := h.SaveContext(context); err != nil {
		return err
	}

	// Step 4: Generate Inverse Output Rules
	if err := h.generateInverseRules(context); err != nil {
		return fmt.Errorf("failed to generate inverse rules: %w", err)
	}
	if err := h.SaveContext(context); err != nil {
		return err
	}

	// Step 5: Generate Tests
	if err := h.generateTests(context); err != nil {
		return fmt.Errorf("failed to generate tests: %w", err)
	}
	if err := h.SaveContext(context); err != nil {
		return err
	}

	// Step 8: Generate Groundtruth (if model specified)
	if h.options.Models.Groundtruth != "" && h.options.Models.Groundtruth != "none" {
		if err := h.generateGroundtruth(context); err != nil {
			return fmt.Errorf("failed to generate groundtruth: %w", err)
		}
		if err := h.SaveContext(context); err != nil {
			return err
		}
	}

	// insert test cases in prompt and write back to file
	if err := h.updatePromptFile(context); err != nil {
		return err
	}
	if err := h.SaveContext(context); err != nil {
		return err
	}

	// Generate summary report
	if err := h.generateSummary(context); err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}
	return nil
}

// generateIntent generates the intent of the prompt
func (h *generateCommandHandler) generateIntent(context *PromptPexContext) error {
	h.WriteStartBox("Intent", "")
	if context.Intent == nil || *context.Intent == "" {
		system := `Analyze the following prompt and describe its intent in 2-3 sentences.`
		prompt := fmt.Sprintf(`<prompt>
%s
</prompt>

Intent:`, RenderMessagesToString(context.Prompt.Messages))

		messages := []azuremodels.ChatMessage{
			{Role: azuremodels.ChatMessageRoleSystem, Content: util.Ptr(systemPromptTextOnly)},
			{Role: azuremodels.ChatMessageRoleSystem, Content: util.Ptr(system)},
			{Role: azuremodels.ChatMessageRoleUser, Content: util.Ptr(prompt)},
		}
		options := azuremodels.ChatCompletionOptions{
			Model:       h.options.Models.Rules, // GitHub Models compatible model
			Messages:    messages,
			Temperature: util.Ptr(0.0),
			Stream:      false,
		}
		intent, err := h.callModelWithRetry("intent", options)
		if err != nil {
			return err
		}
		context.Intent = util.Ptr(intent)
	}

	h.WriteToParagraph(*context.Intent)
	h.WriteEndBox("")

	return nil
}

// generateInputSpec generates the input specification
func (h *generateCommandHandler) generateInputSpec(context *PromptPexContext) error {
	h.WriteStartBox("Input Specification", "")
	if context.InputSpec == nil || *context.InputSpec == "" {
		system := `Analyze the following prompt and generate a specification for its inputs.
List the expected input parameters, their types, constraints, and examples.`
		prompt := fmt.Sprintf(`<prompt>
%s
</prompt>

Input Specification:`, RenderMessagesToString(context.Prompt.Messages))

		messages := []azuremodels.ChatMessage{
			{Role: azuremodels.ChatMessageRoleSystem, Content: util.Ptr(systemPromptTextOnly)},
			{Role: azuremodels.ChatMessageRoleSystem, Content: util.Ptr(system)},
			{Role: azuremodels.ChatMessageRoleUser, Content: util.Ptr(prompt)},
		}

		options := azuremodels.ChatCompletionOptions{
			Model:       h.options.Models.Rules,
			Messages:    messages,
			Temperature: util.Ptr(0.0),
		}

		inputSpec, err := h.callModelWithRetry("input spec", options)
		if err != nil {
			return err
		}
		context.InputSpec = util.Ptr(inputSpec)
	}

	h.WriteToParagraph(*context.InputSpec)
	h.WriteEndBox("")

	return nil
}

// generateOutputRules generates output rules for the prompt
func (h *generateCommandHandler) generateOutputRules(context *PromptPexContext) error {
	h.WriteStartBox("Output rules", "")
	if len(context.Rules) == 0 {
		system := `Analyze the following prompt and generate a list of output rules.
These rules should describe what makes a valid output from this prompt.
List each rule on a separate line starting with a number.`
		prompt := fmt.Sprintf(`<prompt>
%s
</prompt>

Output Rules:`, RenderMessagesToString(context.Prompt.Messages))

		messages := []azuremodels.ChatMessage{
			{Role: azuremodels.ChatMessageRoleSystem, Content: util.Ptr(systemPromptTextOnly)},
			{Role: azuremodels.ChatMessageRoleSystem, Content: util.Ptr(system)},
			{Role: azuremodels.ChatMessageRoleUser, Content: util.Ptr(prompt)},
		}

		options := azuremodels.ChatCompletionOptions{
			Model:       h.options.Models.Rules, // GitHub Models compatible model
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
	}

	h.WriteEndListBox(context.Rules, 16)

	return nil
}

// generateInverseRules generates inverse rules (what makes an invalid output)
func (h *generateCommandHandler) generateInverseRules(context *PromptPexContext) error {
	h.WriteStartBox("Inverse output rules", "")
	if len(context.InverseRules) == 0 {

		system := `Based on the following <output_rules>, generate inverse rules that describe what would make an INVALID output.
These should be the opposite or negation of the original rules.`
		prompt := fmt.Sprintf(`<output_rules>
%s
</output_rules>

Inverse Output Rules:`, strings.Join(context.Rules, "\n"))

		messages := []azuremodels.ChatMessage{
			{Role: azuremodels.ChatMessageRoleSystem, Content: util.Ptr(systemPromptTextOnly)},
			{Role: azuremodels.ChatMessageRoleSystem, Content: util.Ptr(system)},
			{Role: azuremodels.ChatMessageRoleUser, Content: util.Ptr(prompt)},
		}

		options := azuremodels.ChatCompletionOptions{
			Model:       h.options.Models.Rules, // GitHub Models compatible model
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
	}

	h.WriteEndListBox(context.InverseRules, 16)
	return nil
}

// generateTests generates test cases for the prompt
func (h *generateCommandHandler) generateTests(context *PromptPexContext) error {
	h.WriteStartBox("Tests", fmt.Sprintf("%d rules x %d tests per rule", len(context.Rules)+len(context.InverseRules), h.options.TestsPerRule))
	if len(context.Tests) == 0 {
		testsPerRule := 3
		if h.options.TestsPerRule != 0 {
			testsPerRule = h.options.TestsPerRule
		}

		allRules := append(context.Rules, context.InverseRules...)

		nTests := testsPerRule * len(context.Rules)
		// Build dynamic prompt based on the actual content (like TypeScript reference)
		system := `Response in JSON format only.`
		prompt := fmt.Sprintf(`Generate %d test cases for the following prompt based on the intent, input specification, and output rules. Generate %d tests per rule.		

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
5. Avoid whitespace only test inputs

Return only a JSON array with this exact format:
[
  {
    "scenario": "Description of what this test validates",
    "testInput": "The actual input text or data",
    "reasoning": "Why this test is important and what it validates"
  }
]

Generate exactly %d diverse test cases:`, nTests,
			testsPerRule,
			*context.Intent,
			*context.InputSpec,
			strings.Join(allRules, "\n"),
			RenderMessagesToString(context.Prompt.Messages),
			nTests)

		messages := []azuremodels.ChatMessage{
			{Role: azuremodels.ChatMessageRoleSystem, Content: util.Ptr(system)},
			{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
		}

		options := azuremodels.ChatCompletionOptions{
			Model:       h.options.Models.Tests, // GitHub Models compatible model
			Messages:    messages,
			Temperature: util.Ptr(0.3),
		}

		tests, err := h.callModelToGenerateTests(options)
		if err != nil {
			return fmt.Errorf("failed to generate tests: %w", err)
		}
		if len(tests) == 0 {
			return fmt.Errorf("no tests generated, please check your prompt and rules")
		}
		context.Tests = tests
	}

	testViews := make([]string, len(context.Tests)*2)
	for i, test := range context.Tests {
		testViews[i*2] = test.TestInput
		testViews[i*2+1] = fmt.Sprintf("    %s%s", BOX_END, test.Reasoning)
	}
	h.WriteEndListBox(testViews, PREVIEW_TEST_COUNT)
	return nil
}

func (h *generateCommandHandler) callModelToGenerateTests(options azuremodels.ChatCompletionOptions) ([]PromptPexTest, error) {
	// try multiple times to generate tests
	const maxGenerateTestRetry = 3
	for i := 0; i < maxGenerateTestRetry; i++ {
		content, err := h.callModelWithRetry("tests", options)
		if err != nil {
			continue
		}
		tests, err := h.ParseTestsFromLLMResponse(content)
		if err != nil {
			continue
		}
		return tests, nil
	}
	// last attempt without retry
	content, err := h.callModelWithRetry("tests", options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tests: %w", err)
	}
	tests, err := h.ParseTestsFromLLMResponse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test JSON: %w", err)
	}
	return tests, nil
}

// runSingleTestWithContext runs a single test against a model with context
func (h *generateCommandHandler) runSingleTestWithContext(input string, modelName string, context *PromptPexContext) (string, error) {
	// Use the context if provided, otherwise use the stored context
	messages := context.Prompt.Messages

	// Build OpenAI messages from our messages format
	openaiMessages := []azuremodels.ChatMessage{}
	for _, msg := range messages {
		templateData := make(map[string]interface{})
		templateData["input"] = input
		// Replace template variables in content
		content, err := prompt.TemplateString(msg.Content, templateData)
		if err != nil {
			return "", fmt.Errorf("failed to render message content: %w", err)
		}

		// Convert role format
		var role azuremodels.ChatMessageRole
		switch msg.Role {
		case "assistant":
			role = azuremodels.ChatMessageRoleAssistant
		case "system":
			role = azuremodels.ChatMessageRoleSystem
		case "user":
			role = azuremodels.ChatMessageRoleUser
		default:
			return "", fmt.Errorf("unknown role: %s", msg.Role)
		}

		// Handle the openaiMessages array indexing properly
		openaiMessages = append(openaiMessages, azuremodels.ChatMessage{
			Role:    role,
			Content: &content,
		})
	}

	options := azuremodels.ChatCompletionOptions{
		Model:       modelName,
		Messages:    openaiMessages,
		Temperature: util.Ptr(0.0),
	}

	result, err := h.callModelWithRetry("tests", options)
	if err != nil {
		return "", fmt.Errorf("failed to run test input: %w", err)
	}

	return result, nil
}

// generateGroundtruth generates groundtruth outputs using the specified model
func (h *generateCommandHandler) generateGroundtruth(context *PromptPexContext) error {
	groundtruthModel := h.options.Models.Groundtruth
	h.WriteStartBox("Groundtruth", fmt.Sprintf("with %s", groundtruthModel))
	for i := range context.Tests {
		test := &context.Tests[i]
		h.WriteToLine(test.TestInput)
		if test.Groundtruth == "" {
			// Generate groundtruth output
			output, err := h.runSingleTestWithContext(test.TestInput, groundtruthModel, context)
			if err != nil {
				h.cfg.WriteToOut(fmt.Sprintf("Failed to generate groundtruth for test %d: %v", i, err))
				continue
			}
			test.Groundtruth = output
			test.GroundtruthModel = groundtruthModel

			h.SaveContext(context) // Save context after generating groundtruth
		}
		h.WriteToLine(fmt.Sprintf("    %s%s", BOX_END, test.Groundtruth)) // Write groundtruth output
	}

	h.WriteEndBox(fmt.Sprintf("%d items", len(context.Tests)))
	return nil
}

// toGitHubModelsPrompt converts PromptPex context to GitHub Models format
func (h *generateCommandHandler) updatePromptFile(context *PromptPexContext) error {
	// Convert test data
	testData := []prompt.TestDataItem{}
	for _, test := range context.Tests {
		item := prompt.TestDataItem{}
		item["input"] = test.TestInput
		if test.Groundtruth != "" {
			item["expected"] = test.Groundtruth
		}
		testData = append(testData, item)
	}
	context.Prompt.TestData = testData

	// Save updated prompt to file
	if err := context.Prompt.SaveToFile(h.promptFile); err != nil {
		return fmt.Errorf("failed to save updated prompt file: %w", err)
	}

	return nil
}
