package ux

import (
	"slices"
	"sort"
	"strings"

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
			// Otherwise, sort by friendly name
			// Note: sometimes the casing returned by the API is inconsistent, so sort using lowercase values.
			friendlyNameI := strings.ToLower(models[i].FriendlyName)
			friendlyNameJ := strings.ToLower(models[j].FriendlyName)

			return friendlyNameI < friendlyNameJ
		}
	})
}
