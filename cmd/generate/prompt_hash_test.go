package generate

import (
	"testing"

	"github.com/github/gh-models/pkg/prompt"
	"github.com/github/gh-models/pkg/util"
)

func TestComputePromptHash(t *testing.T) {
	tests := []struct {
		name        string
		prompt      *prompt.File
		wantError   bool
		description string
	}{
		{
			name: "basic prompt with minimal data",
			prompt: &prompt.File{
				Model: "gpt-4o",
				Messages: []prompt.Message{
					{
						Role:    "system",
						Content: "You are a helpful assistant.",
					},
				},
				ModelParameters: prompt.ModelParameters{},
			},
			wantError:   false,
			description: "Should compute hash for minimal prompt",
		},
		{
			name: "prompt with model parameters",
			prompt: &prompt.File{
				Model: "gpt-4o",
				Messages: []prompt.Message{
					{
						Role:    "user",
						Content: "Hello world",
					},
				},
				ModelParameters: prompt.ModelParameters{
					MaxTokens:   util.Ptr(1000),
					Temperature: util.Ptr(0.7),
					TopP:        util.Ptr(0.9),
				},
			},
			wantError:   false,
			description: "Should compute hash for prompt with model parameters",
		},
		{
			name: "prompt with multiple messages",
			prompt: &prompt.File{
				Model: "gpt-3.5-turbo",
				Messages: []prompt.Message{
					{
						Role:    "system",
						Content: "You are a helpful assistant.",
					},
					{
						Role:    "user",
						Content: "What is the capital of France?",
					},
					{
						Role:    "assistant",
						Content: "The capital of France is Paris.",
					},
					{
						Role:    "user",
						Content: "What about Germany?",
					},
				},
				ModelParameters: prompt.ModelParameters{
					Temperature: util.Ptr(0.5),
				},
			},
			wantError:   false,
			description: "Should compute hash for prompt with multiple messages",
		},
		{
			name: "prompt with template variables in content",
			prompt: &prompt.File{
				Model: "gpt-4o",
				Messages: []prompt.Message{
					{
						Role:    "system",
						Content: "You are a {{role}} assistant.",
					},
					{
						Role:    "user",
						Content: "Please help me with {{task}}",
					},
				},
				ModelParameters: prompt.ModelParameters{
					MaxTokens: util.Ptr(500),
				},
			},
			wantError:   false,
			description: "Should compute hash for prompt with template variables",
		},
		{
			name: "empty prompt",
			prompt: &prompt.File{
				Model:           "",
				Messages:        []prompt.Message{},
				ModelParameters: prompt.ModelParameters{},
			},
			wantError:   false,
			description: "Should compute hash for empty prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := ComputePromptHash(tt.prompt)

			if tt.wantError {
				if err == nil {
					t.Errorf("ComputePromptHash() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ComputePromptHash() unexpected error: %v", err)
				return
			}

			// Verify hash is not empty
			if hash == "" {
				t.Errorf("ComputePromptHash() returned empty hash")
			}

			// Verify hash is consistent (run twice and compare)
			hash2, err2 := ComputePromptHash(tt.prompt)
			if err2 != nil {
				t.Errorf("ComputePromptHash() second call unexpected error: %v", err2)
				return
			}

			if hash != hash2 {
				t.Errorf("ComputePromptHash() inconsistent results: %s != %s", hash, hash2)
			}

			// Verify hash looks like a SHA256 hex string (64 characters, hex only)
			if len(hash) != 64 {
				t.Errorf("ComputePromptHash() hash length = %d, want 64", len(hash))
			}

			// Check if hash contains only hex characters
			for _, r := range hash {
				if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
					t.Errorf("ComputePromptHash() hash contains non-hex character: %c", r)
					break
				}
			}
		})
	}
}

func TestComputePromptHashDifferentInputs(t *testing.T) {
	// Test that different prompts produce different hashes
	prompt1 := &prompt.File{
		Model: "gpt-4o",
		Messages: []prompt.Message{
			{Role: "user", Content: "Hello"},
		},
		ModelParameters: prompt.ModelParameters{},
	}

	prompt2 := &prompt.File{
		Model: "gpt-4o",
		Messages: []prompt.Message{
			{Role: "user", Content: "Hi"},
		},
		ModelParameters: prompt.ModelParameters{},
	}

	hash1, err1 := ComputePromptHash(prompt1)
	if err1 != nil {
		t.Fatalf("ComputePromptHash() for prompt1 failed: %v", err1)
	}

	hash2, err2 := ComputePromptHash(prompt2)
	if err2 != nil {
		t.Fatalf("ComputePromptHash() for prompt2 failed: %v", err2)
	}

	if hash1 == hash2 {
		t.Errorf("ComputePromptHash() produced same hash for different prompts: %s", hash1)
	}
}

