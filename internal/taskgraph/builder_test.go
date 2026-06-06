package taskgraph

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-task/task/v3/taskfile/ast"

	"github.com/theunrepentantgeek/task-graph/internal/graph"
)

// makeTaskfile creates a minimal ast.Taskfile with the given task elements for use in tests.
func makeTaskfile(elements ...*ast.TaskElement) *ast.Taskfile {
	return &ast.Taskfile{
		Tasks: ast.NewTasks(elements...),
	}
}

// TestBuilder_Build_UndefinedDependency_IsSkipped verifies that deps pointing to tasks
// not defined in the Taskfile are silently skipped (no edge is created).
func TestBuilder_Build_UndefinedDependency_IsSkipped(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tf := makeTaskfile(
		&ast.TaskElement{
			Key: "task-a",
			Value: &ast.Task{
				Deps: []*ast.Dep{
					{Task: "undefined-dep"},
				},
			},
		},
	)

	gr := New(tf).Build()

	node, ok := gr.Node("task-a")
	g.Expect(ok).To(BeTrue(), "task-a node should exist")
	g.Expect(node.Edges()).To(BeEmpty(), "no edges should be created for undefined deps")
}

// TestBuilder_Build_UndefinedCall_IsSkipped verifies that cmd task calls pointing to tasks
// not defined in the Taskfile are silently skipped (no edge is created).
func TestBuilder_Build_UndefinedCall_IsSkipped(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tf := makeTaskfile(
		&ast.TaskElement{
			Key: "task-a",
			Value: &ast.Task{
				Cmds: []*ast.Cmd{
					{Task: "undefined-call"},
				},
			},
		},
	)

	gr := New(tf).Build()

	node, ok := gr.Node("task-a")
	g.Expect(ok).To(BeTrue(), "task-a node should exist")
	g.Expect(node.Edges()).To(BeEmpty(), "no edges should be created for undefined calls")
}

// TestBuilder_Build_ValidEdge_CreatesEdge verifies that both dep and call edges are
// correctly created between two defined tasks.
func TestBuilder_Build_ValidEdge_CreatesEdge(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		taskElement *ast.TaskElement
		wantClass   string
	}{
		"dep": {
			taskElement: &ast.TaskElement{
				Key: "task-a",
				Value: &ast.Task{
					Deps: []*ast.Dep{
						{Task: "task-b"},
					},
				},
			},
			wantClass: "dep",
		},
		"call": {
			taskElement: &ast.TaskElement{
				Key: "task-a",
				Value: &ast.Task{
					Cmds: []*ast.Cmd{
						{Task: "task-b"},
					},
				},
			},
			wantClass: "call",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			tf := makeTaskfile(
				tc.taskElement,
				&ast.TaskElement{
					Key:   "task-b",
					Value: &ast.Task{},
				},
			)

			gr := New(tf).Build()

			nodeA, ok := gr.Node("task-a")
			g.Expect(ok).To(BeTrue())

			edges := nodeA.Edges()

			g.Expect(edges).To(HaveLen(1), "one edge should be created")
			g.Expect(edges[0].Class()).To(Equal(tc.wantClass))
			g.Expect(edges[0].To().ID()).To(Equal("task-b"))
		})
	}
}

// TestBuilder_Build_NonTaskCmd_NoEdge verifies that shell commands (not task calls)
// do not produce edges.
func TestBuilder_Build_NonTaskCmd_NoEdge(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tf := makeTaskfile(
		&ast.TaskElement{
			Key: "task-a",
			Value: &ast.Task{
				Cmds: []*ast.Cmd{
					{Cmd: "echo hello"},
				},
			},
		},
	)

	gr := New(tf).Build()

	node, ok := gr.Node("task-a")
	g.Expect(ok).To(BeTrue())
	g.Expect(node.Edges()).To(BeEmpty(), "shell commands should not produce edges")
}

// TestBuilder_Build_IncludeGlobalVars_NilVars_ProducesNoVariableNodes verifies that
// enabling IncludeGlobalVars with a taskfile that has no global variables (Vars == nil)
// produces a graph with no variable nodes and no errors.
func TestBuilder_Build_IncludeGlobalVars_NilVars_ProducesNoVariableNodes(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tf := makeTaskfile(
		&ast.TaskElement{
			Key:   "task-a",
			Value: &ast.Task{},
		},
	)
	// makeTaskfile does not set Vars, so tf.Vars is nil.

	builder := New(tf)
	builder.IncludeGlobalVars = true
	gr := builder.Build()

	node, ok := gr.Node("task-a")
	g.Expect(ok).To(BeTrue(), "task-a node should still be created")
	g.Expect(node.Edges()).To(BeEmpty(), "no edges expected when there are no global vars")
}

// TestAddEdgesForVarRefs_MissingVarNode_SkipsEdge verifies that when a referenced
// variable name has no corresponding node in the graph, no edge is created (defensive
// guard on the varNode lookup).
func TestAddEdgesForVarRefs_MissingVarNode_SkipsEdge(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	builder := &Builder{}
	gr := graph.New()
	gr.AddNode("task-a")
	// No "var:X" node is added — varNode lookup will fail.

	builder.addEdgesForVarRefs(gr, "task-a", map[string]bool{"X": true})

	node, ok := gr.Node("task-a")
	g.Expect(ok).To(BeTrue())
	g.Expect(node.Edges()).To(BeEmpty(), "no edge should be created when var node is missing")
}

// TestAddEdgesForVarRefs_MissingTaskNode_SkipsEdge verifies that when the task
// referenced by taskName has no node in the graph, no edge is created (defensive
// guard on the taskNode lookup).
func TestAddEdgesForVarRefs_MissingTaskNode_SkipsEdge(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	builder := &Builder{}
	gr := graph.New()
	varNode := gr.AddNode("var:X")
	// "task-missing" is intentionally NOT added to the graph.

	builder.addEdgesForVarRefs(gr, "task-missing", map[string]bool{"X": true})

	g.Expect(varNode.Edges()).To(BeEmpty(), "no edge should be created when task node is missing")
}

// TestAddEdgesForDependencies_MissingSourceNode_IsNoop verifies that calling
// addEdgesForDependencies with a taskID that is not present in the graph is a no-op
// (defensive guard — should not happen in normal Build flow, but must not panic).
func TestAddEdgesForDependencies_MissingSourceNode_IsNoop(t *testing.T) {
	t.Parallel()
	// No assertion needed — we just verify that the function does not panic.
	builder := &Builder{}
	gr := graph.New()
	// "nonexistent-task" is intentionally NOT added to the graph.
	task := &ast.Task{
		Deps: []*ast.Dep{{Task: "other-task"}},
	}

	builder.addEdgesForDependencies("nonexistent-task", task, gr)
	// If we reach here, the defensive guard worked correctly.
}

// TestAddEdgesForCalls_MissingSourceNode_IsNoop verifies that calling
// addEdgesForCalls with a taskID that is not present in the graph is a no-op
// (defensive guard — should not happen in normal Build flow, but must not panic).
func TestAddEdgesForCalls_MissingSourceNode_IsNoop(t *testing.T) {
	t.Parallel()
	// No assertion needed — we just verify that the function does not panic.
	builder := &Builder{}
	gr := graph.New()
	// "nonexistent-task" is intentionally NOT added to the graph.
	task := &ast.Task{
		Cmds: []*ast.Cmd{{Task: "other-task"}},
	}

	builder.addEdgesForCalls("nonexistent-task", task, gr)
	// If we reach here, the defensive guard worked correctly.
}
