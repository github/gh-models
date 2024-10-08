package util

import (
	"fmt"
	"strings"

	"github.com/github/gh-models/internal/azure_models"
)

// GetModelByName returns the model with the specified name, or an error if no such model exists within the given list.
func GetModelByName(modelName string, models []*azure_models.ModelSummary) (*azure_models.ModelSummary, error) {
	for _, model := range models {
		if strings.EqualFold(model.FriendlyName, modelName) || strings.EqualFold(model.Name, modelName) {
			return model, nil
		}
	}
	return nil, fmt.Errorf("the specified model name is not supported: %s", modelName)
}
