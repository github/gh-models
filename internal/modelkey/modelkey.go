package modelkey

import (
	"fmt"
	"strings"
)

type ModelKey struct {
	Provider  string
	Publisher string
	ModelName string
}

func ParseModelKey(modelKey string) (*ModelKey, error) {
	if modelKey == "" {
		return nil, fmt.Errorf("invalid model key format: %s", modelKey)
	}

	parts := strings.Split(modelKey, "/")

	// Check for empty parts
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid model key format: %s", modelKey)
		}
	}

	switch len(parts) {
	case 2:
		// Format: publisher/model-name (provider defaults to "azureml")
		return &ModelKey{
			Provider:  "azureml",
			Publisher: parts[0],
			ModelName: parts[1],
		}, nil
	case 3:
		// Format: provider/publisher/model-name
		return &ModelKey{
			Provider:  parts[0],
			Publisher: parts[1],
			ModelName: parts[2],
		}, nil
	default:
		return nil, fmt.Errorf("invalid model key format: %s", modelKey)
	}
}
