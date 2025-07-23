package generate

import "github.com/github/gh-models/pkg/util"

// GetDefaultOptions returns default options for PromptPex
func GetDefaultOptions() PromptPexOptions {
	return PromptPexOptions{
		Temperature:        util.Ptr(0.0),
		TestsPerRule:       util.Ptr(3),
		RunsPerTest:        util.Ptr(2),
		SplitRules:         util.Ptr(true),
		MaxRulesPerTestGen: util.Ptr(3),
		TestGenerations:    util.Ptr(2),
		TestExpansions:     util.Ptr(0),
		FilterTestCount:    util.Ptr(5),
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
		WorkflowDiagram:    util.Ptr(true),
		LoadContext:        util.Ptr(false),
		LoadContextFile:    util.Ptr("promptpex_context.json"),
		Verbose:            util.Ptr(false),
	}
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
	if result.Verbose == nil && defaults.Verbose != nil {
		result.Verbose = defaults.Verbose
	}
	if result.MaxRules == nil && defaults.MaxRules != nil {
		result.MaxRules = defaults.MaxRules
	}
	if result.MaxTestsToRun == nil && defaults.MaxTestsToRun != nil {
		result.MaxTestsToRun = defaults.MaxTestsToRun
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
