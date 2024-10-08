package list

import (
	"fmt"
	"io"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/github/gh-models/internal/azure_models"
	"github.com/github/gh-models/internal/ux"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available models",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			terminal := term.FromEnv()
			out := terminal.Out()

			token, _ := auth.TokenForHost("github.com")
			if token == "" {
				io.WriteString(out, "No GitHub token found. Please run 'gh auth login' to authenticate.\n")
				return nil
			}

			client := azure_models.NewClient(token)

			models, err := client.ListModels()
			if err != nil {
				return err
			}

			// For now, filter to just chat models.
			// Once other tasks are supported (like embeddings), update the list to show all models, with the task as a column.
			models = ux.FilterToChatModels(models)
			ux.SortModels(models)

			isTTY := terminal.IsTerminalOutput()

			if isTTY {
				io.WriteString(out, "\n")
				io.WriteString(out, fmt.Sprintf("Showing %d available chat models\n", len(models)))
				io.WriteString(out, "\n")
			}

			width, _, _ := terminal.Size()
			printer := tableprinter.New(out, isTTY, width)

			printer.AddHeader([]string{"Display Name", "Model Name"}, tableprinter.WithColor(util.LightGrayUnderline))
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
