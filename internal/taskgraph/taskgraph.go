package taskgraph

import (
	"fmt"
	"slices"

	"github.com/go-task/task/v3/taskfile/ast"

	"github.com/theunrepentantgeek/task-graph/internal/graph"
)

// Builder is responsible for building a graph.Graph from a Taskfile.
type Builder struct {
	taskfile *ast.Taskfile

	// IncludeGlobalVars controls whether global variables are added as nodes
	// to the graph, with edges pointing to the tasks that reference them.
	IncludeGlobalVars bool
}

func New(taskfile *ast.Taskfile) *Builder {
	return &Builder{
		taskfile: taskfile,
	}
}

// Build constructs a graph.Graph from the Taskfile, as follows.
// Each task in the Taskfile is represented as a node in the graph.
// Each task dependency is captured as a directed edge from the dependent task to the task it depends on.
// Direct task calls within task commands are also captured as edges.
func (b *Builder) Build() *graph.Graph {
	g := graph.New()

	// Create nodes for each task
	for taskName, task := range b.taskfile.Tasks.All(alphaNumeric) {
		node := g.AddNode(taskName)
		node.Description = task.Desc
	}

	// Create edges for task dependencies and calls
	for taskName, task := range b.taskfile.Tasks.All(alphaNumeric) {
		b.addEdgesForDependencies(taskName, task, g)
		b.addEdgesForCalls(taskName, task, g)
	}

	if b.IncludeGlobalVars {
		b.addGlobalVariables(g)
	}

	return g
}

func (*Builder) addEdgesForDependencies(
	taskID string,
	task *ast.Task,
	g *graph.Graph,
) {
	taskNode, ok := g.Node(taskID)
	if !ok {
		// This shouldn't happen since we added all tasks as nodes, but we check to be safe
		return
	}

	for _, dep := range task.Deps {
		toNode, ok := g.Node(dep.Task)
		if !ok {
			// If the dependency task is not defined in the Taskfile, we skip adding the edge.
			// Alternatively, we could choose to add the node for the undefined task and then add the edge.
			// We might need to do this for tasknames that include variables that are not resolved at this stage
			continue
		}

		edge := taskNode.AddEdge(toNode)
		edge.SetClass("dep")
	}
}

func (*Builder) addEdgesForCalls(
	taskID string,
	task *ast.Task,
	g *graph.Graph,
) {
	taskNode, ok := g.Node(taskID)
	if !ok {
		return
	}

	for _, cmd := range task.Cmds {
		if cmd.Task != "" {
			toNode, ok := g.Node(cmd.Task)
			if !ok {
				continue
			}

			edge := taskNode.AddEdge(toNode)
			edge.SetClass("call")
		}
	}
}

// alphaNumeric sorts the slice into alphanumeric order.
// Copied from an internal function in the tasks package.
func alphaNumeric(items []string, _ []string) []string {
	slices.Sort(items)

	return items
}

func (b *Builder) addGlobalVariables(g *graph.Graph) {
	if b.taskfile.Vars == nil {
		return
	}

	// Create variable nodes
	for name, v := range b.taskfile.Vars.All() {
		nodeID := "var:" + name
		node := g.AddNode(nodeID)
		node.Kind = graph.NodeKindVariable
		node.Label = name
		node.Description = varDescription(v)
	}

	// Scan tasks for variable references and create edges
	for taskName, task := range b.taskfile.Tasks.All(alphaNumeric) {
		refs := b.scanTaskVarRefs(task)
		for varName := range refs {
			varNodeID := "var:" + varName
			varNode, ok := g.Node(varNodeID)
			if !ok {
				continue
			}

			taskNode, ok := g.Node(taskName)
			if !ok {
				continue
			}

			edge := varNode.AddEdge(taskNode)
			edge.SetClass("var")
		}
	}
}

func varDescription(v ast.Var) string {
	if v.Sh != nil && *v.Sh != "" {
		return "sh: " + *v.Sh
	}

	if v.Value != nil {
		return fmt.Sprintf("%v", v.Value)
	}

	if v.Ref != "" {
		return "ref: " + v.Ref
	}

	return ""
}

func (b *Builder) scanTaskVarRefs(task *ast.Task) map[string]bool {
	refs := make(map[string]bool)
	globalVarNames := b.globalVarNames()

	var strings []string

	// Commands
	for _, cmd := range task.Cmds {
		strings = append(strings, cmd.Cmd, cmd.Task)
	}

	// Dependencies
	for _, dep := range task.Deps {
		strings = append(strings, dep.Task)
	}

	// Other string fields
	strings = append(strings, task.Dir, task.Label)

	// Env vars
	if task.Env != nil {
		for _, v := range task.Env.All() {
			if v.Value != nil {
				strings = append(strings, fmt.Sprintf("%v", v.Value))
			}
		}
	}

	// Task-local vars (values may reference globals)
	if task.Vars != nil {
		for _, v := range task.Vars.All() {
			if v.Value != nil {
				strings = append(strings, fmt.Sprintf("%v", v.Value))
			}
			if v.Sh != nil {
				strings = append(strings, *v.Sh)
			}
		}
	}

	// Sources
	for _, src := range task.Sources {
		strings = append(strings, src.Glob)
	}

	// Generates
	for _, gen := range task.Generates {
		strings = append(strings, gen.Glob)
	}

	// Status
	strings = append(strings, task.Status...)

	// Preconditions
	for _, pre := range task.Preconditions {
		strings = append(strings, pre.Sh, pre.Msg)
	}

	// Extract var refs from all collected strings
	for _, s := range strings {
		for _, name := range extractVarRefs(s) {
			if globalVarNames[name] {
				refs[name] = true
			}
		}
	}

	return refs
}

func (b *Builder) globalVarNames() map[string]bool {
	names := make(map[string]bool)
	if b.taskfile.Vars != nil {
		for name := range b.taskfile.Vars.All() {
			names[name] = true
		}
	}

	return names
}
