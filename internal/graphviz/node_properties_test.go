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
