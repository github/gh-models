package view

import (
	"io"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/github/gh-models/internal/azure_models"
	"github.com/github/gh-models/internal/ux"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/cobra"
)

func NewViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view [model]",
		Short: "View details about a model",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			terminal := term.FromEnv()

			token, _ := auth.TokenForHost("github.com")
			if token == "" {
				io.WriteString(terminal.Out(), "No GitHub token found. Please run 'gh auth login' to authenticate.\n")
				return nil
			}

			client := azure_models.NewClient(token)

			models, err := client.ListModels()
			if err != nil {
				return err
			}

			ux.SortModels(models)

			modelName := ""
			switch {
			case len(args) == 0:
				// Need to prompt for a model
				prompt := &survey.Select{
					Message: "Select a model:",
					Options: []string{},
				}

				for _, model := range models {
					if !ux.IsChatModel(model) {
						continue
					}
					prompt.Options = append(prompt.Options, model.FriendlyName)
				}

				err = survey.AskOne(prompt, &modelName, survey.WithPageSize(10))
				if err != nil {
					return err
				}

			case len(args) >= 1:
				modelName = args[0]
			}

			model, err := util.GetModelByName(modelName, models)
			if err != nil {
				return err
			}

			modelPrinter := newModelPrinter(model, terminal)

			err = modelPrinter.render()
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}
