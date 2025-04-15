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
		model := &ModelSummary{Name: "foo123", Publisher: "bar"}

		require.True(t, model.HasName(FormatIdentifier(model.Publisher, model.Name)))
		require.True(t, model.HasName("BaR/foO123"))
		require.False(t, model.HasName("completely different value"))
		require.False(t, model.HasName("foo"))
		require.False(t, model.HasName("bar"))
	})

	t.Run("SortModels sorts given slice in-place by publisher/name", func(t *testing.T) {
		modelA := &ModelSummary{Publisher: "a", Name: "z"}
		modelB := &ModelSummary{Publisher: "a", Name: "Y"}
		modelC := &ModelSummary{Publisher: "b", Name: "x"}
		models := []*ModelSummary{modelC, modelB, modelA}

		SortModels(models)

		require.Equal(t, 3, len(models))
		require.Equal(t, "Y", models[0].Name)
		require.Equal(t, "z", models[1].Name)
		require.Equal(t, "x", models[2].Name)
	})
}
