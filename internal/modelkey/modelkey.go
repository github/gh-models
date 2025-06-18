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

// String returns the string representation of the ModelKey.
func (mk *ModelKey) String() string {
	provider := formatPart(mk.Provider)
	publisher := formatPart(mk.Publisher)
	modelName := formatPart(mk.ModelName)

	if provider == "azureml" {
		return fmt.Sprintf("%s/%s", publisher, modelName)
	}

	return fmt.Sprintf("%s/%s/%s", provider, publisher, modelName)
}

func formatPart(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

	return s
}

func FormatIdentifier(provider, publisher, name string) string {
	mk := &ModelKey{
		Provider:  provider,
		Publisher: publisher,
		ModelName: name,
	}

	return mk.String()
}
