package config

type Graphviz struct {
	// Font is the font used for labels in the Graphviz output. It can be any valid Graphviz font.
	// https://graphviz.org/docs/attrs/fontname/
	Font string `json:"font,omitempty" yaml:"font,omitempty"`

	// FontSize is the font size used in the Graphviz output, in points.
	// https://graphviz.org/docs/attrs/fontsize/
	FontSize int `json:"fontSize,omitempty" yaml:"fontSize,omitempty"`

	// DependencyEdges is the presentation for dependency edges between tasks
	DependencyEdges *GraphvizEdge `json:"dependencyEdges,omitempty" yaml:"dependencyEdges,omitempty"`

	// CallEdges is the presentation for call edges between tasks
	CallEdges *GraphvizEdge `json:"callEdges,omitempty" yaml:"callEdges,omitempty"`

	// TaskNodes is the presentation for task nodes
	TaskNodes *GraphvizNode `json:"taskNodes,omitempty" yaml:"taskNodes,omitempty"`

	// VariableNodes is the presentation for global variable nodes
	VariableNodes *GraphvizNode `json:"variableNodes,omitempty" yaml:"variableNodes,omitempty"`

	// VariableEdges is the presentation for edges from variables to tasks
	VariableEdges *GraphvizEdge `json:"variableEdges,omitempty" yaml:"variableEdges,omitempty"`
}

type GraphvizNode struct {
	// Color is the color of the node border. It can be any valid Graphviz color.
	// https://graphviz.org/docs/attrs/color/
	Color string `json:"color,omitempty" yaml:"color,omitempty"`

	// FillColor is the fill/background color of the node. It can be any valid Graphviz color.
	// https://graphviz.org/docs/attrs/fillcolor/
	FillColor string `json:"fillColor,omitempty" yaml:"fillColor,omitempty"`

	// Style is the style of the node (e.g., "filled", "dashed", "bold").
	// https://graphviz.org/docs/attr-types/style/
	Style string `json:"style,omitempty" yaml:"style,omitempty"`

	// FontColor is the color of the label text. It can be any valid Graphviz color.
	// https://graphviz.org/docs/attrs/fontcolor/
	FontColor string `json:"fontColor,omitempty" yaml:"fontColor,omitempty"`
}

type GraphvizEdge struct {
	// Color is the color of the edge. It can be any valid Graphviz color.
	// https://graphviz.org/docs/attrs/color/
	Color string `json:"color,omitempty" yaml:"color,omitempty"`

	// Width is the width of the edge. It can be any positive integer.
	Width int `json:"width,omitempty" yaml:"width,omitempty"`

	// Style is the style of the edge. It can be any valid Graphviz style.
	// https://graphviz.org/docs/attr-types/style/
	Style string `json:"style,omitempty" yaml:"style,omitempty"`
}

func newGraphViz() *Graphviz {
	return &Graphviz{
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
	}
}
