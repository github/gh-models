package generate

// EffortConfiguration defines the configuration for different effort levels
type EffortConfiguration struct {
	TestsPerRule              int
	RunsPerTest               int
	MaxRules                  int
	MaxRulesPerTestGeneration int
	RulesPerGen               int
}

// GetEffortConfiguration returns the configuration for a given effort level
// Based on the reference TypeScript implementation in constants.mts
func GetEffortConfiguration(effort string) *EffortConfiguration {
	switch effort {
	case EffortLow:
		return &EffortConfiguration{
			MaxRules:                  3,
			TestsPerRule:              2,
			RunsPerTest:               1,
			MaxRulesPerTestGeneration: 5,
			RulesPerGen:               10,
		}
	case EffortMedium:
		return &EffortConfiguration{
			MaxRules:                  20,
			TestsPerRule:              3,
			RunsPerTest:               1,
			MaxRulesPerTestGeneration: 5,
			RulesPerGen:               5,
		}
	case EffortHigh:
		return &EffortConfiguration{
			MaxRules:                  50,
			MaxRulesPerTestGeneration: 2,
			RulesPerGen:               3,
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
	if options.TestsPerRule == 0 {
		options.TestsPerRule = config.TestsPerRule
	}
	if options.RunsPerTest == 0 {
		options.RunsPerTest = config.RunsPerTest
	}
	if options.MaxRules == 0 {
		options.MaxRules = config.MaxRules
	}
	if options.MaxRulesPerTestGen == 0 {
		options.MaxRulesPerTestGen = config.MaxRulesPerTestGeneration
	}
	if options.RulesPerGen == 0 {
		options.RulesPerGen = config.RulesPerGen
	}
}
