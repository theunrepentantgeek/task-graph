package graph

import (
	"iter"
	"maps"
)

// Graph represents an abstraction for a directed graph structure.
type Graph struct {
	nodes map[string]*Node
}

// New creates and returns a new instance of Graph.
func New() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
	}
}

// AddNode adds a new node with the given ID to the graph and returns it.
// id is a unique identifier for the node. If a node with the same ID already exists,
// it will be overwritten.
func (g *Graph) AddNode(id string) *Node {
	node := NewNode(id)
	g.nodes[id] = node

	return node
}

// Node returns the node with the given ID if it exists in the graph,
// otherwise it returns false.
func (g *Graph) Node(id string) (*Node, bool) {
	node, exists := g.nodes[id]

	return node, exists
}

// Nodes implements an iterator to allow iterating over all nodes in the graph.
func (g *Graph) Nodes() iter.Seq[*Node] {
	return maps.Values(g.nodes)
}

// ReachableFrom returns the set of all node IDs transitively reachable from the
// given seed node IDs, traversing edges in both directions (dependencies and
// dependents). Seeds that do not exist in the graph are silently ignored.
//
//nolint:revive // Difficult to simplify
func (g *Graph) ReachableFrom(
	seeds map[string]bool,
) map[string]bool {
	if len(seeds) == 0 {
		return make(map[string]bool)
	}

	// Build a reverse adjacency list so we can walk backwards (dependents).
	incoming := g.indexByIncomingEdge()

	visited := make(map[string]bool, len(seeds))
	queue := g.createScanningQueue(seeds)

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if visited[cur] {
			continue
		}

		visited[cur] = true
		node := g.nodes[cur]

		// Forward: follow outgoing edges to dependencies.
		for _, edge := range node.Edges() {
			queue = append(queue, edge.To().ID())
		}

		// Backward: follow incoming edges to dependents.
		for _, from := range incoming[cur] {
			queue = append(queue, from.ID())
		}
	}

	return visited
}

// FilterNodes returns a new graph containing only the nodes present in the keep
// set, along with the edges between them. Node metadata (Kind, Label and Description)
// is preserved.
func (g *Graph) FilterNodes(keep map[string]bool) *Graph {
	result := New()

	for id, node := range g.nodes {
		if !keep[id] {
			continue
		}

		newNode := result.AddNode(id)
		newNode.Kind = node.Kind
		newNode.Label = node.Label
		newNode.Description = node.Description
	}

	for id, node := range g.nodes {
		if !keep[id] {
			continue
		}

		copyFilteredEdges(result, keep, id, node)
	}

	return result
}

func (g *Graph) createScanningQueue(seeds map[string]bool) []string {
	queue := make([]string, 0, len(seeds))

	for id := range seeds {
		if _, ok := g.nodes[id]; ok {
			queue = append(queue, id)
		}
	}

	return queue
}

// indexByIncomingEdge builds a map from node ID to the list of nodes that have
// edges pointing to it.
func (g *Graph) indexByIncomingEdge() map[string][]*Node {
	incoming := make(map[string][]*Node, len(g.nodes))
	for _, node := range g.nodes {
		for _, edge := range node.Edges() {
			toID := edge.To().ID()
			incoming[toID] = append(incoming[toID], node)
		}
	}

	return incoming
}

// copyFilteredEdges copies node's edges into result, skipping edges whose
// target is not in the keep set.
func copyFilteredEdges(result *Graph, keep map[string]bool, id string, node *Node) {
	fromNode, _ := result.Node(id)

	for _, edge := range node.Edges() {
		if !keep[edge.To().ID()] {
			continue
		}

		toNode, _ := result.Node(edge.To().ID())
		newEdge := fromNode.AddEdge(toNode)
		newEdge.SetClass(edge.Class())

		if edge.Label() != "" {
			newEdge.SetLabel(edge.Label())
		}
	}
}
