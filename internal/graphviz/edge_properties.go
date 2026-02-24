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

	if cfg.Color != "" {
		p.Add("color", cfg.Color)
	}

	if cfg.Width > 0 {
		p.Addf("penwidth", "%d", cfg.Width)
	}

	if cfg.Style != "" {
		p.Add("style", cfg.Style)
	}
}
