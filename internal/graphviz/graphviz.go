package graphviz

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/indentwriter"
)

func SaveTo(
	path string,
	gr *graph.Graph,
	cfg *config.Config,
) error {
	f, err := os.Create(path)
	if err != nil {
		return eris.Wrapf(err, "failed to create file: %s", path)
	}

	defer f.Close()

	return WriteTo(f, gr, cfg)
}

func WriteTo(
	w io.Writer,
	g *graph.Graph,
	cfg *config.Config,
) error {
	const indent = "  "

	if g == nil {
		return errors.New("graphviz: graph is nil")
	}

	nodes := make([]*graph.Node, 0)
	for node := range g.Nodes() {
		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID() < nodes[j].ID()
	})

	iw := indentwriter.New()
	root := iw.Add("digraph {")

	for i, node := range nodes {
		writeNodeTo(root, node, cfg)

		if i < len(nodes)-1 {
			root.Add("") // blank line between nodes
		}
	}

	iw.Add("}")

	_, err := iw.WriteTo(w, indent)
	if err != nil {
		return eris.Wrap(err, "failed to write graphviz output")
	}

	return nil
}

func writeNodeTo(
	root *indentwriter.Line,
	node *graph.Node,
	cfg *config.Config,
) {
	writeNodeDefinitionTo(root, node, cfg)

	for _, edge := range node.Edges() {
		writeEdgeTo(root, edge, cfg)
	}
}

func writeNodeDefinitionTo(
	root *indentwriter.Line,
	node *graph.Node,
	cfg *config.Config,
) {
	margin := min((len(node.Description)+20)/2, 40)

	rec := newRecord()
	rec.add(nodeLabel(node))
	rec.addWrapped(margin, node.Description)

	props := newNodeProperties()
	props.Addf("shape", "Mrecord")
	props.Add("label", rec.String())

	if cfg != nil && cfg.Graphviz != nil {
		props.AddAttributes(cfg.Graphviz.TaskNodes)
	}

	props.WriteTo(node.ID(), root)
}

func writeEdgeTo(
	root *indentwriter.Line,
	edge *graph.Edge,
	cfg *config.Config,
) {
	props := newEdgeProperties()

	if edge.Label() != "" {
		props.Add("label", edge.Label())
	}

	if cfg != nil && cfg.Graphviz != nil {
		switch edge.Class() {
		case "dep":
			props.AddAttributes(cfg.Graphviz.DependencyEdges)
		case "call":
			props.AddAttributes(cfg.Graphviz.CallEdges)
		default:
			// Nothing
		}
	}

	props.WriteTo(
		fmt.Sprintf(
			"\"%s\" -> \"%s\"",
			edge.From().ID(),
			edge.To().ID()),
		root)
}

func nodeLabel(node *graph.Node) string {
	if node.Label != "" {
		return node.Label
	}

	return node.ID()
}
