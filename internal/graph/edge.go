package graph

// Edge represents a directed connection between two nodes in the graph,
// with an optional label.
type Edge struct {
	label string
	class string
	from  *Node
	to    *Node
}

// NewEdge creates a new edge from the given source node to the target node with an optional label.
func newEdge(from *Node, to *Node) *Edge {
	return &Edge{
		from: from,
		to:   to,
	}
}

// Label returns the label of the edge.
func (e *Edge) Label() string {
	return e.label
}

// SetLabel sets the label of the edge to the given value.
func (e *Edge) SetLabel(label string) {
	e.label = label
}

// Class returns the class of the edge.
// The class can be used to categorize edges into different types or groups.
func (e *Edge) Class() string {
	return e.class
}

// SetClass sets the class of the edge.
func (e *Edge) SetClass(class string) {
	e.class = class
}

// From returns the source node of the edge.
func (e *Edge) From() *Node {
	return e.from
}

// To returns the target node of the edge.
func (e *Edge) To() *Node {
	return e.to
}
