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
			enter a prompt. The extension will then return a response from the model.

			For more information about what you can do with GitHub Models extension, see the manual
			at https://github.com/github/gh-models/blob/main/README.md.
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
