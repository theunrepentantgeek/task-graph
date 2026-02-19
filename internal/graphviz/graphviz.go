package graphviz

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

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
	label := nodeLabel(node)
	description := nodeDescription(node)

	var parts []string
	if label != "" {
		parts = append(parts, label)
	}

	if description != "" {
		parts = append(parts, description)
	}

	if len(parts) == 0 {
		// No explicit label, so we just use the node ID as the label
		parts = append(parts, node.ID())
	}

	content := fmt.Sprintf("{%s}", strings.Join(parts, " | "))

	nested := root.Addf("%q [", node.ID())
	nested.Addf("label=%q", content)
	nested.Addf("shape=%q", "Mrecord")
	root.Add("]")
}

func nodeLabel(node *graph.Node) string {
	if node.Label != "" {
		return node.Label
	}

	return node.ID()
}

func nodeDescription(node *graph.Node) string {
	if node.Description == "" {
		return ""
	}

	margin := min((len(node.Description)+20)/2, 40)

	desc := strings.ReplaceAll(node.Description, "{", "&lbrace;")
	desc = strings.ReplaceAll(desc, "}", "&rbrace;")

	lines := indentwriter.WordWrap(desc, margin)

	return strings.Join(lines, "\n")
}
