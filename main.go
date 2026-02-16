package main

import (
	"github.com/alecthomas/kong"

	"github.com/theunrepentantgeek/task-graph/internal/cmd"
)

func main() {
	// Entry point for the application.
	var cli cmd.CLI

	ctx := kong.Parse(&cli,
		kong.UsageOnError())

	log := cli.CreateLogger()

	flags := &cmd.Flags{
		Verbose: cli.Verbose,
		Log:     log,
	}

	err := ctx.Run(flags)
	if err != nil {
		flags.Log.Error("Error executing command", "error", err)
		ctx.Exit(1)
	}

	ctx.Exit(0)
}
