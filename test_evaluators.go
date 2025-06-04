package main

import (
	"fmt"
	"github.com/github/gh-models/cmd/eval"
)

func main() {
	fmt.Println("Testing built-in evaluators...")

	// Test that all expected evaluators exist
	evaluators := []string{"similarity", "coherence", "fluency", "relevance", "groundedness"}

	for _, name := range evaluators {
		if evaluator, exists := eval.BuiltInEvaluators[name]; exists {
			fmt.Printf("✓ %s evaluator exists with model: %s\n", name, evaluator.ModelID)
		} else {
			fmt.Printf("✗ %s evaluator not found\n", name)
		}
	}

	fmt.Println("Built-in evaluators test completed!")
}
