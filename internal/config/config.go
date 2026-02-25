package config

type Config struct {
	// Graphviz is the configuration for the Graphviz output.
	Graphviz *Graphviz `json:"graphviz"`

	// DotPath is the path to the dot executable, or the folder containing it.
	// If not specified, dot will be looked up on the PATH.
	DotPath string `json:"dotPath,omitempty" yaml:"dotPath,omitempty"`
}

// New creates a new Config with default values.
func New() *Config {
	return &Config{
		Graphviz: &Graphviz{
			Font:     "Verdana",
			FontSize: 16,
			DependencyEdges: &GraphvizEdge{
				Color: "black",
				Width: 1,
				Style: "solid",
			},
			CallEdges: &GraphvizEdge{
				Color: "blue",
				Width: 1,
				Style: "dashed",
			},
			TaskNodes: &GraphvizNode{
				Color: "black",
			},
		},
	}
}
