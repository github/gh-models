package ux

import "github.com/github/gh-models/internal/azuremodels"

func IsChatModel(model *azuremodels.ModelSummary) bool {
	return model.Task == "chat-completion"
}

func FilterToChatModels(models []*azuremodels.ModelSummary) []*azuremodels.ModelSummary {
	var chatModels []*azuremodels.ModelSummary
	for _, model := range models {
		if IsChatModel(model) {
			chatModels = append(chatModels, model)
		}
	}
	return chatModels
}
