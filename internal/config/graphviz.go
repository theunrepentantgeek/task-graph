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
}

type GraphvizNode struct {
	// Color is the color of the node. It can be any valid Graphviz color.
	// https://graphviz.org/docs/attrs/color/
	Color string `json:"color,omitempty" yaml:"color,omitempty"`
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
