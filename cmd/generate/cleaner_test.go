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

func TestUnXml(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes simple XML tags",
			input:    "<tag>content</tag>",
			expected: "content",
		},
		{
			name:     "removes XML tags with content spanning multiple lines",
			input:    "<code>\nline 1\nline 2\nline 3\n</code>",
			expected: "line 1\nline 2\nline 3",
		},
		{
			name:     "removes tags with attributes",
			input:    `<div class="container" id="main">Hello World</div>`,
			expected: "Hello World",
		},
		{
			name:     "preserves content without XML tags",
			input:    "just plain text",
			expected: "just plain text",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles whitespace around XML",
			input:    "  <p>content</p>  ",
			expected: "content",
		},
		{
			name:     "handles content with leading/trailing whitespace inside tags",
			input:    "<div>  \n  content  \n  </div>",
			expected: "content",
		},
		{
			name:     "handles mismatched tag names",
			input:    "<start>content</end>",
			expected: "<start>content</end>",
		},
		{
			name:     "handles missing closing tag",
			input:    "<tag>content without closing",
			expected: "<tag>content without closing",
		},
		{
			name:     "handles missing opening tag",
			input:    "content without opening</tag>",
			expected: "content without opening</tag>",
		},
		{
			name:     "handles nested XML tags (outer only)",
			input:    "<outer><inner>content</inner></outer>",
			expected: "<inner>content</inner>",
		},
		{
			name:     "handles complex content with newlines and special characters",
			input:    "<response>\nHere's some code:\n\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n\nThat should work!\n</response>",
			expected: "Here's some code:\n\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n\nThat should work!",
		},
		{
			name:     "handles tag names with numbers and hyphens",
			input:    "<h1>Heading</h1>",
			expected: "Heading",
		},
		{
			name:     "handles tag names with underscores",
			input:    "<test_tag>content</test_tag>",
			expected: "content",
		},
		{
			name:     "handles empty tag content",
			input:    "<empty></empty>",
			expected: "",
		},
		{
			name:     "handles XML with only whitespace content",
			input:    "<space>   \n   </space>",
			expected: "",
		},
		{
			name:     "handles text that looks like XML but isn't",
			input:    "This < is not > XML < tags >",
			expected: "This < is not > XML < tags >",
		},
		{
			name:     "handles single character tag names",
			input:    "<a>link</a>",
			expected: "link",
		},
		{
			name:     "handles complex attributes with quotes",
			input:    `<tag attr1="value1" attr2='value2' attr3=value3>content</tag>`,
			expected: "content",
		},
		{
			name:     "handles XML declaration-like content (not removed)",
			input:    `<?xml version="1.0"?>content`,
			expected: `<?xml version="1.0"?>content`,
		},
		{
			name:     "handles comment-like content (not removed)",
			input:    `<!-- comment -->content`,
			expected: `<!-- comment -->content`,
		},
		{
			name:     "handles CDATA-like content (not removed)",
			input:    `<![CDATA[content]]>`,
			expected: `<![CDATA[content]]>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Unxml(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
