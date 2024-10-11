// Package view provides a `gh models view` command to view details about a model.
package view

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/internal/ux"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/cobra"
)

// NewViewCommand returns a new command to view details about a model.
func NewViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view [model]",
		Short: "View details about a model",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			terminal := term.FromEnv()

			token, _ := auth.TokenForHost("github.com")
			if token == "" {
				util.WriteToOut(terminal.Out(), "No GitHub token found. Please run 'gh auth login' to authenticate.\n")
				return nil
			}

			client := azuremodels.NewClient(token)
			ctx := cmd.Context()

			models, err := client.ListModels(ctx)
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

			modelSummary, err := getModelByName(modelName, models)
			if err != nil {
				return err
			}

			modelDetails, err := client.GetModelDetails(ctx, modelSummary.RegistryName, modelSummary.Name, modelSummary.Version)
			if err != nil {
				return err
			}

			modelPrinter := newModelPrinter(modelSummary, modelDetails, terminal)

			err = modelPrinter.render()
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}

// getModelByName returns the model with the specified name, or an error if no such model exists within the given list.
func getModelByName(modelName string, models []*azuremodels.ModelSummary) (*azuremodels.ModelSummary, error) {
	for _, model := range models {
		if model.HasName(modelName) {
			return model, nil
		}
	}
	return nil, fmt.Errorf("the specified model name is not supported: %s", modelName)
}
