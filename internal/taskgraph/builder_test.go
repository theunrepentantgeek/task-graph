package taskgraph

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-task/task/v3/taskfile/ast"
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
