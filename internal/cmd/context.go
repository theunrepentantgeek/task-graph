package cmd

import (
	"log/slog"

	"github.com/theunrepentantgeek/task-graph/internal/config"
)

// Flags holds the shared runtime context threaded through CLI command execution.
type Flags struct {
	Verbose bool
	Log     *slog.Logger
	Config  *config.Config
}
