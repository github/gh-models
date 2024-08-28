package ux

import (
	"slices"
	"sort"

	"github.com/github/gh-models/internal/azure_models"
)

var (
	featuredModelNames = []string{}
)

func SortModels(models []*azure_models.ModelSummary) {
	sort.Slice(models, func(i, j int) bool {
		// Sort featured models first, by name
		isFeaturedI := slices.Contains(featuredModelNames, models[i].Name)
		isFeaturedJ := slices.Contains(featuredModelNames, models[j].Name)

		if isFeaturedI && !isFeaturedJ {
			return true
		} else if !isFeaturedI && isFeaturedJ {
			return false
		} else {
			// Otherwise, sort by publisher and then friendly name
			if models[i].Publisher == models[j].Publisher {
				return models[i].FriendlyName < models[j].FriendlyName
			}
			return models[i].Publisher < models[j].Publisher
		}
	})
}
