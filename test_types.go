package main

import (
	"fmt"

	"github.com/github/gh-models/cmd/generate"
)

func main() {
	test := generate.PromptPexTest{
		Scenario:  "test scenario",
		Reasoning: "test reasoning",
		TestInput: "test input",
		RuleID:    1,
		TestID:    2,
		Baseline:  true,
	}

	fmt.Printf("Scenario type: %T, value: %s\n", test.Scenario, test.Scenario)
	fmt.Printf("Reasoning type: %T, value: %s\n", test.Reasoning, test.Reasoning)
	fmt.Printf("RuleID type: %T, value: %d\n", test.RuleID, test.RuleID)
	fmt.Printf("TestID type: %T, value: %d\n", test.TestID, test.TestID)
	fmt.Printf("Baseline type: %T, value: %t\n", test.Baseline, test.Baseline)
}
