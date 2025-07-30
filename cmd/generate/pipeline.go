package generate

import (
	"fmt"
	"slices"
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
		}

		// Add custom instruction if provided
		if h.options.Instructions != nil && h.options.Instructions.Intent != "" {
			messages = append(messages, azuremodels.ChatMessage{
				Role:    azuremodels.ChatMessageRoleSystem,
				Content: util.Ptr(h.options.Instructions.Intent),
			})
		}

		messages = append(messages,
			azuremodels.ChatMessage{Role: azuremodels.ChatMessageRoleUser, Content: util.Ptr(prompt)},
		)

		options := azuremodels.ChatCompletionOptions{
			Model:       h.options.Models.Rules, // GitHub Models compatible model
			Messages:    messages,
			Temperature: util.Ptr(0.0),
			Stream:      false,
			MaxTokens:   util.Ptr(h.options.IntentMaxTokens),
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
		}

		// Add custom instruction if provided
		if h.options.Instructions != nil && h.options.Instructions.InputSpec != "" {
			messages = append(messages, azuremodels.ChatMessage{
				Role:    azuremodels.ChatMessageRoleSystem,
				Content: util.Ptr(h.options.Instructions.InputSpec),
			})
		}

		messages = append(messages,
			azuremodels.ChatMessage{Role: azuremodels.ChatMessageRoleUser, Content: util.Ptr(prompt)},
		)

		options := azuremodels.ChatCompletionOptions{
			Model:       h.options.Models.Rules,
			Messages:    messages,
			Temperature: util.Ptr(0.0),
			MaxTokens:   util.Ptr(h.options.InputSpecMaxTokens),
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
		}

		// Add custom instruction if provided
		if h.options.Instructions != nil && h.options.Instructions.OutputRules != "" {
			messages = append(messages, azuremodels.ChatMessage{
				Role:    azuremodels.ChatMessageRoleSystem,
				Content: util.Ptr(h.options.Instructions.OutputRules),
			})
		}

		messages = append(messages,
			azuremodels.ChatMessage{Role: azuremodels.ChatMessageRoleUser, Content: util.Ptr(prompt)},
		)

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
		}

		// Add custom instruction if provided
		if h.options.Instructions != nil && h.options.Instructions.InverseOutputRules != "" {
			messages = append(messages, azuremodels.ChatMessage{
				Role:    azuremodels.ChatMessageRoleSystem,
				Content: util.Ptr(h.options.Instructions.InverseOutputRules),
			})
		}

		messages = append(messages,
			azuremodels.ChatMessage{Role: azuremodels.ChatMessageRoleUser, Content: util.Ptr(prompt)},
		)

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

		// Generate tests iteratively for groups of rules
		var allTests []PromptPexTest

		rulesPerGen := h.options.RulesPerGen
		// Split rules into groups
		for start := 0; start < len(allRules); start += rulesPerGen {
			end := start + rulesPerGen
			if end > len(allRules) {
				end = len(allRules)
			}
			ruleGroup := allRules[start:end]

			// Generate tests for this group of rules
			groupTests, err := h.generateTestsForRuleGroup(context, ruleGroup, testsPerRule, allTests)
			if err != nil {
				return fmt.Errorf("failed to generate tests for rule group: %w", err)
			}

			// render to terminal
			for _, test := range groupTests {
				h.WriteToLine(test.Input)
				h.WriteToLine(fmt.Sprintf("    %s%s", BOX_END, test.Reasoning))
			}

			// Accumulate tests
			allTests = append(allTests, groupTests...)
		}

		if len(allTests) == 0 {
			return fmt.Errorf("no tests generated, please check your prompt and rules")
		}
		context.Tests = allTests
	}

	h.WriteEndBox(fmt.Sprintf("%d tests", len(context.Tests)))
	return nil
}

