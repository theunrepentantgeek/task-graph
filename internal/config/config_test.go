package config

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestNew_GraphvizFont_ReturnsVerdana(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := New()

	g.Expect(cfg.Graphviz).NotTo(gomega.BeNil())
	g.Expect(cfg.Graphviz.Font).To(gomega.Equal("Verdana"))
}

func TestNew_GraphvizFontSize_Returns16(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := New()

	g.Expect(cfg.Graphviz).NotTo(gomega.BeNil())
	g.Expect(cfg.Graphviz.FontSize).To(gomega.Equal(16))
}

func TestNew_GraphvizDependencyEdges_ReturnsBlackSolidWidth1(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := New()

	g.Expect(cfg.Graphviz).NotTo(gomega.BeNil())
	g.Expect(cfg.Graphviz.DependencyEdges).NotTo(gomega.BeNil())
	g.Expect(cfg.Graphviz.DependencyEdges.Color).To(gomega.Equal("black"))
	g.Expect(cfg.Graphviz.DependencyEdges.Style).To(gomega.Equal("solid"))
	g.Expect(cfg.Graphviz.DependencyEdges.Width).To(gomega.Equal(1))
}

func TestNew_GraphvizCallEdges_ReturnsBlueDashedWidth1(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := New()

	g.Expect(cfg.Graphviz).NotTo(gomega.BeNil())
	g.Expect(cfg.Graphviz.CallEdges).NotTo(gomega.BeNil())
	g.Expect(cfg.Graphviz.CallEdges.Color).To(gomega.Equal("blue"))
	g.Expect(cfg.Graphviz.CallEdges.Style).To(gomega.Equal("dashed"))
	g.Expect(cfg.Graphviz.CallEdges.Width).To(gomega.Equal(1))
}

func TestNew_GraphvizTaskNodes_ReturnsBlackBorder(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := New()

	g.Expect(cfg.Graphviz).NotTo(gomega.BeNil())
	g.Expect(cfg.Graphviz.TaskNodes).NotTo(gomega.BeNil())
	g.Expect(cfg.Graphviz.TaskNodes.Color).To(gomega.Equal("black"))
}

func TestNew_MermaidDirection_ReturnsTD(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := New()

	g.Expect(cfg.Mermaid).NotTo(gomega.BeNil())
	g.Expect(cfg.Mermaid.Direction).To(gomega.Equal("TD"))
}

func TestNew_NodeStyleRules_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := New()

	g.Expect(cfg.NodeStyleRules).To(gomega.BeEmpty())
}

func TestNew_GroupByNamespace_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := New()

	g.Expect(cfg.GroupByNamespace).To(gomega.BeFalse())
}

func TestNew_AutoColor_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := New()

	g.Expect(cfg.AutoColor).To(gomega.BeFalse())
}

func TestNodeStyleRule_Fields_RoundTrip(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	rule := NodeStyleRule{
		Match:     "build*",
		Color:     "red",
		FillColor: "lightyellow",
		Style:     "filled",
		FontColor: "navy",
	}

	g.Expect(rule.Match).To(gomega.Equal("build*"))
	g.Expect(rule.Color).To(gomega.Equal("red"))
	g.Expect(rule.FillColor).To(gomega.Equal("lightyellow"))
	g.Expect(rule.Style).To(gomega.Equal("filled"))
	g.Expect(rule.FontColor).To(gomega.Equal("navy"))
}
