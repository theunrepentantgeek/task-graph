package config

// MermaidStyle holds CSS-like style properties for Mermaid classDef directives.
type MermaidStyle struct {
	// Fill is the background fill color.
	Fill string `json:"fill,omitempty" yaml:"fill,omitempty"`

	// Stroke is the border/line color.
	Stroke string `json:"stroke,omitempty" yaml:"stroke,omitempty"`

	// Color is the text color.
	Color string `json:"color,omitempty" yaml:"color,omitempty"`
}

// Mermaid holds configuration specific to Mermaid flowchart output.
type Mermaid struct {
	// Direction is the direction of the flowchart.
	// Valid values: TD (top-down), LR (left-right), BT (bottom-top), RL (right-left).
	// Defaults to "TD" when not specified.
	Direction string `json:"direction,omitempty" yaml:"direction,omitempty"`

	// VariableNodes holds style properties for variable nodes in the Mermaid output.
	VariableNodes *MermaidStyle `json:"variableNodes,omitempty" yaml:"variableNodes,omitempty"`

	// VariableEdges holds style properties for variable edges in the Mermaid output.
	VariableEdges *MermaidStyle `json:"variableEdges,omitempty" yaml:"variableEdges,omitempty"`
}