// generateTestsForRuleGroup generates test cases for a specific group of rules
func (h *generateCommandHandler) generateTestsForRuleGroup(context *PromptPexContext, ruleGroup []string, testsPerRule int, existingTests []PromptPexTest) ([]PromptPexTest, error) {
	nTests := testsPerRule * len(ruleGroup)

	// Build the prompt for this rule group
	system := `Response in JSON format only.`

	// Build existing tests context if there are any
	existingTestsContext := ""
	if len(existingTests) > 0 {
		var testInputs []string
		for _, test := range existingTests {
			testInputs = append(testInputs, fmt.Sprintf("- %s", test.Input))
		}
		existingTestsContext = fmt.Sprintf(`

The following <existing_tests> inputs have already been generated. Avoid creating duplicates:
<existing_tests>
%s
</existing_tests>`, strings.Join(testInputs, "\n"))
	}

	prompt := fmt.Sprintf(`Generate %d test cases for the following prompt based on the intent, input specification, and output rules. Generate %d tests per rule.%s

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
6. Ensure diversity and avoid duplicating existing test inputs

Return only a JSON array with this exact format:
[
  {
    "scenario": "Description of what this test validates",
    "reasoning": "Why this test is important and what it validates",
    "input": "The actual input text or data"
  }
]

Generate exactly %d diverse test cases:`, nTests,
		testsPerRule,
		existingTestsContext,
		*context.Intent,
		*context.InputSpec,
		strings.Join(ruleGroup, "\n"),
		RenderMessagesToString(context.Prompt.Messages),
		nTests)

	messages := []azuremodels.ChatMessage{
		{Role: azuremodels.ChatMessageRoleSystem, Content: util.Ptr(system)},
	}

	// Add custom instruction if provided
	if h.options.Instructions != nil && h.options.Instructions.Tests != "" {
		messages = append(messages, azuremodels.ChatMessage{
			Role:    azuremodels.ChatMessageRoleSystem,
			Content: util.Ptr(h.options.Instructions.Tests),
		})
	}

	messages = append(messages,
		azuremodels.ChatMessage{Role: azuremodels.ChatMessageRoleUser, Content: &prompt},
	)

	options := azuremodels.ChatCompletionOptions{
		Model:       h.options.Models.Tests, // GitHub Models compatible model
		Messages:    messages,
		Temperature: util.Ptr(0.3),
	}

	tests, err := h.callModelToGenerateTests(options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tests for rule group: %w", err)
	}

	return tests, nil
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
		
		// Add the input variable (backward compatibility)
		templateData["input"] = input
		
		// Add custom variables
		for key, value := range h.templateVars {
			templateData[key] = value
		}
		
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
		h.WriteToLine(test.Input)
		if test.Expected == "" {
			// Generate groundtruth output
			output, err := h.runSingleTestWithContext(test.Input, groundtruthModel, context)
			if err != nil {
				h.cfg.WriteToOut(fmt.Sprintf("Failed to generate groundtruth for test %d: %v", i, err))
				continue
			}
			test.Expected = output

			if err := h.SaveContext(context); err != nil {
				// keep going even if saving fails
				h.cfg.WriteToOut(fmt.Sprintf("Saving context failed: %v", err))
			}
		}
		h.WriteToLine(fmt.Sprintf("    %s%s", BOX_END, test.Expected)) // Write groundtruth output
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
		item["input"] = test.Input
		if test.Expected != "" {
			item["expected"] = test.Expected
		}
		testData = append(testData, item)
	}
	context.Prompt.TestData = testData

	// insert output rule evaluator
	if context.Prompt.Evaluators == nil {
		context.Prompt.Evaluators = make([]prompt.Evaluator, 0)
	}
	evaluator := h.GenerateRulesEvaluator(context)
	context.Prompt.Evaluators = slices.DeleteFunc(context.Prompt.Evaluators, func(e prompt.Evaluator) bool {
		return e.Name == evaluator.Name
	})
	context.Prompt.Evaluators = append(context.Prompt.Evaluators, evaluator)

	// Save updated prompt to file
	if err := context.Prompt.SaveToFile(h.promptFile); err != nil {
		return fmt.Errorf("failed to save updated prompt file: %w", err)
	}

	return nil
}
