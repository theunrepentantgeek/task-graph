package graphviz

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/task-graph/internal/config"
)

func TestNewEdgeProperties_Default_ReturnsEmptyProperties(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	p := newEdgeProperties()

	// Assert
	g.Expect(p.properties).NotTo(BeNil())
	g.Expect(p.properties).To(BeEmpty())
}

//
// AddAttributes tests.
//

func TestEdgePropertiesAddAttributes_WithNilConfig_LeavesPropertiesEmpty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newEdgeProperties()

	// Act
	p.AddAttributes(nil)

	// Assert
	g.Expect(p.properties).To(BeEmpty())
}

func TestEdgePropertiesAddAttributes_WhenColorConfigured_SetsColorAttribute(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newEdgeProperties()
	cfg := &config.GraphvizEdge{
		Color: "red",
	}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).To(HaveKeyWithValue("color", "red"))
}

func TestEdgePropertiesAddAttributes_WhenWidthConfigured_SetsPenwidthAttribute(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newEdgeProperties()
	cfg := &config.GraphvizEdge{
		Width: 3,
	}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).To(HaveKeyWithValue("penwidth", "3"))
}

func TestEdgePropertiesAddAttributes_WhenStyleConfigured_SetsStyleAttribute(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newEdgeProperties()
	cfg := &config.GraphvizEdge{
		Style: "dashed",
	}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).To(HaveKeyWithValue("style", "dashed"))
}

func TestEdgePropertiesAddAttributes_WhenMultipleAttributesConfigured_SetsAllAttributes(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newEdgeProperties()
	cfg := &config.GraphvizEdge{
		Color: "blue",
		Width: 2,
		Style: "bold",
	}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).To(HaveKeyWithValue("color", "blue"))
	g.Expect(p.properties).To(HaveKeyWithValue("penwidth", "2"))
	g.Expect(p.properties).To(HaveKeyWithValue("style", "bold"))
}

func TestEdgePropertiesAddAttributes_WhenWidthNotPositive_OmitsPenwidth(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newEdgeProperties()
	cfg := &config.GraphvizEdge{
		Color: "green",
		Width: 0,
		Style: "solid",
	}

	// Act
	p.AddAttributes(cfg)

	// Assert
	g.Expect(p.properties).NotTo(HaveKey("penwidth"))
	g.Expect(p.properties).To(HaveKeyWithValue("color", "green"))
	g.Expect(p.properties).To(HaveKeyWithValue("style", "solid"))
}
