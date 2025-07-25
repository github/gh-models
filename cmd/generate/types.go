package generate

import "github.com/github/gh-models/pkg/prompt"

// PromptPexModelAliases represents model aliases for different purposes
type PromptPexModelAliases struct {
	Rules       string `yaml:"rules,omitempty" json:"rules,omitempty"`
	Tests       string `yaml:"tests,omitempty" json:"tests,omitempty"`
	Groundtruth string `yaml:"groundtruth,omitempty" json:"groundtruth,omitempty"`
}

// PromptPexPrompts contains custom prompts for different stages
type PromptPexPrompts struct {
	InputSpec          string `yaml:"inputSpec,omitempty" json:"inputSpec,omitempty"`
	OutputRules        string `yaml:"outputRules,omitempty" json:"outputRules,omitempty"`
	InverseOutputRules string `yaml:"inverseOutputRules,omitempty" json:"inverseOutputRules,omitempty"`
	Intent             string `yaml:"intent,omitempty" json:"intent,omitempty"`
}

// PromptPexOptions contains all configuration options for PromptPex
type PromptPexOptions struct {
	// Core options
	Instructions       *PromptPexPrompts      `yaml:"instructions,omitempty" json:"instructions,omitempty"`
	Models             *PromptPexModelAliases `yaml:"models,omitempty" json:"models,omitempty"`
	TestsPerRule       int                    `yaml:"testsPerRule,omitempty" json:"testsPerRule,omitempty"`
	RunsPerTest        int                    `yaml:"runsPerTest,omitempty" json:"runsPerTest,omitempty"`
	MaxRules           int                    `yaml:"maxRules,omitempty" json:"maxRules,omitempty"`
	MaxRulesPerTestGen int                    `yaml:"maxRulesPerTestGeneration,omitempty" json:"maxRulesPerTestGeneration,omitempty"`

	// CLI-specific options
	Effort string `yaml:"effort,omitempty" json:"effort,omitempty"`
	Prompt string `yaml:"prompt,omitempty" json:"prompt,omitempty"`

	// Loader options
	Verbose bool `yaml:"verbose,omitempty" json:"verbose,omitempty"`
}

// PromptPexContext represents the main context for PromptPex operations
type PromptPexContext struct {
	RunID        string            `json:"runId" yaml:"runId"`
	Prompt       *prompt.File      `json:"prompt" yaml:"prompt"`
	PromptHash   string            `json:"promptHash" yaml:"promptHash"`
	Options      *PromptPexOptions `json:"options" yaml:"options"`
	Intent       *string           `json:"intent" yaml:"intent"`
	Rules        []string          `json:"rules" yaml:"rules"`
	InverseRules []string          `json:"inverseRules" yaml:"inverseRules"`
	InputSpec    *string           `json:"inputSpec" yaml:"inputSpec"`
	Tests        []PromptPexTest   `json:"tests" yaml:"tests"`
}

// PromptPexTest represents a single test case
type PromptPexTest struct {
	RuleID            int    `json:"ruleid,omitempty" yaml:"ruleid,omitempty"`
	TestID            int    `json:"testid,omitempty" yaml:"testid,omitempty"`
	Baseline          bool   `json:"baseline,omitempty" yaml:"baseline,omitempty"`
	GroundtruthModel  string `json:"groundtruthModel,omitempty" yaml:"groundtruthModel,omitempty"`
	Groundtruth       string `json:"groundtruth,omitempty" yaml:"groundtruth,omitempty"`
	TestInput         string `json:"testinput" yaml:"testinput"`
	TestInputOriginal string `json:"testinputOriginal,omitempty" yaml:"testinputOriginal,omitempty"`
	ExpectedOutput    string `json:"expectedoutput,omitempty" yaml:"expectedoutput,omitempty"`
	Reasoning         string `json:"reasoning,omitempty" yaml:"reasoning,omitempty"`
	Scenario          string `json:"scenario,omitempty" yaml:"scenario,omitempty"`
}

// Effort levels
const (
	EffortLow    = "low"
	EffortMedium = "medium"
	EffortHigh   = "high"
)
