package generate

import (
	"reflect"
	"testing"

	"github.com/github/gh-models/pkg/util"
)

func TestGetDefaultOptions(t *testing.T) {
	defaults := GetDefaultOptions()

	// Test individual fields to ensure they have expected default values
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"Temperature", defaults.Temperature, util.Ptr(0.0)},
		{"TestsPerRule", defaults.TestsPerRule, util.Ptr(3)},
		{"RunsPerTest", defaults.RunsPerTest, util.Ptr(2)},
		{"MaxRulesPerTestGen", defaults.MaxRulesPerTestGen, util.Ptr(3)},
		{"TestGenerations", defaults.TestGenerations, util.Ptr(2)},
		{"TestExpansions", defaults.TestExpansions, util.Ptr(0)},
		{"FilterTestCount", defaults.FilterTestCount, util.Ptr(5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.actual, tt.expected) {
				t.Errorf("GetDefaultOptions().%s = %+v, want %+v", tt.name, tt.actual, tt.expected)
			}
		})
	}
}

func TestGetDefaultOptions_Consistency(t *testing.T) {
	// Test that calling GetDefaultOptions multiple times returns the same values
	defaults1 := GetDefaultOptions()
	defaults2 := GetDefaultOptions()

	if !reflect.DeepEqual(defaults1, defaults2) {
		t.Errorf("GetDefaultOptions() returned different values on subsequent calls")
	}
}

func TestGetDefaultOptions_NonNilFields(t *testing.T) {
	// Test that all expected fields are non-nil in default options
	defaults := GetDefaultOptions()

	nonNilFields := []struct {
		name  string
		value interface{}
	}{
		{"Temperature", defaults.Temperature},
		{"TestsPerRule", defaults.TestsPerRule},
		{"RunsPerTest", defaults.RunsPerTest},
		{"MaxRulesPerTestGen", defaults.MaxRulesPerTestGen},
		{"TestGenerations", defaults.TestGenerations},
		{"TestExpansions", defaults.TestExpansions},
		{"FilterTestCount", defaults.FilterTestCount},
	}

	for _, field := range nonNilFields {
		t.Run(field.name, func(t *testing.T) {
			if field.value == nil {
				t.Errorf("GetDefaultOptions().%s is nil, expected non-nil value", field.name)
			}
		})
	}
}
