package generate

import (
	"reflect"
	"testing"
)

func TestGetDefaultOptions(t *testing.T) {
	defaults := GetDefaultOptions()

	// Test individual fields to ensure they have expected default values
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"Temperature", defaults.Temperature, Float64Ptr(0.0)},
		{"TestsPerRule", defaults.TestsPerRule, IntPtr(3)},
		{"RunsPerTest", defaults.RunsPerTest, IntPtr(2)},
		{"SplitRules", defaults.SplitRules, BoolPtr(true)},
		{"MaxRulesPerTestGen", defaults.MaxRulesPerTestGen, IntPtr(3)},
		{"TestGenerations", defaults.TestGenerations, IntPtr(2)},
		{"TestExpansions", defaults.TestExpansions, IntPtr(0)},
		{"FilterTestCount", defaults.FilterTestCount, IntPtr(5)},
		{"Evals", defaults.Evals, BoolPtr(false)},
		{"Compliance", defaults.Compliance, BoolPtr(false)},
		{"BaselineTests", defaults.BaselineTests, BoolPtr(false)},
		{"StoreCompletions", defaults.StoreCompletions, BoolPtr(false)},
		{"CreateEvalRuns", defaults.CreateEvalRuns, BoolPtr(false)},
		{"RateTests", defaults.RateTests, BoolPtr(false)},
		{"DisableSafety", defaults.DisableSafety, BoolPtr(false)},
		{"EvalCache", defaults.EvalCache, BoolPtr(false)},
		{"TestRunCache", defaults.TestRunCache, BoolPtr(false)},
		{"OutputPrompts", defaults.OutputPrompts, BoolPtr(false)},
		{"WorkflowDiagram", defaults.WorkflowDiagram, BoolPtr(true)},
		{"LoadContext", defaults.LoadContext, BoolPtr(false)},
		{"LoadContextFile", defaults.LoadContextFile, StringPtr("promptpex_context.json")},
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
		Temperature:  Float64Ptr(1.0),
		TestsPerRule: IntPtr(5),
		SplitRules:   BoolPtr(false),
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
		Temperature:        Float64Ptr(0.0),
		TestsPerRule:       IntPtr(3),
		RunsPerTest:        IntPtr(2),
		SplitRules:         BoolPtr(true),
		MaxRulesPerTestGen: IntPtr(3),
		TestGenerations:    IntPtr(2),
		Evals:              BoolPtr(false),
		WorkflowDiagram:    BoolPtr(true),
	}

	overrides := PromptPexOptions{
		Temperature:     Float64Ptr(1.5),
		TestsPerRule:    IntPtr(10),
		SplitRules:      BoolPtr(false),
		Evals:           BoolPtr(true),
		WorkflowDiagram: BoolPtr(false),
	}

	merged := MergeOptions(defaults, overrides)

	// Test that overridden values take precedence
	if !reflect.DeepEqual(merged.Temperature, Float64Ptr(1.5)) {
		t.Errorf("merged.Temperature = %+v, want %+v", merged.Temperature, Float64Ptr(1.5))
	}
	if !reflect.DeepEqual(merged.TestsPerRule, IntPtr(10)) {
		t.Errorf("merged.TestsPerRule = %+v, want %+v", merged.TestsPerRule, IntPtr(10))
	}
	if !reflect.DeepEqual(merged.SplitRules, BoolPtr(false)) {
		t.Errorf("merged.SplitRules = %+v, want %+v", merged.SplitRules, BoolPtr(false))
	}
	if !reflect.DeepEqual(merged.Evals, BoolPtr(true)) {
		t.Errorf("merged.Evals = %+v, want %+v", merged.Evals, BoolPtr(true))
	}
	if !reflect.DeepEqual(merged.WorkflowDiagram, BoolPtr(false)) {
		t.Errorf("merged.WorkflowDiagram = %+v, want %+v", merged.WorkflowDiagram, BoolPtr(false))
	}

	// Test that non-overridden values come from defaults
	if !reflect.DeepEqual(merged.RunsPerTest, IntPtr(2)) {
		t.Errorf("merged.RunsPerTest = %+v, want %+v", merged.RunsPerTest, IntPtr(2))
	}
	if !reflect.DeepEqual(merged.MaxRulesPerTestGen, IntPtr(3)) {
		t.Errorf("merged.MaxRulesPerTestGen = %+v, want %+v", merged.MaxRulesPerTestGen, IntPtr(3))
	}
	if !reflect.DeepEqual(merged.TestGenerations, IntPtr(2)) {
		t.Errorf("merged.TestGenerations = %+v, want %+v", merged.TestGenerations, IntPtr(2))
	}
}

