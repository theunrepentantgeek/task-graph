package cmd

import "log/slog"

type Flags struct {
	Verbose bool
	Log     *slog.Logger
}
