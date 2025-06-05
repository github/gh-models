package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/internal/sse"
	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/prompt"
	"github.com/stretchr/testify/require"
)

func TestEval(t *testing.T) {
	t.Run("loads and parses evaluation prompt file", func(t *testing.T) {
		const yamlBody = `
name: Test Evaluation
description: A test evaluation
model: openai/gpt-4o
modelParameters:
  temperature: 0.5
  maxTokens: 100
testData:
  - input: "hello"
    expected: "hello world"
  - input: "goodbye"
    expected: "goodbye world"
messages:
  - role: system
    content: You are a helpful assistant.
  - role: user
    content: "Please respond to: {{input}}"
evaluators:
  - name: contains-world
    string:
      contains: "world"
  - name: similarity-check
    uses: github/similarity
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		evalFile, err := prompt.LoadFromFile(promptFile)
		require.NoError(t, err)
		require.Equal(t, "Test Evaluation", evalFile.Name)
		require.Equal(t, "A test evaluation", evalFile.Description)
		require.Equal(t, "openai/gpt-4o", evalFile.Model)
		require.Equal(t, 0.5, *evalFile.ModelParameters.Temperature)
		require.Equal(t, 100, *evalFile.ModelParameters.MaxTokens)
		require.Len(t, evalFile.TestData, 2)
		require.Len(t, evalFile.Messages, 2)
		require.Len(t, evalFile.Evaluators, 2)
	})

	t.Run("templates messages correctly", func(t *testing.T) {
		evalFile := &prompt.File{
			Messages: []prompt.Message{
				{Role: "system", Content: "You are helpful."},
				{Role: "user", Content: "Process {{input}} and return {{expected}}"},
			},
		}

		handler := &evalCommandHandler{evalFile: evalFile}
		testCase := map[string]interface{}{
			"input":    "hello",
			"expected": "world",
		}

		messages, err := handler.templateMessages(testCase)
		require.NoError(t, err)
		require.Len(t, messages, 2)
		require.Equal(t, "You are helpful.", *messages[0].Content)
		require.Equal(t, "Process hello and return world", *messages[1].Content)
	})

	t.Run("string evaluator works correctly", func(t *testing.T) {
		handler := &evalCommandHandler{}

		tests := []struct {
			name      string
			evaluator prompt.StringEvaluator
			response  string
			expected  bool
		}{
			{
				name:      "contains match",
				evaluator: prompt.StringEvaluator{Contains: "world"},
				response:  "hello world",
				expected:  true,
			},
			{
				name:      "contains no match",
				evaluator: prompt.StringEvaluator{Contains: "world"},
				response:  "hello there",
				expected:  false,
			},
			{
				name:      "equals match",
				evaluator: prompt.StringEvaluator{Equals: "exact"},
				response:  "exact",
				expected:  true,
			},
			{
				name:      "equals no match",
				evaluator: prompt.StringEvaluator{Equals: "exact"},
				response:  "not exact",
				expected:  false,
			},
			{
				name:      "starts with match",
				evaluator: prompt.StringEvaluator{StartsWith: "hello"},
				response:  "hello world",
				expected:  true,
			},
			{
				name:      "ends with match",
				evaluator: prompt.StringEvaluator{EndsWith: "world"},
				response:  "hello world",
				expected:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := handler.runStringEvaluator("test", tt.evaluator, tt.response)
				require.NoError(t, err)
				require.Equal(t, tt.expected, result.Passed)
				if tt.expected {
					require.Equal(t, 1.0, result.Score)
				} else {
					require.Equal(t, 0.0, result.Score)
				}
			})
		}
	})

	t.Run("plugin evaluator works with github/similarity", func(t *testing.T) {
		out := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(out, out, client, true, 100)

		// Mock a response that returns "4" for the LLM evaluator
		client.MockGetChatCompletionStream = func(ctx context.Context, req azuremodels.ChatCompletionOptions) (*azuremodels.ChatCompletionResponse, error) {
			reader := sse.NewMockEventReader([]azuremodels.ChatCompletion{
				{
					Choices: []azuremodels.ChatChoice{
						{
							Message: &azuremodels.ChatChoiceMessage{
								Content: func() *string { s := "4"; return &s }(),
							},
						},
					},
				},
			})
			return &azuremodels.ChatCompletionResponse{Reader: reader}, nil
		}

		handler := &evalCommandHandler{
			cfg:    cfg,
			client: client,
		}
		testCase := map[string]interface{}{
			"input":    "test question",
			"expected": "test answer",
		}

		result, err := handler.runPluginEvaluator(context.Background(), "similarity", "github/similarity", testCase, "test response")
		require.NoError(t, err)
		require.Equal(t, "similarity", result.EvaluatorName)
		require.Equal(t, 0.75, result.Score) // Score for choice "4"
		require.True(t, result.Passed)
	})

	t.Run("command creation works", func(t *testing.T) {
		out := new(bytes.Buffer)
		client := azuremodels.NewMockClient()
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewEvalCommand(cfg)
		require.Equal(t, "eval", cmd.Use)
		require.Contains(t, cmd.Short, "Evaluate prompts")
	})

	t.Run("integration test with mock client", func(t *testing.T) {
		const yamlBody = `
name: Mock Test
description: Test with mock client
model: openai/test-model
testData:
  - input: "test input"
    expected: "test response"
messages:
  - role: user
    content: "{{input}}"
evaluators:
  - name: contains-test
    string:
      contains: "test"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		client := azuremodels.NewMockClient()

		// Mock a simple response
		client.MockGetChatCompletionStream = func(ctx context.Context, req azuremodels.ChatCompletionOptions) (*azuremodels.ChatCompletionResponse, error) {
			// Create a mock reader that returns "test response"
			reader := sse.NewMockEventReader([]azuremodels.ChatCompletion{
				{
					Choices: []azuremodels.ChatChoice{
						{
							Message: &azuremodels.ChatChoiceMessage{
								Content: func() *string { s := "test response"; return &s }(),
							},
						},
					},
				},
			})
			return &azuremodels.ChatCompletionResponse{Reader: reader}, nil
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewEvalCommand(cfg)
		cmd.SetArgs([]string{promptFile})

		err = cmd.Execute()
		require.NoError(t, err)

		output := out.String()
		require.Contains(t, output, "Mock Test")
		require.Contains(t, output, "Running test case")
		require.Contains(t, output, "PASSED")
	})

	t.Run("logs model response when test fails", func(t *testing.T) {
		const yamlBody = `
name: Failing Test
description: Test that fails to check model response logging
model: openai/test-model
testData:
  - input: "test input"
    expected: "expected but not returned"
messages:
  - role: user
    content: "{{input}}"
evaluators:
  - name: contains-nonexistent
    string:
      contains: "nonexistent text"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		client := azuremodels.NewMockClient()

		// Mock a response that will fail the evaluator
		client.MockGetChatCompletionStream = func(ctx context.Context, req azuremodels.ChatCompletionOptions) (*azuremodels.ChatCompletionResponse, error) {
			reader := sse.NewMockEventReader([]azuremodels.ChatCompletion{
				{
					Choices: []azuremodels.ChatChoice{
						{
							Message: &azuremodels.ChatChoiceMessage{
								Content: func() *string { s := "actual model response"; return &s }(),
							},
						},
					},
				},
			})
			return &azuremodels.ChatCompletionResponse{Reader: reader}, nil
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewEvalCommand(cfg)
		cmd.SetArgs([]string{promptFile})

		err = cmd.Execute()
		require.NoError(t, err)

		output := out.String()
		require.Contains(t, output, "Failing Test")
		require.Contains(t, output, "Running test case")
		require.Contains(t, output, "FAILED")
		require.Contains(t, output, "Model Response: actual model response")
	})

	t.Run("json output format", func(t *testing.T) {
		const yamlBody = `
name: JSON Test Evaluation
description: Testing JSON output format
model: openai/gpt-4o
testData:
  - input: "hello"
    expected: "hello world"
  - input: "test"
    expected: "test response"
messages:
  - role: user
    content: "Respond to: {{input}}"
evaluators:
  - name: contains-hello
    string:
      contains: "hello"
  - name: exact-match
    string:
      equals: "hello world"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		client := azuremodels.NewMockClient()

		// Mock responses for both test cases
		callCount := 0
		client.MockGetChatCompletionStream = func(ctx context.Context, req azuremodels.ChatCompletionOptions) (*azuremodels.ChatCompletionResponse, error) {
			callCount++
			var response string
			if callCount == 1 {
				response = "hello world" // This will pass both evaluators
			} else {
				response = "test output" // This will fail both evaluators
			}

			reader := sse.NewMockEventReader([]azuremodels.ChatCompletion{
				{
					Choices: []azuremodels.ChatChoice{
						{
							Message: &azuremodels.ChatChoiceMessage{
								Content: &response,
							},
						},
					},
				},
			})
			return &azuremodels.ChatCompletionResponse{Reader: reader}, nil
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewEvalCommand(cfg)
		cmd.SetArgs([]string{"--json", promptFile})

		err = cmd.Execute()
		require.NoError(t, err)

		output := out.String()

		// Verify JSON structure
		var result EvaluationSummary
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		// Verify top-level fields
		require.Equal(t, "JSON Test Evaluation", result.Name)
		require.Equal(t, "Testing JSON output format", result.Description)
		require.Equal(t, "openai/gpt-4o", result.Model)
		require.Len(t, result.TestResults, 2)

		// Verify summary
		require.Equal(t, 2, result.Summary.TotalTests)
		require.Equal(t, 1, result.Summary.PassedTests)
		require.Equal(t, 1, result.Summary.FailedTests)
		require.Equal(t, 50.0, result.Summary.PassRate)

		// Verify first test case (should pass)
		testResult1 := result.TestResults[0]
		require.Equal(t, "hello world", testResult1.ModelResponse)
		require.Len(t, testResult1.EvaluationResults, 2)
		require.True(t, testResult1.EvaluationResults[0].Passed)
		require.True(t, testResult1.EvaluationResults[1].Passed)
		require.Equal(t, 1.0, testResult1.EvaluationResults[0].Score)
		require.Equal(t, 1.0, testResult1.EvaluationResults[1].Score)

		// Verify second test case (should fail)
		testResult2 := result.TestResults[1]
		require.Equal(t, "test output", testResult2.ModelResponse)
		require.Len(t, testResult2.EvaluationResults, 2)
		require.False(t, testResult2.EvaluationResults[0].Passed)
		require.False(t, testResult2.EvaluationResults[1].Passed)
		require.Equal(t, 0.0, testResult2.EvaluationResults[0].Score)
		require.Equal(t, 0.0, testResult2.EvaluationResults[1].Score)

		// Verify that human-readable text is NOT in the output
		require.NotContains(t, output, "Running evaluation:")
		require.NotContains(t, output, "✓ PASSED")
		require.NotContains(t, output, "✗ FAILED")
		require.NotContains(t, output, "Evaluation Summary:")
	})

	t.Run("json output vs human-readable output", func(t *testing.T) {
		const yamlBody = `
name: Output Comparison Test
description: Compare JSON vs human-readable output
model: openai/gpt-4o
testData:
  - input: "hello"
messages:
  - role: user
    content: "Say: {{input}}"
evaluators:
  - name: simple-check
    string:
      contains: "hello"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		client := azuremodels.NewMockClient()
		client.MockGetChatCompletionStream = func(ctx context.Context, req azuremodels.ChatCompletionOptions) (*azuremodels.ChatCompletionResponse, error) {
			response := "hello world"
			reader := sse.NewMockEventReader([]azuremodels.ChatCompletion{
				{
					Choices: []azuremodels.ChatChoice{
						{
							Message: &azuremodels.ChatChoiceMessage{
								Content: &response,
							},
						},
					},
				},
			})
			return &azuremodels.ChatCompletionResponse{Reader: reader}, nil
		}

		// Test human-readable output
		humanOut := new(bytes.Buffer)
		humanCfg := command.NewConfig(humanOut, humanOut, client, true, 100)
		humanCmd := NewEvalCommand(humanCfg)
		humanCmd.SetArgs([]string{promptFile})

		err = humanCmd.Execute()
		require.NoError(t, err)

		humanOutput := humanOut.String()
		require.Contains(t, humanOutput, "Running evaluation:")
		require.Contains(t, humanOutput, "Output Comparison Test")
		require.Contains(t, humanOutput, "✓ PASSED")
		require.Contains(t, humanOutput, "Evaluation Summary:")
		require.Contains(t, humanOutput, "🎉 All tests passed!")

		// Test JSON output
		jsonOut := new(bytes.Buffer)
		jsonCfg := command.NewConfig(jsonOut, jsonOut, client, true, 100)
		jsonCmd := NewEvalCommand(jsonCfg)
		jsonCmd.SetArgs([]string{"--json", promptFile})

		err = jsonCmd.Execute()
		require.NoError(t, err)

		jsonOutput := jsonOut.String()

		// Verify JSON is valid
		var result EvaluationSummary
		err = json.Unmarshal([]byte(jsonOutput), &result)
		require.NoError(t, err)

		// Verify JSON doesn't contain human-readable elements
		require.NotContains(t, jsonOutput, "Running evaluation:")
		require.NotContains(t, jsonOutput, "✓ PASSED")
		require.NotContains(t, jsonOutput, "Evaluation Summary:")
		require.NotContains(t, jsonOutput, "🎉")

		// Verify JSON contains the right data
		require.Equal(t, "Output Comparison Test", result.Name)
		require.Equal(t, 1, result.Summary.TotalTests)
		require.Equal(t, 1, result.Summary.PassedTests)
	})

	t.Run("json flag works with failing tests", func(t *testing.T) {
		const yamlBody = `
name: JSON Failing Test
description: Testing JSON with failing evaluators
model: openai/gpt-4o
testData:
  - input: "hello"
messages:
  - role: user
    content: "{{input}}"
evaluators:
  - name: impossible-check
    string:
      contains: "impossible_string_that_wont_match"
`

		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt.yml")
		err := os.WriteFile(promptFile, []byte(yamlBody), 0644)
		require.NoError(t, err)

		client := azuremodels.NewMockClient()
		client.MockGetChatCompletionStream = func(ctx context.Context, req azuremodels.ChatCompletionOptions) (*azuremodels.ChatCompletionResponse, error) {
			response := "hello world"
			reader := sse.NewMockEventReader([]azuremodels.ChatCompletion{
				{
					Choices: []azuremodels.ChatChoice{
						{
							Message: &azuremodels.ChatChoiceMessage{
								Content: &response,
							},
						},
					},
				},
			})
			return &azuremodels.ChatCompletionResponse{Reader: reader}, nil
		}

		out := new(bytes.Buffer)
		cfg := command.NewConfig(out, out, client, true, 100)

		cmd := NewEvalCommand(cfg)
		cmd.SetArgs([]string{"--json", promptFile})

		err = cmd.Execute()
		require.NoError(t, err)

		output := out.String()

		var result EvaluationSummary
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		// Verify failing test is properly represented
		require.Equal(t, 1, result.Summary.TotalTests)
		require.Equal(t, 0, result.Summary.PassedTests)
		require.Equal(t, 1, result.Summary.FailedTests)
		require.Equal(t, 0.0, result.Summary.PassRate)

		require.Len(t, result.TestResults, 1)
		require.False(t, result.TestResults[0].EvaluationResults[0].Passed)
		require.Equal(t, 0.0, result.TestResults[0].EvaluationResults[0].Score)
	})
}
