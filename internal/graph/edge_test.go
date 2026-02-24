package graph

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestEdge_Label_GivenNewEdge_ReturnsEmptyString(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	edge := newEdge(newNode("from"), newNode("to"))

	g.Expect(edge.Label()).To(gomega.Equal(""))
}

func TestEdge_SetLabel_WithValue_UpdatesLabel(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	edge := newEdge(newNode("from"), newNode("to"))

	edge.SetLabel("edge-label")

	g.Expect(edge.Label()).To(gomega.Equal("edge-label"))
}
