package autocolor

import (
	"cmp"
	"slices"
	"strings"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
)

// Palette is the ordered list of fill colors used for auto-coloring namespaces.
// Colors are selected to be visually distinct and suitable for node backgrounds.
var Palette = []string{
	"lightblue",
	"lightgreen",
	"lightyellow",
	"lightsalmon",
	"lightpink",
	"lightgray",
	"lavender",
	"peachpuff",
}

// GenerateRules generates a NodeStyleRule for each distinct namespace found in the graph.
// Colors are assigned from the Palette in order: shallower namespaces first, then
// alphabetically within the same depth. When there are more namespaces than palette entries,
// colors wrap around.
//
// The generated rules should be prepended to any existing NodeStyleRules so that
// user-defined rules take precedence over the auto-generated ones.
func GenerateRules(gr *graph.Graph) []config.NodeStyleRule {
	if len(Palette) == 0 {
		return nil
	}

	namespaces := collectAllNamespaces(gr)
	sortNamespaces(namespaces)

	rules := make([]config.NodeStyleRule, 0, len(namespaces))
	for i, ns := range namespaces {
		color := Palette[i%len(Palette)]
		rules = append(rules, config.NodeStyleRule{
			Match:     ns + ":*",
			FillColor: color,
			Style:     "filled",
		})
	}

	return rules
}

// collectAllNamespaces returns all distinct namespaces found in the graph.
// A namespace is any prefix prior to a colon in a node ID, including parent namespaces.
// For example, a node "cmd:test:unit" contributes namespaces "cmd:test" and "cmd".
func collectAllNamespaces(gr *graph.Graph) []string {
	seen := make(map[string]bool)

	for node := range gr.Nodes() {
		ns := nodeNamespace(node.ID())
		for ns != "" {
			seen[ns] = true
			ns = parentNamespace(ns)
		}
	}

	result := make([]string, 0, len(seen))
	for ns := range seen {
		result = append(result, ns)
	}

	return result
}

// sortNamespaces sorts namespaces so that shallower ones come first (fewer colons),
// and within the same depth, sorts alphabetically for deterministic color assignment.
func sortNamespaces(namespaces []string) {
	slices.SortFunc(namespaces, func(a, b string) int {
		depthA := strings.Count(a, ":")
		depthB := strings.Count(b, ":")

		if depthA != depthB {
			return cmp.Compare(depthA, depthB)
		}

		return cmp.Compare(a, b)
	})
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
