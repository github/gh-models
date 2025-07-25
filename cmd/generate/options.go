package generate

import "github.com/github/gh-models/pkg/util"

// GetDefaultOptions returns default options for PromptPex
func GetDefaultOptions() *PromptPexOptions {
	return &PromptPexOptions{
		TestsPerRule:       util.Ptr(3),
		RunsPerTest:        util.Ptr(2),
		MaxRulesPerTestGen: util.Ptr(3),
		Verbose:            util.Ptr(false),
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
