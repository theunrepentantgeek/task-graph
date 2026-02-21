package config

type Config struct {
	// Graphviz is the configuration for the Graphviz output.
	Graphviz *Graphviz `json:"graphviz"`
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
