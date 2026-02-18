package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/phsym/console-slog"
	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/task-graph/internal/graphviz"
	"github.com/theunrepentantgeek/task-graph/internal/loader"
	"github.com/theunrepentantgeek/task-graph/internal/taskgraph"
)

type CLI struct {
	Taskfile string `arg:""                          help:"Path to the taskfile to process."`
	Output   string `help:"Path to the output file." long:"output"                           required:"true" short:"o"`
	Verbose  bool   `help:"Enable verbose logging."`
}

// Run executes the clean command for each provided path.
func (c *CLI) Run(
	flags *Flags,
) error {
	flags.Log.Info("Done")

	tf, err := loader.Load(context.Background(), c.Taskfile)
	if err != nil {
		return eris.Wrap(err, "failed to load taskfile")
	}

	flags.Log.Info(
		"Loaded taskfile",
		"taskfile", c.Taskfile,
		"tasks", tf.Tasks.Len())

	gr := taskgraph.New(tf).Build()

	err = graphviz.SaveTo(c.Output, gr)
	if err != nil {
		return eris.Wrap(err, "failed to save graph")
	}

	flags.Log.Info(
		"Saved graph",
		"output", c.Output,
	)

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
