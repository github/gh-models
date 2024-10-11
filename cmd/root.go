// Package cmd represents the base command when called without any subcommands.
package cmd

import (
	"strings"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/github/gh-models/cmd/list"
	"github.com/github/gh-models/cmd/run"
	"github.com/github/gh-models/cmd/view"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/cobra"
)

// NewRootCommand returns a new root command for the gh-models extension.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "GitHub Models extension",
	}

	token, _ := auth.TokenForHost("github.com")
	if token == "" {
		terminal := term.FromEnv()
		util.WriteToOut(terminal.Out(), "No GitHub token found. Please run 'gh auth login' to authenticate.\n")
		return nil
	}

	client := azuremodels.NewAzureClient(token)

	cmd.AddCommand(list.NewListCommand(client))
	cmd.AddCommand(run.NewRunCommand(client))
	cmd.AddCommand(view.NewViewCommand(client))

	// Cobra doesn't have a way to specify a two word command (ie. "gh models"), so set a custom usage template
	// with `gh`` in it. Cobra will use this template for the root and all child commands.
	cmd.SetUsageTemplate(strings.NewReplacer(
		"{{.UseLine}}", "gh {{.UseLine}}",
		"{{.CommandPath}}", "gh {{.CommandPath}}").Replace(cmd.UsageTemplate()))
	return cmd
}
