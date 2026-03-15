package graphviz

import "github.com/theunrepentantgeek/task-graph/internal/config"

type edgeProperties struct {
	properties
}

func newEdgeProperties() edgeProperties {
	return edgeProperties{
		properties: make(map[string]string),
	}
}

// AddAttributes adds the attributes from the given GraphvizEdge configuration to the properties map.
func (p edgeProperties) AddAttributes(
	cfg *config.GraphvizEdge,
) {
	if cfg == nil {
		return
	}

	p.AddIfNotEmpty("color", cfg.Color)
	p.AddIfNotEmpty("style", cfg.Style)

	if cfg.Width > 0 {
		p.Addf("penwidth", "%d", cfg.Width)
	}
}
