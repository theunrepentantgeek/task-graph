package graphviz

import (
	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/namespace"
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

// AddStyleRuleAttributes adds the attributes from the given NodeStyleRule to the properties map
// if the given node ID matches the rule's pattern.
func (p nodeProperties) AddStyleRuleAttributes(
	nodeID string,
	rule config.NodeStyleRule,
) error {
	re, err := namespace.CompileMatchPattern(rule.Match)
	if err != nil {
		return eris.Wrapf(err, "failed to compile match pattern %q", rule.Match)
	}

	if !re.MatchString(nodeID) {
		return nil
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

	return nil
}
