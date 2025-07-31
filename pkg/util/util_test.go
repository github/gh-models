package util

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestParseTemplateVariables(t *testing.T) {
	tests := []struct {
		name      string
		varFlags  []string
		expected  map[string]string
		expectErr bool
	}{
		{
			name:     "empty flags",
			varFlags: []string{},
			expected: map[string]string{},
		},
		{
			name:     "single variable",
			varFlags: []string{"name=Alice"},
			expected: map[string]string{"name": "Alice"},
		},
		{
			name:     "multiple variables",
			varFlags: []string{"name=Alice", "age=30", "city=Boston"},
			expected: map[string]string{"name": "Alice", "age": "30", "city": "Boston"},
		},
		{
			name:     "variable with spaces in value",
			varFlags: []string{"description=Hello World"},
			expected: map[string]string{"description": "Hello World"},
		},
		{
			name:     "variable with equals in value",
			varFlags: []string{"equation=x=y+1"},
			expected: map[string]string{"equation": "x=y+1"},
		},
		{
			name:     "variable with empty value",
			varFlags: []string{"empty="},
			expected: map[string]string{"empty": ""},
		},
		{
			name:     "variable with whitespace around key",
			varFlags: []string{" name =Alice"},
			expected: map[string]string{"name": "Alice"},
		},
		{
			name:     "preserve whitespace in value",
			varFlags: []string{"message= Hello World "},
			expected: map[string]string{"message": " Hello World "},
		},
		{
			name:      "empty string flag is ignored",
			varFlags:  []string{"", "name=Alice"},
			expected:  map[string]string{"name": "Alice"},
			expectErr: false,
		},
		{
			name:      "whitespace only flag is ignored",
			varFlags:  []string{"   ", "name=Alice"},
			expected:  map[string]string{"name": "Alice"},
			expectErr: false,
		},
		{
			name:      "missing equals sign",
			varFlags:  []string{"name"},
			expectErr: true,
		},
		{
			name:      "missing equals sign with multiple vars",
			varFlags:  []string{"name=Alice", "age"},
			expectErr: true,
		},
		{
			name:      "empty key",
			varFlags:  []string{"=value"},
			expectErr: true,
		},
		{
			name:      "whitespace only key",
			varFlags:  []string{" =value"},
			expectErr: true,
		},
		{
			name:      "duplicate keys",
			varFlags:  []string{"name=Alice", "name=Bob"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.StringSlice("var", tt.varFlags, "test flag")

			result, err := ParseTemplateVariables(flags)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
