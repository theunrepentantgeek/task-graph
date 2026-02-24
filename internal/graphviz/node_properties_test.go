package graphviz

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/task-graph/internal/config"
)

func TestNewNodeProperties_Default_ReturnsEmptyProperties(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	p := newNodeProperties()

	// Assert
	g.Expect(p.properties).NotTo(BeNil())
	g.Expect(p.properties).To(BeEmpty())
}

//
// AddAttributes tests.
//

func TestNodePropertiesAddAttributes_WithNilConfig_LeavesPropertiesEmpty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()

	// Act
	p.AddAttributes(nil)

	// Assert
	g.Expect(p.properties).To(BeEmpty())
}

func TestNodePropertiesAddAttributes_WhenColorConfigured_SetsColorAttribute(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	cfg := &config.GraphvizNode{
		Color: "red",
	}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).To(HaveKeyWithValue("color", "red"))
}

func TestNodePropertiesAddAttributes_WhenColorNotConfigured_OmitsColorAttribute(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	cfg := &config.GraphvizNode{
		Color: "",
	}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).NotTo(HaveKey("color"))
}

func TestNodePropertiesAddAttributes_WhenFillColorConfigured_SetsFillColorAttribute(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	cfg := &config.GraphvizNode{
		FillColor: "lightyellow",
	}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).To(HaveKeyWithValue("fillcolor", "lightyellow"))
}

func TestNodePropertiesAddAttributes_WhenFillColorNotConfigured_OmitsFillColorAttribute(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	cfg := &config.GraphvizNode{}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).NotTo(HaveKey("fillcolor"))
}

func TestNodePropertiesAddAttributes_WhenStyleConfigured_SetsStyleAttribute(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	cfg := &config.GraphvizNode{
		Style: "filled",
	}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).To(HaveKeyWithValue("style", "filled"))
}

func TestNodePropertiesAddAttributes_WhenStyleNotConfigured_OmitsStyleAttribute(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	cfg := &config.GraphvizNode{}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).NotTo(HaveKey("style"))
}

func TestNodePropertiesAddAttributes_WhenFontColorConfigured_SetsFontColorAttribute(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	cfg := &config.GraphvizNode{
		FontColor: "blue",
	}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).To(HaveKeyWithValue("fontcolor", "blue"))
}

func TestNodePropertiesAddAttributes_WhenFontColorNotConfigured_OmitsFontColorAttribute(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	cfg := &config.GraphvizNode{}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).NotTo(HaveKey("fontcolor"))
}

//
// AddStyleRuleAttributes tests.
//

func TestNodePropertiesAddStyleRuleAttributes_WhenPatternMatchesExactName_AppliesAttributes(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	rule := config.GraphvizStyleRule{
		Match: "alpha",
		Color: "red",
	}

	// Act
	p.AddStyleRuleAttributes("alpha", rule)

	// Assert
	g.Expect(p.properties).To(HaveKeyWithValue("color", "red"))
}

func TestNodePropertiesAddStyleRuleAttributes_WhenPatternDoesNotMatch_LeavesPropertiesUnchanged(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	rule := config.GraphvizStyleRule{
		Match: "alpha",
		Color: "red",
	}

	// Act
	p.AddStyleRuleAttributes("beta", rule)

	// Assert
	g.Expect(p.properties).NotTo(HaveKey("color"))
}

func TestNodePropertiesAddStyleRuleAttributes_WhenPatternUsesWildcard_MatchesMultipleNames(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		nodeID  string
		pattern string
		matches bool
	}{
		"wildcard matches prefix":          {nodeID: "test-build", pattern: "test-*", matches: true},
		"wildcard matches suffix":          {nodeID: "build-test", pattern: "*-test", matches: true},
		"wildcard matches full name":       {nodeID: "anything", pattern: "*", matches: true},
		"wildcard does not match mismatch": {nodeID: "deploy", pattern: "test-*", matches: false},
		"question mark matches one char":  {nodeID: "ab", pattern: "a?", matches: true},
		"question mark no match":          {nodeID: "abc", pattern: "a?", matches: false},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			// Arrange
			p := newNodeProperties()
			rule := config.GraphvizStyleRule{
				Match: c.pattern,
				Color: "red",
			}

			// Act
			p.AddStyleRuleAttributes(c.nodeID, rule)

			// Assert
			if c.matches {
				g.Expect(p.properties).To(HaveKeyWithValue("color", "red"))
			} else {
				g.Expect(p.properties).NotTo(HaveKey("color"))
			}
		})
	}
}

func TestNodePropertiesAddStyleRuleAttributes_WhenRuleHasAllFields_AppliesAllAttributes(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	rule := config.GraphvizStyleRule{
		Match:     "alpha",
		Color:     "red",
		FillColor: "lightyellow",
		Style:     "filled",
		FontColor: "blue",
	}

	// Act
	p.AddStyleRuleAttributes("alpha", rule)

	// Assert
	g.Expect(p.properties).To(HaveKeyWithValue("color", "red"))
	g.Expect(p.properties).To(HaveKeyWithValue("fillcolor", "lightyellow"))
	g.Expect(p.properties).To(HaveKeyWithValue("style", "filled"))
	g.Expect(p.properties).To(HaveKeyWithValue("fontcolor", "blue"))
}

func TestNodePropertiesAddStyleRuleAttributes_LastRuleWins_WhenMultipleRulesMatch(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newNodeProperties()
	rule1 := config.GraphvizStyleRule{Match: "alpha", Color: "red"}
	rule2 := config.GraphvizStyleRule{Match: "alpha", Color: "blue"}

	// Act
	p.AddStyleRuleAttributes("alpha", rule1)
	p.AddStyleRuleAttributes("alpha", rule2)

	// Assert
	g.Expect(p.properties).To(HaveKeyWithValue("color", "blue"))
}
