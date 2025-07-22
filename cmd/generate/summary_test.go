package generate

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/prompt"
)

func TestGenerateSummary(t *testing.T) {
	tests := []struct {
		name            string
		context         *PromptPexContext
		expectedMessage string
		expectedJSON    map[string]interface{}
	}{
		{
			name: "basic summary with tests",
			context: &PromptPexContext{
				RunID: "run_test_123",
				Prompt: &prompt.File{
					Name: "test-prompt",
				},
				PromptPexTests: []PromptPexTest{
					{TestInput: "test1"},
					{TestInput: "test2"},
					{TestInput: "test3"},
				},
			},
			expectedMessage: "Summary: Generated 3 tests for prompt 'test-prompt'",
			expectedJSON: map[string]interface{}{
				"name":  "test-prompt",
				"tests": float64(3), // JSON unmarshaling converts numbers to float64
				"runId": "run_test_123",
			},
		},
		{
			name: "summary with no tests",
			context: &PromptPexContext{
				RunID: "run_empty_456",
				Prompt: &prompt.File{
					Name: "empty-prompt",
				},
				PromptPexTests: []PromptPexTest{},
			},
			expectedMessage: "Summary: Generated 0 tests for prompt 'empty-prompt'",
			expectedJSON: map[string]interface{}{
				"name":  "empty-prompt",
				"tests": float64(0),
				"runId": "run_empty_456",
			},
		},
		{
			name: "summary with single test",
			context: &PromptPexContext{
				RunID: "run_single_789",
				Prompt: &prompt.File{
					Name: "single-test-prompt",
				},
				PromptPexTests: []PromptPexTest{
					{TestInput: "only test"},
				},
			},
			expectedMessage: "Summary: Generated 1 tests for prompt 'single-test-prompt'",
			expectedJSON: map[string]interface{}{
				"name":  "single-test-prompt",
				"tests": float64(1),
				"runId": "run_single_789",
			},
		},
		{
			name: "summary with complex prompt name",
			context: &PromptPexContext{
				RunID: "run_complex_000",
				Prompt: &prompt.File{
					Name: "my-complex-prompt-with-special-chars",
				},
				PromptPexTests: []PromptPexTest{
					{TestInput: "test1"},
					{TestInput: "test2"},
					{TestInput: "test3"},
					{TestInput: "test4"},
					{TestInput: "test5"},
				},
			},
			expectedMessage: "Summary: Generated 5 tests for prompt 'my-complex-prompt-with-special-chars'",
			expectedJSON: map[string]interface{}{
				"name":  "my-complex-prompt-with-special-chars",
				"tests": float64(5),
				"runId": "run_complex_000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			var outBuf bytes.Buffer
			cfg := &command.Config{
				Out: &outBuf,
			}

			// Create handler
			handler := &generateCommandHandler{
				cfg: cfg,
			}

			// Call GenerateSummary
			jsonResult, err := handler.GenerateSummary(tt.context)

			// Check for no error
			if err != nil {
				t.Errorf("GenerateSummary() returned unexpected error: %v", err)
				return
			}

			// Check output message
			outputMessage := outBuf.String()
			if !strings.Contains(outputMessage, tt.expectedMessage) {
				t.Errorf("Expected output to contain %q, got %q", tt.expectedMessage, outputMessage)
			}

			// Check JSON result
			var actualJSON map[string]interface{}
			err = json.Unmarshal([]byte(jsonResult), &actualJSON)
			if err != nil {
				t.Errorf("Failed to unmarshal JSON result: %v", err)
				return
			}

			// Verify JSON fields
			for key, expectedValue := range tt.expectedJSON {
				actualValue, exists := actualJSON[key]
				if !exists {
					t.Errorf("Expected JSON to contain key %q", key)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("Expected JSON field %q to be %v, got %v", key, expectedValue, actualValue)
				}
			}

			// Check that JSON is properly formatted (indented)
			var compactJSON bytes.Buffer
			err = json.Compact(&compactJSON, []byte(jsonResult))
			if err != nil {
				t.Errorf("JSON result is not valid JSON: %v", err)
			}

			// The result should be indented (not compact)
			if strings.TrimSpace(jsonResult) == compactJSON.String() {
				t.Errorf("Expected JSON to be indented, but it appears to be compact")
			}
		})
	}
}

