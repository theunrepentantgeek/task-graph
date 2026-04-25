package mermaid

import (
	"cmp"
	"errors"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/indentwriter"
	"github.com/theunrepentantgeek/task-graph/internal/namespace"
	"github.com/theunrepentantgeek/task-graph/internal/safe"
)

// SaveTo writes the Mermaid flowchart representation of the graph to the given file path.
func SaveTo(
	filePath string,
	gr *graph.Graph,
	cfg *config.Config,
) error {
	f, err := os.Create(filePath)
	if err != nil {
		return eris.Wrapf(err, "failed to create file: %s", filePath)
	}

	defer f.Close()

	return WriteTo(f, gr, cfg)
}

// WriteTo writes the Mermaid flowchart representation of the graph to the given writer.
func WriteTo(
	w io.Writer,
	g *graph.Graph,
	cfg *config.Config,
) error {
	const indent = "  "

	if g == nil {
		return errors.New("mermaid: graph is nil")
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

	taskNodes, varNodes := splitNodesByKind(nodes)

	iw := indentwriter.New()
	root := iw.Addf("flowchart %s", flowchartDirection(cfg))

	if cfg != nil && cfg.GroupByNamespace {
		writeGroupedNodesTo(root, taskNodes, reg)
	} else {
		writeNodesTo(root, taskNodes, reg)
	}

	if len(varNodes) > 0 {
		writeVariableNodesTo(root, varNodes, reg)
		writeVariableClassDef(root, varNodes, cfg, reg)
	}

	err := writeStyleRulesTo(root, nodes, cfg, reg)
	if err != nil {
		return err
	}

	_, err = iw.WriteTo(w, indent)
	if err != nil {
		return eris.Wrap(err, "failed to write mermaid output")
	}

	return nil
}

// flowchartDirection returns the mermaid flowchart direction from config,
// defaulting to "TD" (top-down).
func flowchartDirection(cfg *config.Config) string {
	if cfg != nil && cfg.Mermaid != nil && cfg.Mermaid.Direction != "" {
		return cfg.Mermaid.Direction
	}

	return "TD"
}

// writeGroupedNodesTo writes nodes organised into namespace subgraph clusters.
func writeGroupedNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	reg *safe.Registry,
) {
	nsToNodes := indexNodesByNamespace(nodes)
	allNS := findAllNamespaces(nsToNodes)

	// Pre-build parent→children map so each lookup is O(1) rather than O(N).
	childrenOf := buildChildrenMap(allNS)

	writeNodesTo(root, nsToNodes[""], reg)

	for _, ns := range childrenOf[""] {
		writeNamespaceSubgraphTo(root, ns, nsToNodes, childrenOf, reg)
	}
}

// findAllNamespaces returns a set of all namespaces found in the given node map.
func findAllNamespaces(nsToNodes map[string][]*graph.Node) map[string]bool {
	allNS := make(map[string]bool)

	for ns := range nsToNodes {
		for current := ns; current != ""; current = namespace.Parent(current) {
			allNS[current] = true
		}
	}

	return allNS
}

// buildChildrenMap builds a parent→sorted-children map from a set of all namespaces.
// The empty-string key ("") holds the sorted list of top-level namespaces.
// Building this map once avoids an O(N) scan of allNS for every namespace during
// the recursive subgraph walk.
func buildChildrenMap(allNS map[string]bool) map[string][]string {
	childrenOf := make(map[string][]string, len(allNS)+1)

	for ns := range allNS {
		parent := namespace.Parent(ns)
		childrenOf[parent] = append(childrenOf[parent], ns)
	}

	for key := range childrenOf {
		slices.Sort(childrenOf[key])
	}

	return childrenOf
}

func indexNodesByNamespace(nodes []*graph.Node) map[string][]*graph.Node {
	nsToNodes := make(map[string][]*graph.Node)

	for _, node := range nodes {
		ns := namespace.Namespace(node.ID())
		nsToNodes[ns] = append(nsToNodes[ns], node)
	}

	return nsToNodes
}

// writeNamespaceSubgraphTo writes a subgraph for the given namespace, recursively handling children.
func writeNamespaceSubgraphTo(
	parent *indentwriter.Line,
	ns string,
	nsToNodes map[string][]*graph.Node,
	childrenOf map[string][]string,
	reg *safe.Registry,
) {
	sg := parent.Addf("subgraph %s[\"%s\"]", reg.IDWithPrefix("sg_", ns), ns)

	writeNodesTo(sg, nsToNodes[ns], reg)

	for _, child := range childrenOf[ns] {
		writeNamespaceSubgraphTo(sg, child, nsToNodes, childrenOf, reg)
	}

	parent.Add("end")
}

// writeNodesTo writes all nodes and their edges.
func writeNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	reg *safe.Registry,
) {
	for _, node := range nodes {
		writeNodeTo(root, node, reg)
	}
}

// writeNodeTo writes a single node definition and its outgoing edges.
func writeNodeTo(
	root *indentwriter.Line,
	node *graph.Node,
	reg *safe.Registry,
) {
	writeNodeDefinitionTo(root, node, reg)

	for _, edge := range node.Edges() {
		writeEdgeTo(root, edge, reg)
	}

	root.Add("")
}

func writeNodeDefinitionTo(
	root *indentwriter.Line,
	node *graph.Node,
	reg *safe.Registry,
) {
	label := safe.Label(node.DisplayLabel())
	root.Addf("%s[\"%s\"]", reg.ID(node.ID()), label)
}

