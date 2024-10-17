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
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

// NewRootCommand returns a new root command for the gh-models extension.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "GitHub Models extension",
		Long: heredoc.Docf(`
			GitHub Models CLI extension allows you to experiment with AI models from the command line.

			To see a list of all available commands, run %[1]sgh models help%[1]s. To run the extension in
			interactive mode, run %[1]sgh models run%[1]s. This will prompt you to select a model and then
			to enter a prompt. The extension will then return a response from the model.

			For more information about what you can do with GitHub Models extension see the manual
			at https://cli.github.com/manual/gh_models
		`, "`"),
	}

	terminal := term.FromEnv()
	out := terminal.Out()
	token, _ := auth.TokenForHost("github.com")

	var client azuremodels.Client

	if token == "" {
		util.WriteToOut(out, "No GitHub token found. Please run 'gh auth login' to authenticate.\n")
		client = azuremodels.NewUnauthenticatedClient()
	} else {
		client = azuremodels.NewAzureClient(token)
	}

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
