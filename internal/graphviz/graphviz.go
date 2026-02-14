package graphviz

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/indentwriter"
)

func SaveTo(
	path string,
	gr *graph.Graph,
) error {
	f, err := os.Create(path)
	if err != nil {
		return err
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
		return fmt.Errorf("graphviz: graph is nil")
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
		nodeLine := root.Addf("\"%s\"", node.ID())

		for _, edge := range node.Edges() {
			if edge.Label() == "" {
				nodeLine.Addf("\"%s\" -> \"%s\"", edge.From().ID(), edge.To().ID())
				continue
			}

			nodeLine.Addf("\"%s\" -> \"%s\" [label=\"%s\"]", edge.From().ID(), edge.To().ID(), edge.Label())
		}

		if i < len(nodes)-1 {
			root.Add("") // blank line between nodes
		}
	}

	iw.Add("}")

	_, err := iw.WriteTo(w, indent)
	return err

}
