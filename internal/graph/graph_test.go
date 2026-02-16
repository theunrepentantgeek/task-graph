package graph

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestGraph_AddNode_NewID_NodeStored(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	graph := New()

	node := graph.AddNode("A")

	g.Expect(node).NotTo(gomega.BeNil())
	g.Expect(node.ID()).To(gomega.Equal("A"))

	retrieved, ok := graph.Node("A")

	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(retrieved).To(gomega.BeIdenticalTo(node))
}

func TestGraph_AddNode_DuplicateID_OverwritesExistingNode(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	graph := New()
	first := graph.AddNode("dup")
	first.Label = "first"

	second := graph.AddNode("dup")

	retrieved, ok := graph.Node("dup")

	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(second).NotTo(gomega.BeNil())
	g.Expect(second).NotTo(gomega.BeIdenticalTo(first))
	g.Expect(retrieved).To(gomega.BeIdenticalTo(second))
	g.Expect(retrieved.Label).To(gomega.Equal(""))
}

func TestGraph_Node_ExistingID_ReturnsNodeAndTrue(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	graph := New()
	expected := graph.AddNode("node-1")

	node, ok := graph.Node("node-1")

	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(node).To(gomega.BeIdenticalTo(expected))
}

func TestGraph_Node_MissingID_ReturnsNilAndFalse(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	graph := New()

	node, ok := graph.Node("missing")

	g.Expect(ok).To(gomega.BeFalse())
	g.Expect(node).To(gomega.BeNil())
}

func TestGraph_Nodes_WithMultipleNodes_IteratesAllNodes(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	graph := New()
	graph.AddNode("alpha")
	graph.AddNode("beta")
	graph.AddNode("gamma")

	var ids []string

	graph.Nodes()(func(n *Node) bool {
		ids = append(ids, n.ID())

		return true
	})

	g.Expect(ids).To(gomega.ConsistOf("alpha", "beta", "gamma"))
}
