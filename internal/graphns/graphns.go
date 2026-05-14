// Package graphns provides shared utilities for namespace-aware graph operations.
// It is used by both the graphviz and mermaid rendering packages to avoid
// duplicating the logic for partitioning and organising nodes by namespace.
package graphns

import (
	"slices"

	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/namespace"
)

// SplitByKind partitions nodes into task nodes and variable nodes.
//
//nolint:revive // Choosing to return two unnamed slices
func SplitByKind(nodes []*graph.Node) ([]*graph.Node, []*graph.Node) {
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

// IndexByNamespace groups nodes by their namespace, returning a map from
// namespace string to the slice of nodes directly within that namespace.
// Nodes with no namespace are stored under the empty-string key.
func IndexByNamespace(nodes []*graph.Node) map[string][]*graph.Node {
	nsToNodes := make(map[string][]*graph.Node)

	for _, node := range nodes {
		ns := namespace.Namespace(node.ID())
		nsToNodes[ns] = append(nsToNodes[ns], node)
	}

	return nsToNodes
}

// FindAllNamespaces returns the complete set of all namespace strings present
// in nsToNodes, including all intermediate parent namespaces.
func FindAllNamespaces(nsToNodes map[string][]*graph.Node) map[string]bool {
	allNS := make(map[string]bool)

	for ns := range nsToNodes {
		for current := ns; current != ""; current = namespace.Parent(current) {
			allNS[current] = true
		}
	}

	return allNS
}

// BuildChildrenMap builds a parent→sorted-children map from a set of all
// namespaces. The empty-string key ("") holds the sorted list of top-level
// namespaces. Building this map once avoids an O(N) scan for every namespace
// during the recursive subgraph walk.
func BuildChildrenMap(allNS map[string]bool) map[string][]string {
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
