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
		{"SplitRules", defaults.SplitRules, util.Ptr(true)},
		{"MaxRulesPerTestGen", defaults.MaxRulesPerTestGen, util.Ptr(3)},
		{"TestGenerations", defaults.TestGenerations, util.Ptr(2)},
		{"TestExpansions", defaults.TestExpansions, util.Ptr(0)},
		{"FilterTestCount", defaults.FilterTestCount, util.Ptr(5)},
		{"Evals", defaults.Evals, util.Ptr(false)},
		{"Compliance", defaults.Compliance, util.Ptr(false)},
		{"BaselineTests", defaults.BaselineTests, util.Ptr(false)},
		{"StoreCompletions", defaults.StoreCompletions, util.Ptr(false)},
		{"CreateEvalRuns", defaults.CreateEvalRuns, util.Ptr(false)},
		{"RateTests", defaults.RateTests, util.Ptr(false)},
		{"DisableSafety", defaults.DisableSafety, util.Ptr(false)},
		{"EvalCache", defaults.EvalCache, util.Ptr(false)},
		{"TestRunCache", defaults.TestRunCache, util.Ptr(false)},
		{"OutputPrompts", defaults.OutputPrompts, util.Ptr(false)},
		{"WorkflowDiagram", defaults.WorkflowDiagram, util.Ptr(true)},
		{"LoadContext", defaults.LoadContext, util.Ptr(false)},
		{"LoadContextFile", defaults.LoadContextFile, util.Ptr("promptpex_context.json")},
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
		{"SplitRules", defaults.SplitRules},
		{"MaxRulesPerTestGen", defaults.MaxRulesPerTestGen},
		{"TestGenerations", defaults.TestGenerations},
		{"TestExpansions", defaults.TestExpansions},
		{"FilterTestCount", defaults.FilterTestCount},
		{"Evals", defaults.Evals},
		{"Compliance", defaults.Compliance},
		{"BaselineTests", defaults.BaselineTests},
		{"StoreCompletions", defaults.StoreCompletions},
		{"CreateEvalRuns", defaults.CreateEvalRuns},
		{"RateTests", defaults.RateTests},
		{"DisableSafety", defaults.DisableSafety},
		{"EvalCache", defaults.EvalCache},
		{"TestRunCache", defaults.TestRunCache},
		{"OutputPrompts", defaults.OutputPrompts},
		{"WorkflowDiagram", defaults.WorkflowDiagram},
		{"LoadContext", defaults.LoadContext},
		{"LoadContextFile", defaults.LoadContextFile},
	}

	for _, field := range nonNilFields {
		t.Run(field.name, func(t *testing.T) {
			if field.value == nil {
				t.Errorf("GetDefaultOptions().%s is nil, expected non-nil value", field.name)
			}
		})
	}
}

func TestMergeOptions_EmptyOverrides(t *testing.T) {
	// Test merging with empty overrides - should return defaults
	defaults := GetDefaultOptions()
	overrides := PromptPexOptions{}

	merged := MergeOptions(defaults, overrides)

	if !reflect.DeepEqual(merged, defaults) {
		t.Errorf("MergeOptions with empty overrides should return defaults")
	}
}

func TestMergeOptions_EmptyDefaults(t *testing.T) {
	// Test merging with empty defaults - should return overrides
	defaults := PromptPexOptions{}
	overrides := PromptPexOptions{
		Temperature:  util.Ptr(1.0),
		TestsPerRule: util.Ptr(5),
		SplitRules:   util.Ptr(false),
	}

	merged := MergeOptions(defaults, overrides)

	expected := overrides
	if !reflect.DeepEqual(merged, expected) {
		t.Errorf("MergeOptions with empty defaults = %+v, want %+v", merged, expected)
	}
}

