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
	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/cobra"
)

// NewRootCommand returns a new root command for the gh-models extension.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "GitHub Models extension",
	}

	terminal := term.FromEnv()
	out := terminal.Out()
	token, _ := auth.TokenForHost("github.com")
	if token == "" {
		util.WriteToOut(out, "No GitHub token found. Please run 'gh auth login' to authenticate.\n")
		return nil
	}

	client := azuremodels.NewAzureClient(token)
	cfg := command.NewConfigWithTerminal(terminal, client)

	cmd.AddCommand(list.NewListCommand(cfg))
	cmd.AddCommand(run.NewRunCommand(cfg))
	cmd.AddCommand(view.NewViewCommand(cfg))

	// Cobra doesn't have a way to specify a two word command (ie. "gh models"), so set a custom usage template
	// with `gh`` in it. Cobra will use this template for the root and all child commands.
	cmd.SetUsageTemplate(strings.NewReplacer(
		"{{.UseLine}}", "gh {{.UseLine}}",
		"{{.CommandPath}}", "gh {{.CommandPath}}").Replace(cmd.UsageTemplate()))
	return cmd
}
