package run

import (
	"context"
	"testing"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/internal/sse"
	"github.com/github/gh-models/pkg/command"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestHttpLogPassthrough(t *testing.T) {
	// Test that the httpLog parameter is correctly passed through the call chain via context
	var capturedHttpLog string
	
	client := azuremodels.NewMockClient()
	client.MockGetChatCompletionStream = func(ctx context.Context, opt azuremodels.ChatCompletionOptions, org string) (*azuremodels.ChatCompletionResponse, error) {
		capturedHttpLog = azuremodels.HTTPLogFileFromContext(ctx)
		reader := sse.NewMockEventReader([]azuremodels.ChatCompletion{})
		return &azuremodels.ChatCompletionResponse{Reader: reader}, nil
	}
	
	cfg := command.NewConfig(nil, nil, client, false, 80)
	
	// Create a command with the http-log flag
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background()) // Set a context for the command
	cmd.Flags().String("http-log", "", "Path to log HTTP requests to (optional)")
	cmd.Flags().Set("http-log", "/tmp/test.log")
	
	// Create handler
	handler := newRunCommandHandler(cmd, cfg, []string{})
	
	// Test that httpLog is correctly stored in context
	require.Equal(t, "/tmp/test.log", azuremodels.HTTPLogFileFromContext(handler.ctx))
	
	// Test that it's passed to the client call via context
	req := azuremodels.ChatCompletionOptions{
		Model: "test-model",
		Messages: []azuremodels.ChatMessage{
			{Role: azuremodels.ChatMessageRoleUser, Content: &[]string{"test"}[0]},
		},
	}
	
	_, err := handler.getChatCompletionStreamReader(req, "")
	require.NoError(t, err)
	require.Equal(t, "/tmp/test.log", capturedHttpLog)
}