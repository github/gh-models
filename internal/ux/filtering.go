// Package ux provides utility functions around presentation and user experience.
package ux

import "github.com/github/gh-models/internal/azuremodels"

// IsChatModel returns true if the given model is for chat completions.
func IsChatModel(model *azuremodels.ModelSummary) bool {
	return model.Task == "chat-completion"
}
