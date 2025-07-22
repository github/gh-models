package generate

import (
	"encoding/json"
	"fmt"
)

// ParseTestsFromLLMResponse parses test cases from LLM response with robust error handling
func (h *generateCommandHandler) ParseTestsFromLLMResponse(content string) ([]PromptPexTest, error) {
	jsonStr := ExtractJSON(content)

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
