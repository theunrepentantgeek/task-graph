package taskgraph

import (
	"slices"

	"github.com/go-task/task/v3/taskfile/ast"

	"github.com/theunrepentantgeek/task-graph/internal/graph"
)

// Builder is responsible for building a graph.Graph from a Taskfile.
type Builder struct {
	taskfile *ast.Taskfile
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

	return g
}

func (*Builder) addEdgesForDependencies(
	taskID string,
	task *ast.Task,
	g *graph.Graph,
) {
	taskNode, ok := g.Node(taskID)
	if !ok {
		// This shouldn'thappen since we added all tasks as nodes, but we check to be safe
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
