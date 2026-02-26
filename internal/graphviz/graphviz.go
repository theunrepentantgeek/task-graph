package graphviz

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

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

	if cfg != nil && cfg.GroupByNamespace {
		writeGroupedNodesTo(root, nodes, cfg)
	} else {
		writeNodesTo(root, nodes, cfg)
	}

	iw.Add("}")

	_, err := iw.WriteTo(w, indent)
	if err != nil {
		return eris.Wrap(err, "failed to write graphviz output")
	}

	return nil
}

// writeGroupedNodesTo writes nodes organized into namespace subgraph clusters.
func writeGroupedNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	cfg *config.Config,
) {
	// Build map: namespace -> nodes directly in that namespace
	nsToNodes := make(map[string][]*graph.Node)
	for _, node := range nodes {
		ns := nodeNamespace(node.ID())
		nsToNodes[ns] = append(nsToNodes[ns], node)
	}

	// Collect all unique namespaces, including intermediate ones
	allNS := make(map[string]bool)
	for ns := range nsToNodes {
		for current := ns; current != ""; current = parentNamespace(current) {
			allNS[current] = true
		}
	}

	// Find and sort top-level namespaces (those with no parent)
	topLevel := make([]string, 0, len(allNS))
	for ns := range allNS {
		if parentNamespace(ns) == "" {
			topLevel = append(topLevel, ns)
		}
	}
	sort.Strings(topLevel)

	writeNodesTo(root, nsToNodes[""], cfg)

	}

	for _, ns := range topLevel {
		if needsBlankLine {
			root.Add("")
		}
		writeNamespaceSubgraphTo(root, ns, nsToNodes, allNS, cfg)
		needsBlankLine = true
	}
}

// writeNamespaceSubgraphTo writes a subgraph cluster for the given namespace, recursively
// handling child namespaces.
func writeNamespaceSubgraphTo(
	parent *indentwriter.Line,
	ns string,
	nsToNodes map[string][]*graph.Node,
	allNS map[string]bool,
	cfg *config.Config,
) {
	subgraph := parent.Addf("subgraph %s {", clusterID(ns))
	subgraph.Addf("label=%q", ns)

	// Find and sort child namespaces
	children := make([]string, 0, len(allNS))
	for candidate := range allNS {
		if parentNamespace(candidate) == ns {
			children = append(children, candidate)
		}
	}
	sort.Strings(children)

	// Write nodes directly in this namespace, then child subgraphs, with blank lines between items
	writeNodesTo(subgraph, nsToNodes[ns], cfg)

	for _, child := range children {
		if needsBlankLine {
			subgraph.Add("")
		}
		writeNamespaceSubgraphTo(subgraph, child, nsToNodes, allNS, cfg)
		needsBlankLine = true
	}

	parent.Add("}")
}

// nodeNamespace returns the namespace portion of a node ID (everything before the last colon).
// Returns an empty string if the ID has no colon.
func nodeNamespace(id string) string {
	idx := strings.LastIndex(id, ":")
	if idx < 0 {
		return ""
	}

	return id[:idx]
}

// parentNamespace returns the parent namespace of a namespace.
// Returns an empty string if the namespace has no parent (i.e., no colon).
func parentNamespace(ns string) string {
	idx := strings.LastIndex(ns, ":")
	if idx < 0 {
		return ""
	}

	return ns[:idx]
}

// clusterID converts a namespace string into a valid Graphviz cluster subgraph identifier.
// Colons are replaced with underscores to ensure the ID is valid.
func clusterID(ns string) string {
	return "cluster_" + strings.ReplaceAll(ns, ":", "_")
}

// writeNodesTo writes all nodes and their edges to the graphviz output.
func writeNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	cfg *config.Config,
) {
	for _, node := range nodes {
		writeNodeTo(root, node, cfg)
	}
}

// writeNodeTo writes a single node and its edges to the graphviz output.
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

		for _, rule := range cfg.Graphviz.StyleRules {
			props.AddStyleRuleAttributes(node.ID(), rule)
		}
	}

	id := fmt.Sprintf("\"%s\"", node.ID())
	props.WriteTo(id, root)
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
