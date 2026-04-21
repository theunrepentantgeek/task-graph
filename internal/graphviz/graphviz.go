package graphviz

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/indentwriter"
	"github.com/theunrepentantgeek/task-graph/internal/namespace"
	"github.com/theunrepentantgeek/task-graph/internal/safe"
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

	nodes := slices.Collect(g.Nodes())
	slices.SortFunc(
		nodes,
		func(left *graph.Node, right *graph.Node) int {
			return cmp.Compare(left.ID(), right.ID())
		})

	nodeIDs := make([]string, len(nodes))
	for i, n := range nodes {
		nodeIDs[i] = n.ID()
	}

	reg := safe.NewRegistry()
	reg.Prepare(nodeIDs)

	iw := indentwriter.New()
	root := iw.Add("digraph {")

	err := writeAllNodesTo(root, nodes, cfg, reg)
	if err != nil {
		return err
	}

	iw.Add("}")

	_, err = iw.WriteTo(w, indent)
	if err != nil {
		return eris.Wrap(err, "failed to write graphviz output")
	}

	return nil
}

func writeAllNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	if cfg != nil && cfg.GroupByNamespace {
		return writeGroupedNodesTo(root, nodes, cfg, reg)
	}

	return writeNodesTo(root, nodes, cfg, reg)
}

// writeGroupedNodesTo writes nodes organized into namespace subgraph clusters.
func writeGroupedNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	// Build map: namespace -> nodes directly in that namespace
	nsToNodes := indexNodesByNamespace(nodes)

	// Collect all unique namespaces, including intermediate ones
	allNS := findAllNamespaces(nsToNodes)

	// Find and sort top-level namespaces (those with no parent)
	topLevel := make([]string, 0, len(allNS))
	for ns := range allNS {
		if namespace.Parent(ns) == "" {
			topLevel = append(topLevel, ns)
		}
	}

	slices.Sort(topLevel)

	err := writeNodesTo(root, nsToNodes[""], cfg, reg)
	if err != nil {
		return err
	}

	for _, ns := range topLevel {
		err := writeNamespaceSubgraphTo(root, ns, nsToNodes, allNS, cfg, reg)
		if err != nil {
			return err
		}
	}

	return nil
}

// findAllNamespaces takes a map of namespaces to their directly contained nodes
// and returns a set of all namespaces.
func findAllNamespaces(nsToNodes map[string][]*graph.Node) map[string]bool {
	allNS := make(map[string]bool)

	for ns := range nsToNodes {
		for current := ns; current != ""; current = namespace.Parent(current) {
			allNS[current] = true
		}
	}

	return allNS
}

func indexNodesByNamespace(nodes []*graph.Node) map[string][]*graph.Node {
	nsToNodes := make(map[string][]*graph.Node)

	for _, node := range nodes {
		ns := namespace.Namespace(node.ID())
		nsToNodes[ns] = append(nsToNodes[ns], node)
	}

	return nsToNodes
}

// writeNamespaceSubgraphTo writes a subgraph cluster for the given namespace, recursively
// handling child namespaces.
func writeNamespaceSubgraphTo(
	parent *indentwriter.Line,
	ns string,
	nsToNodes map[string][]*graph.Node,
	allNS map[string]bool,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	subgraph := parent.Addf("subgraph %s {", reg.IDWithPrefix("cluster_", ns))
	subgraph.Addf("label=%q", ns)

	children := childNamespaces(allNS, ns)

	// Write nodes directly in this namespace, then child subgraphs, with blank lines between items
	err := writeNodesTo(subgraph, nsToNodes[ns], cfg, reg)
	if err != nil {
		return err
	}

	for _, child := range children {
		err := writeNamespaceSubgraphTo(subgraph, child, nsToNodes, allNS, cfg, reg)
		if err != nil {
			return err
		}
	}

	parent.Add("}")

	return nil
}

// childNamespaces returns sorted child namespaces of the given namespace.
func childNamespaces(allNS map[string]bool, ns string) []string {
	children := make([]string, 0, len(allNS))

	for candidate := range allNS {
		if namespace.Parent(candidate) == ns {
			children = append(children, candidate)
		}
	}

	slices.Sort(children)

	return children
}

// writeNodesTo writes all nodes and their edges to the graphviz output.
func writeNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	for _, node := range nodes {
		err := writeNodeTo(root, node, cfg, reg)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeNodeTo writes a single node and its edges to the graphviz output.
func writeNodeTo(
	root *indentwriter.Line,
	node *graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	err := writeNodeDefinitionTo(root, node, cfg, reg)
	if err != nil {
		return err
	}

	for _, edge := range node.Edges() {
		writeEdgeTo(root, edge, cfg, reg)
	}

	// Finish with a blank line for readability
	root.Add("")

	return nil
}

func writeNodeDefinitionTo(
	root *indentwriter.Line,
	node *graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	margin := min((len(node.Description)+20)/2, 40)

	rec := newRecord()
	rec.add(nodeLabel(node))
	rec.addWrapped(margin, node.Description)

	props := newNodeProperties()
	props.Addf("shape", "Mrecord")
	props.Add("label", rec.String())

	err := applyNodeConfig(&props, node, cfg)
	if err != nil {
		return err
	}

	if props.ContainsKey("fillcolor") && !props.ContainsKey("style") {
		props.Add("style", "filled")
	}

	id := fmt.Sprintf("\"%s\"", reg.ID(node.ID()))
	props.WriteTo(id, root)

	return nil
}

func applyNodeConfig(props *nodeProperties, node *graph.Node, cfg *config.Config) error {
	if cfg == nil || cfg.Graphviz == nil {
		return nil
	}

	props.AddAttributes(cfg.Graphviz.TaskNodes)

	for _, rule := range cfg.NodeStyleRules {
		err := props.AddStyleRuleAttributes(node.ID(), rule)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeEdgeTo(
	root *indentwriter.Line,
	edge *graph.Edge,
	cfg *config.Config,
	reg *safe.Registry,
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
			reg.ID(edge.From().ID()),
			reg.ID(edge.To().ID())),
		root)
}

func nodeLabel(node *graph.Node) string {
	if node.Label != "" {
		return node.Label
	}

	return node.ID()
}
