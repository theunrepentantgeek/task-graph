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
	"github.com/theunrepentantgeek/task-graph/internal/safe"
)

const testDescription = "A test node"

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
	cfg.NodeStyleRules = []config.NodeStyleRule{
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
	node.Description = testDescription

	cfg := config.New()
	cfg.Graphviz.TaskNodes.FillColor = "lightblue"

	iw := indentwriter.New()
	root := iw.Add("digraph {")

	err := writeNodeDefinitionTo(root, node, cfg, safe.NewRegistry())
	g.Expect(err).NotTo(gomega.HaveOccurred())

	root.Add("}")

	_, err = iw.WriteTo(&buf, "  ")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	output := buf.String()
	g.Expect(output).To(gomega.ContainSubstring(`style="filled"`))
	g.Expect(output).To(gomega.ContainSubstring(`fillcolor="lightblue"`))
}

func TestWriteNodeDefinitionTo_WithFillColorAndStyle_DoesNotOverrideStyle(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	node := graph.NewNode("test-node")
	node.Description = testDescription

	cfg := config.New()
	cfg.Graphviz.TaskNodes.FillColor = "lightblue"
	cfg.Graphviz.TaskNodes.Style = "rounded,filled"

	iw := indentwriter.New()
	root := iw.Add("digraph {")

	err := writeNodeDefinitionTo(root, node, cfg, safe.NewRegistry())
	g.Expect(err).NotTo(gomega.HaveOccurred())

	root.Add("}")

	_, err = iw.WriteTo(&buf, "  ")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	output := buf.String()
	g.Expect(output).To(gomega.ContainSubstring(`style="rounded,filled"`))
	g.Expect(output).To(gomega.ContainSubstring(`fillcolor="lightblue"`))
}

func TestWriteNodeDefinitionTo_WithNodeLabel_UsesLabel(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	node := graph.NewNode("test-node")
	node.Label = "Custom Label"
	node.Description = testDescription

	cfg := config.New()

	iw := indentwriter.New()
	root := iw.Add("digraph {")

	err := writeNodeDefinitionTo(root, node, cfg, safe.NewRegistry())
	g.Expect(err).NotTo(gomega.HaveOccurred())

	root.Add("}")

	_, err = iw.WriteTo(&buf, "  ")
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
	node.Description = longDesc

	cfg := config.New()

	iw := indentwriter.New()
	root := iw.Add("digraph {")

	err := writeNodeDefinitionTo(root, node, cfg, safe.NewRegistry())
	g.Expect(err).NotTo(gomega.HaveOccurred())

	root.Add("}")

	_, err = iw.WriteTo(&buf, "  ")
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

func TestWriteTo_WithVariableNodes_WritesVariableNodesWithRankSink(t *testing.T) {
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

func TestWriteTo_WithVariableNodesAndMatchingStyleRule_AppliesStyleToVariableNode(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	buf := bytes.Buffer{}
	gr := buildGraphWithVariables(t)

	cfg := config.New()
	cfg.NodeStyleRules = []config.NodeStyleRule{
		{Match: "var:*", Color: "purple", FillColor: "lavender"},
	}

	// Act
	err := WriteTo(&buf, gr, cfg)

	// Assert
	g.Expect(err).NotTo(gomega.HaveOccurred())

	gg := goldie.New(t)
	g.Expect(gg.WithFixtureDir("testdata")).To(gomega.Succeed())

	gg.Assert(t, "sample_graph_variables_with_style_rules", buf.Bytes())
}

func TestWriteTo_WithNilGraphvizConfig_WritesNodesWithoutGraphvizStyling(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := &config.Config{} // Graphviz is nil

	// Act
	err := WriteTo(&buf, gr, cfg)

	// Assert
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(buf.String()).To(gomega.ContainSubstring("digraph {"))
	g.Expect(buf.String()).NotTo(gomega.ContainSubstring(`color="black"`))
}

func TestWriteTo_WithTaskNodesAndInvalidStyleRulePattern_ReturnsError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := config.New()
	cfg.NodeStyleRules = []config.NodeStyleRule{
		{Match: "[invalid", Color: "red"},
	}

	// Act
	err := WriteTo(&buf, gr, cfg)

	// Assert
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestWriteTo_WithVariableNodesAndInvalidStyleRulePattern_ReturnsError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	buf := bytes.Buffer{}
	gr := buildGraphWithVariables(t)

	cfg := config.New()
	cfg.NodeStyleRules = []config.NodeStyleRule{
		{Match: "[invalid", Color: "red"},
	}

	// Act
	err := WriteTo(&buf, gr, cfg)

	// Assert
	g.Expect(err).To(gomega.HaveOccurred())
}

// TestWriteTo_WithVariableNodes_NilConfig_WritesVariableNodes verifies that
// WriteTo handles nil config when variable nodes are present (exercises the
// cfg == nil early-return path in applyVariableNodeConfig).
func TestWriteTo_WithVariableNodes_NilConfig_WritesVariableNodes(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	buf := bytes.Buffer{}
	gr := buildGraphWithVariables(t)

	// Act – nil config must not panic and should produce valid output
	err := WriteTo(&buf, gr, nil)

	// Assert
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(buf.String()).To(gomega.ContainSubstring("digraph {"))
	g.Expect(buf.String()).To(gomega.ContainSubstring("rank=sink"))
}

// TestWriteTo_WithCallEdge_AppliesCallEdgeStyle verifies that edges with
// class "call" receive the call-edge styling from the graphviz config.
func TestWriteTo_WithCallEdge_AppliesCallEdgeStyle(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	buf := bytes.Buffer{}
	gr := graph.New()

	build := gr.AddNode("build")
	deploy := gr.AddNode("deploy")
	edge := build.AddEdge(deploy)
	edge.SetClass("call")

	cfg := config.New()
	cfg.Graphviz.CallEdges = &config.GraphvizEdge{
		Color: "green",
		Style: "dotted",
	}

	// Act
	err := WriteTo(&buf, gr, cfg)

	// Assert
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(buf.String()).To(gomega.ContainSubstring(`color="green"`))
	g.Expect(buf.String()).To(gomega.ContainSubstring(`style="dotted"`))
}

// TestSaveTo_InvalidPath_ReturnsError verifies that SaveTo returns a wrapped
// error when the output file cannot be created (e.g. non-existent directory).
func TestSaveTo_InvalidPath_ReturnsError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange – path into a directory that does not exist
	gr := buildSampleGraph(t)
	path := filepath.Join(t.TempDir(), "nonexistent", "output.dot")

	// Act
	err := SaveTo(path, gr, config.New())

	// Assert
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("failed to create file"))
}

