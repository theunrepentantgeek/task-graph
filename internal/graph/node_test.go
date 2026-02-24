package graph

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestNode_ID_ReturnsProvidedIdentifier(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	node := newNode("identifier")

	g.Expect(node.ID()).To(gomega.Equal("identifier"))
}

func TestNode_AddEdge_ToTargetNode_AppendsEdge(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	source := newNode("source")
	target := newNode("target")

	edge := source.AddEdge(target)

	g.Expect(source.edges).To(gomega.HaveLen(1))
	g.Expect(source.edges[0]).To(gomega.BeIdenticalTo(edge))
	g.Expect(edge.From()).To(gomega.BeIdenticalTo(source))
	g.Expect(edge.To()).To(gomega.BeIdenticalTo(target))
}
