package azuremodels

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// ModelSummary includes basic information about a model.
type ModelSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Registry     string `json:"registry"`
	FriendlyName string `json:"friendly_name"`
	Task         string `json:"task"`
	Publisher    string `json:"publisher"`
	Summary      string `json:"summary"`
	Version      string `json:"version"`
}

// IsChatModel returns true if the model is for chat completions.
func (m *ModelSummary) IsChatModel() bool {
	return m.Task == "chat-completion"
}

// HasName checks if the model has the given name.
func (m *ModelSummary) HasName(name string) bool {
	return strings.EqualFold(m.ID, name)
}

var (
	featuredModelNames = []string{}
)

// SortModels sorts the given models in place, with featured models first, and then by friendly name.
func SortModels(models []*ModelSummary) {
	sort.Slice(models, func(i, j int) bool {
		// Sort featured models first, by name
		isFeaturedI := slices.Contains(featuredModelNames, models[i].Name)
		isFeaturedJ := slices.Contains(featuredModelNames, models[j].Name)

		if isFeaturedI && !isFeaturedJ {
			return true
		}

		if !isFeaturedI && isFeaturedJ {
			return false
		}

		// Otherwise, sort by friendly name
		// Note: sometimes the casing returned by the API is inconsistent, so sort using lowercase values.
		idI := strings.ToLower(fmt.Sprintf("%s/%s", models[i].Publisher, models[i].Name))
		idJ := strings.ToLower(fmt.Sprintf("%s/%s", models[j].Publisher, models[j].Name))

		return idI < idJ
	})
}
