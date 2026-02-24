package cmd

import (
	"log/slog"

	"github.com/theunrepentantgeek/task-graph/internal/config"
)

type Flags struct {
	Verbose bool
	Log     *slog.Logger
	Config  *config.Config
}
