package graphviz

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/graphns"
	"github.com/theunrepentantgeek/task-graph/internal/indentwriter"
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

	bw := bufio.NewWriter(f)

	err = WriteTo(bw, gr, cfg)
	if err != nil {
		return err
	}

	return eris.Wrap(bw.Flush(), "failed to flush graphviz output")
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

	taskNodes, varNodes := graphns.SplitByKind(nodes)

	iw := indentwriter.New()
	root := iw.Add("digraph {")

	err := writeAllNodesTo(root, taskNodes, cfg, reg)
	if err != nil {
		return err
	}

	if len(varNodes) > 0 {
		err = writeVariableNodesTo(root, varNodes, cfg, reg)
		if err != nil {
			return err
		}
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
	nsToNodes := graphns.IndexByNamespace(nodes)

	// Collect all unique namespaces, including intermediate ones
	allNS := graphns.FindAllNamespaces(nsToNodes)

	// Pre-build parent→children map so each lookup is O(1) rather than O(N).
	childrenOf := graphns.BuildChildrenMap(allNS)

	// Find and sort top-level namespaces (those with no parent)
	topLevel := childrenOf[""]

	err := writeNodesTo(root, nsToNodes[""], cfg, reg)
	if err != nil {
		return err
	}

	for _, ns := range topLevel {
		err := writeNamespaceSubgraphTo(root, ns, nsToNodes, childrenOf, cfg, reg)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeNamespaceSubgraphTo writes a subgraph cluster for the given namespace, recursively
// handling child namespaces.
func writeNamespaceSubgraphTo(
	parent *indentwriter.Line,
	ns string,
	nsToNodes map[string][]*graph.Node,
	childrenOf map[string][]string,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	subgraph := parent.Addf("subgraph %s {", reg.IDWithPrefix("cluster_", ns))
	subgraph.Addf("label=%q", ns)

	// Write nodes directly in this namespace, then child subgraphs, with blank lines between items
	err := writeNodesTo(subgraph, nsToNodes[ns], cfg, reg)
	if err != nil {
		return err
	}

	for _, child := range childrenOf[ns] {
		err := writeNamespaceSubgraphTo(subgraph, child, nsToNodes, childrenOf, cfg, reg)
		if err != nil {
			return err
		}
	}

	parent.Add("}")

	return nil
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
	return writeNodeDefinitionWithShapeTo(root, node, cfg, reg, "Mrecord", applyNodeConfig)
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
		case "var":
			props.AddAttributes(cfg.Graphviz.VariableEdges)
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

func writeVariableNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	for _, node := range nodes {
		err := writeVariableNodeDefinitionTo(root, node, cfg, reg)
		if err != nil {
			return err
		}

		for _, edge := range node.Edges() {
			writeEdgeTo(root, edge, cfg, reg)
		}

		root.Add("")
	}

	// Add rank=sink to force variables to bottom
	sink := root.Add("{ rank=sink")

	for _, node := range nodes {
		sink.Addf("\"%s\"", reg.ID(node.ID()))
	}

	root.Add("}")

	return nil
}

func writeVariableNodeDefinitionTo(
	root *indentwriter.Line,
	node *graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	return writeNodeDefinitionWithShapeTo(root, node, cfg, reg, "record", applyVariableNodeConfig)
}

func writeNodeDefinitionWithShapeTo(
	root *indentwriter.Line,
	node *graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
	shape string,
	configFunc func(*nodeProperties, *graph.Node, *config.Config) error,
) error {
	margin := min((len(node.Description)+20)/2, 40)

	rec := newRecord()
	rec.add(node.DisplayLabel())
	rec.addWrapped(margin, node.Description)

	props := newNodeProperties()
	props.Add("shape", shape)
	props.Add("label", rec.String())

	err := configFunc(&props, node, cfg)
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

func applyVariableNodeConfig(props *nodeProperties, node *graph.Node, cfg *config.Config) error {
	if cfg == nil {
		return nil
	}

	if cfg.Graphviz != nil {
		props.AddAttributes(cfg.Graphviz.VariableNodes)
	}

	for _, rule := range cfg.NodeStyleRules {
		err := props.AddStyleRuleAttributes(node.ID(), rule)
		if err != nil {
			return err
		}
	}

	return nil
}
