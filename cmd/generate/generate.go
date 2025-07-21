// Package generate provides a gh command to generate tests.
package generate

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/github/gh-models/pkg/command"
	"github.com/spf13/cobra"
)

// NewListCommand returns a new command to list available GitHub models.
func NewListCommand(cfg *command.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate tests using PromptPex",
		Long: heredoc.Docf(`
			Augment prompt.yml file with generated test cases.
		`, "`"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}

	return cmd
}