func TestGenerateSummaryNilContext(t *testing.T) {
	// Test with nil context - this should panic or handle gracefully
	// depending on the intended behavior
	var outBuf bytes.Buffer
	cfg := &command.Config{
		Out: &outBuf,
	}

	handler := &generateCommandHandler{
		cfg: cfg,
	}

	defer func() {
		if r := recover(); r != nil {
			// If it panics, that's expected behavior
			// We're just documenting this test case
			t.Logf("GenerateSummary panicked with nil context (expected): %v", r)
		}
	}()

	_, err := handler.GenerateSummary(nil)
	if err == nil {
		t.Errorf("Expected error or panic with nil context")
	}
}

func TestGenerateSummaryNilPrompt(t *testing.T) {
	// Test with nil prompt in context
	var outBuf bytes.Buffer
	cfg := &command.Config{
		Out: &outBuf,
	}

	handler := &generateCommandHandler{
		cfg: cfg,
	}

	context := &PromptPexContext{
		RunID:          "run_nil_prompt",
		Prompt:         nil, // nil prompt
		PromptPexTests: []PromptPexTest{},
	}

	defer func() {
		if r := recover(); r != nil {
			// If it panics, that's expected behavior
			t.Logf("GenerateSummary panicked with nil prompt (expected): %v", r)
		}
	}()

	_, err := handler.GenerateSummary(context)
	if err == nil {
		t.Errorf("Expected error or panic with nil prompt")
	}
}

func TestGenerateSummaryJSONFormat(t *testing.T) {
	// Test specifically the JSON formatting aspects
	var outBuf bytes.Buffer
	cfg := &command.Config{
		Out: &outBuf,
	}

	handler := &generateCommandHandler{
		cfg: cfg,
	}

	context := &PromptPexContext{
		RunID: "run_json_test",
		Prompt: &prompt.File{
			Name: "json-format-test",
		},
		PromptPexTests: []PromptPexTest{
			{TestInput: "test1"},
			{TestInput: "test2"},
		},
	}

	jsonResult, err := handler.GenerateSummary(context)
	if err != nil {
		t.Fatalf("GenerateSummary() returned unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var jsonObj map[string]interface{}
	err = json.Unmarshal([]byte(jsonResult), &jsonObj)
	if err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}

	// Verify formatting - should contain newlines (indented)
	if !strings.Contains(jsonResult, "\n") {
		t.Errorf("Expected JSON to be indented with newlines")
	}

	// Verify structure
	expectedKeys := []string{"name", "tests", "runId"}
	for _, key := range expectedKeys {
		if _, exists := jsonObj[key]; !exists {
			t.Errorf("Expected JSON to contain key %q", key)
		}
	}

	// Verify that returned string and console output are consistent
	expectedMessage := "Summary: Generated 2 tests for prompt 'json-format-test'"
	outputMessage := outBuf.String()
	if !strings.Contains(outputMessage, expectedMessage) {
		t.Errorf("Expected output message %q, got %q", expectedMessage, outputMessage)
	}
}

func TestGenerateSummaryLargeNumberOfTests(t *testing.T) {
	// Test with a large number of tests
	var outBuf bytes.Buffer
	cfg := &command.Config{
		Out: &outBuf,
	}

	handler := &generateCommandHandler{
		cfg: cfg,
	}

	// Create a large number of tests
	const numTests = 1000
	tests := make([]PromptPexTest, numTests)
	for i := 0; i < numTests; i++ {
		tests[i] = PromptPexTest{TestInput: "test" + string(rune(i))}
	}

	context := &PromptPexContext{
		RunID: "run_large_test",
		Prompt: &prompt.File{
			Name: "large-test-prompt",
		},
		PromptPexTests: tests,
	}

	jsonResult, err := handler.GenerateSummary(context)
	if err != nil {
		t.Errorf("GenerateSummary() returned unexpected error: %v", err)
	}

	// Verify JSON result
	var actualJSON map[string]interface{}
	err = json.Unmarshal([]byte(jsonResult), &actualJSON)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON result: %v", err)
	}

	// Check test count
	if actualJSON["tests"] != float64(numTests) {
		t.Errorf("Expected test count to be %d, got %v", numTests, actualJSON["tests"])
	}

	// Check output message
	expectedMessage := "Summary: Generated 1000 tests for prompt 'large-test-prompt'"
	outputMessage := outBuf.String()
	if !strings.Contains(outputMessage, expectedMessage) {
		t.Errorf("Expected output to contain %q, got %q", expectedMessage, outputMessage)
	}
}
