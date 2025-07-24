package generate

import "github.com/github/gh-models/pkg/util"

// GetDefaultOptions returns default options for PromptPex
func GetDefaultOptions() *PromptPexOptions {
	return &PromptPexOptions{
		Temperature:        util.Ptr(0.0),
		Models:             &PromptPexModelAliases{},
		TestsPerRule:       util.Ptr(3),
		RunsPerTest:        util.Ptr(2),
		MaxRulesPerTestGen: util.Ptr(3),
		TestGenerations:    util.Ptr(2),
		TestExpansions:     util.Ptr(0),
		FilterTestCount:    util.Ptr(5),
		Evals:              util.Ptr(false),
		Compliance:         util.Ptr(false),
		Verbose:            util.Ptr(false),
	}
}

// GetOptions returns the current options for testing purposes
func (h *generateCommandHandler) GetOptions() *PromptPexOptions {
	return h.options
}
