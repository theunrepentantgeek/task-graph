package graphviz

import (
	"errors"
	"io"
	"os"
	"sort"

	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/indentwriter"
)

func SaveTo(
	path string,
	gr *graph.Graph,
) error {
	f, err := os.Create(path)
	if err != nil {
		return eris.Wrapf(err, "failed to create file: %s", path)
	}

	defer f.Close()

	return WriteTo(f, gr)
}

func WriteTo(
	w io.Writer,
	g *graph.Graph,
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
		writeNodeTo(root, node)

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
) {
	root.Addf("\"%s\"", node.ID())

	for _, edge := range node.Edges() {
		if edge.Label() == "" {
			root.Addf(
				"\"%s\" -> \"%s\"",
				edge.From().ID(),
				edge.To().ID())

			continue
		}

		root.Addf(
			"\"%s\" -> \"%s\" [label=\"%s\"]",
			edge.From().ID(),
			edge.To().ID(),
			edge.Label())
	}
}
