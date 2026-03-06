package config

// Mermaid holds configuration specific to Mermaid flowchart output.
type Mermaid struct {
	// Direction is the direction of the flowchart.
	// Valid values: TD (top-down), LR (left-right), BT (bottom-top), RL (right-left).
	// Defaults to "TD" when not specified.
	Direction string `json:"direction,omitempty" yaml:"direction,omitempty"`
}
