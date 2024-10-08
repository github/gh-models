package util

import (
	"strings"

	"github.com/github/gh-models/internal/azure_models"
)

// GetValidModelName checks whether the given raw model name is a valid one, based on the provided list of models.
// If the given name does not represent a valid model, it returns nil.
func GetValidModelName(candidateModelName string, models []*azure_models.ModelSummary) *string {
	for _, model := range models {
		if strings.EqualFold(model.FriendlyName, candidateModelName) || strings.EqualFold(model.Name, candidateModelName) {
			return &model.Name
		}
	}
	return nil
}
