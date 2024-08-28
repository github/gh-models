package ux

import (
	"github.com/github/gh-models/internal/azure_models"
)

func IsChatModel(model *azure_models.ModelSummary) bool {
	return model.Task == "chat-completion"
}
