package eval

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimilarityEvaluator(t *testing.T) {
	evaluator := NewSimilarityEvaluator()

	t.Run("exact match", func(t *testing.T) {
		testCase := map[string]interface{}{
			"expected": "hello world",
		}

		result, err := evaluator.Evaluate("similarity", testCase, "hello world")
		require.NoError(t, err)
		require.True(t, result.Passed)
		require.Equal(t, 1.0, result.Score)
		require.Contains(t, result.Details, "1.00")
	})

	t.Run("high similarity", func(t *testing.T) {
		testCase := map[string]interface{}{
			"expected": "hello world test",
		}

		result, err := evaluator.Evaluate("similarity", testCase, "hello world example")
		require.NoError(t, err)
		require.True(t, result.Passed)
		require.True(t, result.Score > 0.4)
	})

	t.Run("low similarity", func(t *testing.T) {
		testCase := map[string]interface{}{
			"expected": "hello world",
		}

		result, err := evaluator.Evaluate("similarity", testCase, "completely different text")
		require.NoError(t, err)
		require.False(t, result.Passed)
		require.True(t, result.Score < 0.7)
	})

	t.Run("missing expected value", func(t *testing.T) {
		testCase := map[string]interface{}{}

		result, err := evaluator.Evaluate("similarity", testCase, "hello world")
		require.NoError(t, err)
		require.False(t, result.Passed)
		require.Equal(t, 0.0, result.Score)
		require.Contains(t, result.Details, "No 'expected' value found")
	})

	t.Run("non-string expected value", func(t *testing.T) {
		testCase := map[string]interface{}{
			"expected": 123,
		}

		result, err := evaluator.Evaluate("similarity", testCase, "hello world")
		require.NoError(t, err)
		require.False(t, result.Passed)
		require.Equal(t, 0.0, result.Score)
		require.Contains(t, result.Details, "Expected value is not a string")
	})

	t.Run("empty strings", func(t *testing.T) {
		testCase := map[string]interface{}{
			"expected": "",
		}

		result, err := evaluator.Evaluate("similarity", testCase, "")
		require.NoError(t, err)
		require.True(t, result.Passed)
		require.Equal(t, 1.0, result.Score)
	})

	t.Run("case insensitive", func(t *testing.T) {
		testCase := map[string]interface{}{
			"expected": "Hello World",
		}

		result, err := evaluator.Evaluate("similarity", testCase, "hello world")
		require.NoError(t, err)
		require.True(t, result.Passed)
		require.Equal(t, 1.0, result.Score)
	})
}

func TestCalculateSimpleSimilarity(t *testing.T) {
	evaluator := NewSimilarityEvaluator()

	tests := []struct {
		name     string
		expected string
		actual   string
		minScore float64
		maxScore float64
	}{
		{
			name:     "identical strings",
			expected: "hello world",
			actual:   "hello world",
			minScore: 1.0,
			maxScore: 1.0,
		},
		{
			name:     "partial overlap",
			expected: "hello world test",
			actual:   "hello world example",
			minScore: 0.4,
			maxScore: 0.6,
		},
		{
			name:     "no overlap",
			expected: "hello world",
			actual:   "foo bar",
			minScore: 0.0,
			maxScore: 0.0,
		},
		{
			name:     "empty strings",
			expected: "",
			actual:   "",
			minScore: 1.0,
			maxScore: 1.0,
		},
		{
			name:     "one empty",
			expected: "hello",
			actual:   "",
			minScore: 0.0,
			maxScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := evaluator.calculateSimpleSimilarity(tt.expected, tt.actual)
			require.GreaterOrEqual(t, score, tt.minScore)
			require.LessOrEqual(t, score, tt.maxScore)
		})
	}
}
