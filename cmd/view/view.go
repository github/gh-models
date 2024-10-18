// Package view provides a `gh models view` command to view details about a model.
package view

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

// NewViewCommand returns a new command to view details about a model.
func NewViewCommand(cfg *command.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view [model]",
		Short: "View details about a model",
		Long: heredoc.Docf(`
			Returns details about the specified model.

			Use %[1]sgh models view%[1]s to run in interactive mode. It will provide a list of the current
			models and allow you to select the one you want information about.

			If you know which model you want information for, you can run the request in a single command
			as %[1]sgh models view [model]%[1]s
		`, "`"),
		Example: "gh models view gpt-4o",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cfg.Client
			models, err := client.ListModels(ctx)
			if err != nil {
				return err
			}

			azuremodels.SortModels(models)

			modelName := ""
			switch {
			case len(args) == 0:
				// Need to prompt for a model
				prompt := &survey.Select{
					Message: "Select a model:",
					Options: []string{},
				}

				for _, model := range models {
					if !model.IsChatModel() {
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

			modelPrinter := newModelPrinter(modelSummary, modelDetails, cfg)

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
