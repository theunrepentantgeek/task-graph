package config

type Config struct {
	// GroupByNamespace controls whether tasks in the same namespace are grouped together
	// in the output. Namespace is defined by a common prefix prior to a colon (`:`).
	GroupByNamespace bool `json:"groupByNamespace,omitempty" yaml:"groupByNamespace,omitempty"`

	// GraphType is the type of graph to generate. Valid values: dot, mermaid.
	// Defaults to "dot" when not specified.
	GraphType string `json:"graphType,omitempty" yaml:"graphType,omitempty"`

	// HighlightColor is the fill color used for highlighted task nodes (from --highlight flag).
	// Defaults to "yellow" when not specified.
	HighlightColor string `json:"highlightColor,omitempty" yaml:"highlightColor,omitempty"`

	// AutoColor controls whether nodes are automatically colored by namespace.
	// When true, a distinct fill color from a built-in palette is assigned to each
	// namespace found in the taskfile. User-defined NodeStyleRules take precedence.
	AutoColor bool `json:"autoColor,omitempty" yaml:"autoColor,omitempty"`

	// IncludeGlobalVars controls whether global Taskfile variables are included
	// as nodes in the generated graph, with edges to the tasks that reference them.
	IncludeGlobalVars bool `json:"includeGlobalVars,omitempty" yaml:"includeGlobalVars,omitempty"`

	// NodeStyleRules are additional style rules applied to matching task nodes, in order.
	// All matching rules are applied; in case of conflicts, the last matching rule wins.
	// These rules work across all graph types.
	NodeStyleRules []NodeStyleRule `json:"nodeStyleRules,omitempty" yaml:"nodeStyleRules,omitempty"`

	// Graphviz is the configuration for the Graphviz dot output.
	Graphviz *Graphviz `json:"graphviz,omitempty" yaml:"graphviz,omitempty"`

	// Mermaid is the configuration for the Mermaid flowchart output.
	Mermaid *Mermaid `json:"mermaid,omitempty" yaml:"mermaid,omitempty"`

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
			VariableNodes: &GraphvizNode{
				Color:     "#666666",
				FillColor: "#e8e8e8",
				Style:     "filled",
			},
			VariableEdges: &GraphvizEdge{
				Color: "#228B22",
				Width: 1,
				Style: "dotted",
			},
		},
		Mermaid: &Mermaid{
			Direction: "TD",
		},
	}
}
