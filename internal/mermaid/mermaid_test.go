package mermaid

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"
	"github.com/sebdah/goldie/v2"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
)

func TestWriteTo_WithNodesAndEdges_WritesSortedMermaid(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := config.New()
	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	gg := goldie.New(t)
	g.Expect(gg.WithFixtureDir("testdata")).To(gomega.Succeed())

	gg.Assert(t, "sample_graph", buf.Bytes())
}

func TestWriteTo_WithGroupByNamespace_WritesNamespaceSubgraphs(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildNamespacedGraph(t)

	cfg := config.New()
	cfg.GroupByNamespace = true
	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	gg := goldie.New(t)
	g.Expect(gg.WithFixtureDir("testdata")).To(gomega.Succeed())

	gg.Assert(t, "namespace_graph", buf.Bytes())
}

func TestWriteTo_WithStyleRules_AppliesClassDefs(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := config.New()
	cfg.NodeStyleRules = []config.NodeStyleRule{
		{Match: "alpha", Color: "red", FillColor: "lightyellow"},
		{Match: "b*", FontColor: "blue"},
	}

	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	gg := goldie.New(t)
	g.Expect(gg.WithFixtureDir("testdata")).To(gomega.Succeed())

	gg.Assert(t, "sample_graph_with_style_rules", buf.Bytes())
}

func TestWriteTo_NilGraph_ReturnsError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	cfg := config.New()

	err := WriteTo(&buf, nil, cfg)

	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("mermaid: graph is nil")))
}

func TestSaveTo_WritesFileToDisk(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "graph.mmd")
	cfg := config.New()

	err := SaveTo(filePath, buildSampleGraph(t), cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	content, readErr := os.ReadFile(filePath)
	g.Expect(readErr).NotTo(gomega.HaveOccurred())

	gg := goldie.New(t)
	g.Expect(gg.WithFixtureDir("testdata")).To(gomega.Succeed())

	gg.Assert(t, "sample_graph", content)
}

func TestWriteTo_WithCustomDirection_UsesConfiguredDirection(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := config.New()
	cfg.Mermaid.Direction = "LR"
	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(buf.String()).To(gomega.HavePrefix("flowchart LR\n"))
}

func TestWriteTo_WithCallEdge_UsesDashedArrow(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := graph.New()
	alpha := gr.AddNode("alpha")
	beta := gr.AddNode("beta")
	e := alpha.AddEdge(beta)
	e.SetClass("call")

	cfg := config.New()
	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(buf.String()).To(gomega.ContainSubstring("alpha -.-> beta"))
}

func TestWriteTo_WithNodeLabel_UsesLabel(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := graph.New()
	n := gr.AddNode("test-node")
	n.Label = "Custom Label"

	cfg := config.New()
	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(buf.String()).To(gomega.ContainSubstring(`test-node["Custom Label"]`))
}

func TestWriteTo_WithVariableNodes_WritesVariableNodesAfterTasks(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildGraphWithVariables(t)

	cfg := config.New()
	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	gg := goldie.New(t)
	g.Expect(gg.WithFixtureDir("testdata")).To(gomega.Succeed())

	gg.Assert(t, "sample_graph_with_variables", buf.Bytes())
}

func buildSampleGraph(t *testing.T) *graph.Graph {
	t.Helper()

	gr := graph.New()

	alpha := gr.AddNode("alpha")
	beta := gr.AddNode("beta")
	gamma := gr.AddNode("gamma")

	alpha.AddEdge(beta)
	alpha.AddEdge(gamma)

	labeled := beta.AddEdge(gamma)
	labeled.SetLabel("next")

	return gr
}

func TestWriteTo_WithNilConfig_WritesTDFlowchart(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	err := WriteTo(&buf, gr, nil)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(buf.String()).To(gomega.HavePrefix("flowchart TD\n"))
}

func TestWriteTo_WithNilMermaidConfig_WritesTDFlowchart(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := config.New()
	cfg.Mermaid = nil

	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(buf.String()).To(gomega.HavePrefix("flowchart TD\n"))
}

