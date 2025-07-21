package generate

import (
	"testing"
)

func TestFloat32Ptr(t *testing.T) {
	tests := []struct {
		name     string
		input    float32
		expected float32
	}{
		{
			name:     "positive value",
			input:    3.14,
			expected: 3.14,
		},
		{
			name:     "negative value",
			input:    -2.5,
			expected: -2.5,
		},
		{
			name:     "zero value",
			input:    0.0,
			expected: 0.0,
		},
		{
			name:     "large value",
			input:    999999.99,
			expected: 999999.99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Float32Ptr(tt.input)
			if result == nil {
				t.Fatalf("Float32Ptr returned nil")
			}
			if *result != tt.expected {
				t.Errorf("Float32Ptr(%f) = %f, want %f", tt.input, *result, tt.expected)
			}
		})
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain JSON object",
			input:    `{"key": "value", "number": 42}`,
			expected: `{"key": "value", "number": 42}`,
		},
		{
			name:     "plain JSON array",
			input:    `[{"id": 1}, {"id": 2}]`,
			expected: `[{"id": 1}, {"id": 2}]`,
		},
		{
			name:     "JSON wrapped in markdown code block",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON wrapped in generic code block",
			input:    "```\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with extra whitespace",
			input:    "   \n  {\"key\": \"value\"}  \n  ",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON embedded in text",
			input:    "Here is some JSON: {\"key\": \"value\"} and some more text",
			expected: `{"key": "value"}`,
		},
		{
			name:     "array embedded in text",
			input:    "The data is: [{\"id\": 1}, {\"id\": 2}] as shown above",
			expected: `[{"id": 1}, {"id": 2}]`,
		},
		{
			name:     "JavaScript string concatenation",
			input:    `{"message": "Hello" + "World"}`,
			expected: `{"message": "HelloWorld"}`,
		},
		{
			name:     "multiline string concatenation",
			input:    "{\n\"message\": \"Hello\" +\n\"World\"\n}",
			expected: "{\n\"message\": \"HelloWorld\"\n}",
		},
		{
			name:     "complex JavaScript expression",
			input:    `{"text": "A" + "B" * 1998}`,
			expected: `{"text": "AB_repeated"}`,
		},
		{
			name:     "JavaScript comments",
			input:    "{\n// This is a comment\n\"key\": \"value\"\n}",
			expected: "{\n\n\"key\": \"value\"\n}",
		},
		{
			name:     "multiple string concatenations",
			input:    `{"a": "Hello" + "World", "b": "Foo" + "Bar"}`,
			expected: `{"a": "HelloWorld", "b": "FooBar"}`,
		},
		{
			name:     "no JSON content",
			input:    "This is just plain text with no JSON",
			expected: "This is just plain text with no JSON",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "nested object",
			input:    `{"outer": {"inner": "value"}}`,
			expected: `{"outer": {"inner": "value"}}`,
		},
		{
			name:     "complex nested with concatenation",
			input:    "```json\n{\n  \"message\": \"Start\" + \"End\",\n  \"data\": {\n    \"value\": \"A\" + \"B\"\n  }\n}\n```",
			expected: "{\n  \"message\": \"StartEnd\",\n  \"data\": {\n    \"value\": \"AB\"\n  }\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractJSON(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractJSON(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanJavaScriptStringConcat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple concatenation",
			input:    `"Hello" + "World"`,
			expected: `"HelloWorld"`,
		},
		{
			name:     "concatenation with spaces",
			input:    `"Hello"  +  "World"`,
			expected: `"HelloWorld"`,
		},
		{
			name:     "multiline concatenation",
			input:    "\"Hello\" +\n\"World\"",
			expected: `"HelloWorld"`,
		},
		{
			name:     "multiple concatenations",
			input:    `"A" + "B" + "C"`,
			expected: `"ABC"`,
		},
		{
			name:     "complex expression",
			input:    `"Prefix" + "Suffix" * 1998`,
			expected: `"PrefixSuffix_repeated"`,
		},
		{
			name:     "with JavaScript comments",
			input:    "// Comment\n\"Hello\" + \"World\"",
			expected: "\n\"HelloWorld\"",
		},
		{
			name:     "no concatenation",
			input:    `"Just a string"`,
			expected: `"Just a string"`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "concatenation in JSON context",
			input:    `{"key": "Value1" + "Value2"}`,
			expected: `{"key": "Value1Value2"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanJavaScriptStringConcat(tt.input)
			if result != tt.expected {
				t.Errorf("cleanJavaScriptStringConcat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStringSliceContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{
			name:     "value exists in slice",
			slice:    []string{"apple", "banana", "cherry"},
			value:    "banana",
			expected: true,
		},
		{
			name:     "value does not exist in slice",
			slice:    []string{"apple", "banana", "cherry"},
			value:    "orange",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			value:    "apple",
			expected: false,
		},
		{
			name:     "nil slice",
			slice:    nil,
			value:    "apple",
			expected: false,
		},
		{
			name:     "single element slice - match",
			slice:    []string{"only"},
			value:    "only",
			expected: true,
		},
		{
			name:     "single element slice - no match",
			slice:    []string{"only"},
			value:    "other",
			expected: false,
		},
		{
			name:     "empty string in slice",
			slice:    []string{"", "apple", "banana"},
			value:    "",
			expected: true,
		},
		{
			name:     "case sensitive match",
			slice:    []string{"Apple", "Banana"},
			value:    "apple",
			expected: false,
		},
		{
			name:     "duplicate values in slice",
			slice:    []string{"apple", "apple", "banana"},
			value:    "apple",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringSliceContains(tt.slice, tt.value)
			if result != tt.expected {
				t.Errorf("StringSliceContains(%v, %q) = %t, want %t", tt.slice, tt.value, result, tt.expected)
			}
		})
	}
}

func TestMergeStringMaps(t *testing.T) {
	tests := []struct {
		name     string
		maps     []map[string]string
		expected map[string]string
	}{
		{
			name: "merge two maps",
			maps: []map[string]string{
				{"a": "1", "b": "2"},
				{"c": "3", "d": "4"},
			},
			expected: map[string]string{"a": "1", "b": "2", "c": "3", "d": "4"},
		},
		{
			name: "later map overwrites earlier",
			maps: []map[string]string{
				{"a": "1", "b": "2"},
				{"b": "overwritten", "c": "3"},
			},
			expected: map[string]string{"a": "1", "b": "overwritten", "c": "3"},
		},
		{
			name:     "empty maps",
			maps:     []map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "single map",
			maps: []map[string]string{
				{"a": "1", "b": "2"},
			},
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			name: "nil map in slice",
			maps: []map[string]string{
				{"a": "1"},
				nil,
				{"b": "2"},
			},
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			name: "empty map in slice",
			maps: []map[string]string{
				{"a": "1"},
				{},
				{"b": "2"},
			},
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			name: "three maps with overwrites",
			maps: []map[string]string{
				{"a": "1", "b": "2", "c": "3"},
				{"b": "overwritten1", "d": "4"},
				{"b": "final", "e": "5"},
			},
			expected: map[string]string{"a": "1", "b": "final", "c": "3", "d": "4", "e": "5"},
		},
		{
			name: "empty string values",
			maps: []map[string]string{
				{"a": "", "b": "2"},
				{"a": "1", "c": ""},
			},
			expected: map[string]string{"a": "1", "b": "2", "c": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeStringMaps(tt.maps...)

			// Check if the maps have the same length
			if len(result) != len(tt.expected) {
				t.Errorf("MergeStringMaps() result length = %d, want %d", len(result), len(tt.expected))
				return
			}

			// Check each key-value pair
			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("MergeStringMaps() missing key %q", key)
				} else if actualValue != expectedValue {
					t.Errorf("MergeStringMaps() key %q = %q, want %q", key, actualValue, expectedValue)
				}
			}

			// Check for unexpected keys
			for key := range result {
				if _, exists := tt.expected[key]; !exists {
					t.Errorf("MergeStringMaps() unexpected key %q with value %q", key, result[key])
				}
			}
		})
	}
}
