package generate

import "github.com/github/gh-models/pkg/prompt"

// PromptPexModelAliases represents model aliases for different purposes
type PromptPexModelAliases string

const (
	ModelAliasRules       PromptPexModelAliases = "rules"
	ModelAliasEval        PromptPexModelAliases = "eval"
	ModelAliasLarge       PromptPexModelAliases = "large"
	ModelAliasBaseline    PromptPexModelAliases = "baseline"
	ModelAliasGroundtruth PromptPexModelAliases = "groundtruth"
)

// PromptPexPrompts contains custom prompts for different stages
type PromptPexPrompts struct {
	InputSpec          *string `yaml:"inputSpec,omitempty" json:"inputSpec,omitempty"`
	OutputRules        *string `yaml:"outputRules,omitempty" json:"outputRules,omitempty"`
	InverseOutputRules *string `yaml:"inverseOutputRules,omitempty" json:"inverseOutputRules,omitempty"`
	Intent             *string `yaml:"intent,omitempty" json:"intent,omitempty"`
	TestExpansion      *string `yaml:"testExpansion,omitempty" json:"testExpansion,omitempty"`
}

// WorkspaceFile represents a file in the workspace
type WorkspaceFile struct {
	Filename string `json:"filename" yaml:"filename"`
	Content  string `json:"content" yaml:"content"`
}

// PromptPexOptions contains all configuration options for PromptPex
type PromptPexOptions struct {
	// Core options
	Temperature           *float64                         `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	OutputPrompts         *bool                            `yaml:"outputPrompts,omitempty" json:"outputPrompts,omitempty"`
	WorkflowDiagram       *bool                            `yaml:"workflowDiagram,omitempty" json:"workflowDiagram,omitempty"`
	Instructions          *PromptPexPrompts                `yaml:"instructions,omitempty" json:"instructions,omitempty"`
	ModelAliases          map[PromptPexModelAliases]string `yaml:"modelAliases,omitempty" json:"modelAliases,omitempty"`
	EvalCache             *bool                            `yaml:"evalCache,omitempty" json:"evalCache,omitempty"`
	Evals                 *bool                            `yaml:"evals,omitempty" json:"evals,omitempty"`
	TestRunCache          *bool                            `yaml:"testRunCache,omitempty" json:"testRunCache,omitempty"`
	RulesModel            *string                          `yaml:"rulesModel,omitempty" json:"rulesModel,omitempty"`
	StoreModel            *string                          `yaml:"storeModel,omitempty" json:"storeModel,omitempty"`
	GroundtruthModel      *string                          `yaml:"groundtruthModel,omitempty" json:"groundtruthModel,omitempty"`
	BaselineModel         *string                          `yaml:"baselineModel,omitempty" json:"baselineModel,omitempty"`
	TestsPerRule          *int                             `yaml:"testsPerRule,omitempty" json:"testsPerRule,omitempty"`
	RunsPerTest           *int                             `yaml:"runsPerTest,omitempty" json:"runsPerTest,omitempty"`
	Compliance            *bool                            `yaml:"compliance,omitempty" json:"compliance,omitempty"`
	BaselineTests         *bool                            `yaml:"baselineTests,omitempty" json:"baselineTests,omitempty"`
	MaxTestsToRun         *int                             `yaml:"maxTestsToRun,omitempty" json:"maxTestsToRun,omitempty"`
	MaxRules              *int                             `yaml:"maxRules,omitempty" json:"maxRules,omitempty"`
	Cache                 interface{}                      `yaml:"cache,omitempty" json:"cache,omitempty"` // can be bool or string
	StoreCompletions      *bool                            `yaml:"storeCompletions,omitempty" json:"storeCompletions,omitempty"`
	ModelsUnderTest       []string                         `yaml:"modelsUnderTest,omitempty" json:"modelsUnderTest,omitempty"`
	SplitRules            *bool                            `yaml:"splitRules,omitempty" json:"splitRules,omitempty"`
	MaxRulesPerTestGen    *int                             `yaml:"maxRulesPerTestGeneration,omitempty" json:"maxRulesPerTestGeneration,omitempty"`
	TestGenerations       *int                             `yaml:"testGenerations,omitempty" json:"testGenerations,omitempty"`
	CreateEvalRuns        *bool                            `yaml:"createEvalRuns,omitempty" json:"createEvalRuns,omitempty"`
	TestExpansions        *int                             `yaml:"testExpansions,omitempty" json:"testExpansions,omitempty"`
	RateTests             *bool                            `yaml:"rateTests,omitempty" json:"rateTests,omitempty"`
	FilterTestCount       *int                             `yaml:"filterTestCount,omitempty" json:"filterTestCount,omitempty"`
	EvalModels            []string                         `yaml:"evalModels,omitempty" json:"evalModels,omitempty"`
	EvalModelsGroundtruth []string                         `yaml:"evalModelsGroundtruth,omitempty" json:"evalModelsGroundtruth,omitempty"`

	// CLI-specific options
	Effort                         *string `yaml:"effort,omitempty" json:"effort,omitempty"`
	CustomMetric                   *string `yaml:"customMetric,omitempty" json:"customMetric,omitempty"`
	Prompt                         *string `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	InputSpecInstructions          *string `yaml:"inputSpecInstructions,omitempty" json:"inputSpecInstructions,omitempty"`
	OutputRulesInstructions        *string `yaml:"outputRulesInstructions,omitempty" json:"outputRulesInstructions,omitempty"`
	InverseOutputRulesInstructions *string `yaml:"inverseOutputRulesInstructions,omitempty" json:"inverseOutputRulesInstructions,omitempty"`
	TestExpansionInstructions      *string `yaml:"testExpansionInstructions,omitempty" json:"testExpansionInstructions,omitempty"`

	// Loader options
	DisableSafety      *bool   `yaml:"disableSafety,omitempty" json:"disableSafety,omitempty"`
	TestSamplesCount   *int    `yaml:"testSamplesCount,omitempty" json:"testSamplesCount,omitempty"`
	TestSamplesShuffle *bool   `yaml:"testSamplesShuffle,omitempty" json:"testSamplesShuffle,omitempty"`
	LoadContext        *bool   `yaml:"loadContext,omitempty" json:"loadContext,omitempty"`
	LoadContextFile    *string `yaml:"loadContextFile,omitempty" json:"loadContextFile,omitempty"`
	Verbose            *bool   `yaml:"verbose,omitempty" json:"verbose,omitempty"`
}

