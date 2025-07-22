package generate

/*
// NewPromptPex creates a new PromptPex instance
func NewPromptPex(cmd *cobra.Command, cfg *command.Config, args []string) *generateCommandHandler {
	// Merge with default options
	defaultOptions := GetDefaultOptions()
	mergedOptions := mergeOptions(defaultOptions, options)

	// Create LLM client
	return &PromptPex{
		options: mergedOptions,
		client:  cfg.Client,
		ctx:     context.Background(),
		logger:  log.New(os.Stdout, "[PromptPex] ", log.LstdFlags),
	}
}

// Run executes the PromptPex pipeline
func (h *generateCommandHandler) Run(inputFile string) error {
	h.cfg.WriteToOut("Starting PromptPex with input: %s", inputFile)

	// Load or create context
	var context *PromptPexContext
	var err error

	if p.options.LoadContext != nil && *p.options.LoadContext {
		// Load existing context
		contextFile := "promptpex_context.json"
		if p.options.LoadContextFile != nil {
			contextFile = *p.options.LoadContextFile
		}
		context, err = p.loadContext(contextFile)
		if err != nil {
			return fmt.Errorf("failed to load context: %w", err)
		}
		p.logger.Printf("Loaded context from %s", contextFile)
	} else {
		// Create new context from prompt file
		context, err = p.createContext(inputFile)
		if err != nil {
			return fmt.Errorf("failed to create context: %w", err)
		}
	}

	// Run the PromptPex pipeline
	return p.runPipeline(context)
}


// runPipeline executes the main PromptPex pipeline
func (p *PromptPex) runPipeline(context *PromptPexContext) error {
	p.logger.Printf("Running pipeline for prompt: %s", context.Name)

	// Step 1: Generate Intent
	if err := p.generateIntent(context); err != nil {
		return fmt.Errorf("failed to generate intent: %w", err)
	}

	// Step 2: Generate Input Specification
	if err := p.generateInputSpec(context); err != nil {
		return fmt.Errorf("failed to generate input specification: %w", err)
	}

	// Step 3: Generate Output Rules
	if err := p.generateOutputRules(context); err != nil {
		return fmt.Errorf("failed to generate output rules: %w", err)
	}

	// Step 4: Generate Inverse Output Rules
	if err := p.generateInverseRules(context); err != nil {
		return fmt.Errorf("failed to generate inverse rules: %w", err)
	}

	// Step 5: Generate Tests
	if err := p.generateTests(context); err != nil {
		return fmt.Errorf("failed to generate tests: %w", err)
	}

	// Step 6: Test Expansions (if enabled)
	if p.options.TestExpansions != nil && *p.options.TestExpansions > 0 {
		if err := p.expandTests(context); err != nil {
			return fmt.Errorf("failed to expand tests: %w", err)
		}
	}

	// Step 7: Rate Tests (if enabled)
	if p.options.RateTests != nil && *p.options.RateTests {
		if err := p.rateTests(context); err != nil {
			return fmt.Errorf("failed to rate tests: %w", err)
		}
	}

	// Step 8: Generate Groundtruth (if model specified)
	if p.options.GroundtruthModel != nil {
		if err := p.generateGroundtruth(context); err != nil {
			return fmt.Errorf("failed to generate groundtruth: %w", err)
		}
	}

	// Step 9: Run Tests (if models specified)
	if len(p.options.ModelsUnderTest) > 0 {
		if err := p.runTests(context); err != nil {
			return fmt.Errorf("failed to run tests: %w", err)
		}
	}

	// Step 10: Evaluate Results (if enabled)
	if p.options.Evals != nil && *p.options.Evals && len(p.options.EvalModels) > 0 {
		if err := p.evaluateResults(context); err != nil {
			return fmt.Errorf("failed to evaluate results: %w", err)
		}
	}

	// Step 11: Generate GitHub Models Evals
	if err := p.githubModelsEvalsGenerate(context); err != nil {
		return fmt.Errorf("failed to generate GitHub Models evals: %w", err)
	}

	// Save context
	if err := p.saveContext(context); err != nil {
		return fmt.Errorf("failed to save context: %w", err)
	}

	// Generate summary report
	if err := p.generateSummary(context); err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	p.logger.Printf("Pipeline completed successfully. Results saved to: %s", *context.Dir)
	return nil
}

// generateSummary generates a summary report
func (p *PromptPex) generateSummary(context *PromptPexContext) error {
	p.logger.Printf("Summary: Generated %d tests for prompt '%s'", len(context.PromptPexTests), context.Name)

	summary := map[string]interface{}{
		"name":      context.Name,
		"tests":     len(context.PromptPexTests),
		"outputDir": *context.Dir,
		"runId":     context.RunID,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	data, _ := json.MarshalIndent(summary, "", "  ")
	summaryFile := filepath.Join(*context.Dir, "summary.json")

	if context.WriteResults != nil && *context.WriteResults {
		return os.WriteFile(summaryFile, data, 0644)
	}

	return nil
}

// generateIntent generates the intent of the prompt
func (p *PromptPex) generateIntent(context *PromptPexContext) error {
	p.logger.Println("Generating intent...")

	prompt := fmt.Sprintf(`Analyze the following prompt and describe its intent in 2-3 sentences.

Prompt:
%s

Intent:`, context.Prompt.Content)

	response, err := p.llmClient.ChatCompletion(p.ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini", // GitHub Models compatible model
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: *utils.Float32Ptr(0.0),
	})

	if err != nil {
		return err
	}

	intent := response.Choices[0].Message.Content
	context.Intent.Content = intent

	// Write to file if needed
	if context.WriteResults != nil && *context.WriteResults {
		return os.WriteFile(context.Intent.Filename, []byte(intent), 0644)
	}

	return nil
}

// generateInputSpec generates the input specification
func (p *PromptPex) generateInputSpec(context *PromptPexContext) error {
	p.logger.Println("Generating input specification...")

	prompt := fmt.Sprintf(`Analyze the following prompt and generate a specification for its inputs.
List the expected input parameters, their types, constraints, and examples.

Prompt:
%s

Input Specification:`, context.Prompt.Content)

	response, err := p.llmClient.ChatCompletion(p.ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini", // GitHub Models compatible model
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: *utils.Float32Ptr(0.0),
	})

	if err != nil {
		return err
	}

	inputSpec := response.Choices[0].Message.Content
	context.InputSpec.Content = inputSpec

	// Write to file if needed
	if context.WriteResults != nil && *context.WriteResults {
		return os.WriteFile(context.InputSpec.Filename, []byte(inputSpec), 0644)
	}

	return nil
}

// generateOutputRules generates output rules for the prompt
func (p *PromptPex) generateOutputRules(context *PromptPexContext) error {
	p.logger.Println("Generating output rules...")

	prompt := fmt.Sprintf(`Analyze the following prompt and generate a list of output rules.
These rules should describe what makes a valid output from this prompt.
List each rule on a separate line starting with a number.

Prompt:
%s

Output Rules:`, context.Prompt.Content)

	response, err := p.llmClient.ChatCompletion(p.ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini", // GitHub Models compatible model
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: *utils.Float32Ptr(0.0),
	})

	if err != nil {
		return err
	}

	rules := response.Choices[0].Message.Content
	context.Rules.Content = rules

	// Write to file if needed
	if context.WriteResults != nil && *context.WriteResults {
		return os.WriteFile(context.Rules.Filename, []byte(rules), 0644)
	}

	return nil
}

// generateInverseRules generates inverse rules (what makes an invalid output)
func (p *PromptPex) generateInverseRules(context *PromptPexContext) error {
	p.logger.Println("Generating inverse rules...")

	prompt := fmt.Sprintf(`Based on the following output rules, generate inverse rules that describe what would make an INVALID output.
These should be the opposite or negation of the original rules.

Original Rules:
%s

Inverse Rules:`, context.Rules.Content)

	response, err := p.llmClient.ChatCompletion(p.ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini", // GitHub Models compatible model
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: *utils.Float32Ptr(0.0),
	})

	if err != nil {
		return err
	}

	inverseRules := response.Choices[0].Message.Content
	context.InverseRules.Content = inverseRules

	// Write to file if needed
	if context.WriteResults != nil && *context.WriteResults {
		return os.WriteFile(context.InverseRules.Filename, []byte(inverseRules), 0644)
	}

	return nil
}

// generateTests generates test cases for the prompt
func (p *PromptPex) generateTests(context *PromptPexContext) error {
	p.logger.Println("Generating tests...")

	testsPerRule := 3
	if p.options.TestsPerRule != nil {
		testsPerRule = *p.options.TestsPerRule
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
		context.Intent.Content,
		context.InputSpec.Content,
		context.Rules.Content,
		context.Prompt.Content,
		testsPerRule*3)

	response, err := p.llmClient.ChatCompletion(p.ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini", // GitHub Models compatible model
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: *utils.Float32Ptr(0.3),
	})

	if err != nil {
		return err
	}

	// Parse the JSON response
	content := response.Choices[0].Message.Content
	p.logger.Printf("LLM Response for tests: %s", content)

	tests, err := p.parseTestsFromLLMResponse(content)
	if err != nil {
		return fmt.Errorf("failed to parse test JSON: %w", err)
	}

	context.PromptPexTests = tests

	// Serialize tests to JSON
	testsJSON, err := json.MarshalIndent(tests, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tests: %w", err)
	}
	context.Tests.Content = string(testsJSON)

	// Create test data file
	context.TestData.Content = string(testsJSON)

	// Write to files if needed
	if context.WriteResults != nil && *context.WriteResults {
		if err := os.WriteFile(context.Tests.Filename, testsJSON, 0644); err != nil {
			return err
		}
		return os.WriteFile(context.TestData.Filename, testsJSON, 0644)
	}

	return nil
}

// runTests executes tests against the specified models
func (p *PromptPex) runTests(context *PromptPexContext) error {
	p.logger.Println("Running tests against models...")

	var results []PromptPexTestResult
	runsPerTest := 1
	if p.options.RunsPerTest != nil {
		runsPerTest = *p.options.RunsPerTest
	}

	for _, modelName := range p.options.ModelsUnderTest {
		p.logger.Printf("Running tests with model: %s", modelName)

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
				output, err := p.runSingleTestWithContext(test.TestInput, modelName, context)
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
	context.TestOutputs.Content = string(resultsJSON)

	if context.WriteResults != nil && *context.WriteResults {
		return os.WriteFile(context.TestOutputs.Filename, resultsJSON, 0644)
	}

	return nil
}

// runSingleTest runs a single test against a model
func (p *PromptPex) runSingleTest(input, modelName string) (string, error) {
	return p.runSingleTestWithContext(input, modelName, nil)
}

// runSingleTestWithContext runs a single test against a model with context
func (p *PromptPex) runSingleTestWithContext(input, modelName string, context *PromptPexContext) (string, error) {
	// Use the context if provided, otherwise use the stored context
	var messages []ChatMessage
	if context != nil {
		messages = context.Messages
	} else {
		// Fallback to basic sentiment analysis prompt
		messages = []ChatMessage{
			{Role: "system", Content: "You are a sentiment analysis expert. Classify the sentiment of the given text."},
			{Role: "user", Content: "Classify the sentiment of this text as positive, negative, or neutral: {{text}}\n\nRespond with only the sentiment word."},
		}
	}

	// Build OpenAI messages from our messages format
	var openaiMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		// Replace template variables in content
		content := strings.ReplaceAll(msg.Content, "{{text}}", input)

		// Convert role format
		role := msg.Role
		if role == "A" || role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		} else if role == "system" {
			role = openai.ChatMessageRoleSystem
		} else {
			role = openai.ChatMessageRoleUser
		}

		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    role,
			Content: content,
		})
	}

	response, err := p.llmClient.ChatCompletion(p.ctx, openai.ChatCompletionRequest{
		Model:       "gpt-4o-mini", // GitHub Models compatible model
		Messages:    openaiMessages,
		Temperature: *utils.Float32Ptr(0.0),
	})

	if err != nil {
		return "", err
	}

	return response.Choices[0].Message.Content, nil
}

// evaluateResults evaluates test results using the specified evaluation models
func (p *PromptPex) evaluateResults(context *PromptPexContext) error {
	p.logger.Println("Evaluating test results...")

	// Parse existing test results
	var results []PromptPexTestResult
	if err := json.Unmarshal([]byte(context.TestOutputs.Content), &results); err != nil {
		return fmt.Errorf("failed to parse test results: %w", err)
	}

	// Evaluate each result
	for i := range results {
		if results[i].Error != nil {
			continue // Skip failed tests
		}

		// Evaluate against output rules
		compliance, err := p.evaluateCompliance(results[i].Output, context.Rules.Content)
		if err != nil {
			p.logger.Printf("Failed to evaluate compliance for test %s: %v", results[i].ID, err)
		} else {
			results[i].Compliance = &compliance
		}

		// Add custom metrics evaluation
		if p.options.CustomMetric != nil {
			score, err := p.evaluateCustomMetric(results[i].Output, *p.options.CustomMetric)
			if err != nil {
				p.logger.Printf("Failed to evaluate custom metric for test %s: %v", results[i].ID, err)
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
	context.TestOutputs.Content = string(resultsJSON)

	if context.WriteResults != nil && *context.WriteResults {
		return os.WriteFile(context.TestOutputs.Filename, resultsJSON, 0644)
	}

	return nil
}

// evaluateCompliance evaluates if an output complies with the given rules
func (p *PromptPex) evaluateCompliance(output, rules string) (PromptPexEvalResultType, error) {
	prompt := fmt.Sprintf(`Evaluate if the following output complies with the given rules.
Respond with only one word: "ok" if it complies, "err" if it doesn't, or "unknown" if uncertain.

Rules:
%s

Output to evaluate:
%s

Compliance:`, rules, output)

	response, err := p.llmClient.ChatCompletion(p.ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini", // GitHub Models compatible model
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: *utils.Float32Ptr(0.0),
	})

	if err != nil {
		return EvalResultUnknown, err
	}

	result := strings.ToLower(strings.TrimSpace(response.Choices[0].Message.Content))
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
func (p *PromptPex) evaluateCustomMetric(output, metric string) (float64, error) {
	prompt := fmt.Sprintf(`%s

Output to evaluate:
%s

Score (0-1):`, metric, output)

	response, err := p.llmClient.ChatCompletion(p.ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini", // GitHub Models compatible model
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: *utils.Float32Ptr(0.0),
	})

	if err != nil {
		return 0.0, err
	}

	// Parse the score from the response
	scoreStr := strings.TrimSpace(response.Choices[0].Message.Content)
	var score float64
	if _, err := fmt.Sscanf(scoreStr, "%f", &score); err != nil {
		return 0.0, fmt.Errorf("failed to parse score: %w", err)
	}

	return score, nil
}

// generateGroundtruth generates groundtruth outputs using the specified model
func (p *PromptPex) generateGroundtruth(context *PromptPexContext) error {
	p.logger.Printf("Generating groundtruth with model: %s", *p.options.GroundtruthModel)

	for i := range context.PromptPexTests {
		test := &context.PromptPexTests[i]

		// Generate groundtruth output
		output, err := p.runSingleTestWithContext(test.TestInput, *p.options.GroundtruthModel, context)
		if err != nil {
			p.logger.Printf("Failed to generate groundtruth for test %d: %v", i, err)
			continue
		}

		test.Groundtruth = &output
		test.GroundtruthModel = p.options.GroundtruthModel
	}

	// Update test data
	testData, _ := json.MarshalIndent(context.PromptPexTests, "", "  ")
	context.TestData.Content = string(testData)

	if context.WriteResults != nil && *context.WriteResults {
		return os.WriteFile(context.TestData.Filename, testData, 0644)
	}

	return nil
}

// expandTests implements test expansion functionality
func (p *PromptPex) expandTests(context *PromptPexContext) error {
	p.logger.Printf("Expanding tests with %d expansion phases", *p.options.TestExpansions)

	originalTestCount := len(context.PromptPexTests)

	for phase := 0; phase < *p.options.TestExpansions; phase++ {
		p.logger.Printf("Test expansion phase %d/%d", phase+1, *p.options.TestExpansions)

		var newTests []PromptPexTest

		for _, test := range context.PromptPexTests {
			// Generate expanded versions of each test
			expandedTests, err := p.expandSingleTest(test, context)
			if err != nil {
				p.logger.Printf("Failed to expand test: %v", err)
				continue
			}

			newTests = append(newTests, expandedTests...)
		}

		// Add new tests to the collection
		context.PromptPexTests = append(context.PromptPexTests, newTests...)
	}

	p.logger.Printf("Expanded from %d to %d tests", originalTestCount, len(context.PromptPexTests))

	// Update test data
	testData, _ := json.MarshalIndent(context.PromptPexTests, "", "  ")
	context.TestData.Content = string(testData)

	if context.WriteResults != nil && *context.WriteResults {
		return os.WriteFile(context.TestData.Filename, testData, 0644)
	}

	return nil
}

// expandSingleTest expands a single test into multiple variations
func (p *PromptPex) expandSingleTest(test PromptPexTest, context *PromptPexContext) ([]PromptPexTest, error) {
	prompt := fmt.Sprintf(`Given this test case, generate 2-3 variations that test similar scenarios but with different inputs.
Keep the same scenario type but vary the specific details.

Original test:
Scenario: %s
Input: %s
Reasoning: %s

Generate variations in JSON format as an array of objects with "scenario", "testinput", and "reasoning" fields.`,
		*test.Scenario, test.TestInput, *test.Reasoning)

	response, err := p.llmClient.ChatCompletion(p.ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini", // GitHub Models compatible model
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: *utils.Float32Ptr(0.5),
	})

	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var expandedTests []PromptPexTest
	content := response.Choices[0].Message.Content
	jsonStr := utils.ExtractJSON(content)

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
func (p *PromptPex) rateTests(context *PromptPexContext) error {
	p.logger.Println("Rating test collection quality...")

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

	response, err := p.llmClient.ChatCompletion(p.ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini", // GitHub Models compatible model
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: *utils.Float32Ptr(0.2),
	})

	if err != nil {
		return err
	}

	rating := response.Choices[0].Message.Content
	context.RateTests.Content = rating

	if context.WriteResults != nil && *context.WriteResults {
		return os.WriteFile(context.RateTests.Filename, []byte(rating), 0644)
	}

	return nil
}

// parseTestsFromLLMResponse parses test cases from LLM response with robust error handling
func (p *PromptPex) parseTestsFromLLMResponse(content string) ([]PromptPexTest, error) {
	jsonStr := utils.ExtractJSON(content)

	// First try to parse as our expected structure
	var tests []PromptPexTest
	if err := json.Unmarshal([]byte(jsonStr), &tests); err == nil {
		return tests, nil
	}

	// If that fails, try to parse as a more flexible structure
	var rawTests []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawTests); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	// Convert to our structure
	for _, rawTest := range rawTests {
		test := PromptPexTest{}

		if scenario, ok := rawTest["scenario"].(string); ok {
			test.Scenario = &scenario
		}

		// Handle testinput - can be string or structured object
		if testinput, ok := rawTest["testinput"].(string); ok {
			test.TestInput = testinput
		} else if testinputObj, ok := rawTest["testinput"].(map[string]interface{}); ok {
			// Convert structured object to JSON string
			if jsonBytes, err := json.Marshal(testinputObj); err == nil {
				test.TestInput = string(jsonBytes)
			}
		} else if testInput, ok := rawTest["testInput"].(string); ok {
			test.TestInput = testInput
		} else if testInputObj, ok := rawTest["testInput"].(map[string]interface{}); ok {
			// Convert structured object to JSON string
			if jsonBytes, err := json.Marshal(testInputObj); err == nil {
				test.TestInput = string(jsonBytes)
			}
		} else if input, ok := rawTest["input"].(string); ok {
			test.TestInput = input
		} else if inputObj, ok := rawTest["input"].(map[string]interface{}); ok {
			// Convert structured object to JSON string
			if jsonBytes, err := json.Marshal(inputObj); err == nil {
				test.TestInput = string(jsonBytes)
			}
		}

		if reasoning, ok := rawTest["reasoning"].(string); ok {
			test.Reasoning = &reasoning
		}

		tests = append(tests, test)
	}

	return tests, nil
}
*/
