package graphviz

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/onsi/gomega"
	"github.com/sebdah/goldie/v2"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/indentwriter"
)

func TestWriteTo_WithNodesAndEdges_WritesSortedGraphviz(t *testing.T) {
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

func TestWriteTo_WithStyleRules_AppliesMatchingStyles(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := config.New()
	cfg.Graphviz.NodeStyleRules = []config.GraphvizStyleRule{
		{Match: "alpha", Color: "red", FillColor: "lightyellow", Style: "filled"},
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

	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("graphviz: graph is nil")))
}

func TestSaveTo_WritesFileToDisk(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "graph.dot")
	cfg := config.New()

	err := SaveTo(path, buildSampleGraph(t), cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	content, readErr := os.ReadFile(path)
	g.Expect(readErr).NotTo(gomega.HaveOccurred())

	gg := goldie.New(t)
	g.Expect(gg.WithFixtureDir("testdata")).To(gomega.Succeed())

	gg.Assert(t, "sample_graph", content)
}

func TestWriteNodeDefinitionTo_WithFillColorNoStyle_AddsFilled(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	node := graph.NewNode("test-node")
	node.SetDescription("A test node")

	cfg := config.New()
	cfg.Graphviz.TaskNodes.FillColor = "lightblue"

	iw := indentwriter.New()
	root := iw.Add("digraph {")
	writeNodeDefinitionTo(root, node, cfg)
	root.Add("}")

	_, err := iw.WriteTo(&buf, "  ")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	output := buf.String()
	g.Expect(output).To(gomega.ContainSubstring("style=filled"))
	g.Expect(output).To(gomega.ContainSubstring("fillcolor=lightblue"))
}

func TestWriteNodeDefinitionTo_WithFillColorAndStyle_DoesNotOverrideStyle(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	node := graph.NewNode("test-node")
	node.SetDescription("A test node")

	cfg := config.New()
	cfg.Graphviz.TaskNodes.FillColor = "lightblue"
	cfg.Graphviz.TaskNodes.Style = "rounded,filled"

	iw := indentwriter.New()
	root := iw.Add("digraph {")
	writeNodeDefinitionTo(root, node, cfg)
	root.Add("}")

	_, err := iw.WriteTo(&buf, "  ")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	output := buf.String()
	g.Expect(output).To(gomega.ContainSubstring("style=\"rounded,filled\""))
	g.Expect(output).To(gomega.ContainSubstring("fillcolor=lightblue"))
}

func TestWriteNodeDefinitionTo_WithNodeLabel_UsesLabel(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	node := graph.NewNode("test-node")
	node.Label = "Custom Label"
	node.SetDescription("A test node")

	cfg := config.New()

	iw := indentwriter.New()
	root := iw.Add("digraph {")
	writeNodeDefinitionTo(root, node, cfg)
	root.Add("}")

	_, err := iw.WriteTo(&buf, "  ")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	output := buf.String()
	g.Expect(output).To(gomega.ContainSubstring("Custom Label"))
}

func TestWriteNodeDefinitionTo_WithLongDescription_WrapsText(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	node := graph.NewNode("test-node")
	longDesc := strings.Repeat("x", 100)
	node.SetDescription(longDesc)

	cfg := config.New()

	iw := indentwriter.New()
	root := iw.Add("digraph {")
	writeNodeDefinitionTo(root, node, cfg)
	root.Add("}")

	_, err := iw.WriteTo(&buf, "  ")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	output := buf.String()
	g.Expect(output).To(gomega.ContainSubstring("Mrecord"))
	g.Expect(output).To(gomega.ContainSubstring(longDesc))
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
