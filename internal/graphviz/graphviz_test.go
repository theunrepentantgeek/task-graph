package graphviz

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

func TestWriteTo_WithStyleRules_AppliesMatchingStyles(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildSampleGraph(t)

	cfg := config.New()
	cfg.Graphviz.StyleRules = []config.GraphvizStyleRule{
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
