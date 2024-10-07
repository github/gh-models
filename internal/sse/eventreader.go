// Forked from https://github.com/Azure/azure-sdk-for-go/blob/4661007ca1fd68b2e31f3502d4282904014fd731/sdk/ai/azopenai/event_reader.go#L18

package sse

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

type EventReaderInterface[T any] interface {
	Read() (T, error)
	Close() error
}

// Reader is an interface for reading events from an SSE stream.
type Reader[T any] interface {
	// Read reads the next event from the stream.
	// Returns io.EOF when there are no further events.
	Read() (T, error)
	// Close closes the Reader and any applicable inner stream state.
	Close() error
}

// EventReader streams events dynamically from an OpenAI endpoint.
type EventReader[T any] struct {
	reader  io.ReadCloser // Required for Closing
	scanner *bufio.Scanner
}

// NewEventReader creates an EventReader that provides access to messages of
// type T from r.
func NewEventReader[T any](r io.ReadCloser) *EventReader[T] {
	return &EventReader[T]{reader: r, scanner: bufio.NewScanner(r)}
}

// Read reads the next event from the stream.
// Returns io.EOF when there are no further events.
func (er *EventReader[T]) Read() (T, error) {
	// https://html.spec.whatwg.org/multipage/server-sent-events.html
	for er.scanner.Scan() { // Scan while no error
		line := er.scanner.Text() // Get the line & interpret the event stream:

		if line == "" || line[0] == ':' { // If the line is blank or is a comment, skip it
			continue
		}

		if strings.Contains(line, ":") { // If the line contains a U+003A COLON character (:), process the field
			tokens := strings.SplitN(line, ":", 2)
			tokens[0], tokens[1] = strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
			var data T
			switch tokens[0] {
			case "data": // return the deserialized JSON object
				if tokens[1] == "[DONE]" { // If data is [DONE], end of stream was reached
					return data, io.EOF
				}
				err := json.Unmarshal([]byte(tokens[1]), &data)
				return data, err
			default: // Any other event type is an unexpected
				return data, errors.New("unexpected event type: " + tokens[0])
			}
			// Unreachable
		}
	}

	scannerErr := er.scanner.Err()

	if scannerErr == nil {
		return *new(T), errors.New("incomplete stream")
	}

	return *new(T), scannerErr
}

// Close closes the EventReader and any applicable inner stream state.
func (er *EventReader[T]) Close() error {
	return er.reader.Close()
}
