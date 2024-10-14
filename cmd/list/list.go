// Package list provides a gh command to list available models.
package list

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/internal/ux"
	"github.com/github/gh-models/pkg/command"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
)

var (
	lightGrayUnderline = ansi.ColorFunc("white+du")
)

// NewListCommand returns a new command to list available GitHub models.
func NewListCommand(cfg *command.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available models",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cfg.Client
			models, err := client.ListModels(ctx)
			if err != nil {
				return err
			}

			// For now, filter to just chat models.
			// Once other tasks are supported (like embeddings), update the list to show all models, with the task as a column.
			models = filterToChatModels(models)
			ux.SortModels(models)

			if cfg.IsTerminalOutput {
				cfg.WriteToOut("\n")
				cfg.WriteToOut(fmt.Sprintf("Showing %d available chat models\n", len(models)))
				cfg.WriteToOut("\n")
			}

			printer := cfg.NewTablePrinter()

			printer.AddHeader([]string{"DISPLAY NAME", "MODEL NAME"}, tableprinter.WithColor(lightGrayUnderline))
			printer.EndRow()

			for _, model := range models {
				printer.AddField(model.FriendlyName)
				printer.AddField(model.Name)
				printer.EndRow()
			}

			err = printer.Render()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func filterToChatModels(models []*azuremodels.ModelSummary) []*azuremodels.ModelSummary {
	var chatModels []*azuremodels.ModelSummary
	for _, model := range models {
		if ux.IsChatModel(model) {
			chatModels = append(chatModels, model)
		}
	}
	return chatModels
}
