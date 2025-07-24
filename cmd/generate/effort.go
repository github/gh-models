package generate

import "github.com/github/gh-models/pkg/util"

// EffortConfiguration defines the configuration for different effort levels
type EffortConfiguration struct {
	TestGenerations           *int  `json:"testGenerations,omitempty"`
	TestsPerRule              *int  `json:"testsPerRule,omitempty"`
	RunsPerTest               *int  `json:"runsPerTest,omitempty"`
	TestExpansions            *int  `json:"testExpansions,omitempty"`
	MaxRules                  *int  `json:"maxRules,omitempty"`
	MaxRulesPerTestGeneration *int  `json:"maxRulesPerTestGeneration,omitempty"`
	MaxTestsToRun             *int  `json:"maxTestsToRun,omitempty"`
	Compliance                *bool `json:"compliance,omitempty"`
}

// GetEffortConfiguration returns the configuration for a given effort level
// Based on the reference TypeScript implementation in constants.mts
func GetEffortConfiguration(effort string) *EffortConfiguration {
	switch effort {
	case EffortMin:
		return &EffortConfiguration{
			TestGenerations:           util.Ptr(1),
			TestsPerRule:              util.Ptr(1),
			RunsPerTest:               util.Ptr(1),
			TestExpansions:            util.Ptr(0),
			MaxRules:                  util.Ptr(6),
			MaxRulesPerTestGeneration: util.Ptr(100),
			MaxTestsToRun:             util.Ptr(10),
			Compliance:                util.Ptr(false),
		}
	case EffortLow:
		return &EffortConfiguration{
			TestExpansions:            util.Ptr(0),
			TestGenerations:           util.Ptr(1),
			MaxRules:                  util.Ptr(3),
			TestsPerRule:              util.Ptr(2),
			RunsPerTest:               util.Ptr(1),
			MaxRulesPerTestGeneration: util.Ptr(5),
			MaxTestsToRun:             util.Ptr(20),
		}
	case EffortMedium:
		return &EffortConfiguration{
			TestExpansions:            util.Ptr(0),
			MaxRules:                  util.Ptr(20),
			TestsPerRule:              util.Ptr(3),
			RunsPerTest:               util.Ptr(1),
			MaxRulesPerTestGeneration: util.Ptr(5),
			TestGenerations:           util.Ptr(1),
		}
	case EffortHigh:
		return &EffortConfiguration{
			TestExpansions:            util.Ptr(1),
			MaxRules:                  util.Ptr(50),
			MaxRulesPerTestGeneration: util.Ptr(2),
			TestGenerations:           util.Ptr(2),
		}
	default:
		return nil
	}
}

// ApplyEffortConfiguration applies effort configuration to options
func ApplyEffortConfiguration(options *PromptPexOptions, effort string) {
	if options == nil || effort == "" {
		return
	}

	config := GetEffortConfiguration(effort)
	if config == nil {
		return
	}

	// Apply configuration settings only if not already set
	if config.TestGenerations != nil && options.TestGenerations == nil {
		options.TestGenerations = config.TestGenerations
	}
	if config.TestsPerRule != nil && options.TestsPerRule == nil {
		options.TestsPerRule = config.TestsPerRule
	}
	if config.RunsPerTest != nil && options.RunsPerTest == nil {
		options.RunsPerTest = config.RunsPerTest
	}
	if config.TestExpansions != nil && options.TestExpansions == nil {
		options.TestExpansions = config.TestExpansions
	}
	if config.MaxRules != nil && options.MaxRules == nil {
		options.MaxRules = config.MaxRules
	}
	if config.MaxRulesPerTestGeneration != nil && options.MaxRulesPerTestGen == nil {
		options.MaxRulesPerTestGen = config.MaxRulesPerTestGeneration
	}
	if config.MaxTestsToRun != nil && options.MaxTestsToRun == nil {
		options.MaxTestsToRun = config.MaxTestsToRun
	}
}
