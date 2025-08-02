package generate

import (
	"regexp"
	"strings"
)

// ExtractJSON extracts JSON content from a string that might be wrapped in markdown
func ExtractJSON(content string) string {
	// Remove markdown code blocks
	content = strings.TrimSpace(content)

	// Remove ```json and ``` markers
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}

	content = strings.TrimSpace(content)

	// Clean up JavaScript string concatenation syntax
	content = cleanJavaScriptStringConcat(content)

	// If it starts with [ or {, likely valid JSON
	if strings.HasPrefix(content, "[") || strings.HasPrefix(content, "{") {
		return content
	}

	// Find JSON array or object with more robust regex
	jsonPattern := regexp.MustCompile(`(\[[\s\S]*\]|\{[\s\S]*\})`)
	matches := jsonPattern.FindString(content)
	if matches != "" {
		return cleanJavaScriptStringConcat(matches)
	}

	return content
}

// cleanJavaScriptStringConcat removes JavaScript string concatenation syntax from JSON
func cleanJavaScriptStringConcat(content string) string {
	// Remove JavaScript comments first
	commentPattern := regexp.MustCompile(`//[^\n]*`)
	content = commentPattern.ReplaceAllString(content, "")

	// Handle complex JavaScript expressions that look like: "A" + "B" * 1998
	// Replace with a simple fallback string
	complexExprPattern := regexp.MustCompile(`"([^"]*)"[ \t]*\+[ \t]*"([^"]*)"[ \t]*\*[ \t]*\d+`)
	content = complexExprPattern.ReplaceAllString(content, `"${1}${2}_repeated"`)

	// Find and fix JavaScript string concatenation (e.g., "text" + "more text")
	// This is a common issue when LLMs generate JSON with JS-style string concatenation
	concatPattern := regexp.MustCompile(`"([^"]*)"[ \t]*\+[ \t\n]*"([^"]*)"`)
	for concatPattern.MatchString(content) {
		content = concatPattern.ReplaceAllString(content, `"$1$2"`)
	}

	// Handle multiline concatenation
	multilinePattern := regexp.MustCompile(`"([^"]*)"[ \t]*\+[ \t]*\n[ \t]*"([^"]*)"`)
	for multilinePattern.MatchString(content) {
		content = multilinePattern.ReplaceAllString(content, `"$1$2"`)
	}

	return content
}

// StringSliceContains checks if a string slice contains a value
func StringSliceContains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// MergeStringMaps merges multiple string maps, with later maps taking precedence
func MergeStringMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
