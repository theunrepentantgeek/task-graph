package graph

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestNode_ID_ReturnsProvidedIdentifier(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	node := NewNode("identifier")

	g.Expect(node.ID()).To(gomega.Equal("identifier"))
}

func TestNode_AddEdge_ToTargetNode_AppendsEdge(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	source := NewNode("source")
	target := NewNode("target")

	edge := source.AddEdge(target)

	g.Expect(source.edges).To(gomega.HaveLen(1))
	g.Expect(source.edges[0]).To(gomega.BeIdenticalTo(edge))
	g.Expect(edge.From()).To(gomega.BeIdenticalTo(source))
	g.Expect(edge.To()).To(gomega.BeIdenticalTo(target))
}

func TestNode_Kind_DefaultsToTask(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	node := NewNode("test")

	// Assert
	g.Expect(node.Kind).To(gomega.Equal(NodeKindTask))
}

func TestNode_Kind_CanBeSetToVariable(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	node := NewNode("var:FOO")

	// Act
	node.Kind = NodeKindVariable

	// Assert
	g.Expect(node.Kind).To(gomega.Equal(NodeKindVariable))
}

func TestNode_Edges_WithNoEdges_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	node := NewNode("alone")

	g.Expect(node.Edges()).To(gomega.BeEmpty())
}

func TestNode_Edges_WithMultipleEdges_ReturnsAll(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	source := NewNode("source")
	target1 := NewNode("target1")
	target2 := NewNode("target2")

	edge1 := source.AddEdge(target1)
	edge2 := source.AddEdge(target2)

	edges := source.Edges()

	g.Expect(edges).To(gomega.HaveLen(2))
	g.Expect(edges[0]).To(gomega.BeIdenticalTo(edge1))
	g.Expect(edges[1]).To(gomega.BeIdenticalTo(edge2))
}
