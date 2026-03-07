package mermaid

import (
	"cmp"
	"errors"
	"io"
	"os"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/indentwriter"
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

	direction := "TD"
	if cfg != nil && cfg.Mermaid != nil && cfg.Mermaid.Direction != "" {
		direction = cfg.Mermaid.Direction
	}

	iw := indentwriter.New()
	root := iw.Addf("flowchart %s", direction)

	if cfg != nil && cfg.GroupByNamespace {
		writeGroupedNodesTo(root, nodes, reg)
	} else {
		writeNodesTo(root, nodes, reg)
	}

	writeStyleRulesTo(root, nodes, cfg, reg)

	_, err := iw.WriteTo(w, indent)
	if err != nil {
		return eris.Wrap(err, "failed to write mermaid output")
	}

	return nil
}

// writeGroupedNodesTo writes nodes organised into namespace subgraph clusters.
func writeGroupedNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	reg *safe.Registry,
) {
	nsToNodes := indexNodesByNamespace(nodes)
	allNS := findAllNamespaces(nsToNodes)

	topLevel := make([]string, 0, len(allNS))
	for ns := range allNS {
		if parentNamespace(ns) == "" {
			topLevel = append(topLevel, ns)
		}
	}

	sort.Strings(topLevel)

	writeNodesTo(root, nsToNodes[""], reg)

	for _, ns := range topLevel {
		writeNamespaceSubgraphTo(root, ns, nsToNodes, allNS, reg)
	}
}

// findAllNamespaces returns a set of all namespaces found in the given node map.
func findAllNamespaces(nsToNodes map[string][]*graph.Node) map[string]bool {
	allNS := make(map[string]bool)

	for ns := range nsToNodes {
		for current := ns; current != ""; current = parentNamespace(current) {
			allNS[current] = true
		}
	}

	return allNS
}

func indexNodesByNamespace(nodes []*graph.Node) map[string][]*graph.Node {
	nsToNodes := make(map[string][]*graph.Node)

	for _, node := range nodes {
		ns := nodeNamespace(node.ID())
		nsToNodes[ns] = append(nsToNodes[ns], node)
	}

	return nsToNodes
}

// writeNamespaceSubgraphTo writes a subgraph for the given namespace, recursively handling children.
func writeNamespaceSubgraphTo(
	parent *indentwriter.Line,
	ns string,
	nsToNodes map[string][]*graph.Node,
	allNS map[string]bool,
	reg *safe.Registry,
) {
	sg := parent.Addf("subgraph %s[\"%s\"]", reg.IDWithPrefix("sg_", ns), ns)

	children := make([]string, 0, len(allNS))
	for candidate := range allNS {
		if parentNamespace(candidate) == ns {
			children = append(children, candidate)
		}
	}

	sort.Strings(children)

	writeNodesTo(sg, nsToNodes[ns], reg)

	for _, child := range children {
		writeNamespaceSubgraphTo(sg, child, nsToNodes, allNS, reg)
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
	label := safe.Label(nodeDisplayLabel(node))
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
) {
	if cfg == nil || len(cfg.NodeStyleRules) == 0 {
		return
	}

	for i, rule := range cfg.NodeStyleRules {
		writeStyleRuleTo(root, nodes, i, rule, reg)
	}
}

func writeStyleRuleTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	index int,
	rule config.NodeStyleRule,
	reg *safe.Registry,
) {
	classDef := buildClassDef(rule)
	if classDef == "" {
		return
	}

	matchingIDs := findMatchingNodeIDs(nodes, rule.Match, reg)
	if len(matchingIDs) > 0 {
		sort.Strings(matchingIDs)
		root.Addf("classDef rule%d %s", index, classDef)
		root.Addf("class %s rule%d", strings.Join(matchingIDs, ","), index)
	}
}

func findMatchingNodeIDs(nodes []*graph.Node, pattern string, reg *safe.Registry) []string {
	var matchingIDs []string

	for _, node := range nodes {
		matched, err := path.Match(pattern, node.ID())
		if err == nil && matched {
			matchingIDs = append(matchingIDs, reg.ID(node.ID()))
		}
	}

	return matchingIDs
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

// nodeDisplayLabel returns the display label for a node.
func nodeDisplayLabel(node *graph.Node) string {
	if node.Label != "" {
		return node.Label
	}

	return node.ID()
}
