// Package util provides utility functions for the gh-models extension.
package util

import (
	"fmt"
	"io"
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
