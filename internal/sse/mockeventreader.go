package sse

import (
	"bufio"
	"bytes"
	"io"
)

// MockEventReader is a mock implementation of the sse.EventReader. This lets us use EventReader as a common interface
// for models that support streaming (like gpt-4o) and models that do not (like the o1 class of models).
type MockEventReader[T any] struct {
	reader  io.ReadCloser
	scanner *bufio.Scanner
	events  []T
	index   int
}

// NewMockEventReader creates a new MockEventReader with the given events.
func NewMockEventReader[T any](events []T) *MockEventReader[T] {
	data := []byte{}
	reader := io.NopCloser(bytes.NewReader(data))
	scanner := bufio.NewScanner(reader)
	return &MockEventReader[T]{reader: reader, scanner: scanner, events: events, index: 0}
}

// Read reads the next event from the stream.
func (mer *MockEventReader[T]) Read() (T, error) {
	if mer.index >= len(mer.events) {
		var zero T
		return zero, io.EOF
	}
	event := mer.events[mer.index]
	mer.index++
	return event, nil
}

// Close closes the Reader and any applicable inner stream state.
func (mer *MockEventReader[T]) Close() error {
	return mer.reader.Close()
}
