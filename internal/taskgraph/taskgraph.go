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

	globalVars map[string]bool

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
		edge.SetClass(graph.EdgeClassDep)
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
			edge.SetClass(graph.EdgeClassCall)
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

	b.addVariableNodes(g)
	b.addVariableEdges(g)
}

func (b *Builder) addVariableNodes(g *graph.Graph) {
	for name, v := range b.taskfile.Vars.All() {
		nodeID := "var:" + name
		node := g.AddNode(nodeID)
		node.Kind = graph.NodeKindVariable
		node.Label = name
		node.Description = varDescription(v)
	}
}

func (b *Builder) addVariableEdges(g *graph.Graph) {
	for taskName, task := range b.taskfile.Tasks.All(alphaNumeric) {
		refs := b.scanTaskVarRefs(task)
		b.addEdgesForVarRefs(g, taskName, refs)
	}
}

func (*Builder) addEdgesForVarRefs(g *graph.Graph, taskName string, refs map[string]bool) {
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
		edge.SetClass(graph.EdgeClassVar)
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

	for _, s := range collectTaskStrings(task) {
		for _, name := range extractVarRefs(s) {
			if globalVarNames[name] {
				refs[name] = true
			}
		}
	}

	return refs
}

func collectTaskStrings(task *ast.Task) []string {
	var result []string

	for _, cmd := range task.Cmds {
		result = appendNonEmpty(result, cmd.Cmd, cmd.Task)
	}

	for _, dep := range task.Deps {
		result = appendNonEmpty(result, dep.Task)
	}

	result = appendNonEmpty(result, task.Dir, task.Label)
	result = collectEnvStrings(result, task.Env)
	result = collectVarStrings(result, task.Vars)

	for _, src := range task.Sources {
		result = appendNonEmpty(result, src.Glob)
	}

	for _, gen := range task.Generates {
		result = appendNonEmpty(result, gen.Glob)
	}

	result = append(result, task.Status...)

	for _, pre := range task.Preconditions {
		result = appendNonEmpty(result, pre.Sh, pre.Msg)
	}

	return result
}

// appendNonEmpty appends only non-empty strings from ss to result.
func appendNonEmpty(result []string, ss ...string) []string {
	for _, s := range ss {
		if s != "" {
			result = append(result, s)
		}
	}

	return result
}

func collectEnvStrings(result []string, env *ast.Vars) []string {
	if env == nil {
		return result
	}

	for _, v := range env.All() {
		if v.Value != nil {
			result = append(result, fmt.Sprintf("%v", v.Value))
		}
	}

	return result
}

func collectVarStrings(result []string, vars *ast.Vars) []string {
	if vars == nil {
		return result
	}

	for _, v := range vars.All() {
		if v.Value != nil {
			result = append(result, fmt.Sprintf("%v", v.Value))
		}

		if v.Sh != nil {
			result = append(result, *v.Sh)
		}
	}

	return result
}

func (b *Builder) globalVarNames() map[string]bool {
	if b.globalVars != nil {
		return b.globalVars
	}

	b.globalVars = make(map[string]bool)

	if b.taskfile.Vars != nil {
		for name := range b.taskfile.Vars.All() {
			b.globalVars[name] = true
		}
	}

	return b.globalVars
}
