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

func TestGraph_Nodes_WithEarlyTermination_StopsIteration(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	graph := New()
	graph.AddNode("alpha")
	graph.AddNode("beta")
	graph.AddNode("gamma")

	var count int

	graph.Nodes()(func(_ *Node) bool {
		count++

		return false // stop after first node
	})

	g.Expect(count).To(gomega.Equal(1))
}

func TestGraph_Nodes_EmptyGraph_IteratesNothing(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	graph := New()

	var count int

	graph.Nodes()(func(_ *Node) bool {
		count++

		return true
	})

	g.Expect(count).To(gomega.Equal(0))
}

// ReachableFrom tests

func TestGraph_ReachableFrom_EmptySeeds_ReturnsEmptySet(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	graph := New()
	graph.AddNode("A")
	graph.AddNode("B")

	// Act
	result := graph.ReachableFrom(map[string]bool{})

	// Assert
	g.Expect(result).To(gomega.BeEmpty())
}

func TestGraph_ReachableFrom_SingleIsolatedSeed_ReturnsSeedOnly(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	graph := New()
	graph.AddNode("A")
	graph.AddNode("B")

	// Act
	result := graph.ReachableFrom(map[string]bool{"A": true})

	// Assert
	g.Expect(result).To(gomega.Equal(map[string]bool{"A": true}))
}

func TestGraph_ReachableFrom_SeedWithDependency_IncludesTransitiveDependency(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange: build -> compile -> link
	graph := New()
	build := graph.AddNode("build")
	compile := graph.AddNode("compile")
	link := graph.AddNode("link")

	build.AddEdge(compile)
	compile.AddEdge(link)

	// Act: focus on build
	result := graph.ReachableFrom(map[string]bool{"build": true})

	// Assert: entire forward chain is included
	g.Expect(result).To(gomega.Equal(map[string]bool{
		"build":   true,
		"compile": true,
		"link":    true,
	}))
}

func TestGraph_ReachableFrom_SeedWithDependent_IncludesDependent(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange: ci -> test -> unit-test
	graph := New()
	ci := graph.AddNode("ci")
	test := graph.AddNode("test")
	unitTest := graph.AddNode("unit-test")

	ci.AddEdge(test)
	test.AddEdge(unitTest)

	// Act: focus on an intermediate node
	result := graph.ReachableFrom(map[string]bool{"test": true})

	// Assert: upstream dependent (ci) and downstream dependency (unit-test) are both included
	g.Expect(result).To(gomega.Equal(map[string]bool{
		"ci":        true,
		"test":      true,
		"unit-test": true,
	}))
}

func TestGraph_ReachableFrom_SeedNotInGraph_ReturnsEmptySet(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	graph := New()
	graph.AddNode("A")

	// Act
	result := graph.ReachableFrom(map[string]bool{"missing": true})

	// Assert
	g.Expect(result).To(gomega.BeEmpty())
}

func TestGraph_ReachableFrom_DisconnectedSubgraph_DoesNotCrossComponent(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange: A -> B, C -> D (two disconnected components)
	graph := New()
	a := graph.AddNode("A")
	b := graph.AddNode("B")
	c := graph.AddNode("C")
	d := graph.AddNode("D")

	a.AddEdge(b)
	c.AddEdge(d)

	// Act: focus on A
	result := graph.ReachableFrom(map[string]bool{"A": true})

	// Assert: only A and its dependency B
	g.Expect(result).To(gomega.Equal(map[string]bool{
		"A": true,
		"B": true,
	}))
}

// FilterNodes tests

func TestGraph_FilterNodes_EmptyKeep_ReturnsEmptyGraph(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	graph := New()
	graph.AddNode("A")
	graph.AddNode("B")

	// Act
	result := graph.FilterNodes(map[string]bool{})

	// Assert
	var ids []string
	for n := range result.Nodes() {
		ids = append(ids, n.ID())
	}

	g.Expect(ids).To(gomega.BeEmpty())
}

func TestGraph_FilterNodes_AllNodes_ReturnsEquivalentGraph(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	graph := New()
	a := graph.AddNode("A")
	b := graph.AddNode("B")
	a.AddEdge(b).SetClass("dep")

	// Act
	result := graph.FilterNodes(map[string]bool{"A": true, "B": true})

	// Assert: both nodes present
	_, okA := result.Node("A")
	_, okB := result.Node("B")

	g.Expect(okA).To(gomega.BeTrue())
	g.Expect(okB).To(gomega.BeTrue())

	// Edge is preserved
	resA, _ := result.Node("A")
	g.Expect(resA.Edges()).To(gomega.HaveLen(1))
	g.Expect(resA.Edges()[0].Class()).To(gomega.Equal("dep"))
}

func TestGraph_FilterNodes_SubsetOfNodes_EdgesToExcludedNodesAreDropped(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange: A -> B -> C; keep only A and B
	graph := New()
	a := graph.AddNode("A")
	b := graph.AddNode("B")
	c := graph.AddNode("C")

	a.AddEdge(b)
	b.AddEdge(c)

	// Act
	result := graph.FilterNodes(map[string]bool{"A": true, "B": true})

	// Assert: C is excluded; B has no edges
	_, okC := result.Node("C")
	g.Expect(okC).To(gomega.BeFalse())

	resB, _ := result.Node("B")
	g.Expect(resB.Edges()).To(gomega.BeEmpty())
}

func TestGraph_FilterNodes_NodeMetadata_IsPreserved(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	graph := New()
	node := graph.AddNode("A")
	node.Label = "Build"
	node.Description = "Builds the project"

	// Act
	result := graph.FilterNodes(map[string]bool{"A": true})

	// Assert
	resA, _ := result.Node("A")
	g.Expect(resA.Label).To(gomega.Equal("Build"))
	g.Expect(resA.Description).To(gomega.Equal("Builds the project"))
}

func TestGraph_FilterNodes_VariableNodeKind_IsPreserved(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	graph := New()
	varNode := graph.AddNode("MY_VAR")
	varNode.Kind = NodeKindVariable

	// Act
	result := graph.FilterNodes(map[string]bool{"MY_VAR": true})

	// Assert
	resVar, _ := result.Node("MY_VAR")
	g.Expect(resVar.Kind).To(gomega.Equal(NodeKindVariable))
}

func TestGraph_FilterNodes_EdgeLabel_IsPreserved(t *testing.T) {
t.Parallel()
g := gomega.NewWithT(t)

// Arrange: A -[next]-> B -> C; keep only A and B
gr := New()
a := gr.AddNode("A")
b := gr.AddNode("B")
c := gr.AddNode("C")

labeled := a.AddEdge(b)
labeled.SetLabel("next")
b.AddEdge(c)

// Act
result := gr.FilterNodes(map[string]bool{"A": true, "B": true})

// Assert: the labeled edge A->B is present and its label is preserved
resA, _ := result.Node("A")
g.Expect(resA.Edges()).To(gomega.HaveLen(1))
g.Expect(resA.Edges()[0].Label()).To(gomega.Equal("next"))
}
