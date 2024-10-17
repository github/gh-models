package azuremodels

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelSummary(t *testing.T) {
	t.Run("IsChatModel", func(t *testing.T) {
		embeddingModel := &ModelSummary{Task: "embeddings"}
		chatCompletionModel := &ModelSummary{Task: "chat-completion"}
		otherModel := &ModelSummary{Task: "something-else"}

		require.False(t, embeddingModel.IsChatModel())
		require.True(t, chatCompletionModel.IsChatModel())
		require.False(t, otherModel.IsChatModel())
	})

	t.Run("HasName", func(t *testing.T) {
		model := &ModelSummary{Name: "foo123", FriendlyName: "Foo 123"}

		require.True(t, model.HasName(model.Name))
		require.True(t, model.HasName("FOO123"))
		require.True(t, model.HasName(model.FriendlyName))
		require.True(t, model.HasName("fOo 123"))
		require.False(t, model.HasName("completely different value"))
		require.False(t, model.HasName("foo"))
	})

	t.Run("SortModels sorts given slice in-place by friendly name, case-insensitive", func(t *testing.T) {
		modelA := &ModelSummary{Name: "z", FriendlyName: "AARDVARK"}
		modelB := &ModelSummary{Name: "y", FriendlyName: "betta"}
		modelC := &ModelSummary{Name: "x", FriendlyName: "Cat"}
		models := []*ModelSummary{modelB, modelA, modelC}

		SortModels(models)

		require.Equal(t, 3, len(models))
		require.Equal(t, "AARDVARK", models[0].FriendlyName)
		require.Equal(t, "betta", models[1].FriendlyName)
		require.Equal(t, "Cat", models[2].FriendlyName)
	})
}
