// cmd represents the base command when called without any subcommands.
package cmd

import (
	"github.com/github/gh-models/cmd/list"
	"github.com/github/gh-models/cmd/run"
	"github.com/github/gh-models/cmd/view"
	"github.com/spf13/cobra"
	"strings"
)

// NewRootCommand returns a new root command for the gh-models extension.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "GitHub Models extension",
	}

	cmd.AddCommand(list.NewListCommand())
	cmd.AddCommand(run.NewRunCommand())
	cmd.AddCommand(view.NewViewCommand())

	// Cobra doesn't have a way to specify a two word command (ie. "gh models"), so set a custom usage template
	// with `gh`` in it. Cobra will use this template for the root and all child commands.
	cmd.SetUsageTemplate(strings.NewReplacer(
		"{{.UseLine}}", "gh {{.UseLine}}",
		"{{.CommandPath}}", "gh {{.CommandPath}}").Replace(cmd.UsageTemplate()))
	return cmd
}