func writeEdgeTo(
	root *indentwriter.Line,
	edge *graph.Edge,
	reg *safe.Registry,
) {
	from := reg.ID(edge.From().ID())
	to := reg.ID(edge.To().ID())

	connector := "-->"
	if edge.Class() == "call" {
		connector = "-.->"
	}

	if edge.Label() != "" {
		label := safe.Label(edge.Label())
		root.Addf("%s %s|\"%s\"| %s", from, connector, label, to)
	} else {
		root.Addf("%s %s %s", from, connector, to)
	}
}

// writeStyleRulesTo writes Mermaid classDef and class directives for any matching NodeStyleRules.
func writeStyleRulesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	if cfg == nil || len(cfg.NodeStyleRules) == 0 {
		return nil
	}

	for i, rule := range cfg.NodeStyleRules {
		err := writeStyleRuleTo(root, nodes, i, rule, reg)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeStyleRuleTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	index int,
	rule config.NodeStyleRule,
	reg *safe.Registry,
) error {
	classDef := buildClassDef(rule)
	if classDef == "" {
		return nil
	}

	matchingIDs, err := findMatchingNodeIDs(nodes, rule.Match, reg)
	if err != nil {
		return err
	}

	if len(matchingIDs) > 0 {
		slices.Sort(matchingIDs)
		root.Addf("classDef rule%d %s", index, classDef)
		root.Addf("class %s rule%d", strings.Join(matchingIDs, ","), index)
	}

	return nil
}

func findMatchingNodeIDs(nodes []*graph.Node, pattern string, reg *safe.Registry) ([]string, error) {
	re, err := namespace.CompileMatchPattern(pattern)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to compile match pattern %q", pattern)
	}

	var matchingIDs []string

	for _, node := range nodes {
		if re.MatchString(node.ID()) {
			matchingIDs = append(matchingIDs, reg.ID(node.ID()))
		}
	}

	return matchingIDs, nil
}

// buildClassDef constructs a Mermaid classDef value string from a NodeStyleRule.
// Returns empty string if no visual properties are set.
func buildClassDef(rule config.NodeStyleRule) string {
	var parts []string

	if rule.FillColor != "" {
		parts = append(parts, "fill:"+rule.FillColor)
	}

	if rule.Color != "" {
		parts = append(parts, "stroke:"+rule.Color)
	}

	if rule.FontColor != "" {
		parts = append(parts, "color:"+rule.FontColor)
	}

	return strings.Join(parts, ",")
}
func writeVariableNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	reg *safe.Registry,
) {
	for _, node := range nodes {
		writeVariableNodeDefinitionTo(root, node, reg)
		writeVariableEdgesTo(root, node, reg)
		root.Add("")
	}
}

func writeVariableNodeDefinitionTo(
	root *indentwriter.Line,
	node *graph.Node,
	reg *safe.Registry,
) {
	label := safe.Label(variableDisplayLabel(node))
	root.Addf("%s(\"%s\")", reg.ID(node.ID()), label)
}

func writeVariableEdgesTo(
	root *indentwriter.Line,
	node *graph.Node,
	reg *safe.Registry,
) {
	for _, edge := range node.Edges() {
		// Reverse edge direction visually: write as task ==> variable
		// so Mermaid's layout pushes variables below tasks
		from := reg.ID(edge.To().ID())
		to := reg.ID(edge.From().ID())
		root.Addf("%s ==> %s", from, to)
	}
}

func variableDisplayLabel(node *graph.Node) string {
	label := node.DisplayLabel()
	if node.Description != "" {
		return label + ": " + node.Description
	}

	return label
}

func writeVariableClassDef(
	root *indentwriter.Line,
	nodes []*graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) {
	parts := variableClassDefParts(cfg)
	classDef := strings.Join(parts, ",")

	ids := make([]string, 0, len(nodes))
	for _, n := range nodes {
		ids = append(ids, reg.ID(n.ID()))
	}

	slices.Sort(ids)
	root.Addf("classDef varStyle %s", classDef)
	root.Addf("class %s varStyle", strings.Join(ids, ","))
}

func variableClassDefParts(cfg *config.Config) []string {
	if cfg == nil || cfg.Mermaid == nil || cfg.Mermaid.VariableNodes == nil {
		return []string{"fill:#e8e8e8", "stroke:#666"}
	}

	vs := cfg.Mermaid.VariableNodes

	var parts []string

	if vs.Fill != "" {
		parts = append(parts, "fill:"+vs.Fill)
	}

	if vs.Stroke != "" {
		parts = append(parts, "stroke:"+vs.Stroke)
	}

	if vs.Color != "" {
		parts = append(parts, "color:"+vs.Color)
	}

	if len(parts) > 0 {
		return parts
	}

	return []string{"fill:#e8e8e8", "stroke:#666"}
}

//nolint:revive // Choosing to return two unnamed slices
func splitNodesByKind(nodes []*graph.Node) ([]*graph.Node, []*graph.Node) {
	var (
		taskNodes []*graph.Node
		varNodes  []*graph.Node
	)

	for _, n := range nodes {
		if n.Kind == graph.NodeKindVariable {
			varNodes = append(varNodes, n)
		} else {
			taskNodes = append(taskNodes, n)
		}
	}

	return taskNodes, varNodes
}
