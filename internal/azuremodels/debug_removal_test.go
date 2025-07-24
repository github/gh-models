package azuremodels

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHttpLoggingParameterReplacement(t *testing.T) {
	// Test that HTTP logging now uses context instead of function parameters
	// This test ensures we moved HTTP log configuration to context
	
	// Create a mock client to test the interface
	client := NewMockClient()
	
	// Test that the interface accepts context and extracts HTTP log filename
	var capturedHttpLogFile string
	client.MockGetChatCompletionStream = func(ctx context.Context, req ChatCompletionOptions, org string) (*ChatCompletionResponse, error) {
		capturedHttpLogFile = HTTPLogFileFromContext(ctx)
		return &ChatCompletionResponse{}, nil
	}
	
	// Test with context without HTTP log file
	ctx := context.Background()
	_, _ = client.GetChatCompletionStream(ctx, ChatCompletionOptions{}, "")
	require.Equal(t, "", capturedHttpLogFile)
	
	// Test with context containing HTTP log file
	testLogFile := "/tmp/test.log"
	ctxWithLog := WithHTTPLogFile(ctx, testLogFile)
	_, _ = client.GetChatCompletionStream(ctxWithLog, ChatCompletionOptions{}, "")
	require.Equal(t, testLogFile, capturedHttpLogFile)
}