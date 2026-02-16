package cmd

import (
	"log/slog"
	"os"

	"github.com/phsym/console-slog"
)

type CLI struct {
	Taskfile string `arg:""                         help:"Path to the taskfile to process."`
	Verbose  bool   `help:"Enable verbose logging."`
}

// Run executes the clean command for each provided path.
func (c *CLI) Run(flags *Flags) error {
	flags.Log.Info("Done")

	return nil
}

// CreateLogger builds a slog logger configured from the CLI flags.
func (c *CLI) CreateLogger() *slog.Logger {
	level := slog.LevelInfo
	if c.Verbose {
		level = slog.LevelDebug
	}

	opts := &console.HandlerOptions{
		Level: level,
	}
	handler := console.NewHandler(os.Stderr, opts)

	return slog.New(handler)
}
