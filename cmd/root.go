// Package cmd represents the base command when called without any subcommands.
package cmd

import (
	"fmt"
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

	var client azuremodels.Client

	if token == "" {
		util.WriteToOut(out, "No GitHub token found. Please run 'gh auth login' to authenticate.\n")
		client = azuremodels.NewUnauthenticatedClient()
	} else {
		var err error
		client, err = azuremodels.NewDefaultAzureClient(token)
		if err != nil {
			util.WriteToOut(terminal.ErrOut(), "Error creating Azure client: "+err.Error())
			return nil
		}
	}

	cfg := command.NewConfigWithTerminal(terminal, client)

	cmd.AddCommand(list.NewListCommand(cfg))
	cmd.AddCommand(run.NewRunCommand(cfg))
	cmd.AddCommand(view.NewViewCommand(cfg))

	// Cobra does not have a nice way to inject "global" help text, so we have to do it manually.
	// Copied from https://github.com/spf13/cobra/blob/e94f6d0dd9a5e5738dca6bce03c4b1207ffbc0ec/command.go#L595-L597
	cmd.SetHelpTemplate(fmt.Sprintf(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

%s

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`, azuremodels.NOTICE))

	// Cobra doesn't have a way to specify a two word command (ie. "gh models"), so set a custom usage template
	// with `gh`` in it. Cobra will use this template for the root and all child commands.
	cmd.SetUsageTemplate(strings.NewReplacer(
		"{{.UseLine}}", "gh {{.UseLine}}",
		"{{.CommandPath}}", "gh {{.CommandPath}}").Replace(cmd.UsageTemplate()))
	return cmd
}
