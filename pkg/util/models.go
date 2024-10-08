package util

import (
	"fmt"
	"strings"

	"github.com/github/gh-models/internal/azure_models"
)

// ValidateModelName checks whether the given model name is a valid one, based on the provided list of models.
func ValidateModelName(modelName string, models []*azure_models.ModelSummary) (string, error) {
	for _, model := range models {
		if strings.EqualFold(model.FriendlyName, modelName) || strings.EqualFold(model.Name, modelName) {
			return model.Name, nil
		}
	}
	return "", fmt.Errorf("the specified model name is not supported: %s", modelName)
}
