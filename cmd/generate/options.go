package generate

// GetDefaultOptions returns default options for PromptPex
func GetDefaultOptions() *PromptPexOptions {
	return &PromptPexOptions{
		TestsPerRule:       3,
		RulesPerGen:        3,
		MaxRulesPerTestGen: 3,
		Verbose:            false,
		IntentMaxTokens:    100,
		InputSpecMaxTokens: 500,
		Models: &PromptPexModelAliases{
			Rules:       "openai/gpt-4o",
			Tests:       "openai/gpt-4o",
			Groundtruth: "openai/gpt-4o",
			Eval:        "openai/gpt-4o",
		},
	}
}
