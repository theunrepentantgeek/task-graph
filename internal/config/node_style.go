package config

// NodeStyleRule defines a style rule that is applied to task nodes whose names match the given pattern.
// These rules work across all graph types (dot, mermaid, etc.).
type NodeStyleRule struct {
	// Match is the pattern used to match task names. Supports wildcards (* and ?).
	Match string `json:"match,omitempty" yaml:"match,omitempty"`

	// Color is the color of the node border.
	Color string `json:"color,omitempty" yaml:"color,omitempty"`

	// FillColor is the fill/background color of the node.
	FillColor string `json:"fillColor,omitempty" yaml:"fillColor,omitempty"`

	// Style is the style of the node (e.g., "filled", "dashed", "bold").
	Style string `json:"style,omitempty" yaml:"style,omitempty"`

	// FontColor is the color of the label text.
	FontColor string `json:"fontColor,omitempty" yaml:"fontColor,omitempty"`
}
