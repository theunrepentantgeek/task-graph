package autocolor

import (
	"cmp"
	"slices"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/namespace"
)

// palette is the ordered list of fill colors used for auto-coloring namespaces.
// Colors are selected to be visually distinct and suitable for node backgrounds.
var palette = []string{
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
	if len(palette) == 0 {
		return nil
	}

	namespaces := collectAllNamespaces(gr)
	sortNamespaces(namespaces)

	rules := make([]config.NodeStyleRule, 0, len(namespaces)*2)
	for i, ns := range namespaces {
		color := palette[i%len(palette)]

		// Exact match for the namespace task itself (e.g. "tidy")
		rules = append(rules, config.NodeStyleRule{
			Match:     ns,
			FillColor: color,
			Style:     "filled",
		})

		// Children match for subtasks (e.g. "tidy:gofumpt", "tidy:lint")
		rules = append(rules, config.NodeStyleRule{
			Match:     namespace.MatchPattern(ns),
			FillColor: color,
			Style:     "filled",
		})
	}

	return rules
}

// collectAllNamespaces returns all distinct namespaces found in the graph,
// including parent namespaces.
func collectAllNamespaces(gr *graph.Graph) []string {
	seen := make(map[string]bool)

	for node := range gr.Nodes() {
		ns := namespace.Namespace(node.ID())
		for ns != "" {
			seen[ns] = true
			ns = namespace.Parent(ns)
		}
	}

	result := make([]string, 0, len(seen))
	for ns := range seen {
		result = append(result, ns)
	}

	return result
}

// sortNamespaces sorts namespaces so that shallower ones come first,
// and within the same depth, sorts alphabetically for deterministic color assignment.
func sortNamespaces(namespaces []string) {
	slices.SortFunc(namespaces, func(a, b string) int {
		depthA := namespace.Depth(a)
		depthB := namespace.Depth(b)

		if depthA != depthB {
			return cmp.Compare(depthA, depthB)
		}

		return cmp.Compare(a, b)
	})
}