func TestMergeOptions_OverridesPrecedence(t *testing.T) {
	// Test that overrides take precedence over defaults
	defaults := PromptPexOptions{
		Temperature:        util.Ptr(0.0),
		TestsPerRule:       util.Ptr(3),
		RunsPerTest:        util.Ptr(2),
		SplitRules:         util.Ptr(true),
		MaxRulesPerTestGen: util.Ptr(3),
		TestGenerations:    util.Ptr(2),
		Evals:              util.Ptr(false),
		WorkflowDiagram:    util.Ptr(true),
	}

	overrides := PromptPexOptions{
		Temperature:     util.Ptr(1.5),
		TestsPerRule:    util.Ptr(10),
		SplitRules:      util.Ptr(false),
		Evals:           util.Ptr(true),
		WorkflowDiagram: util.Ptr(false),
	}

	merged := MergeOptions(defaults, overrides)

	// Test that overridden values take precedence
	if !reflect.DeepEqual(merged.Temperature, util.Ptr(1.5)) {
		t.Errorf("merged.Temperature = %+v, want %+v", merged.Temperature, util.Ptr(1.5))
	}
	if !reflect.DeepEqual(merged.TestsPerRule, util.Ptr(10)) {
		t.Errorf("merged.TestsPerRule = %+v, want %+v", merged.TestsPerRule, util.Ptr(10))
	}
	if !reflect.DeepEqual(merged.SplitRules, util.Ptr(false)) {
		t.Errorf("merged.SplitRules = %+v, want %+v", merged.SplitRules, util.Ptr(false))
	}
	if !reflect.DeepEqual(merged.Evals, util.Ptr(true)) {
		t.Errorf("merged.Evals = %+v, want %+v", merged.Evals, util.Ptr(true))
	}
	if !reflect.DeepEqual(merged.WorkflowDiagram, util.Ptr(false)) {
		t.Errorf("merged.WorkflowDiagram = %+v, want %+v", merged.WorkflowDiagram, util.Ptr(false))
	}

	// Test that non-overridden values come from defaults
	if !reflect.DeepEqual(merged.RunsPerTest, util.Ptr(2)) {
		t.Errorf("merged.RunsPerTest = %+v, want %+v", merged.RunsPerTest, util.Ptr(2))
	}
	if !reflect.DeepEqual(merged.MaxRulesPerTestGen, util.Ptr(3)) {
		t.Errorf("merged.MaxRulesPerTestGen = %+v, want %+v", merged.MaxRulesPerTestGen, util.Ptr(3))
	}
	if !reflect.DeepEqual(merged.TestGenerations, util.Ptr(2)) {
		t.Errorf("merged.TestGenerations = %+v, want %+v", merged.TestGenerations, util.Ptr(2))
	}
}

func TestMergeOptions_PartialOverrides(t *testing.T) {
	// Test merging with partial overrides
	defaults := GetDefaultOptions()
	overrides := PromptPexOptions{
		Temperature:      util.Ptr(0.8),
		TestExpansions:   util.Ptr(5),
		DisableSafety:    util.Ptr(true),
		LoadContextFile:  util.Ptr("custom_context.json"),
		ModelsUnderTest:  []string{"model1", "model2"},
		EvalModels:       []string{"eval1", "eval2"},
		GroundtruthModel: util.Ptr("groundtruth_model"),
		Prompt:           util.Ptr("test_prompt"),
	}

	merged := MergeOptions(defaults, overrides)

	// Test overridden values
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"Temperature", merged.Temperature, util.Ptr(0.8)},
		{"TestExpansions", merged.TestExpansions, util.Ptr(5)},
		{"DisableSafety", merged.DisableSafety, util.Ptr(true)},
		{"LoadContextFile", merged.LoadContextFile, util.Ptr("custom_context.json")},
		{"ModelsUnderTest", merged.ModelsUnderTest, []string{"model1", "model2"}},
		{"EvalModels", merged.EvalModels, []string{"eval1", "eval2"}},
		{"GroundtruthModel", merged.GroundtruthModel, util.Ptr("groundtruth_model")},
		{"Prompt", merged.Prompt, util.Ptr("test_prompt")},
	}

	for _, tt := range tests {
		t.Run("override_"+tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.actual, tt.expected) {
				t.Errorf("merged.%s = %+v, want %+v", tt.name, tt.actual, tt.expected)
			}
		})
	}

	// Test that non-overridden values come from defaults
	defaultTests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"TestsPerRule", merged.TestsPerRule, defaults.TestsPerRule},
		{"RunsPerTest", merged.RunsPerTest, defaults.RunsPerTest},
		{"SplitRules", merged.SplitRules, defaults.SplitRules},
		{"MaxRulesPerTestGen", merged.MaxRulesPerTestGen, defaults.MaxRulesPerTestGen},
		{"TestGenerations", merged.TestGenerations, defaults.TestGenerations},
		{"FilterTestCount", merged.FilterTestCount, defaults.FilterTestCount},
		{"Evals", merged.Evals, defaults.Evals},
		{"Compliance", merged.Compliance, defaults.Compliance},
		{"BaselineTests", merged.BaselineTests, defaults.BaselineTests},
		{"StoreCompletions", merged.StoreCompletions, defaults.StoreCompletions},
		{"CreateEvalRuns", merged.CreateEvalRuns, defaults.CreateEvalRuns},
		{"RateTests", merged.RateTests, defaults.RateTests},
		{"EvalCache", merged.EvalCache, defaults.EvalCache},
		{"TestRunCache", merged.TestRunCache, defaults.TestRunCache},
		{"OutputPrompts", merged.OutputPrompts, defaults.OutputPrompts},
		{"WorkflowDiagram", merged.WorkflowDiagram, defaults.WorkflowDiagram},
		{"LoadContext", merged.LoadContext, defaults.LoadContext},
	}

	for _, tt := range defaultTests {
		t.Run("default_"+tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.actual, tt.expected) {
				t.Errorf("merged.%s = %+v, want %+v", tt.name, tt.actual, tt.expected)
			}
		})
	}
}

