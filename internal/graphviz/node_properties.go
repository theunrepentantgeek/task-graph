package graphviz

import (
	"sync"

	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/namespace"
)

type matchPattern interface {
	MatchString(s string) bool
}

var nodeStylePatternCache sync.Map // map[string]matchPattern

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

	p.AddIfNotEmpty("color", cfg.Color)
	p.AddIfNotEmpty("fillcolor", cfg.FillColor)
	p.AddIfNotEmpty("style", cfg.Style)
	p.AddIfNotEmpty("fontcolor", cfg.FontColor)
}

// AddStyleRuleAttributes adds the attributes from the given NodeStyleRule to the properties map
// if the given node ID matches the rule's pattern.
func (p nodeProperties) AddStyleRuleAttributes(
	nodeID string,
	rule config.NodeStyleRule,
) error {
	value, ok := nodeStylePatternCache.Load(rule.Match)
	if !ok {
		re, err := namespace.CompileMatchPattern(rule.Match)
		if err != nil {
			return eris.Wrapf(err, "failed to compile match pattern %q", rule.Match)
		}

		nodeStylePatternCache.Store(rule.Match, re)
		value = re
	}

	pattern, ok := value.(matchPattern)
	if !ok {
		return eris.New("cached node style pattern does not implement MatchString")
	}

	if !pattern.MatchString(nodeID) {
		return nil
	}

	p.AddIfNotEmpty("color", rule.Color)
	p.AddIfNotEmpty("fillcolor", rule.FillColor)
	p.AddIfNotEmpty("style", rule.Style)
	p.AddIfNotEmpty("fontcolor", rule.FontColor)

	return nil
}
