// Package prompt provides shared types and utilities for working with .prompt.yml files
package prompt

import (
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// File represents the structure of a .prompt.yml file
type File struct {
	Name            string          `yaml:"name"`
	Description     string          `yaml:"description"`
	Model           string          `yaml:"model"`
	ModelParameters ModelParameters `yaml:"modelParameters"`
	Messages        []Message       `yaml:"messages"`
	// TestData and Evaluators are only used by eval command
	TestData   []map[string]interface{} `yaml:"testData,omitempty"`
	Evaluators []Evaluator              `yaml:"evaluators,omitempty"`
}

// ModelParameters represents model configuration parameters
type ModelParameters struct {
	MaxTokens   *int     `yaml:"maxTokens"`
	Temperature *float64 `yaml:"temperature"`
	TopP        *float64 `yaml:"topP"`
}

// Message represents a conversation message
type Message struct {
	Role    string `yaml:"role"`
	Content string `yaml:"content"`
}

// Evaluator represents an evaluation method (only used by eval command)
type Evaluator struct {
	Name   string           `yaml:"name"`
	String *StringEvaluator `yaml:"string,omitempty"`
	LLM    *LLMEvaluator    `yaml:"llm,omitempty"`
	Uses   string           `yaml:"uses,omitempty"`
}

// StringEvaluator represents string-based evaluation
type StringEvaluator struct {
	EndsWith   string `yaml:"endsWith,omitempty"`
	StartsWith string `yaml:"startsWith,omitempty"`
	Contains   string `yaml:"contains,omitempty"`
	Equals     string `yaml:"equals,omitempty"`
}

// LLMEvaluator represents LLM-based evaluation
type LLMEvaluator struct {
	ModelID      string   `yaml:"modelId"`
	Prompt       string   `yaml:"prompt"`
	Choices      []Choice `yaml:"choices"`
	SystemPrompt string   `yaml:"systemPrompt,omitempty"`
}

// Choice represents a scoring choice for LLM evaluation
type Choice struct {
	Choice string  `yaml:"choice"`
	Score  float64 `yaml:"score"`
}

// LoadFromFile loads and parses a prompt file from the given path
func LoadFromFile(filePath string) (*File, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var promptFile File
	if err := yaml.Unmarshal(data, &promptFile); err != nil {
		return nil, err
	}

	return &promptFile, nil
}

// TemplateString templates a string with the given data using Go's text/template
func TemplateString(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("template").Option("missingkey=zero").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", err
	}

	// Replace "<no value>" with empty string for missing template variables
	resultStr := result.String()
	resultStr = strings.ReplaceAll(resultStr, "<no value>", "")

	return resultStr, nil
}