// TestWriteTo_WithGroupByNamespace_RootNodeInvalidStyleRule_ReturnsError verifies
// that writeGroupedNodesTo propagates an error from writeNodesTo when processing
// root-level (un-namespaced) nodes with an invalid style rule pattern.
func TestWriteTo_WithGroupByNamespace_RootNodeInvalidStyleRule_ReturnsError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange – graph has a root node ("build") and namespaced nodes
	buf := bytes.Buffer{}
	gr := buildNamespacedGraph(t)

	cfg := config.New()
	cfg.GroupByNamespace = true
	cfg.NodeStyleRules = []config.NodeStyleRule{
		{Match: "[invalid", Color: "red"},
	}

	// Act
	err := WriteTo(&buf, gr, cfg)

	// Assert
	g.Expect(err).To(gomega.HaveOccurred())
}

// TestWriteTo_WithGroupByNamespace_DeepNamespaceInvalidStyleRule_ReturnsError
// verifies that errors propagate correctly through the recursive
// writeNamespaceSubgraphTo chain. With only a deeply-nested node ("cmd:test:unit"),
// the "cmd" subgraph's writeNodesTo is a no-op but the recursive child call for
// "cmd:test" fails, exercising lines 128–130 and 151–153 and 156–159.
func TestWriteTo_WithGroupByNamespace_DeepNamespaceInvalidStyleRule_ReturnsError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange – only a deeply-nested node; no root or direct-cmd nodes
	gr := graph.New()
	gr.AddNode("cmd:test:unit")

	buf := bytes.Buffer{}
	cfg := config.New()
	cfg.GroupByNamespace = true
	cfg.NodeStyleRules = []config.NodeStyleRule{
		{Match: "[invalid", Color: "red"},
	}

	// Act
	err := WriteTo(&buf, gr, cfg)

	// Assert
	g.Expect(err).To(gomega.HaveOccurred())
}

// TestWriteTo_WithVariableNodesOnly_InvalidStyleRule_ReturnsError verifies that
// writeVariableNodesTo propagates an error when a variable node's style rule
// pattern is invalid. With no task nodes, writeAllNodesTo is a no-op and the
// error originates inside writeVariableNodesTo (line 273).
func TestWriteTo_WithVariableNodesOnly_InvalidStyleRule_ReturnsError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange – a graph with only a variable node and an invalid style rule
	gr := graph.New()
	v := gr.AddNode("var:FOO")
	v.Kind = graph.NodeKindVariable
	v.Label = "FOO"
	v.Description = "some value"

	buf := bytes.Buffer{}
	cfg := config.New()
	cfg.NodeStyleRules = []config.NodeStyleRule{
		{Match: "[invalid", Color: "red"},
	}

	// Act
	err := WriteTo(&buf, gr, cfg)

	// Assert
	g.Expect(err).To(gomega.HaveOccurred())
}

// TestWriteTo_WithVariableNodesAndNilGraphvizBlock_WritesOutput verifies that
// WriteTo succeeds when cfg is non-nil but cfg.Graphviz is nil and the graph
// contains variable nodes. This exercises the `cfg.Graphviz != nil` false-branch
// inside applyVariableNodeConfig.
func TestWriteTo_WithVariableNodesAndNilGraphvizBlock_WritesOutput(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange – cfg present but Graphviz sub-config absent
	buf := bytes.Buffer{}
	gr := buildGraphWithVariables(t)
	cfg := &config.Config{} // Graphviz field is nil

	// Act
	err := WriteTo(&buf, gr, cfg)

	// Assert
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(buf.String()).To(gomega.ContainSubstring("rank=sink"))
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

	// Variable edges
	pkgEdge := pkg.AddEdge(build)
	pkgEdge.SetClass("var")

	verEdge := ver.AddEdge(build)
	verEdge.SetClass("var")

	return gr
}

func TestWriteTo_WithFooter_WritesFooterAttributes(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := config.New()
	cfg.Footer = true
	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	output := buf.String()
	g.Expect(output).To(gomega.ContainSubstring("label="))
	g.Expect(output).To(gomega.ContainSubstring("task-graph"))
	g.Expect(output).To(gomega.ContainSubstring("labelloc=b"))
	g.Expect(output).To(gomega.ContainSubstring("labeljust=r"))
}

func TestWriteTo_WithoutFooter_DoesNotWriteFooterAttributes(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := config.New()
	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	output := buf.String()
	g.Expect(output).NotTo(gomega.ContainSubstring("labelloc=b"))
	g.Expect(output).NotTo(gomega.ContainSubstring("labeljust=r"))
}
