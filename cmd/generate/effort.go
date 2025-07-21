package generate

// EffortConfiguration defines the configuration for different effort levels
type EffortConfiguration struct {
	SplitRules                *bool `json:"splitRules,omitempty"`
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
			SplitRules:                BoolPtr(false),
			TestGenerations:           IntPtr(1),
			TestsPerRule:              IntPtr(1),
			RunsPerTest:               IntPtr(1),
			TestExpansions:            IntPtr(0),
			MaxRules:                  IntPtr(6),
			MaxRulesPerTestGeneration: IntPtr(100),
			MaxTestsToRun:             IntPtr(10),
			Compliance:                BoolPtr(false),
		}
	case EffortLow:
		return &EffortConfiguration{
			TestExpansions:            IntPtr(0),
			TestGenerations:           IntPtr(1),
			MaxRules:                  IntPtr(3),
			TestsPerRule:              IntPtr(2),
			RunsPerTest:               IntPtr(1),
			MaxRulesPerTestGeneration: IntPtr(5),
			SplitRules:                BoolPtr(true),
			MaxTestsToRun:             IntPtr(20),
		}
	case EffortMedium:
		return &EffortConfiguration{
			TestExpansions:            IntPtr(0),
			MaxRules:                  IntPtr(20),
			TestsPerRule:              IntPtr(3),
			RunsPerTest:               IntPtr(1),
			MaxRulesPerTestGeneration: IntPtr(5),
			SplitRules:                BoolPtr(true),
			TestGenerations:           IntPtr(1),
		}
	case EffortHigh:
		return &EffortConfiguration{
			TestExpansions:            IntPtr(1),
			MaxRules:                  IntPtr(50),
			MaxRulesPerTestGeneration: IntPtr(2),
			SplitRules:                BoolPtr(true),
			TestGenerations:           IntPtr(2),
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
	if config.SplitRules != nil && options.SplitRules == nil {
		options.SplitRules = config.SplitRules
	}
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
	if config.Compliance != nil && options.Compliance == nil {
		options.Compliance = config.Compliance
	}
}
