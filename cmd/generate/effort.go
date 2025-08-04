package generate

// EffortConfiguration defines the configuration for different effort levels
type EffortConfiguration struct {
	MaxRules     int
	TestsPerRule int
	RulesPerGen  int
}

// GetEffortConfiguration returns the configuration for a given effort level
// Based on the reference TypeScript implementation in constants.mts
func GetEffortConfiguration(effort string) *EffortConfiguration {
	switch effort {
	case EffortMin:
		return &EffortConfiguration{
			MaxRules:     3,
			TestsPerRule: 1,
			RulesPerGen:  100,
		}
	case EffortLow:
		return &EffortConfiguration{
			MaxRules:     10,
			TestsPerRule: 1,
			RulesPerGen:  10,
		}
	case EffortMedium:
		return &EffortConfiguration{
			MaxRules:     20,
			TestsPerRule: 3,
			RulesPerGen:  5,
		}
	case EffortHigh:
		return &EffortConfiguration{
			TestsPerRule: 4,
			RulesPerGen:  3,
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

	effortConfig := GetEffortConfiguration(effort)
	if effortConfig == nil {
		return
	}
	// Apply effort if set
	if effortConfig.TestsPerRule != 0 {
		options.TestsPerRule = effortConfig.TestsPerRule
	}
	if effortConfig.MaxRules != 0 {
		options.MaxRules = effortConfig.MaxRules
	}
	if effortConfig.RulesPerGen != 0 {
		options.RulesPerGen = effortConfig.RulesPerGen
	}
}
