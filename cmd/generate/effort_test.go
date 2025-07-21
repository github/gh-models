package generate

import (
	"reflect"
	"testing"
)

func TestGetEffortConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		effort   string
		expected *EffortConfiguration
	}{
		{
			name:   "EffortMin configuration",
			effort: EffortMin,
			expected: &EffortConfiguration{
				SplitRules:                BoolPtr(false),
				TestGenerations:           IntPtr(1),
				TestsPerRule:              IntPtr(1),
				RunsPerTest:               IntPtr(1),
				TestExpansions:            IntPtr(0),
				MaxRules:                  IntPtr(6),
				MaxRulesPerTestGeneration: IntPtr(100),
				MaxTestsToRun:             IntPtr(10),
				Compliance:                BoolPtr(false),
			},
		},
		{
			name:   "EffortLow configuration",
			effort: EffortLow,
			expected: &EffortConfiguration{
				TestExpansions:            IntPtr(0),
				TestGenerations:           IntPtr(1),
				MaxRules:                  IntPtr(3),
				TestsPerRule:              IntPtr(2),
				RunsPerTest:               IntPtr(1),
				MaxRulesPerTestGeneration: IntPtr(5),
				SplitRules:                BoolPtr(true),
				MaxTestsToRun:             IntPtr(20),
			},
		},
		{
			name:   "EffortMedium configuration",
			effort: EffortMedium,
			expected: &EffortConfiguration{
				TestExpansions:            IntPtr(0),
				MaxRules:                  IntPtr(20),
				TestsPerRule:              IntPtr(3),
				RunsPerTest:               IntPtr(1),
				MaxRulesPerTestGeneration: IntPtr(5),
				SplitRules:                BoolPtr(true),
				TestGenerations:           IntPtr(1),
			},
		},
		{
			name:   "EffortHigh configuration",
			effort: EffortHigh,
			expected: &EffortConfiguration{
				TestExpansions:            IntPtr(1),
				MaxRules:                  IntPtr(50),
				MaxRulesPerTestGeneration: IntPtr(2),
				SplitRules:                BoolPtr(true),
				TestGenerations:           IntPtr(2),
			},
		},
		{
			name:     "unknown effort level",
			effort:   "unknown",
			expected: nil,
		},
		{
			name:     "empty effort level",
			effort:   "",
			expected: nil,
		},
		{
			name:     "case sensitive effort level",
			effort:   "MIN",
			expected: nil,
		},
		{
			name:     "partial match effort level",
			effort:   "mi",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEffortConfiguration(tt.effort)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("GetEffortConfiguration(%q) = %+v, want nil", tt.effort, result)
				}
				return
			}

			if result == nil {
				t.Errorf("GetEffortConfiguration(%q) = nil, want %+v", tt.effort, tt.expected)
				return
			}

			// Use reflect.DeepEqual for comprehensive comparison
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetEffortConfiguration(%q) = %+v, want %+v", tt.effort, result, tt.expected)
			}
		})
	}
}

func TestGetEffortConfiguration_FieldComparison(t *testing.T) {
	// Test individual fields for EffortMin to ensure correctness
	config := GetEffortConfiguration(EffortMin)
	if config == nil {
		t.Fatal("GetEffortConfiguration(EffortMin) returned nil")
	}

	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"SplitRules", config.SplitRules, BoolPtr(false)},
		{"TestGenerations", config.TestGenerations, IntPtr(1)},
		{"TestsPerRule", config.TestsPerRule, IntPtr(1)},
		{"RunsPerTest", config.RunsPerTest, IntPtr(1)},
		{"TestExpansions", config.TestExpansions, IntPtr(0)},
		{"MaxRules", config.MaxRules, IntPtr(6)},
		{"MaxRulesPerTestGeneration", config.MaxRulesPerTestGeneration, IntPtr(100)},
		{"MaxTestsToRun", config.MaxTestsToRun, IntPtr(10)},
		{"Compliance", config.Compliance, BoolPtr(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.actual, tt.expected) {
				t.Errorf("EffortMin.%s = %+v, want %+v", tt.name, tt.actual, tt.expected)
			}
		})
	}
}

func TestApplyEffortConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		initialOptions  *PromptPexOptions
		effort          string
		expectedChanges map[string]interface{}
		description     string
	}{
		{
			name:           "apply to empty options with EffortMin",
			initialOptions: &PromptPexOptions{},
			effort:         EffortMin,
			expectedChanges: map[string]interface{}{
				"SplitRules":         BoolPtr(false),
				"TestGenerations":    IntPtr(1),
				"TestsPerRule":       IntPtr(1),
				"RunsPerTest":        IntPtr(1),
				"TestExpansions":     IntPtr(0),
				"MaxRules":           IntPtr(6),
				"MaxRulesPerTestGen": IntPtr(100),
				"MaxTestsToRun":      IntPtr(10),
				"Compliance":         BoolPtr(false),
			},
			description: "All fields should be set from EffortMin configuration",
		},
		{
			name: "apply to options with existing values",
			initialOptions: &PromptPexOptions{
				SplitRules:      BoolPtr(true), // Already set, should not change
				TestGenerations: IntPtr(5),     // Already set, should not change
				TestsPerRule:    nil,           // Not set, should be applied
				MaxRules:        nil,           // Not set, should be applied
			},
			effort: EffortMin,
			expectedChanges: map[string]interface{}{
				"SplitRules":         BoolPtr(true),  // Should remain unchanged
				"TestGenerations":    IntPtr(5),      // Should remain unchanged
				"TestsPerRule":       IntPtr(1),      // Should be applied from EffortMin
				"RunsPerTest":        IntPtr(1),      // Should be applied from EffortMin
				"TestExpansions":     IntPtr(0),      // Should be applied from EffortMin
				"MaxRules":           IntPtr(6),      // Should be applied from EffortMin
				"MaxRulesPerTestGen": IntPtr(100),    // Should be applied from EffortMin
				"MaxTestsToRun":      IntPtr(10),     // Should be applied from EffortMin
				"Compliance":         BoolPtr(false), // Should be applied from EffortMin
			},
			description: "Only unset fields should be applied from configuration",
		},
		{
			name:            "apply with empty effort string",
			initialOptions:  &PromptPexOptions{},
			effort:          "",
			expectedChanges: map[string]interface{}{},
			description:     "No changes should be made with empty effort",
		},
		{
			name:            "apply with unknown effort level",
			initialOptions:  &PromptPexOptions{},
			effort:          "unknown",
			expectedChanges: map[string]interface{}{},
			description:     "No changes should be made with unknown effort level",
		},
		{
			name:           "apply EffortLow configuration",
			initialOptions: &PromptPexOptions{},
			effort:         EffortLow,
			expectedChanges: map[string]interface{}{
				"TestExpansions":     IntPtr(0),
				"TestGenerations":    IntPtr(1),
				"MaxRules":           IntPtr(3),
				"TestsPerRule":       IntPtr(2),
				"RunsPerTest":        IntPtr(1),
				"MaxRulesPerTestGen": IntPtr(5),
				"SplitRules":         BoolPtr(true),
				"MaxTestsToRun":      IntPtr(20),
			},
			description: "All fields should be set from EffortLow configuration",
		},
		{
			name:           "apply EffortHigh configuration",
			initialOptions: &PromptPexOptions{},
			effort:         EffortHigh,
			expectedChanges: map[string]interface{}{
				"TestExpansions":     IntPtr(1),
				"MaxRules":           IntPtr(50),
				"MaxRulesPerTestGen": IntPtr(2),
				"SplitRules":         BoolPtr(true),
				"TestGenerations":    IntPtr(2),
			},
			description: "All fields should be set from EffortHigh configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the initial options to avoid modifying the test data
			options := &PromptPexOptions{}
			if tt.initialOptions != nil {
				*options = *tt.initialOptions
			}

			// Apply the effort configuration
			ApplyEffortConfiguration(options, tt.effort)

			// Check each expected change
			for fieldName, expectedValue := range tt.expectedChanges {
				var actualValue interface{}

				switch fieldName {
				case "SplitRules":
					actualValue = options.SplitRules
				case "TestGenerations":
					actualValue = options.TestGenerations
				case "TestsPerRule":
					actualValue = options.TestsPerRule
				case "RunsPerTest":
					actualValue = options.RunsPerTest
				case "TestExpansions":
					actualValue = options.TestExpansions
				case "MaxRules":
					actualValue = options.MaxRules
				case "MaxRulesPerTestGen":
					actualValue = options.MaxRulesPerTestGen
				case "MaxTestsToRun":
					actualValue = options.MaxTestsToRun
				case "Compliance":
					actualValue = options.Compliance
				default:
					t.Errorf("Unknown field name in test: %s", fieldName)
					continue
				}

				if !reflect.DeepEqual(actualValue, expectedValue) {
					t.Errorf("After applying effort %q, field %s = %+v, want %+v", tt.effort, fieldName, actualValue, expectedValue)
				}
			}

			// If no changes expected, verify that the options remain empty/unchanged
			if len(tt.expectedChanges) == 0 {
				if !isOptionsEmpty(options, tt.initialOptions) {
					t.Errorf("Expected no changes but options were modified: %+v", options)
				}
			}
		})
	}
}

func TestApplyEffortConfiguration_NilOptions(t *testing.T) {
	// Test that the function handles nil options gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ApplyEffortConfiguration panicked with nil options: %v", r)
		}
	}()

	// This should not panic and should handle nil gracefully
	ApplyEffortConfiguration(nil, EffortMin)
	// If we get here without panicking, the test passes
}

