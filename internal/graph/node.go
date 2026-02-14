package graph

// NodeID represents a unique identifier for a node in the graph.
type NodeID struct {
	// ID returns the unique identifier for the node.
	id string
}

// Node represents a vertex in the graph, containing a unique identifier.
type Node struct {
	NodeID
	// Label returns the label of the node.
	Label string

	// Edges holds the outgoing edges from this node to other nodes in the graph.
	edges []*Edge
}

// newNode creates a new node with the given ID and returns it.
func newNode(id string) *Node {
	return &Node{
		NodeID: NodeID{id: id},
	}
}

// ID returns the unique identifier of the node.
func (n *Node) ID() string {
	return n.id
}

// Edges returns a slice of all outgoing edges from this node.
func (n *Node) AddEdge(to *Node) *Edge {
	edge := newEdge(n, to)
	n.edges = append(n.edges, edge)
	return edge
}

// Edges returns the outgoing edges for this node.
func (n *Node) Edges() []*Edge {
	return n.edges
}
