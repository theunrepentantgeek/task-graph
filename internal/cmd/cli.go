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

	"github.com/theunrepentantgeek/task-graph/internal/autocolor"
	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/dot"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/graphviz"
	"github.com/theunrepentantgeek/task-graph/internal/loader"
	"github.com/theunrepentantgeek/task-graph/internal/mermaid"
	"github.com/theunrepentantgeek/task-graph/internal/taskgraph"
)

//nolint:tagalign // Not useful here because different members have different tags.
type CLI struct {
	Taskfile string `arg:"" help:"Path to the taskfile to process."`
	Output   string `help:"Path to the output file." long:"output" required:"true" short:"o"`
	Config   string `help:"Path to a config file (YAML or JSON)." long:"config" short:"c"`

	GroupByNamespace bool `help:"Group tasks in the same namespace together in the output." long:"group-by-namespace"`

	AutoColor bool `help:"Automatically color nodes by namespace using a built-in palette." long:"auto-color"`

	GraphType string `help:"Type of graph to generate (dot or mermaid). Defaults to dot." long:"graph-type"`

	//nolint:revive // Intentially long line for clarity in the CLI help.
	Highlight string `help:"Highlight specific tasks in the graph. Accepts task names or glob patterns, separated by commas or semicolons." long:"highlight"`

	//nolint:revive // Intentially long name for clarity in the CLI help.
	RenderImage string `help:"Render the graph as an image using graphviz dot. Specify the file type (e.g. png, svg)." long:"render-image"`

	//nolint:revive // Intentially long name for clarity in the CLI help.
	ExportConfig string `help:"Export the effective configuration to a file (YAML or JSON based on file extension)." long:"export-config"`
	Verbose      bool   `help:"Enable verbose logging."`
}

// Run executes the CLI command with the given flags.
func (c *CLI) Run(
	flags *Flags,
) error {
	flags.Log.Info("Done")

	ctx := context.Background()

	tf, err := loader.Load(ctx, c.Taskfile)
	if err != nil {
		return eris.Wrap(err, "failed to load taskfile")
	}

	flags.Log.Info(
		"Loaded taskfile",
		"taskfile", c.Taskfile,
		"tasks", tf.Tasks.Len())

	gr := taskgraph.New(tf).Build()

	applyAutoColor(flags.Config, gr)

	graphType := c.resolveGraphType(flags)

	switch graphType {
	case "dot":
		err = graphviz.SaveTo(c.Output, gr, flags.Config)
	case "mermaid":
		err = mermaid.SaveTo(c.Output, gr, flags.Config)
	default:
		return eris.Errorf("unsupported graph type: %q, must be dot or mermaid", graphType)
	}

	if err != nil {
		return eris.Wrap(err, "failed to save graph")
	}

	flags.Log.Info(
		"Saved graph",
		"output", c.Output,
	)

	if c.RenderImage != "" {
		err = c.renderImage(ctx, flags)
		if err != nil {
			return err
		}
	}

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
		err := c.loadConfigFile(cfg)
		if err != nil {
			return nil, err
		}
	}

	c.applyConfigOverrides(cfg)

	return cfg, nil
}

// ExportConfigToFile writes the effective configuration to the given file path.
// The format is determined by the file extension (.yaml, .yml, or .json).
func (c *CLI) ExportConfigToFile(cfg *config.Config) error {
	if c.ExportConfig == "" {
		return nil
	}

	ext := strings.ToLower(filepath.Ext(c.ExportConfig))

	var (
		data []byte
		err  error
	)

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(cfg)
		if err != nil {
			return eris.Wrap(err, "failed to marshal config as YAML")
		}

	case ".json":
		data, err = json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return eris.Wrap(err, "failed to marshal config as JSON")
		}

	default:
		return eris.Errorf("unsupported file extension for config export: %s", ext)
	}

	err = os.WriteFile(c.ExportConfig, data, 0o600)
	if err != nil {
		return eris.Wrapf(err, "failed to write config file: %s", c.ExportConfig)
	}

	return nil
}

func (c *CLI) resolveGraphType(flags *Flags) string {
	graphType := c.GraphType
	if graphType == "" {
		graphType = flags.Config.GraphType
	}

	if graphType == "" {
		graphType = "dot"
	}

	return graphType
}

func (c *CLI) renderImage(ctx context.Context, flags *Flags) error {
	dotPath := ""
	if flags.Config != nil {
		dotPath = flags.Config.DotPath
	}

	dotExe, err := dot.FindExecutable(dotPath)
	if err != nil {
		return eris.Wrap(err, "failed to find dot executable")
	}

	ext := filepath.Ext(c.Output)
	imageFile := strings.TrimSuffix(c.Output, ext) + "." + c.RenderImage

	err = dot.RenderImage(ctx, dotExe, c.Output, imageFile, c.RenderImage)
	if err != nil {
		return eris.Wrap(err, "failed to render image")
	}

	flags.Log.Info(
		"Rendered image",
		"output", imageFile,
	)

	return nil
}

// applyConfigOverrides applies CLI flag overrides to the configuration.
func (c *CLI) applyConfigOverrides(cfg *config.Config) {
	if c.GroupByNamespace {
		cfg.GroupByNamespace = true
	}

	if c.AutoColor {
		cfg.AutoColor = true
	}

	if c.GraphType != "" {
		cfg.GraphType = c.GraphType
	}

	if c.Highlight != "" {
		c.applyHighlightOverrides(cfg)
	}
}

// applyHighlightOverrides parses the --highlight flag and appends matching style rules.
func (c *CLI) applyHighlightOverrides(cfg *config.Config) {
	color := "yellow"
	if cfg.HighlightColor != "" {
		color = cfg.HighlightColor
	}

	patterns := strings.FieldsFunc(
		c.Highlight,
		func(r rune) bool {
			return r == ',' || r == ';'
		})

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		rule := config.NodeStyleRule{
			Match:     pattern,
			FillColor: color,
			Style:     "filled",
		}
		cfg.NodeStyleRules = append(cfg.NodeStyleRules, rule)
	}
}

// applyAutoColor generates auto-color rules from the graph's namespaces and
// prepends them to cfg.NodeStyleRules so that user-defined rules take precedence.
func applyAutoColor(cfg *config.Config, gr *graph.Graph) {
	if !cfg.AutoColor {
		return
	}

	autoRules := autocolor.GenerateRules(gr)
	cfg.NodeStyleRules = append(autoRules, cfg.NodeStyleRules...)
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
		yamlErr := yaml.Unmarshal(raw, cfg)
		if yamlErr == nil {
			return nil
		}

		err = json.Unmarshal(raw, cfg)
	}

	if err != nil {
		return eris.Wrapf(err, "failed to parse config file: %s", c.Config)
	}

	return nil
}