func TestEffortConfigurationConstants(t *testing.T) {
	// Test that the effort constants are properly defined
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"EffortMin constant", EffortMin, "min"},
		{"EffortLow constant", EffortLow, "low"},
		{"EffortMedium constant", EffortMedium, "medium"},
		{"EffortHigh constant", EffortHigh, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestEffortConfiguration_AllLevelsHaveUniqueValues(t *testing.T) {
	// Test that each effort level produces a unique configuration
	configs := map[string]*EffortConfiguration{
		EffortMin:    GetEffortConfiguration(EffortMin),
		EffortLow:    GetEffortConfiguration(EffortLow),
		EffortMedium: GetEffortConfiguration(EffortMedium),
		EffortHigh:   GetEffortConfiguration(EffortHigh),
	}

	// Verify all configurations are non-nil
	for effort, config := range configs {
		if config == nil {
			t.Errorf("GetEffortConfiguration(%q) returned nil", effort)
		}
	}

	// Check that configurations are different from each other
	efforts := []string{EffortMin, EffortLow, EffortMedium, EffortHigh}
	for i := 0; i < len(efforts); i++ {
		for j := i + 1; j < len(efforts); j++ {
			effort1, effort2 := efforts[i], efforts[j]
			config1, config2 := configs[effort1], configs[effort2]

			if reflect.DeepEqual(config1, config2) {
				t.Errorf("Configurations for %q and %q are identical: %+v", effort1, effort2, config1)
			}
		}
	}
}

func TestEffortConfiguration_ProgressiveComplexity(t *testing.T) {
	// Test that effort levels generally increase in complexity
	// Note: This is a heuristic test based on the assumption that higher effort means more resources

	minConfig := GetEffortConfiguration(EffortMin)
	lowConfig := GetEffortConfiguration(EffortLow)
	mediumConfig := GetEffortConfiguration(EffortMedium)
	highConfig := GetEffortConfiguration(EffortHigh)

	// Test that MaxRules generally increases with effort level
	if *minConfig.MaxRules > *mediumConfig.MaxRules {
		t.Errorf("Expected EffortMin.MaxRules (%d) <= EffortMedium.MaxRules (%d)", *minConfig.MaxRules, *mediumConfig.MaxRules)
	}

	if *mediumConfig.MaxRules > *highConfig.MaxRules {
		t.Errorf("Expected EffortMedium.MaxRules (%d) <= EffortHigh.MaxRules (%d)", *mediumConfig.MaxRules, *highConfig.MaxRules)
	}

	// Test that TestGenerations increases with effort
	if *lowConfig.TestGenerations > *highConfig.TestGenerations {
		t.Errorf("Expected EffortLow.TestGenerations (%d) <= EffortHigh.TestGenerations (%d)", *lowConfig.TestGenerations, *highConfig.TestGenerations)
	}

	// Test that EffortHigh has the only non-zero TestExpansions
	if *minConfig.TestExpansions != 0 {
		t.Errorf("Expected EffortMin.TestExpansions to be 0, got %d", *minConfig.TestExpansions)
	}
	if *lowConfig.TestExpansions != 0 {
		t.Errorf("Expected EffortLow.TestExpansions to be 0, got %d", *lowConfig.TestExpansions)
	}
	if *mediumConfig.TestExpansions != 0 {
		t.Errorf("Expected EffortMedium.TestExpansions to be 0, got %d", *mediumConfig.TestExpansions)
	}
	if *highConfig.TestExpansions != 1 {
		t.Errorf("Expected EffortHigh.TestExpansions to be 1, got %d", *highConfig.TestExpansions)
	}
}

// Helper function to check if options are empty or unchanged
func isOptionsEmpty(options *PromptPexOptions, original *PromptPexOptions) bool {
	if original == nil {
		return options.SplitRules == nil &&
			options.TestGenerations == nil &&
			options.TestsPerRule == nil &&
			options.RunsPerTest == nil &&
			options.TestExpansions == nil &&
			options.MaxRules == nil &&
			options.MaxRulesPerTestGen == nil &&
			options.MaxTestsToRun == nil &&
			options.Compliance == nil
	}

	// Compare with original values
	return reflect.DeepEqual(options.SplitRules, original.SplitRules) &&
		reflect.DeepEqual(options.TestGenerations, original.TestGenerations) &&
		reflect.DeepEqual(options.TestsPerRule, original.TestsPerRule) &&
		reflect.DeepEqual(options.RunsPerTest, original.RunsPerTest) &&
		reflect.DeepEqual(options.TestExpansions, original.TestExpansions) &&
		reflect.DeepEqual(options.MaxRules, original.MaxRules) &&
		reflect.DeepEqual(options.MaxRulesPerTestGen, original.MaxRulesPerTestGen) &&
		reflect.DeepEqual(options.MaxTestsToRun, original.MaxTestsToRun) &&
		reflect.DeepEqual(options.Compliance, original.Compliance)
}
