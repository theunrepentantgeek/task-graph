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
	"github.com/theunrepentantgeek/task-graph/internal/namespace"
	"github.com/theunrepentantgeek/task-graph/internal/taskgraph"
)

// graphTypeDot and graphTypeMermaid are the supported graph output formats.
const (
	graphTypeDot     = "dot"
	graphTypeMermaid = "mermaid"
)

//nolint:tagalign // Not useful here because different members have different tags.
type CLI struct {
	Taskfile string `arg:"" help:"Path to the taskfile to process."`
	Output   string `help:"Path to the output file." long:"output" required:"true" short:"o"`
	Config   string `help:"Path to a config file (YAML or JSON)." long:"config" short:"c"`

	GroupByNamespace bool `help:"Group tasks in the same namespace together in the output." long:"group-by-namespace"`

	AutoColor bool `help:"Automatically color nodes by namespace using a built-in palette." long:"auto-color"`

	//nolint:revive // Intentionally long name for clarity in the CLI help.
	IncludeGlobalVars bool `help:"Include global variables as nodes in the graph, with edges to consuming tasks." long:"include-global-vars"`

	GraphType string `help:"Type of graph to generate (dot or mermaid). Defaults to dot." long:"graph-type"`

	//nolint:revive // Intentionally long line for clarity in the CLI help.
	Highlight string `help:"Highlight specific tasks in the graph. Accepts task names or glob patterns, separated by commas or semicolons." long:"highlight"`

	//nolint:revive // Intentionally long name for clarity in the CLI help.
	RenderImage string `help:"Render the graph as an image using graphviz dot. Specify the file type (e.g. png, svg)." long:"render-image"`

	//nolint:revive // Intentionally long name for clarity in the CLI help.
	ExportConfig string `help:"Export the effective configuration to a file (YAML or JSON based on file extension)." long:"export-config"`

	//nolint:revive // Intentionally long name for clarity in the CLI help.
	Focus   string `help:"Show only tasks matching the given patterns together with all their transitive dependencies and dependents. Accepts task names or glob patterns, separated by commas or semicolons." long:"focus"`
	Verbose bool   `help:"Enable verbose logging."`
}

// Run executes the CLI command with the given flags.
//
//nolint:revive // Difficult to simplify
func (c *CLI) Run(
	flags *Flags,
) error {
	ctx := context.Background()

	tf, err := loader.Load(ctx, c.Taskfile)
	if err != nil {
		return eris.Wrap(err, "failed to load taskfile")
	}

	flags.Log.Info(
		"Loaded taskfile",
		"taskfile", c.Taskfile,
		"tasks", tf.Tasks.Len())

	builder := taskgraph.New(tf)
	builder.IncludeGlobalVars = flags.Config.IncludeGlobalVars
	gr := builder.Build()

	if c.Focus != "" {
		gr, err = applyFocus(gr, c.Focus)
		if err != nil {
			return eris.Wrap(err, "failed to apply focus filter")
		}
	}

	applyAutoColor(flags.Config, gr)

	err = c.saveGraph(gr, flags)
	if err != nil {
		return err
	}

	if c.RenderImage != "" {
		err = c.renderImage(ctx, flags)
		if err != nil {
			return err
		}
	}

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

func (c *CLI) saveGraph(
	gr *graph.Graph,
	flags *Flags,
) error {
	graphType := c.resolveGraphType(flags)

	var err error

	switch graphType {
	case graphTypeDot:
		err = graphviz.SaveTo(c.Output, gr, flags.Config)
	case graphTypeMermaid:
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

	return nil
}

func (c *CLI) resolveGraphType(flags *Flags) string {
	graphType := c.GraphType
	if graphType == "" {
		graphType = flags.Config.GraphType
	}

	if graphType == "" {
		graphType = graphTypeDot
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

	if c.IncludeGlobalVars {
		cfg.IncludeGlobalVars = true
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

// applyFocus returns a new graph containing only the nodes that match any of the
// given comma-or-semicolon-separated patterns (glob-style), together with all
// nodes transitively reachable from them in either direction.
//
//nolint:revive // Difficult to simplify
func applyFocus(
	gr *graph.Graph,
	focusPatterns string,
) (*graph.Graph, error) {
	patterns := strings.FieldsFunc(
		focusPatterns,
		func(r rune) bool {
			return r == ',' || r == ';'
		})

	seeds := make(map[string]bool)

	for _, raw := range patterns {
		pattern := strings.TrimSpace(raw)
		if pattern == "" {
			continue
		}

		re, err := namespace.CompileMatchPattern(pattern)
		if err != nil {
			return nil, eris.Wrapf(err, "invalid focus pattern %q", pattern)
		}

		for node := range gr.Nodes() {
			if re.MatchString(node.ID()) {
				seeds[node.ID()] = true
			}
		}
	}

	if len(seeds) == 0 {
		return gr, nil
	}

	reachable := gr.ReachableFrom(seeds)

	return gr.FilterNodes(reachable), nil
}
