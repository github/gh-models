package generate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsUnassistedResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "detects 'i can't assist with that' lowercase",
			input:    "i can't assist with that request",
			expected: true,
		},
		{
			name:     "detects 'i can't assist with that' mixed case",
			input:    "I Can't Assist With That Request",
			expected: true,
		},
		{
			name:     "detects 'i'm sorry' lowercase",
			input:    "i'm sorry, but i cannot help",
			expected: true,
		},
		{
			name:     "detects 'i'm sorry' mixed case",
			input:    "I'm Sorry, But I Cannot Help",
			expected: true,
		},
		{
			name:     "detects phrase within larger text",
			input:    "Unfortunately, I can't assist with that particular request. Please try something else.",
			expected: true,
		},
		{
			name:     "detects 'i'm sorry' within larger text",
			input:    "Well, I'm sorry to say this but I cannot proceed.",
			expected: true,
		},
		{
			name:     "returns false for regular response",
			input:    "Here is the code you requested",
			expected: false,
		},
		{
			name:     "returns false for empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "returns false for similar but different phrases",
			input:    "i can assist with that",
			expected: false,
		},
		{
			name:     "returns false for partial matches",
			input:    "sorry for the delay",
			expected: false,
		},
		{
			name:     "handles apostrophe variations",
			input:    "i can't assist with that",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUnassistedResponse(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestUnfence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes code fences with language",
			input:    "```go\npackage main\nfunc main() {}\n```",
			expected: "package main\nfunc main() {}\n",
		},
		{
			name:     "removes code fences without language",
			input:    "```\nsome code\nmore code\n```",
			expected: "some code\nmore code\n",
		},
		{
			name:     "handles text without code fences",
			input:    "just plain text",
			expected: "just plain text",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles whitespace around text",
			input:    "  \n  some text  \n  ",
			expected: "some text",
		},
		{
			name:     "handles only opening fence",
			input:    "```go\ncode without closing",
			expected: "code without closing",
		},
		{
			name:     "handles fence with no content",
			input:    "```\n```",
			expected: "",
		},
		{
			name:     "handles fence with only language - no newline",
			input:    "```python",
			expected: "```python",
		},
		{
			name:     "preserves content that looks like fences but isn't at start",
			input:    "some text\n```\nmore text",
			expected: "some text\n```\nmore text",
		},
		{
			name:     "handles multiple lines after fence",
			input:    "```javascript\nfunction test() {\n  return 'hello';\n}\nconsole.log('world');\n```",
			expected: "function test() {\n  return 'hello';\n}\nconsole.log('world');\n",
		},
		{
			name:     "handles single line with fences - no newline",
			input:    "```const x = 5;```",
			expected: "```const x = 5;",
		},
		{
			name:     "handles content with leading/trailing whitespace inside fences",
			input:    "```\n  \n  code content  \n  \n```",
			expected: "  \n  code content  \n  \n",
		},
		{
			name:     "handles fence with language and content on same line",
			input:    "```go func main() {}```",
			expected: "```go func main() {}",
		},
		{
			name:     "removes only trailing fence markers",
			input:    "```\ncode with ``` in middle\n```",
			expected: "code with ``` in middle\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Unfence(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "splits multi-line text",
			input:    "line 1\nline 2\nline 3",
			expected: []string{"line 1", "line 2", "line 3"},
		},
		{
			name:     "handles single line",
			input:    "single line",
			expected: []string{"single line"},
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "handles string with only newlines",
			input:    "\n\n\n",
			expected: []string{"", "", "", ""},
		},
		{
			name:     "handles text with trailing newline",
			input:    "line 1\nline 2\n",
			expected: []string{"line 1", "line 2", ""},
		},
		{
			name:     "handles text with leading newline",
			input:    "\nline 1\nline 2",
			expected: []string{"", "line 1", "line 2"},
		},
		{
			name:     "handles mixed line endings and content",
			input:    "start\n\nmiddle\n\nend",
			expected: []string{"start", "", "middle", "", "end"},
		},
		{
			name:     "handles single newline",
			input:    "\n",
			expected: []string{"", ""},
		},
		{
			name:     "preserves empty lines between content",
			input:    "first\n\n\nsecond",
			expected: []string{"first", "", "", "second"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitLines(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
