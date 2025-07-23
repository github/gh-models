package generate

import (
	"fmt"
	"strings"

	"github.com/github/gh-models/pkg/prompt"
)

// RenderMessagesToString converts a slice of Messages to a human-readable string representation
func RenderMessagesToString(messages []prompt.Message) string {
	if len(messages) == 0 {
		return ""
	}

	var builder strings.Builder

	for i, msg := range messages {
		// Add role header
		roleUpper := strings.ToUpper(msg.Role)
		builder.WriteString(fmt.Sprintf("[%s]\n", roleUpper))

		// Add content with proper indentation
		content := strings.TrimSpace(msg.Content)
		if content != "" {
			// Split content into lines and indent each line
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				builder.WriteString(fmt.Sprintf("%s\n", line))
			}
		}

		// Add separator between messages (except for the last one)
		if i < len(messages)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}