// PromptPexTestGenerationScenario represents a test generation scenario
type PromptPexTestGenerationScenario struct {
	Name         string                 `yaml:"name" json:"name"`
	Instructions *string                `yaml:"instructions,omitempty" json:"instructions,omitempty"`
	Parameters   map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// PromptPexPromptyFrontmatter represents the frontmatter of a prompty file
type PromptPexPromptyFrontmatter struct {
	Name         *string                           `yaml:"name,omitempty" json:"name,omitempty"`
	Description  *string                           `yaml:"description,omitempty" json:"description,omitempty"`
	Tags         []string                          `yaml:"tags,omitempty" json:"tags,omitempty"`
	Inputs       map[string]interface{}            `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Outputs      map[string]interface{}            `yaml:"outputs,omitempty" json:"outputs,omitempty"`
	Instructions *PromptPexPrompts                 `yaml:"instructions,omitempty" json:"instructions,omitempty"`
	Scenarios    []PromptPexTestGenerationScenario `yaml:"scenarios,omitempty" json:"scenarios,omitempty"`
	TestSamples  []interface{}                     `yaml:"testSamples,omitempty" json:"testSamples,omitempty"`
	Imported     map[string]interface{}            `yaml:"imported,omitempty" json:"imported,omitempty"`
}

// PromptPexContext represents the main context for PromptPex operations
type PromptPexContext struct {
	RunID             string                   `json:"runId" yaml:"runId"`
	WriteResults      *bool                    `json:"writeResults,omitempty" yaml:"writeResults,omitempty"`
	Prompt            *prompt.File             `json:"prompt" yaml:"prompt"`
	Intent            string                   `json:"intent" yaml:"intent"`
	Rules             string                   `json:"rules" yaml:"rules"`
	InverseRules      string                   `json:"inverseRules" yaml:"inverseRules"`
	InputSpec         string                   `json:"inputSpec" yaml:"inputSpec"`
	BaselineTests     string                   `json:"baselineTests" yaml:"baselineTests"`
	Tests             string                   `json:"tests" yaml:"tests"`
	PromptPexTests    []PromptPexTest          `json:"promptPexTests" yaml:"promptPexTests"`
	TestData          string                   `json:"testData" yaml:"testData"`
	RateTests         string                   `json:"rateTests" yaml:"rateTests"`
	TestOutputs       string                   `json:"testOutputs" yaml:"testOutputs"`
	TestEvals         string                   `json:"testEvals" yaml:"testEvals"`
	RuleEvals         string                   `json:"ruleEvals" yaml:"ruleEvals"`
	RuleCoverages     string                   `json:"ruleCoverages" yaml:"ruleCoverages"`
	BaselineTestEvals string                   `json:"baselineTestEvals" yaml:"baselineTestEvals"`
	TestSamples       []map[string]interface{} `json:"testSamples,omitempty" yaml:"testSamples,omitempty"`
	ReuseResults      *bool                    `json:"reuseResults,omitempty" yaml:"reuseResults,omitempty"`
	Options           *PromptPexOptions        `json:"options" yaml:"options"`
}

// PromptPexTest represents a single test case
type PromptPexTest struct {
	RuleID            *int     `json:"ruleid,omitempty" yaml:"ruleid,omitempty"`
	TestID            *int     `json:"testid,omitempty" yaml:"testid,omitempty"`
	Baseline          *bool    `json:"baseline,omitempty" yaml:"baseline,omitempty"`
	GroundtruthModel  *string  `json:"groundtruthModel,omitempty" yaml:"groundtruthModel,omitempty"`
	Groundtruth       *string  `json:"groundtruth,omitempty" yaml:"groundtruth,omitempty"`
	GroundtruthScore  *float64 `json:"groundtruthScore,omitempty" yaml:"groundtruthScore,omitempty"`
	TestInput         string   `json:"testinput" yaml:"testinput"`
	TestInputOriginal *string  `json:"testinputOriginal,omitempty" yaml:"testinputOriginal,omitempty"`
	ExpectedOutput    *string  `json:"expectedoutput,omitempty" yaml:"expectedoutput,omitempty"`
	Reasoning         *string  `json:"reasoning,omitempty" yaml:"reasoning,omitempty"`
	Scenario          *string  `json:"scenario,omitempty" yaml:"scenario,omitempty"`
	Generation        *int     `json:"generation,omitempty" yaml:"generation,omitempty"`
}

// PromptPexEvalResultType represents the result of an evaluation
type PromptPexEvalResultType string

const (
	EvalResultOK      PromptPexEvalResultType = "ok"
	EvalResultError   PromptPexEvalResultType = "err"
	EvalResultUnknown PromptPexEvalResultType = "unknown"
)

// PromptPexEvaluation represents an evaluation result
type PromptPexEvaluation struct {
	Content     string                   `json:"content" yaml:"content"`
	Uncertainty *float64                 `json:"uncertainty,omitempty" yaml:"uncertainty,omitempty"`
	Perplexity  *float64                 `json:"perplexity,omitempty" yaml:"perplexity,omitempty"`
	Outcome     *PromptPexEvalResultType `json:"outcome,omitempty" yaml:"outcome,omitempty"`
	Score       *float64                 `json:"score,omitempty" yaml:"score,omitempty"`
}

// PromptPexTestResult represents the result of running a test
type PromptPexTestResult struct {
	ID               string                         `json:"id" yaml:"id"`
	PromptID         string                         `json:"promptid" yaml:"promptid"`
	RuleID           int                            `json:"ruleid" yaml:"ruleid"`
	Rule             string                         `json:"rule" yaml:"rule"`
	Scenario         string                         `json:"scenario" yaml:"scenario"`
	TestInput        string                         `json:"testinput" yaml:"testinput"`
	Inverse          *bool                          `json:"inverse,omitempty" yaml:"inverse,omitempty"`
	Baseline         *bool                          `json:"baseline,omitempty" yaml:"baseline,omitempty"`
	Model            string                         `json:"model" yaml:"model"`
	Input            string                         `json:"input" yaml:"input"`
	Output           string                         `json:"output" yaml:"output"`
	Error            *string                        `json:"error,omitempty" yaml:"error,omitempty"`
	IsGroundtruth    *bool                          `json:"isGroundtruth,omitempty" yaml:"isGroundtruth,omitempty"`
	Groundtruth      *string                        `json:"groundtruth,omitempty" yaml:"groundtruth,omitempty"`
	GroundtruthModel *string                        `json:"groundtruthModel,omitempty" yaml:"groundtruthModel,omitempty"`
	Compliance       *PromptPexEvalResultType       `json:"compliance,omitempty" yaml:"compliance,omitempty"`
	ComplianceText   *string                        `json:"complianceText,omitempty" yaml:"complianceText,omitempty"`
	Metrics          map[string]PromptPexEvaluation `json:"metrics" yaml:"metrics"`
}

// PromptPexTestEval represents test evaluation results
type PromptPexTestEval struct {
	ID                  string                   `json:"id" yaml:"id"`
	PromptID            string                   `json:"promptid" yaml:"promptid"`
	Model               *string                  `json:"model,omitempty" yaml:"model,omitempty"`
	Rule                string                   `json:"rule" yaml:"rule"`
	Inverse             *bool                    `json:"inverse,omitempty" yaml:"inverse,omitempty"`
	Input               string                   `json:"input" yaml:"input"`
	Coverage            *PromptPexEvalResultType `json:"coverage,omitempty" yaml:"coverage,omitempty"`
	CoverageEvalText    *string                  `json:"coverageEvalText,omitempty" yaml:"coverageEvalText,omitempty"`
	CoverageText        *string                  `json:"coverageText,omitempty" yaml:"coverageText,omitempty"`
	CoverageUncertainty *float64                 `json:"coverageUncertainty,omitempty" yaml:"coverageUncertainty,omitempty"`
	Validity            *PromptPexEvalResultType `json:"validity,omitempty" yaml:"validity,omitempty"`
	ValidityText        *string                  `json:"validityText,omitempty" yaml:"validityText,omitempty"`
	ValidityUncertainty *float64                 `json:"validityUncertainty,omitempty" yaml:"validityUncertainty,omitempty"`
	Error               *string                  `json:"error,omitempty" yaml:"error,omitempty"`
}

// PromptPexRule represents a rule
type PromptPexRule struct {
	Rule    string `json:"rule" yaml:"rule"`
	Inverse *bool  `json:"inverse,omitempty" yaml:"inverse,omitempty"`
}

// PromptPexRuleEval represents rule evaluation results
type PromptPexRuleEval struct {
	ID           string                   `json:"id" yaml:"id"`
	PromptID     string                   `json:"promptid" yaml:"promptid"`
	RuleID       int                      `json:"ruleid" yaml:"ruleid"`
	Rule         string                   `json:"rule" yaml:"rule"`
	GroundedText *string                  `json:"groundedText,omitempty" yaml:"groundedText,omitempty"`
	Grounded     *PromptPexEvalResultType `json:"grounded,omitempty" yaml:"grounded,omitempty"`
	Error        *string                  `json:"error,omitempty" yaml:"error,omitempty"`
}

// PromptPexConstants contains constant values used throughout the application
type PromptPexConstants struct {
	PromptPexContext string
	ModelAliasRules  string
	ModelAliasStore  string
}

var Constants = PromptPexConstants{
	PromptPexContext: "promptpex_context.json",
	ModelAliasRules:  "rules",
	ModelAliasStore:  "store",
}

// Effort levels
const (
	EffortMin    = "min"
	EffortLow    = "low"
	EffortMedium = "medium"
	EffortHigh   = "high"
)