func TestMergeOptions_WithEffort(t *testing.T) {
	// Test merging options with effort configuration
	defaults := GetDefaultOptions()
	overrides := PromptPexOptions{
		Effort:      util.Ptr(EffortHigh),
		Temperature: util.Ptr(0.9),
		Evals:       util.Ptr(true),
	}

	merged := MergeOptions(defaults, overrides)

	// Test that effort was applied (checking some effort-specific values)
	if merged.TestExpansions == nil || *merged.TestExpansions != 1 {
		t.Errorf("merged.TestExpansions = %+v, want %d (from EffortHigh)", merged.TestExpansions, 1)
	}
	if merged.MaxRules == nil || *merged.MaxRules != 50 {
		t.Errorf("merged.MaxRules = %+v, want %d (from EffortHigh)", merged.MaxRules, 50)
	}
	if merged.SplitRules == nil || !*merged.SplitRules {
		t.Errorf("merged.SplitRules = %+v, want %t (from EffortHigh)", merged.SplitRules, true)
	}

	// Test that explicit overrides still take precedence over effort
	if !reflect.DeepEqual(merged.Temperature, util.Ptr(0.9)) {
		t.Errorf("merged.Temperature = %+v, want %+v (explicit override)", merged.Temperature, util.Ptr(0.9))
	}
	if !reflect.DeepEqual(merged.Evals, util.Ptr(true)) {
		t.Errorf("merged.Evals = %+v, want %+v (explicit override)", merged.Evals, util.Ptr(true))
	}

	// Test that defaults are still applied for non-effort, non-override fields
	if !reflect.DeepEqual(merged.Compliance, defaults.Compliance) {
		t.Errorf("merged.Compliance = %+v, want %+v (from defaults)", merged.Compliance, defaults.Compliance)
	}
}

func TestMergeOptions_NilValues(t *testing.T) {
	// Test merging with nil values in various combinations
	tests := []struct {
		name                string
		defaultTemperature  *float64
		overrideTemperature *float64
		expectedTemperature *float64
	}{
		{
			name:                "both nil",
			defaultTemperature:  nil,
			overrideTemperature: nil,
			expectedTemperature: nil,
		},
		{
			name:                "default set, override nil",
			defaultTemperature:  util.Ptr(0.5),
			overrideTemperature: nil,
			expectedTemperature: util.Ptr(0.5),
		},
		{
			name:                "default nil, override set",
			defaultTemperature:  nil,
			overrideTemperature: util.Ptr(0.8),
			expectedTemperature: util.Ptr(0.8),
		},
		{
			name:                "both set",
			defaultTemperature:  util.Ptr(0.5),
			overrideTemperature: util.Ptr(0.8),
			expectedTemperature: util.Ptr(0.8),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaults := PromptPexOptions{Temperature: tt.defaultTemperature}
			overrides := PromptPexOptions{Temperature: tt.overrideTemperature}

			merged := MergeOptions(defaults, overrides)

			if !reflect.DeepEqual(merged.Temperature, tt.expectedTemperature) {
				t.Errorf("merged.Temperature = %+v, want %+v", merged.Temperature, tt.expectedTemperature)
			}
		})
	}
}

