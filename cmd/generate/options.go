package generate

// GetDefaultOptions returns default options for PromptPex
func GetDefaultOptions() PromptPexOptions {
	return PromptPexOptions{
		Temperature:        Float64Ptr(0.0),
		TestsPerRule:       IntPtr(3),
		RunsPerTest:        IntPtr(2),
		SplitRules:         BoolPtr(true),
		MaxRulesPerTestGen: IntPtr(3),
		TestGenerations:    IntPtr(2),
		TestExpansions:     IntPtr(0),
		FilterTestCount:    IntPtr(5),
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
		WorkflowDiagram:    BoolPtr(true),
		LoadContext:        BoolPtr(false),
		LoadContextFile:    StringPtr("promptpex_context.json"),
	}
}

// Helper functions to create pointers
func BoolPtr(b bool) *bool {
	return &b
}

func IntPtr(i int) *int {
	return &i
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func StringPtr(s string) *string {
	return &s
}

// GetOptions returns the current options for testing purposes
func (h *generateCommandHandler) GetOptions() PromptPexOptions {
	return h.options
}

// mergeOptions merges two option structs, with the second taking precedence
func MergeOptions(defaults PromptPexOptions, overrides PromptPexOptions) PromptPexOptions {
	// Start with overrides as the base
	result := overrides

	// Apply effort configuration first, only to fields not explicitly set in overrides
	if overrides.Effort != nil {
		ApplyEffortConfiguration(&result, *overrides.Effort)
	}

	// Then apply defaults for any fields still not set
	if result.Temperature == nil && defaults.Temperature != nil {
		result.Temperature = defaults.Temperature
	}
	if result.TestsPerRule == nil && defaults.TestsPerRule != nil {
		result.TestsPerRule = defaults.TestsPerRule
	}
	if result.RunsPerTest == nil && defaults.RunsPerTest != nil {
		result.RunsPerTest = defaults.RunsPerTest
	}
	if result.SplitRules == nil && defaults.SplitRules != nil {
		result.SplitRules = defaults.SplitRules
	}
	if result.MaxRulesPerTestGen == nil && defaults.MaxRulesPerTestGen != nil {
		result.MaxRulesPerTestGen = defaults.MaxRulesPerTestGen
	}
	if result.TestGenerations == nil && defaults.TestGenerations != nil {
		result.TestGenerations = defaults.TestGenerations
	}
	if result.TestExpansions == nil && defaults.TestExpansions != nil {
		result.TestExpansions = defaults.TestExpansions
	}
	if result.FilterTestCount == nil && defaults.FilterTestCount != nil {
		result.FilterTestCount = defaults.FilterTestCount
	}
	if result.Evals == nil && defaults.Evals != nil {
		result.Evals = defaults.Evals
	}
	if result.Compliance == nil && defaults.Compliance != nil {
		result.Compliance = defaults.Compliance
	}
	if result.BaselineTests == nil && defaults.BaselineTests != nil {
		result.BaselineTests = defaults.BaselineTests
	}
	if result.StoreCompletions == nil && defaults.StoreCompletions != nil {
		result.StoreCompletions = defaults.StoreCompletions
	}
	if result.CreateEvalRuns == nil && defaults.CreateEvalRuns != nil {
		result.CreateEvalRuns = defaults.CreateEvalRuns
	}
	if result.RateTests == nil && defaults.RateTests != nil {
		result.RateTests = defaults.RateTests
	}
	if result.DisableSafety == nil && defaults.DisableSafety != nil {
		result.DisableSafety = defaults.DisableSafety
	}
	if result.EvalCache == nil && defaults.EvalCache != nil {
		result.EvalCache = defaults.EvalCache
	}
	if result.TestRunCache == nil && defaults.TestRunCache != nil {
		result.TestRunCache = defaults.TestRunCache
	}
	if result.OutputPrompts == nil && defaults.OutputPrompts != nil {
		result.OutputPrompts = defaults.OutputPrompts
	}
	if result.WorkflowDiagram == nil && defaults.WorkflowDiagram != nil {
		result.WorkflowDiagram = defaults.WorkflowDiagram
	}
	if result.LoadContext == nil && defaults.LoadContext != nil {
		result.LoadContext = defaults.LoadContext
	}
	if result.LoadContextFile == nil && defaults.LoadContextFile != nil {
		result.LoadContextFile = defaults.LoadContextFile
	}
	if result.MaxRules == nil && defaults.MaxRules != nil {
		result.MaxRules = defaults.MaxRules
	}
	if result.MaxTestsToRun == nil && defaults.MaxTestsToRun != nil {
		result.MaxTestsToRun = defaults.MaxTestsToRun
	}
	if result.Out == nil && defaults.Out != nil {
		result.Out = defaults.Out
	}
	if result.ModelsUnderTest == nil && defaults.ModelsUnderTest != nil {
		result.ModelsUnderTest = defaults.ModelsUnderTest
	}
	if result.EvalModels == nil && defaults.EvalModels != nil {
		result.EvalModels = defaults.EvalModels
	}
	if result.GroundtruthModel == nil && defaults.GroundtruthModel != nil {
		result.GroundtruthModel = defaults.GroundtruthModel
	}
	if result.Prompt == nil && defaults.Prompt != nil {
		result.Prompt = defaults.Prompt
	}

	return result
}
