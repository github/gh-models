package eval

import (
	"fmt"
	"strings"
)

// SimilarityEvaluator handles similarity-based evaluation
type SimilarityEvaluator struct{}

// NewSimilarityEvaluator creates a new similarity evaluator
func NewSimilarityEvaluator() *SimilarityEvaluator {
	return &SimilarityEvaluator{}
}

// Evaluate runs similarity evaluation between expected and actual values
func (s *SimilarityEvaluator) Evaluate(name string, testCase map[string]interface{}, response string) (EvaluationResult, error) {
	// Simple similarity check using expected value if present
	expected, ok := testCase["expected"]
	if !ok {
		return EvaluationResult{
			EvaluatorName: name,
			Score:         0.0,
			Passed:        false,
			Details:       "No 'expected' value found in test case for similarity evaluation",
		}, nil
	}

	expectedStr, ok := expected.(string)
	if !ok {
		return EvaluationResult{
			EvaluatorName: name,
			Score:         0.0,
			Passed:        false,
			Details:       "Expected value is not a string",
		}, nil
	}

	// Simple similarity metric (could be enhanced with more sophisticated algorithms)
	similarity := s.calculateSimpleSimilarity(expectedStr, response)
	passed := similarity > 0.4 // 40% similarity threshold

	return EvaluationResult{
		EvaluatorName: name,
		Score:         similarity,
		Passed:        passed,
		Details:       fmt.Sprintf("Similarity score: %.2f (threshold: 0.4)", similarity),
	}, nil
}

// calculateSimpleSimilarity computes a simple word-based similarity score
func (s *SimilarityEvaluator) calculateSimpleSimilarity(expected, actual string) float64 {
	// Simple word-based similarity
	expectedWords := strings.Fields(strings.ToLower(expected))
	actualWords := strings.Fields(strings.ToLower(actual))

	if len(expectedWords) == 0 && len(actualWords) == 0 {
		return 1.0
	}
	if len(expectedWords) == 0 || len(actualWords) == 0 {
		return 0.0
	}

	// Count matching words
	expectedWordSet := make(map[string]bool)
	for _, word := range expectedWords {
		expectedWordSet[word] = true
	}

	matchingWords := 0
	for _, word := range actualWords {
		if expectedWordSet[word] {
			matchingWords++
		}
	}

	// Jaccard similarity
	totalWords := len(expectedWords) + len(actualWords) - matchingWords
	if totalWords == 0 {
		return 1.0
	}

	return float64(matchingWords) / float64(totalWords)
}
