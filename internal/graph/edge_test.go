package graph

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestEdge_Label_GivenNewEdge_ReturnsEmptyString(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	edge := newEdge(NewNode("from"), NewNode("to"))

	g.Expect(edge.Label()).To(gomega.Equal(""))
}

func TestEdge_SetLabel_WithValue_UpdatesLabel(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	edge := newEdge(NewNode("from"), NewNode("to"))

	edge.SetLabel("edge-label")

	g.Expect(edge.Label()).To(gomega.Equal("edge-label"))
}

func TestEdge_Class_GivenNewEdge_ReturnsEmptyString(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	edge := newEdge(NewNode("from"), NewNode("to"))

	g.Expect(edge.Class()).To(gomega.Equal(""))
}

func TestEdge_SetClass_WithValue_UpdatesClass(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	edge := newEdge(NewNode("from"), NewNode("to"))

	edge.SetClass("dependency")

	g.Expect(edge.Class()).To(gomega.Equal("dependency"))
}

func TestEdge_From_ReturnsSourceNode(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	source := NewNode("source")
	target := NewNode("target")

	edge := newEdge(source, target)

	g.Expect(edge.From()).To(gomega.BeIdenticalTo(source))
}

func TestEdge_To_ReturnsTargetNode(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	source := NewNode("source")
	target := NewNode("target")

	edge := newEdge(source, target)

	g.Expect(edge.To()).To(gomega.BeIdenticalTo(target))
}
