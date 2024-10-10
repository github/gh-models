// Package main provides the entry point for the gh-models extension.
package main

import (
	"context"
	"os"

	"github.com/github/gh-models/cmd"
)

type exitCode int

const (
	exitOK    exitCode = 0
	exitError exitCode = 1
)

func main() {
	code := mainRun()
	os.Exit(int(code))
}

func mainRun() exitCode {
	rootCmd := cmd.NewRootCommand()
	exitCode := exitOK

	ctx := context.Background()

	if _, err := rootCmd.ExecuteContextC(ctx); err != nil {
		exitCode = exitError
	}

	return exitCode
}
