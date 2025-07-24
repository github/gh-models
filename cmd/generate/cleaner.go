package generate

import (
	"regexp"
	"strings"
)

// IsUnassistedResponse returns true if the text is an unassisted response, like "i'm sorry" or "i can't assist with that".
func IsUnassistedResponse(text string) bool {
	re := regexp.MustCompile(`i can't assist with that|i'm sorry`)
	return re.MatchString(strings.ToLower(text))
}

// unfence removes code fences and splits text into lines.
func Unfence(text string) string {
	text = strings.TrimSpace(text)
	// Remove triple backtick code fences if present
	if strings.HasPrefix(text, "```") {
		parts := strings.SplitN(text, "\n", 2)
		if len(parts) == 2 {
			text = parts[1]
		}
		text = strings.TrimSuffix(text, "```")
	}
	return text
}

// splits text into lines.
func SplitLines(text string) []string {
	lines := strings.Split(text, "\n")
	return lines
}

func UnBacket(text string) string {
	// Remove leading and trailing square brackets
	if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") {
		text = strings.TrimPrefix(text, "[")
		text = strings.TrimSuffix(text, "]")
	}
	return text
}

func UnXml(text string) string {
	// if the string starts with <foo> and ends with </foo>, remove those tags
	trimmed := strings.TrimSpace(text)

	// Use regex to extract tag name and content
	// First, extract the opening tag and tag name
	openTagRe := regexp.MustCompile(`(?s)^<([^>\s]+)[^>]*>(.*)$`)
	openMatches := openTagRe.FindStringSubmatch(trimmed)
	if len(openMatches) != 3 {
		return text
	}

	tagName := openMatches[1]
	content := openMatches[2]

	// Check if it ends with the corresponding closing tag
	closingTag := "</" + tagName + ">"
	if strings.HasSuffix(content, closingTag) {
		content = strings.TrimSuffix(content, closingTag)
		return strings.TrimSpace(content)
	}

	return text
}
