package list

import (
	"io"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/github/gh-models/internal/azure_models"
	"github.com/github/gh-models/internal/ux"
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

			ux.SortModels(models)

			width, _, _ := terminal.Size()
			printer := tableprinter.New(out, terminal.IsTerminalOutput(), width)

			printer.AddHeader([]string{"Name", "Friendly Name", "Publisher"})
			printer.EndRow()

			for _, model := range models {
				if !ux.IsChatModel(model) {
					continue
				}

				printer.AddField(model.Name)
				printer.AddField(model.FriendlyName)
				printer.AddField(model.Publisher)
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
