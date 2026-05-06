package mermaid

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
)

// variableClassDefParts tests

func TestVariableClassDefParts_NilConfig_ReturnsDefaults(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	parts := variableClassDefParts(nil)

	g.Expect(parts).To(Equal([]string{"fill:#e8e8e8", "stroke:#666"}))
}

func TestVariableClassDefParts_NilMermaidConfig_ReturnsDefaults(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := config.New()
	cfg.Mermaid = nil

	parts := variableClassDefParts(cfg)

	g.Expect(parts).To(Equal([]string{"fill:#e8e8e8", "stroke:#666"}))
}

func TestVariableClassDefParts_NilVariableNodes_ReturnsDefaults(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := config.New()
	cfg.Mermaid.VariableNodes = nil

	parts := variableClassDefParts(cfg)

	g.Expect(parts).To(Equal([]string{"fill:#e8e8e8", "stroke:#666"}))
}

func TestVariableClassDefParts_WithFillOnly_ReturnsFillPart(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := config.New()
	cfg.Mermaid.VariableNodes = &config.MermaidStyle{Fill: "#ff0000"}

	parts := variableClassDefParts(cfg)

	g.Expect(parts).To(Equal([]string{"fill:#ff0000"}))
}

func TestVariableClassDefParts_WithStrokeOnly_ReturnsStrokePart(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := config.New()
	cfg.Mermaid.VariableNodes = &config.MermaidStyle{Stroke: "#0000ff"}

	parts := variableClassDefParts(cfg)

	g.Expect(parts).To(Equal([]string{"stroke:#0000ff"}))
}

func TestVariableClassDefParts_WithColorOnly_ReturnsColorPart(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := config.New()
	cfg.Mermaid.VariableNodes = &config.MermaidStyle{Color: "#ffffff"}

	parts := variableClassDefParts(cfg)

	g.Expect(parts).To(Equal([]string{"color:#ffffff"}))
}

func TestVariableClassDefParts_WithAllFields_ReturnsAllParts(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := config.New()
	cfg.Mermaid.VariableNodes = &config.MermaidStyle{
		Fill:   "#ff0000",
		Stroke: "#0000ff",
		Color:  "#ffffff",
	}

	parts := variableClassDefParts(cfg)

	g.Expect(parts).To(Equal([]string{"fill:#ff0000", "stroke:#0000ff", "color:#ffffff"}))
}

func TestVariableClassDefParts_EmptyVariableNodes_ReturnsDefaults(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := config.New()
	cfg.Mermaid.VariableNodes = &config.MermaidStyle{}

	parts := variableClassDefParts(cfg)

	g.Expect(parts).To(Equal([]string{"fill:#e8e8e8", "stroke:#666"}))
}

// variableDisplayLabel tests

func TestVariableDisplayLabel_NoDescription_ReturnsDisplayLabel(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	node := graph.New().AddNode("var:FOO")
	node.Kind = graph.NodeKindVariable
	node.Label = "FOO"

	label := variableDisplayLabel(node)

	g.Expect(label).To(Equal("FOO"))
}

func TestVariableDisplayLabel_WithDescription_ReturnsLabelColonDescription(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	node := graph.New().AddNode("var:FOO")
	node.Kind = graph.NodeKindVariable
	node.Label = "FOO"
	node.Description = "sh: echo hi"

	label := variableDisplayLabel(node)

	g.Expect(label).To(Equal("FOO: sh: echo hi"))
}
