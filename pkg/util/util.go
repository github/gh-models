// Package util provides utility functions for the gh-models extension.
package util

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/pflag"
)

// WriteToOut writes a message to the given io.Writer.
func WriteToOut(out io.Writer, message string) {
	_, err := io.WriteString(out, message)
	if err != nil {
		fmt.Println("Error writing message:", err)
	}
}

// Ptr returns a pointer to the given value.
func Ptr[T any](value T) *T {
	return &value
}

// ParseTemplateVariables parses template variables from the --var flags
func ParseTemplateVariables(flags *pflag.FlagSet) (map[string]string, error) {
	varFlags, err := flags.GetStringArray("var")
	if err != nil {
		return nil, err
	}

	templateVars := make(map[string]string)
	for _, varFlag := range varFlags {
		// Handle empty strings
		if strings.TrimSpace(varFlag) == "" {
			continue
		}

		parts := strings.SplitN(varFlag, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid variable format '%s', expected 'key=value'", varFlag)
		}

		key := strings.TrimSpace(parts[0])
		value := parts[1] // Don't trim value to preserve intentional whitespace

		if key == "" {
			return nil, fmt.Errorf("variable key cannot be empty in '%s'", varFlag)
		}

		// Check for duplicate keys
		if _, exists := templateVars[key]; exists {
			return nil, fmt.Errorf("duplicate variable key '%s'", key)
		}

		templateVars[key] = value
	}

	return templateVars, nil
}
