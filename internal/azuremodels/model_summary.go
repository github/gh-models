package azuremodels

import "strings"

// ModelSummary includes basic information about a model.
type ModelSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	FriendlyName string `json:"friendly_name"`
	Task         string `json:"task"`
	Publisher    string `json:"publisher"`
	Summary      string `json:"summary"`
	Version      string `json:"version"`
	RegistryName string `json:"registry_name"`
}

// IsChatModel returns true if the model is for chat completions.
func (m *ModelSummary) IsChatModel() bool {
	return m.Task == "chat-completion"
}

// HasName checks if the model has the given name.
func (m *ModelSummary) HasName(name string) bool {
	return strings.EqualFold(m.FriendlyName, name) || strings.EqualFold(m.Name, name)
}
