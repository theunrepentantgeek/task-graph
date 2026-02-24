package graphviz

import "github.com/theunrepentantgeek/task-graph/internal/config"

type nodeProperties struct {
	properties
}

func newNodeProperties() nodeProperties {
	return nodeProperties{
		properties: make(map[string]string),
	}
}

// AddAttributes adds the attributes from the given GraphvizNode configuration to the properties map.
func (p nodeProperties) AddAttributes(
	cfg *config.GraphvizNode,
) {
	if cfg == nil {
		return
	}

	if cfg.Color != "" {
		p.Add("color", cfg.Color)
	}
}