func TestMergeOptions_PartialOverrides(t *testing.T) {
	// Test merging with partial overrides
	defaults := GetDefaultOptions()
	overrides := PromptPexOptions{
		Temperature:      Float64Ptr(0.8),
		TestExpansions:   IntPtr(5),
		DisableSafety:    BoolPtr(true),
		LoadContextFile:  StringPtr("custom_context.json"),
		ModelsUnderTest:  []string{"model1", "model2"},
		EvalModels:       []string{"eval1", "eval2"},
		GroundtruthModel: StringPtr("groundtruth_model"),
		Prompt:           StringPtr("test_prompt"),
	}

	merged := MergeOptions(defaults, overrides)

	// Test overridden values
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"Temperature", merged.Temperature, Float64Ptr(0.8)},
		{"TestExpansions", merged.TestExpansions, IntPtr(5)},
		{"DisableSafety", merged.DisableSafety, BoolPtr(true)},
		{"LoadContextFile", merged.LoadContextFile, StringPtr("custom_context.json")},
		{"ModelsUnderTest", merged.ModelsUnderTest, []string{"model1", "model2"}},
		{"EvalModels", merged.EvalModels, []string{"eval1", "eval2"}},
		{"GroundtruthModel", merged.GroundtruthModel, StringPtr("groundtruth_model")},
		{"Prompt", merged.Prompt, StringPtr("test_prompt")},
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
		Effort:      StringPtr(EffortHigh),
		Temperature: Float64Ptr(0.9),
		Evals:       BoolPtr(true),
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
	if !reflect.DeepEqual(merged.Temperature, Float64Ptr(0.9)) {
		t.Errorf("merged.Temperature = %+v, want %+v (explicit override)", merged.Temperature, Float64Ptr(0.9))
	}
	if !reflect.DeepEqual(merged.Evals, BoolPtr(true)) {
		t.Errorf("merged.Evals = %+v, want %+v (explicit override)", merged.Evals, BoolPtr(true))
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
			defaultTemperature:  Float64Ptr(0.5),
			overrideTemperature: nil,
			expectedTemperature: Float64Ptr(0.5),
		},
		{
			name:                "default nil, override set",
			defaultTemperature:  nil,
			overrideTemperature: Float64Ptr(0.8),
			expectedTemperature: Float64Ptr(0.8),
		},
		{
			name:                "both set",
			defaultTemperature:  Float64Ptr(0.5),
			overrideTemperature: Float64Ptr(0.8),
			expectedTemperature: Float64Ptr(0.8),
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
		Temperature:        Float64Ptr(0.1),
		TestsPerRule:       IntPtr(1),
		RunsPerTest:        IntPtr(1),
		SplitRules:         BoolPtr(false),
		MaxRulesPerTestGen: IntPtr(1),
		TestGenerations:    IntPtr(1),
		TestExpansions:     IntPtr(1),
		FilterTestCount:    IntPtr(1),
		Evals:              BoolPtr(false),
		Compliance:         BoolPtr(false),
		BaselineTests:      BoolPtr(false),
		StoreCompletions:   BoolPtr(false),
		CreateEvalRuns:     BoolPtr(false),
		RateTests:          BoolPtr(false),
		DisableSafety:      BoolPtr(false),
		EvalCache:          BoolPtr(false),
		TestRunCache:       BoolPtr(false),
		OutputPrompts:      BoolPtr(false),
		WorkflowDiagram:    BoolPtr(false),
		LoadContext:        BoolPtr(false),
		LoadContextFile:    StringPtr("default.json"),
		MaxRules:           IntPtr(1),
		MaxTestsToRun:      IntPtr(1),
		ModelsUnderTest:    []string{"default_model"},
		EvalModels:         []string{"default_eval"},
		GroundtruthModel:   StringPtr("default_groundtruth"),
		Prompt:             StringPtr("default_prompt"),
	}

	overrides := PromptPexOptions{
		Temperature:        Float64Ptr(0.9),
		TestsPerRule:       IntPtr(10),
		RunsPerTest:        IntPtr(5),
		SplitRules:         BoolPtr(true),
		MaxRulesPerTestGen: IntPtr(20),
		TestGenerations:    IntPtr(3),
		TestExpansions:     IntPtr(2),
		FilterTestCount:    IntPtr(15),
		Evals:              BoolPtr(true),
		Compliance:         BoolPtr(true),
		BaselineTests:      BoolPtr(true),
		StoreCompletions:   BoolPtr(true),
		CreateEvalRuns:     BoolPtr(true),
		RateTests:          BoolPtr(true),
		DisableSafety:      BoolPtr(true),
		EvalCache:          BoolPtr(true),
		TestRunCache:       BoolPtr(true),
		OutputPrompts:      BoolPtr(true),
		WorkflowDiagram:    BoolPtr(true),
		LoadContext:        BoolPtr(true),
		LoadContextFile:    StringPtr("override.json"),
		MaxRules:           IntPtr(100),
		MaxTestsToRun:      IntPtr(50),
		ModelsUnderTest:    []string{"override_model1", "override_model2"},
		EvalModels:         []string{"override_eval1", "override_eval2"},
		GroundtruthModel:   StringPtr("override_groundtruth"),
		Prompt:             StringPtr("override_prompt"),
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
func TestBoolPtr(t *testing.T) {
	tests := []bool{true, false}

	for _, val := range tests {
		ptr := BoolPtr(val)
		if ptr == nil {
			t.Errorf("BoolPtr(%t) returned nil", val)
		}
		if *ptr != val {
			t.Errorf("BoolPtr(%t) = %t, want %t", val, *ptr, val)
		}
	}
}

func TestIntPtr(t *testing.T) {
	tests := []int{0, 1, -1, 100, -100}

	for _, val := range tests {
		ptr := IntPtr(val)
		if ptr == nil {
			t.Errorf("IntPtr(%d) returned nil", val)
		}
		if *ptr != val {
			t.Errorf("IntPtr(%d) = %d, want %d", val, *ptr, val)
		}
	}
}

func TestFloat64Ptr(t *testing.T) {
	tests := []float64{0.0, 1.0, -1.0, 3.14159, -2.71828}

	for _, val := range tests {
		ptr := Float64Ptr(val)
		if ptr == nil {
			t.Errorf("Float64Ptr(%f) returned nil", val)
		}
		if *ptr != val {
			t.Errorf("Float64Ptr(%f) = %f, want %f", val, *ptr, val)
		}
	}
}

func TestStringPtr(t *testing.T) {
	tests := []string{"", "hello", "world", "test string with spaces", "special!@#$%^&*()chars"}

	for _, val := range tests {
		ptr := StringPtr(val)
		if ptr == nil {
			t.Errorf("StringPtr(%q) returned nil", val)
		}
		if *ptr != val {
			t.Errorf("StringPtr(%q) = %q, want %q", val, *ptr, val)
		}
	}
}

// Test the GetOptions method if we can access generateCommandHandler
func TestGetOptions(t *testing.T) {
	// This test assumes we can create a generateCommandHandler for testing
	// If the struct is not accessible for testing, this test can be removed
	handler := &generateCommandHandler{
		options: PromptPexOptions{
			Temperature:  Float64Ptr(0.5),
			TestsPerRule: IntPtr(7),
		},
	}

	options := handler.GetOptions()

	if !reflect.DeepEqual(options.Temperature, Float64Ptr(0.5)) {
		t.Errorf("GetOptions().Temperature = %+v, want %+v", options.Temperature, Float64Ptr(0.5))
	}
	if !reflect.DeepEqual(options.TestsPerRule, IntPtr(7)) {
		t.Errorf("GetOptions().TestsPerRule = %+v, want %+v", options.TestsPerRule, IntPtr(7))
	}
}
