package ux

import (
	"github.com/github/gh-models/internal/azure_models"
)

func IsChatModel(model *azure_models.ModelSummary) bool {
	return model.Task == "chat-completion"
}

func FilterToChatModels(models []*azure_models.ModelSummary) []*azure_models.ModelSummary {
	var chatModels []*azure_models.ModelSummary
	for _, model := range models {
		if IsChatModel(model) {
			chatModels = append(chatModels, model)
		}
	}
	return chatModels
}
