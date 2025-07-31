package generate

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ParseRules removes numbering, bullets, and extraneous "Rules:" lines from a rules text block.
func ParseRules(text string) []string {
	if IsUnassistedResponse(text) {
		return nil
	}
	lines := SplitLines(Unbracket(Unxml(Unfence(text))))
	itemsRe := regexp.MustCompile(`^\s*(\d+\.|_|-|\*)\s+`) // remove leading item numbers or bullets
	rulesRe := regexp.MustCompile(`^\s*(Inverse\s+(Output\s+)?)?Rules:\s*$`)
	pythonWrapRe := regexp.MustCompile(`^\["?(.*?)"?\]$`)
	var cleaned []string
	for _, line := range lines {
		// Remove leading numbering or bullets
		replaced := itemsRe.ReplaceAllString(line, "")
		// Skip empty lines
		if strings.TrimSpace(replaced) == "" {
			continue
		}
		// Skip "Rules:" header lines
		if rulesRe.MatchString(replaced) {
			continue
		}
		// Remove ["..."] wrapping
		replaced = pythonWrapRe.ReplaceAllString(replaced, "$1")
		cleaned = append(cleaned, replaced)
	}
	return cleaned
}

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

		for _, key := range []string{"testInput", "testinput", "input"} {
			if input, ok := rawTest[key].(string); ok {
				test.Input = input
				break
			} else if inputObj, ok := rawTest[key].(map[string]interface{}); ok {
				// Convert structured object to JSON string
				if jsonBytes, err := json.Marshal(inputObj); err == nil {
					test.Input = string(jsonBytes)
				}
				break
			}
		}

		if scenario, ok := rawTest["scenario"].(string); ok {
			test.Scenario = scenario
		}
		if reasoning, ok := rawTest["reasoning"].(string); ok {
			test.Reasoning = reasoning
		}

		if test.Input == "" && test.Scenario == "" && test.Reasoning == "" {
			// If all fields are empty, skip this test
			continue
		} else if strings.TrimSpace(test.Input) == "" && (test.Scenario != "" || test.Reasoning != "") {
			// ignore whitespace-only inputs
			continue
		}

		tests = append(tests, test)
	}

	return tests, nil
}
