package graphns_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/graphns"
)

// helpers

func makeTaskNode(id string) *graph.Node {
	n := graph.NewNode(id)

	return n
}

func makeVarNode(id string) *graph.Node {
	n := graph.NewNode(id)
	n.Kind = graph.NodeKindVariable

	return n
}

// SplitByKind tests

func TestSplitByKind_EmptySlice_ReturnsTwoEmptySlices(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	tasks, vars := graphns.SplitByKind(nil)

	// Assert
	g.Expect(tasks).To(BeEmpty())
	g.Expect(vars).To(BeEmpty())
}

func TestSplitByKind_MixedNodes_PartitionsCorrectly(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	nodes := []*graph.Node{
		makeTaskNode("build"),
		makeVarNode("var:VERSION"),
		makeTaskNode("test"),
		makeVarNode("var:IMAGE"),
	}

	// Act
	tasks, vars := graphns.SplitByKind(nodes)

	// Assert
	g.Expect(tasks).To(HaveLen(2))
	g.Expect(vars).To(HaveLen(2))
	g.Expect(tasks[0].ID()).To(Equal("build"))
	g.Expect(tasks[1].ID()).To(Equal("test"))
	g.Expect(vars[0].ID()).To(Equal("var:VERSION"))
	g.Expect(vars[1].ID()).To(Equal("var:IMAGE"))
}

func TestSplitByKind_AllTaskNodes_ReturnsAllInTaskSlice(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	nodes := []*graph.Node{makeTaskNode("build"), makeTaskNode("test")}

	// Act
	tasks, vars := graphns.SplitByKind(nodes)

	// Assert
	g.Expect(tasks).To(HaveLen(2))
	g.Expect(vars).To(BeEmpty())
}

// IndexByNamespace tests

func TestIndexByNamespace_NoNamespace_StoredUnderEmptyKey(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	nodes := []*graph.Node{makeTaskNode("build"), makeTaskNode("test")}

	// Act
	result := graphns.IndexByNamespace(nodes)

	// Assert
	g.Expect(result[""]).To(HaveLen(2))
}

func TestIndexByNamespace_WithNamespaces_GroupedCorrectly(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	nodes := []*graph.Node{
		makeTaskNode("cmd:build"),
		makeTaskNode("cmd:test"),
		makeTaskNode("deploy"),
	}

	// Act
	result := graphns.IndexByNamespace(nodes)

	// Assert
	g.Expect(result["cmd"]).To(HaveLen(2))
	g.Expect(result[""]).To(HaveLen(1))
}

// FindAllNamespaces tests

func TestFindAllNamespaces_NestedNamespace_IncludesIntermediates(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange — node a:b:c yields namespaces a:b and a
	nsToNodes := map[string][]*graph.Node{
		"a:b": {makeTaskNode("a:b:task")},
	}

	// Act
	result := graphns.FindAllNamespaces(nsToNodes)

	// Assert
	g.Expect(result).To(HaveKey("a:b"))
	g.Expect(result).To(HaveKey("a"))
}

func TestFindAllNamespaces_TopLevelOnly_ReturnsSingleEntry(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	nsToNodes := map[string][]*graph.Node{
		"cmd": {makeTaskNode("cmd:build")},
	}

	// Act
	result := graphns.FindAllNamespaces(nsToNodes)

	// Assert
	g.Expect(result).To(HaveLen(1))
	g.Expect(result).To(HaveKey("cmd"))
}

// BuildChildrenMap tests

func TestBuildChildrenMap_SimpleHierarchy_MapsParentToChildren(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	allNS := map[string]bool{
		"cmd":       true,
		"cmd:build": true,
		"cmd:test":  true,
	}

	// Act
	result := graphns.BuildChildrenMap(allNS)

	// Assert
	g.Expect(result[""]).To(ConsistOf("cmd"))
	g.Expect(result["cmd"]).To(ConsistOf("cmd:build", "cmd:test"))
}

func TestBuildChildrenMap_ChildrenAreSorted(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	allNS := map[string]bool{
		"b": true,
		"a": true,
		"c": true,
	}

	// Act
	result := graphns.BuildChildrenMap(allNS)

	// Assert
	g.Expect(result[""]).To(Equal([]string{"a", "b", "c"}))
}
