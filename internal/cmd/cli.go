package cmd

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/phsym/console-slog"
	"github.com/rotisserie/eris"
	"gopkg.in/yaml.v3"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graphviz"
	"github.com/theunrepentantgeek/task-graph/internal/loader"
	"github.com/theunrepentantgeek/task-graph/internal/taskgraph"
)

//nolint:tagalign // Not useful here because different members have different tags.
type CLI struct {
	Taskfile           string `arg:"" help:"Path to the taskfile to process."`
	Output             string `help:"Path to the output file." long:"output" required:"true" short:"o"`
	Config             string `help:"Path to a config file (YAML or JSON)." long:"config" short:"c"`
	GroupByNamespace   bool   `help:"Group tasks in the same namespace together in the output." long:"group-by-namespace"`
	Verbose            bool   `help:"Enable verbose logging."`
}

// Run executes the CLI command with the given flags.
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

	err = graphviz.SaveTo(c.Output, gr, flags.Config)
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

func (c *CLI) CreateConfig() (*config.Config, error) {
	cfg := config.New()

	if c.Config != "" {
		if err := c.loadConfigFile(cfg); err != nil {
			return nil, err
		}
	}

	if c.GroupByNamespace {
		cfg.GroupByNamespace = true
	}

	return cfg, nil
}

func (c *CLI) loadConfigFile(cfg *config.Config) error {
	raw, err := os.ReadFile(c.Config)
	if err != nil {
		return eris.Wrapf(err, "failed to read config file: %s", c.Config)
	}

	ext := strings.ToLower(filepath.Ext(c.Config))

	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(raw, cfg)
	case ".json":
		err = json.Unmarshal(raw, cfg)
	default:
		// Attempt YAML first, then JSON, for unknown extensions.
		if yamlErr := yaml.Unmarshal(raw, cfg); yamlErr == nil {
			return nil
		}

		err = json.Unmarshal(raw, cfg)
	}

	if err != nil {
		return eris.Wrapf(err, "failed to parse config file: %s", c.Config)
	}

	return nil
}
