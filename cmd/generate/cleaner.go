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
