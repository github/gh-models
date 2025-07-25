package generate

// GetDefaultOptions returns default options for PromptPex
func GetDefaultOptions() *PromptPexOptions {
	return &PromptPexOptions{
		TestsPerRule:       3,
		RunsPerTest:        2,
		MaxRulesPerTestGen: 3,
		Verbose:            false,
		Models: &PromptPexModelAliases{
			Rules:       "openai/gpt-4o",
			Tests:       "openai/gpt-4o",
			Groundtruth: "openai/gpt-4o",
		},
	}
}

// GetOptions returns the current options for testing purposes
func (h *generateCommandHandler) GetOptions() *PromptPexOptions {
	return h.options
}
