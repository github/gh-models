// Package command provides shared configuration for sub-commands in the gh-models extension.
package command

import (
	"io"

	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/github/gh-models/internal/azuremodels"
)

// Config represents configurable settings for a command.
type Config struct {
	// Out is where standard output is written.
	Out io.Writer
	// ErrOut is where error output is written.
	ErrOut io.Writer
	// Client is the client for interacting with the models service.
	Client azuremodels.Client
	// IsTerminalOutput is true if the output should be formatted for a terminal.
	IsTerminalOutput bool
	// TerminalWidth is the width of the terminal.
	TerminalWidth int
}

// NewConfig returns a new command configuration.
func NewConfig(out, errOut io.Writer, client azuremodels.Client, isTerminalOutput bool, width int) *Config {
	return &Config{Out: out, ErrOut: errOut, Client: client, IsTerminalOutput: isTerminalOutput, TerminalWidth: width}
}

// NewConfigWithTerminal returns a new command configuration using the given terminal.
func NewConfigWithTerminal(terminal term.Term, client azuremodels.Client) *Config {
	width, _, _ := terminal.Size()
	return &Config{
		Out:              terminal.Out(),
		ErrOut:           terminal.ErrOut(),
		Client:           client,
		IsTerminalOutput: terminal.IsTerminalOutput(),
		TerminalWidth:    width,
	}
}

// NewTablePrinter initializes a table printer with terminal mode and terminal width.
func (c *Config) NewTablePrinter() tableprinter.TablePrinter {
	return tableprinter.New(c.Out, c.IsTerminalOutput, c.TerminalWidth)
}
