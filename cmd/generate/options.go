package generate

import "github.com/github/gh-models/pkg/util"

// GetDefaultOptions returns default options for PromptPex
func GetDefaultOptions() *PromptPexOptions {
	return &PromptPexOptions{
		Temperature:        util.Ptr(0.0),
		TestsPerRule:       util.Ptr(3),
		RunsPerTest:        util.Ptr(2),
		MaxRulesPerTestGen: util.Ptr(3),
		TestGenerations:    util.Ptr(2),
		TestExpansions:     util.Ptr(0),
		FilterTestCount:    util.Ptr(5),
		Verbose:            util.Ptr(false),
		Models: &PromptPexModelAliases{
			Rules:       util.Ptr("openai/gpt-4o"),
			Tests:       util.Ptr("openai/gpt-4o"),
			Groundtruth: util.Ptr("openai/gpt-4o"),
		},
	}
}

// GetOptions returns the current options for testing purposes
func (h *generateCommandHandler) GetOptions() *PromptPexOptions {
	return h.options
}