func TestMergeOptions_AllFields(t *testing.T) {
	// Comprehensive test covering all fields in PromptPexOptions
	defaults := PromptPexOptions{
		Temperature:        util.Ptr(0.1),
		TestsPerRule:       util.Ptr(1),
		RunsPerTest:        util.Ptr(1),
		SplitRules:         util.Ptr(false),
		MaxRulesPerTestGen: util.Ptr(1),
		TestGenerations:    util.Ptr(1),
		TestExpansions:     util.Ptr(1),
		FilterTestCount:    util.Ptr(1),
		Evals:              util.Ptr(false),
		Compliance:         util.Ptr(false),
		BaselineTests:      util.Ptr(false),
		StoreCompletions:   util.Ptr(false),
		CreateEvalRuns:     util.Ptr(false),
		RateTests:          util.Ptr(false),
		DisableSafety:      util.Ptr(false),
		EvalCache:          util.Ptr(false),
		TestRunCache:       util.Ptr(false),
		OutputPrompts:      util.Ptr(false),
		WorkflowDiagram:    util.Ptr(false),
		LoadContext:        util.Ptr(false),
		LoadContextFile:    util.Ptr("default.json"),
		MaxRules:           util.Ptr(1),
		MaxTestsToRun:      util.Ptr(1),
		ModelsUnderTest:    []string{"default_model"},
		EvalModels:         []string{"default_eval"},
		GroundtruthModel:   util.Ptr("default_groundtruth"),
		Prompt:             util.Ptr("default_prompt"),
	}

	overrides := PromptPexOptions{
		Temperature:        util.Ptr(0.9),
		TestsPerRule:       util.Ptr(10),
		RunsPerTest:        util.Ptr(5),
		SplitRules:         util.Ptr(true),
		MaxRulesPerTestGen: util.Ptr(20),
		TestGenerations:    util.Ptr(3),
		TestExpansions:     util.Ptr(2),
		FilterTestCount:    util.Ptr(15),
		Evals:              util.Ptr(true),
		Compliance:         util.Ptr(true),
		BaselineTests:      util.Ptr(true),
		StoreCompletions:   util.Ptr(true),
		CreateEvalRuns:     util.Ptr(true),
		RateTests:          util.Ptr(true),
		DisableSafety:      util.Ptr(true),
		EvalCache:          util.Ptr(true),
		TestRunCache:       util.Ptr(true),
		OutputPrompts:      util.Ptr(true),
		WorkflowDiagram:    util.Ptr(true),
		LoadContext:        util.Ptr(true),
		LoadContextFile:    util.Ptr("override.json"),
		MaxRules:           util.Ptr(100),
		MaxTestsToRun:      util.Ptr(50),
		ModelsUnderTest:    []string{"override_model1", "override_model2"},
		EvalModels:         []string{"override_eval1", "override_eval2"},
		GroundtruthModel:   util.Ptr("override_groundtruth"),
		Prompt:             util.Ptr("override_prompt"),
	}

	merged := MergeOptions(defaults, overrides)

	// All fields should match the overrides since they are all set
	if !reflect.DeepEqual(merged, overrides) {
		t.Errorf("MergeOptions with all overrides set should equal overrides")
	}
}