func TestComputePromptHashModelDifference(t *testing.T) {
	// Test that different models produce different hashes
	baseMessages := []prompt.Message{
		{Role: "user", Content: "Hello world"},
	}
	baseParams := prompt.ModelParameters{
		Temperature: util.Ptr(0.7),
	}

	prompt1 := &prompt.File{
		Model:           "gpt-4o",
		Messages:        baseMessages,
		ModelParameters: baseParams,
	}

	prompt2 := &prompt.File{
		Model:           "gpt-3.5-turbo",
		Messages:        baseMessages,
		ModelParameters: baseParams,
	}

	hash1, err1 := ComputePromptHash(prompt1)
	if err1 != nil {
		t.Fatalf("ComputePromptHash() for gpt-4o failed: %v", err1)
	}

	hash2, err2 := ComputePromptHash(prompt2)
	if err2 != nil {
		t.Fatalf("ComputePromptHash() for gpt-3.5-turbo failed: %v", err2)
	}

	if hash1 == hash2 {
		t.Errorf("ComputePromptHash() produced same hash for different models: %s", hash1)
	}
}

func TestComputePromptHashParameterDifference(t *testing.T) {
	// Test that different model parameters produce different hashes
	baseMessages := []prompt.Message{
		{Role: "user", Content: "Hello world"},
	}

	prompt1 := &prompt.File{
		Model:    "gpt-4o",
		Messages: baseMessages,
		ModelParameters: prompt.ModelParameters{
			Temperature: util.Ptr(0.5),
		},
	}

	prompt2 := &prompt.File{
		Model:    "gpt-4o",
		Messages: baseMessages,
		ModelParameters: prompt.ModelParameters{
			Temperature: util.Ptr(0.7),
		},
	}

	hash1, err1 := ComputePromptHash(prompt1)
	if err1 != nil {
		t.Fatalf("ComputePromptHash() for temp 0.5 failed: %v", err1)
	}

	hash2, err2 := ComputePromptHash(prompt2)
	if err2 != nil {
		t.Fatalf("ComputePromptHash() for temp 0.7 failed: %v", err2)
	}

	if hash1 == hash2 {
		t.Errorf("ComputePromptHash() produced same hash for different temperatures: %s", hash1)
	}
}

func TestComputePromptHashIgnoresOtherFields(t *testing.T) {
	// Test that fields not included in hash computation don't affect the result
	prompt1 := &prompt.File{
		Name:        "Test Prompt 1",
		Description: "This is a test prompt",
		Model:       "gpt-4o",
		Messages: []prompt.Message{
			{Role: "user", Content: "Hello"},
		},
		ModelParameters: prompt.ModelParameters{
			Temperature: util.Ptr(0.7),
		},
		TestData: []prompt.TestDataItem{
			{"input": "test"},
		},
		Evaluators: []prompt.Evaluator{
			{Name: "test-eval"},
		},
	}

	prompt2 := &prompt.File{
		Name:        "Test Prompt 2",
		Description: "This is another test prompt",
		Model:       "gpt-4o",
		Messages: []prompt.Message{
			{Role: "user", Content: "Hello"},
		},
		ModelParameters: prompt.ModelParameters{
			Temperature: util.Ptr(0.7),
		},
		TestData: []prompt.TestDataItem{
			{"input": "different"},
		},
		Evaluators: []prompt.Evaluator{
			{Name: "different-eval"},
		},
	}

	hash1, err1 := ComputePromptHash(prompt1)
	if err1 != nil {
		t.Fatalf("ComputePromptHash() for prompt1 failed: %v", err1)
	}

	hash2, err2 := ComputePromptHash(prompt2)
	if err2 != nil {
		t.Fatalf("ComputePromptHash() for prompt2 failed: %v", err2)
	}

	if hash1 != hash2 {
		t.Errorf("ComputePromptHash() produced different hashes for prompts that should be identical (ignoring non-hash fields): %s != %s", hash1, hash2)
	}
}
