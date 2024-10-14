package ux

import (
	"testing"

	"github.com/github/gh-models/internal/azuremodels"
	"github.com/stretchr/testify/require"
)

func TestSorting(t *testing.T) {
	t.Run("SortModels sorts given slice in-place by friendly name, case-insensitive", func(t *testing.T) {
		modelA := &azuremodels.ModelSummary{Name: "z", FriendlyName: "AARDVARK"}
		modelB := &azuremodels.ModelSummary{Name: "y", FriendlyName: "betta"}
		modelC := &azuremodels.ModelSummary{Name: "x", FriendlyName: "Cat"}
		models := []*azuremodels.ModelSummary{modelB, modelA, modelC}

		SortModels(models)

		require.Equal(t, 3, len(models))
		require.Equal(t, "AARDVARK", models[0].FriendlyName)
		require.Equal(t, "betta", models[1].FriendlyName)
		require.Equal(t, "Cat", models[2].FriendlyName)
	})
}
