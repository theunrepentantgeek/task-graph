package graphviz

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"
	"github.com/sebdah/goldie/v2"

	"github.com/theunrepentantgeek/task-graph/internal/graph"
)

func TestWriteTo_WithNodesAndEdges_WritesSortedGraphviz(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	graph := buildSampleGraph(t)

	err := WriteTo(&buf, graph)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	gg := goldie.New(t)
	gg.WithFixtureDir("testdata")
	gg.Assert(t, "sample_graph", buf.Bytes())
}

func TestWriteTo_NilGraph_ReturnsError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}

	err := WriteTo(&buf, nil)

	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("graphviz: graph is nil")))
}

func TestSaveTo_WritesFileToDisk(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "graph.dot")

	err := SaveTo(path, buildSampleGraph(t))

	g.Expect(err).NotTo(gomega.HaveOccurred())

	content, readErr := os.ReadFile(path)
	g.Expect(readErr).NotTo(gomega.HaveOccurred())

	gg := goldie.New(t)
	gg.WithFixtureDir("testdata")
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