func TestMergeOptions_SliceFields(t *testing.T) {
	// Test specific behavior for slice fields
	defaults := PromptPexOptions{
		ModelsUnderTest: []string{"default1", "default2"},
		EvalModels:      []string{"eval_default"},
	}

	overrides := PromptPexOptions{
		ModelsUnderTest: []string{"override1", "override2", "override3"},
		// EvalModels intentionally not set
	}

	merged := MergeOptions(defaults, overrides)

	// Override slice should replace default slice completely
	expectedModels := []string{"override1", "override2", "override3"}
	if !reflect.DeepEqual(merged.ModelsUnderTest, expectedModels) {
		t.Errorf("merged.ModelsUnderTest = %+v, want %+v", merged.ModelsUnderTest, expectedModels)
	}

	// Default slice should be preserved when not overridden
	expectedEvalModels := []string{"eval_default"}
	if !reflect.DeepEqual(merged.EvalModels, expectedEvalModels) {
		t.Errorf("merged.EvalModels = %+v, want %+v", merged.EvalModels, expectedEvalModels)
	}
}

func TestMergeOptions_EmptySlices(t *testing.T) {
	// Test behavior with empty slices vs nil slices
	defaults := PromptPexOptions{
		ModelsUnderTest: []string{"default1", "default2"},
		EvalModels:      nil, // nil slice
	}

	overrides := PromptPexOptions{
		ModelsUnderTest: []string{}, // empty slice
		EvalModels:      []string{"override_eval"},
	}

	merged := MergeOptions(defaults, overrides)

	// Empty slice should override default slice
	if merged.ModelsUnderTest == nil || len(merged.ModelsUnderTest) != 0 {
		t.Errorf("merged.ModelsUnderTest = %+v, want empty slice", merged.ModelsUnderTest)
	}

	// Non-nil override should replace nil default
	expectedEvalModels := []string{"override_eval"}
	if !reflect.DeepEqual(merged.EvalModels, expectedEvalModels) {
		t.Errorf("merged.EvalModels = %+v, want %+v", merged.EvalModels, expectedEvalModels)
	}
}

// Helper function tests
func Testutil.Ptr(t *testing.T) {
	tests := []bool{true, false}

	for _, val := range tests {
		ptr := util.Ptr(val)
		if ptr == nil {
			t.Errorf("util.Ptr(%t) returned nil", val)
		}
		if *ptr != val {
			t.Errorf("util.Ptr(%t) = %t, want %t", val, *ptr, val)
		}
	}
}

func Testutil.Ptr(t *testing.T) {
	tests := []int{0, 1, -1, 100, -100}

	for _, val := range tests {
		ptr := util.Ptr(val)
		if ptr == nil {
			t.Errorf("util.Ptr(%d) returned nil", val)
		}
		if *ptr != val {
			t.Errorf("util.Ptr(%d) = %d, want %d", val, *ptr, val)
		}
	}
}

func Testutil.Ptr(t *testing.T) {
	tests := []float64{0.0, 1.0, -1.0, 3.14159, -2.71828}

	for _, val := range tests {
		ptr := util.Ptr(val)
		if ptr == nil {
			t.Errorf("util.Ptr(%f) returned nil", val)
		}
		if *ptr != val {
			t.Errorf("util.Ptr(%f) = %f, want %f", val, *ptr, val)
		}
	}
}

func Testutil.Ptr(t *testing.T) {
	tests := []string{"", "hello", "world", "test string with spaces", "special!@#$%^&*()chars"}

	for _, val := range tests {
		ptr := util.Ptr(val)
		if ptr == nil {
			t.Errorf("util.Ptr(%q) returned nil", val)
		}
		if *ptr != val {
			t.Errorf("util.Ptr(%q) = %q, want %q", val, *ptr, val)
		}
	}
}

// Test the GetOptions method if we can access generateCommandHandler
func TestGetOptions(t *testing.T) {
	// This test assumes we can create a generateCommandHandler for testing
	// If the struct is not accessible for testing, this test can be removed
	handler := &generateCommandHandler{
		options: PromptPexOptions{
			Temperature:  util.Ptr(0.5),
			TestsPerRule: util.Ptr(7),
		},
	}

	options := handler.GetOptions()

	if !reflect.DeepEqual(options.Temperature, util.Ptr(0.5)) {
		t.Errorf("GetOptions().Temperature = %+v, want %+v", options.Temperature, util.Ptr(0.5))
	}
	if !reflect.DeepEqual(options.TestsPerRule, util.Ptr(7)) {
		t.Errorf("GetOptions().TestsPerRule = %+v, want %+v", options.TestsPerRule, util.Ptr(7))
	}
}
