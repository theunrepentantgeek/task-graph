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
	writeNodeDefinitionTo(root, node)

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

func writeNodeDefinitionTo(
	root *indentwriter.Line,
	node *graph.Node,
) {
	margin := min((len(node.Description)+20)/2, 40)

	rec := newRecord()
	rec.add(nodeLabel(node))
	rec.addWrapped(margin, node.Description)

	props := newProperties()
	props.Addf("shape", "Mrecord")
	props.Add("label", rec.String())

	props.WriteTo(node.ID(), root)
}

func nodeLabel(node *graph.Node) string {
	if node.Label != "" {
		return node.Label
	}

	return node.ID()
}
