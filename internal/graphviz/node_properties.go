package graphviz

import (
	"path"

	"github.com/theunrepentantgeek/task-graph/internal/config"
)

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

	if cfg.FillColor != "" {
		p.Add("fillcolor", cfg.FillColor)
	}

	if cfg.Style != "" {
		p.Add("style", cfg.Style)
	}

	if cfg.FontColor != "" {
		p.Add("fontcolor", cfg.FontColor)
	}
}

// AddStyleRuleAttributes adds the attributes from the given GraphvizStyleRule to the properties map
// if the given node ID matches the rule's pattern.
func (p nodeProperties) AddStyleRuleAttributes(
	nodeID string,
	rule config.GraphvizStyleRule,
) {
	matched, err := path.Match(rule.Match, nodeID)
	if err != nil || !matched {
		return
	}

	if rule.Color != "" {
		p.Add("color", rule.Color)
	}

	if rule.FillColor != "" {
		p.Add("fillcolor", rule.FillColor)
	}

	if rule.Style != "" {
		p.Add("style", rule.Style)
	}

	if rule.FontColor != "" {
		p.Add("fontcolor", rule.FontColor)
	}
}
