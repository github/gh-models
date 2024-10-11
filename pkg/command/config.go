// Package command provides shared configuration for sub-commands in the gh-models extension.
package command

import (
	"io"

	"github.com/github/gh-models/internal/azuremodels"
)

// Config represents configurable settings for a command.
type Config struct {
	Out              io.Writer
	ErrOut           io.Writer
	Client           azuremodels.Client
	IsTerminalOutput bool
	TerminalWidth    int
}

// NewConfig returns a new command configuration.
func NewConfig(out io.Writer, errOut io.Writer, client azuremodels.Client, isTerminalOutput bool, width int) *Config {
	return &Config{Out: out, ErrOut: errOut, Client: client, IsTerminalOutput: isTerminalOutput, TerminalWidth: width}
}
