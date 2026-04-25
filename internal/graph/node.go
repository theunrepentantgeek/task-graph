package graph

// NodeKind represents the type of a node in the graph.
type NodeKind string

const (
	// NodeKindTask represents a task node.
	NodeKindTask NodeKind = "task"

	// NodeKindVariable represents a global variable node.
	NodeKindVariable NodeKind = "variable"
)

// NodeID represents a unique identifier for a node in the graph.
type NodeID struct {
	// ID returns the unique identifier for the node.
	id string
}

// Node represents a vertex in the graph, containing a unique identifier.
type Node struct {
	NodeID

	// Kind identifies the type of this node (task or variable).
	Kind NodeKind

	// Label returns the label of the node.
	Label string

	// Description returns the description of the node.
	Description string

	// Edges holds the outgoing edges from this node to other nodes in the graph.
	edges []*Edge
}

// NewNode creates a new node with the given ID and returns it.
func NewNode(id string) *Node {
	return &Node{
		NodeID: NodeID{id: id},
		Kind:   NodeKindTask,
	}
}

// ID returns the unique identifier of the node.
func (n *Node) ID() string {
	return n.id
}

// AddEdge creates a new edge from this node to the specified node and returns it.
func (n *Node) AddEdge(to *Node) *Edge {
	edge := newEdge(n, to)
	n.edges = append(n.edges, edge)

	return edge
}

// Edges returns the outgoing edges for this node.
func (n *Node) Edges() []*Edge {
	return n.edges
}

// DisplayLabel returns the label to use when displaying this node.
// If the node has an explicit Label set, that is returned; otherwise the node ID is used.
func (n *Node) DisplayLabel() string {
	if n.Label != "" {
		return n.Label
	}

	return n.id
}