func TestWriteTo_WithStyleRuleNoVisualProperties_SkipsClassDef(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := config.New()
	cfg.NodeStyleRules = []config.NodeStyleRule{
		{Match: "alpha"}, // no Fill, Color, or FontColor set
	}

	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(buf.String()).NotTo(gomega.ContainSubstring("classDef"))
}

func TestWriteTo_WithInvalidStylePattern_ReturnsError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := config.New()
	cfg.NodeStyleRules = []config.NodeStyleRule{
		{Match: "[unclosed", FillColor: "red"},
	}

	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("[unclosed"))
}

func TestWriteTo_WithVariableNodeNoDescription_UsesLabelOnly(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := graph.New()

	build := gr.AddNode("build")

	noDesc := gr.AddNode("var:TAG")
	noDesc.Kind = graph.NodeKindVariable
	noDesc.Label = "TAG"
	// Description intentionally left empty

	noDesc.AddEdge(build).SetClass("var")

	cfg := config.New()
	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(buf.String()).To(gomega.ContainSubstring(`var_TAG("TAG")`))
}

func TestWriteTo_WithCustomVariableNodeStyle_UsesConfiguredStyle(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildGraphWithVariables(t)

	cfg := config.New()
	cfg.Mermaid.VariableNodes = &config.MermaidStyle{
		Fill:   "#ffd700",
		Stroke: "#333",
		Color:  "#000",
	}

	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	output := buf.String()
	g.Expect(output).To(gomega.ContainSubstring("fill:#ffd700"))
	g.Expect(output).To(gomega.ContainSubstring("stroke:#333"))
	g.Expect(output).To(gomega.ContainSubstring("color:#000"))
}

func TestWriteTo_WithPartialVariableNodeStyle_UsesDefaultsForMissingFields(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildGraphWithVariables(t)

	cfg := config.New()
	// Only Fill is set; empty Stroke and Color fields are omitted from the class definition
	cfg.Mermaid.VariableNodes = &config.MermaidStyle{
		Fill: "#ffd700",
	}

	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	output := buf.String()
	g.Expect(output).To(gomega.ContainSubstring("fill:#ffd700"))
	g.Expect(output).NotTo(gomega.ContainSubstring("stroke:"))
	g.Expect(output).NotTo(gomega.ContainSubstring("color:"))
}

func TestWriteTo_WithAllEmptyVariableNodeStyle_UsesDefaults(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildGraphWithVariables(t)

	cfg := config.New()
	// All fields empty → variableClassDefParts returns hardcoded defaults
	cfg.Mermaid.VariableNodes = &config.MermaidStyle{}

	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	output := buf.String()
	g.Expect(output).To(gomega.ContainSubstring("fill:#e8e8e8"))
	g.Expect(output).To(gomega.ContainSubstring("stroke:#666"))
}

func buildNamespacedGraph(t *testing.T) *graph.Graph {
	t.Helper()

	gr := graph.New()

	build := gr.AddNode("build")
	cmdBuild := gr.AddNode("cmd:build")
	cmdTestUnit := gr.AddNode("cmd:test:unit")
	cmdTestGolden := gr.AddNode("cmd:test:golden")

	cmdBuild.AddEdge(build)

	dep1 := cmdBuild.AddEdge(cmdTestUnit)
	dep1.SetClass("dep")

	dep2 := cmdBuild.AddEdge(cmdTestGolden)
	dep2.SetClass("dep")

	return gr
}

func buildGraphWithVariables(t *testing.T) *graph.Graph {
	t.Helper()

	gr := graph.New()

	// Task nodes
	build := gr.AddNode("build")
	build.Description = "Build the project"
	test := gr.AddNode("test")

	build.AddEdge(test).SetClass("dep")

	// Variable nodes
	pkg := gr.AddNode("var:PACKAGE")
	pkg.Kind = graph.NodeKindVariable
	pkg.Label = "PACKAGE"
	pkg.Description = "github.com/example/project"

	ver := gr.AddNode("var:VERSION")
	ver.Kind = graph.NodeKindVariable
	ver.Label = "VERSION"
	ver.Description = "sh: git describe --tags"

	// Variable edges (variable -> task in graph model)
	pkgEdge := pkg.AddEdge(build)
	pkgEdge.SetClass("var")

	verEdge := ver.AddEdge(build)
	verEdge.SetClass("var")

	return gr
}
