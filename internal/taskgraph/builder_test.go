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

// TestBuilder_Build_VarDescription_ShellCommand verifies that a global var with a shell
// command produces a "sh: ..." description on the variable node.
func TestBuilder_Build_VarDescription_ShellCommand(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	sh := "git describe --tags"
	tf := &ast.Taskfile{
		Tasks: ast.NewTasks(),
		Vars: ast.NewVars(
			&ast.VarElement{Key: "VERSION", Value: ast.Var{Sh: &sh}},
		),
	}

	builder := New(tf)
	builder.IncludeGlobalVars = true
	gr := builder.Build()

	varNode, ok := gr.Node("var:VERSION")
	g.Expect(ok).To(BeTrue())
	g.Expect(varNode.Description).To(Equal("sh: git describe --tags"))
}

// TestBuilder_Build_VarDescription_StaticValue verifies that a global var with a static
// value produces the formatted value as the node description.
func TestBuilder_Build_VarDescription_StaticValue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tf := &ast.Taskfile{
		Tasks: ast.NewTasks(),
		Vars: ast.NewVars(
			&ast.VarElement{Key: "PACKAGE", Value: ast.Var{Value: "github.com/example/project"}},
		),
	}

	builder := New(tf)
	builder.IncludeGlobalVars = true
	gr := builder.Build()

	varNode, ok := gr.Node("var:PACKAGE")
	g.Expect(ok).To(BeTrue())
	g.Expect(varNode.Description).To(Equal("github.com/example/project"))
}

// TestBuilder_Build_VarDescription_Reference verifies that a global var with a ref
// produces a "ref: ..." description on the variable node.
func TestBuilder_Build_VarDescription_Reference(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tf := &ast.Taskfile{
		Tasks: ast.NewTasks(),
		Vars: ast.NewVars(
			&ast.VarElement{Key: "ALIAS", Value: ast.Var{Ref: "OTHER_VAR"}},
		),
	}

	builder := New(tf)
	builder.IncludeGlobalVars = true
	gr := builder.Build()

	varNode, ok := gr.Node("var:ALIAS")
	g.Expect(ok).To(BeTrue())
	g.Expect(varNode.Description).To(Equal("ref: OTHER_VAR"))
}

// TestBuilder_Build_VarDescription_Empty verifies that a global var with no sh/value/ref
// produces an empty description on the variable node.
func TestBuilder_Build_VarDescription_Empty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tf := &ast.Taskfile{
		Tasks: ast.NewTasks(),
		Vars: ast.NewVars(
			&ast.VarElement{Key: "EMPTY", Value: ast.Var{}},
		),
	}

	builder := New(tf)
	builder.IncludeGlobalVars = true
	gr := builder.Build()

	varNode, ok := gr.Node("var:EMPTY")
	g.Expect(ok).To(BeTrue())
	g.Expect(varNode.Description).To(BeEmpty())
}

// TestBuilder_Build_TaskStringFieldsReferencingGlobal verifies that references to global
// vars found in task env and local vars produce var→task edges. This exercises both
// collectEnvStrings and collectVarStrings.
func TestBuilder_Build_TaskStringFieldsReferencingGlobal(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		task      *ast.Task
		globalKey string
		globalVal string
	}{
		"env var references global": {
			task: &ast.Task{
				Env: ast.NewVars(
					&ast.VarElement{Key: "OUT", Value: ast.Var{Value: "{{.OUTPUT_DIR}}/app"}},
				),
			},
			globalKey: "OUTPUT_DIR",
			globalVal: "./build",
		},
		"local var references global": {
			task: &ast.Task{
				Vars: ast.NewVars(
					&ast.VarElement{Key: "TARGET", Value: ast.Var{Value: "{{.PACKAGE}}/cmd"}},
				),
			},
			globalKey: "PACKAGE",
			globalVal: "github.com/example",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			tf := &ast.Taskfile{
				Tasks: ast.NewTasks(
					&ast.TaskElement{Key: "build", Value: tc.task},
				),
				Vars: ast.NewVars(
					&ast.VarElement{Key: tc.globalKey, Value: ast.Var{Value: tc.globalVal}},
				),
			}

			builder := New(tf)
			builder.IncludeGlobalVars = true
			gr := builder.Build()

			varNode, ok := gr.Node("var:" + tc.globalKey)
			g.Expect(ok).To(BeTrue())

			buildNode, ok := gr.Node("build")
			g.Expect(ok).To(BeTrue())

			hasEdge := hasEdgeTo(varNode.Edges(), buildNode)
			g.Expect(hasEdge).To(BeTrue(), "expected edge from var:%s to build", tc.globalKey)
		})
	}
}

// hasEdgeTo checks whether any edge in the provided list points to target.
func hasEdgeTo(edges []*graph.Edge, target *graph.Node) bool {
	for _, edge := range edges {
		if edge.To() == target {
			return true
		}
	}

	return false
}
