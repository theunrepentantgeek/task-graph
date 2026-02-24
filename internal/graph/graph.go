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
	node := newNode(id)
	g.nodes[id] = node

	return node
}

// Node returns the node with the given ID if it exists in the graph,
// otherwise it returns false.
func (g *Graph) Node(id string) (*Node, bool) {
	node, exists := g.nodes[id]

	return node, exists
}

// Nodes implements an interator to allow iterating over all nodes in the graph.
func (g *Graph) Nodes() iter.Seq[*Node] {
	return maps.Values(g.nodes)
}
