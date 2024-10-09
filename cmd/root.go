// cmd represents the base command when called without any subcommands.
package cmd

import (
	"github.com/github/gh-models/cmd/list"
	"github.com/github/gh-models/cmd/run"
	"github.com/spf13/cobra"
)

// NewRootCommand returns a new root command for the gh-models extension.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gh models",
		Short: "GitHub Models extension",
	}

	cmd.AddCommand(list.NewListCommand())
	cmd.AddCommand(run.NewRunCommand())

	return cmd
}
