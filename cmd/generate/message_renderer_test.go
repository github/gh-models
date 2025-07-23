package generate

import (
	"strings"
	"testing"

	"github.com/github/gh-models/pkg/prompt"
)

func TestRenderMessagesToString(t *testing.T) {
	tests := []struct {
		name     string
		messages []prompt.Message
		expected string
	}{
		{
			name:     "empty messages",
			messages: []prompt.Message{},
			expected: "",
		},
		{
			name: "single system message",
			messages: []prompt.Message{
				{Role: "system", Content: "You are a helpful assistant."},
			},
			expected: "[SYSTEM]\nYou are a helpful assistant.\n",
		},
		{
			name: "single user message",
			messages: []prompt.Message{
				{Role: "user", Content: "Hello, how are you?"},
			},
			expected: "[USER]\nHello, how are you?\n",
		},
		{
			name: "single assistant message",
			messages: []prompt.Message{
				{Role: "assistant", Content: "I'm doing well, thank you!"},
			},
			expected: "[ASSISTANT]\nI'm doing well, thank you!\n",
		},
		{
			name: "multiple messages",
			messages: []prompt.Message{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "What is 2+2?"},
				{Role: "assistant", Content: "2+2 equals 4."},
			},
			expected: "[SYSTEM]\nYou are a helpful assistant.\n\n[USER]\nWhat is 2+2?\n\n[ASSISTANT]\n2+2 equals 4.\n",
		},
		{
			name: "message with empty content",
			messages: []prompt.Message{
				{Role: "user", Content: ""},
			},
			expected: "[USER]\n",
		},
		{
			name: "message with whitespace only content",
			messages: []prompt.Message{
				{Role: "user", Content: "   \n\t  "},
			},
			expected: "[USER]\n",
		},
		{
			name: "message with multiline content",
			messages: []prompt.Message{
				{Role: "user", Content: "This is line 1\nThis is line 2\nThis is line 3"},
			},
			expected: "[USER]\nThis is line 1\nThis is line 2\nThis is line 3\n",
		},
		{
			name: "message with leading and trailing whitespace",
			messages: []prompt.Message{
				{Role: "user", Content: "  \n  Hello world  \n  "},
			},
			expected: "[USER]\nHello world\n",
		},
		{
			name: "mixed roles and content types",
			messages: []prompt.Message{
				{Role: "system", Content: "You are a code assistant."},
				{Role: "user", Content: "Write a function:\n\nfunc add(a, b int) int {\n    return a + b\n}"},
				{Role: "assistant", Content: "Here's the function you requested."},
			},
			expected: "[SYSTEM]\nYou are a code assistant.\n\n[USER]\nWrite a function:\n\nfunc add(a, b int) int {\n    return a + b\n}\n\n[ASSISTANT]\nHere's the function you requested.\n",
		},
		{
			name: "lowercase role names",
			messages: []prompt.Message{
				{Role: "system", Content: "System message"},
				{Role: "user", Content: "User message"},
				{Role: "assistant", Content: "Assistant message"},
			},
			expected: "[SYSTEM]\nSystem message\n\n[USER]\nUser message\n\n[ASSISTANT]\nAssistant message\n",
		},
		{
			name: "uppercase role names",
			messages: []prompt.Message{
				{Role: "SYSTEM", Content: "System message"},
				{Role: "USER", Content: "User message"},
				{Role: "ASSISTANT", Content: "Assistant message"},
			},
			expected: "[SYSTEM]\nSystem message\n\n[USER]\nUser message\n\n[ASSISTANT]\nAssistant message\n",
		},
		{
			name: "mixed case role names",
			messages: []prompt.Message{
				{Role: "System", Content: "System message"},
				{Role: "User", Content: "User message"},
				{Role: "Assistant", Content: "Assistant message"},
			},
			expected: "[SYSTEM]\nSystem message\n\n[USER]\nUser message\n\n[ASSISTANT]\nAssistant message\n",
		},
		{
			name: "custom role name",
			messages: []prompt.Message{
				{Role: "custom", Content: "Custom role message"},
			},
			expected: "[CUSTOM]\nCustom role message\n",
		},
		{
			name: "message with only newlines",
			messages: []prompt.Message{
				{Role: "user", Content: "\n\n\n"},
			},
			expected: "[USER]\n",
		},
		{
			name: "message with mixed whitespace and content",
			messages: []prompt.Message{
				{Role: "user", Content: "\n  Hello  \n\n  World  \n"},
			},
			expected: "[USER]\nHello  \n\n  World\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderMessagesToString(tt.messages)
			if result != tt.expected {
				t.Errorf("renderMessagesToString() = %q, expected %q", result, tt.expected)

				// Print detailed comparison for debugging
				t.Logf("Expected lines:")
				for i, line := range strings.Split(tt.expected, "\n") {
					t.Logf("  %d: %q", i, line)
				}
				t.Logf("Actual lines:")
				for i, line := range strings.Split(result, "\n") {
					t.Logf("  %d: %q", i, line)
				}
			}
		})
	}
}

func TestRenderMessagesToString_EdgeCases(t *testing.T) {
	t.Run("nil messages slice", func(t *testing.T) {
		var messages []prompt.Message
		result := RenderMessagesToString(messages)
		if result != "" {
			t.Errorf("renderMessagesToString(nil) = %q, expected empty string", result)
		}
	})

	t.Run("single message with very long content", func(t *testing.T) {
		longContent := strings.Repeat("This is a very long line of text. ", 100)
		messages := []prompt.Message{
			{Role: "user", Content: longContent},
		}
		result := RenderMessagesToString(messages)
		expected := "[USER]\n" + strings.TrimSpace(longContent) + "\n"
		if result != expected {
			t.Errorf("renderMessagesToString() failed with long content")
		}
	})

	t.Run("message with unicode characters", func(t *testing.T) {
		messages := []prompt.Message{
			{Role: "user", Content: "Hello 🌍! How are you? 你好 مرحبا"},
		}
		result := RenderMessagesToString(messages)
		expected := "[USER]\nHello 🌍! How are you? 你好 مرحبا\n"
		if result != expected {
			t.Errorf("renderMessagesToString() = %q, expected %q", result, expected)
		}
	})

	t.Run("message with special characters", func(t *testing.T) {
		messages := []prompt.Message{
			{Role: "user", Content: "Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?`~"},
		}
		result := RenderMessagesToString(messages)
		expected := "[USER]\nSpecial chars: !@#$%^&*()_+-=[]{}|;':\",./<>?`~\n"
		if result != expected {
			t.Errorf("renderMessagesToString() = %q, expected %q", result, expected)
		}
	})
}
