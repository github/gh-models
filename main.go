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
	cmd := cmd.NewRootCommand()
	exitCode := exitOK

	ctx := context.Background()

	if _, err := cmd.ExecuteContextC(ctx); err != nil {
		exitCode = exitError
	}

	return exitCode
}
