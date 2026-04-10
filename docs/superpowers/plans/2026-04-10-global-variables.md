# Global Variables Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `--include-global-vars` support that renders Taskfile global variables as graph nodes with edges to consuming tasks.

**Architecture:** Add `Kind` field to `graph.Node` to distinguish variable nodes from task nodes. The `taskgraph.Builder` scans task template fields for `{{.VAR}}` references and creates variable→task edges. Both Graphviz and Mermaid renderers filter by node kind for styling and layout.

**Tech Stack:** Go, go-task/ast, gomega, goldie (golden files)

---

## File Structure

| Action | File                                                            | Responsibility                                                         |
| ------ | --------------------------------------------------------------- | ---------------------------------------------------------------------- |
| Modify | `internal/graph/node.go`                                        | Add `NodeKind` type and `Kind` field to `Node`                         |
| Modify | `internal/graph/node_test.go`                                   | Tests for `Kind` field                                                 |
| Create | `internal/taskgraph/scanner.go`                                 | Template reference scanner (extract var names from strings)            |
| Create | `internal/taskgraph/scanner_test.go`                            | Table-driven tests for scanner                                         |
| Modify | `internal/taskgraph/taskgraph.go`                               | Add `IncludeGlobalVars` option to Builder, variable node/edge creation |
| Modify | `internal/taskgraph/testgraph_test.go`                          | Golden file test for builder with variables                            |
| Create | `internal/taskgraph/testdata/global-vars-taskfile.yml`          | Test taskfile with global variables                                    |
| Create | `internal/taskgraph/testdata/global-vars-taskfile.golden`       | Golden file for builder output with variables                          |
| Modify | `internal/config/config.go`                                     | Add `IncludeGlobalVars` field and defaults for variable styling        |
| Modify | `internal/config/graphviz.go`                                   | Add `VariableNodes` and `VariableEdges` fields                         |
| Modify | `internal/config/mermaid.go`                                    | Add `VariableNodes` and `VariableEdges` fields                         |
| Modify | `internal/cmd/cli.go`                                           | Add `--include-global-vars` flag, wire through to builder              |
| Modify | `internal/cmd/cli_test.go`                                      | Test CLI flag                                                          |
| Modify | `internal/graphviz/graphviz.go`                                 | Variable node rendering, `rank=sink` layout, variable edge styling     |
| Modify | `internal/graphviz/graphviz_test.go`                            | Golden file tests for variable nodes                                   |
| Create | `internal/graphviz/testdata/sample_graph_with_variables.golden` | Golden file                                                            |
| Modify | `internal/mermaid/mermaid.go`                                   | Variable node rendering, stadium shape, thick edges                    |
| Modify | `internal/mermaid/mermaid_test.go`                              | Golden file tests for variable nodes                                   |
| Create | `internal/mermaid/testdata/sample_graph_with_variables.golden`  | Golden file                                                            |

---

### Task 1: Add NodeKind to Graph Model

**Files:**
- Modify: `internal/graph/node.go`
- Modify: `internal/graph/node_test.go`

- [ ] **Step 1: Write failing tests for NodeKind**

Add these tests to `internal/graph/node_test.go`:

```go
func TestNode_Kind_DefaultsToTask(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	node := NewNode("test")

	// Assert
	g.Expect(node.Kind).To(gomega.Equal(NodeKindTask))
}

func TestNode_Kind_CanBeSetToVariable(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	node := NewNode("var:FOO")

	// Act
	node.Kind = NodeKindVariable

	// Assert
	g.Expect(node.Kind).To(gomega.Equal(NodeKindVariable))
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/graph/ -run "TestNode_Kind" -v`
Expected: FAIL — `NodeKindTask`, `NodeKindVariable`, `Kind` are undefined.

- [ ] **Step 3: Implement NodeKind type and Kind field**

Add to `internal/graph/node.go`, before the `Node` struct:

```go
// NodeKind represents the type of a node in the graph.
type NodeKind string

const (
	// NodeKindTask represents a task node.
	NodeKindTask NodeKind = "task"

	// NodeKindVariable represents a global variable node.
	NodeKindVariable NodeKind = "variable"
)
```

Add a `Kind` field to the `Node` struct:

```go
type Node struct {
	NodeID

	// Kind identifies the type of this node (task or variable).
	Kind NodeKind

	// Label returns the label of the node.
	Label string

	// Description returns the description of the node.
	Description string

	// Edges holds the outgoing edges from this node to other nodes in the graph.
	edges []*Edge
}
```

Update `NewNode` to default `Kind` to `NodeKindTask`:

```go
func NewNode(id string) *Node {
	return &Node{
		NodeID: NodeID{id: id},
		Kind:   NodeKindTask,
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/graph/ -v`
Expected: All tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/graph/node.go internal/graph/node_test.go
git commit -m "feat: add NodeKind type and Kind field to graph.Node"
```

---

### Task 2: Template Reference Scanner

**Files:**
- Create: `internal/taskgraph/scanner.go`
- Create: `internal/taskgraph/scanner_test.go`

- [ ] **Step 1: Write failing tests for the scanner**

Create `internal/taskgraph/scanner_test.go`:

```go
package taskgraph

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestExtractVarRefs_SimpleReference_ReturnsVarName(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("{{.FOO}}")

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("FOO"))
}

func TestExtractVarRefs_PipedExpression_ReturnsVarName(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("{{.FOO | lowercase}}")

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("FOO"))
}

func TestExtractVarRefs_MultipleVariables_ReturnsAllVarNames(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs(`{{printf "%s/%s" .FOO .BAR}}`)

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("FOO", "BAR"))
}

func TestExtractVarRefs_Conditional_ReturnsVarName(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("{{if .FOO}}yes{{end}}")

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("FOO"))
}

func TestExtractVarRefs_PlainString_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("just a plain string")

	// Assert
	g.Expect(refs).To(gomega.BeEmpty())
}

func TestExtractVarRefs_EmptyString_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("")

	// Assert
	g.Expect(refs).To(gomega.BeEmpty())
}

func TestExtractVarRefs_CurrentContext_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("{{.}}")

	// Assert
	g.Expect(refs).To(gomega.BeEmpty())
}

func TestExtractVarRefs_MultipleTemplateBlocks_ReturnsAllVarNames(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("go build -ldflags {{.LDFLAGS}} -o {{.OUTPUT}}")

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("LDFLAGS", "OUTPUT"))
}

func TestExtractVarRefs_DuplicateReferences_ReturnsUniqueNames(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("{{.FOO}} and {{.FOO}}")

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("FOO"))
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/taskgraph/ -run "TestExtractVarRefs" -v`
Expected: FAIL — `extractVarRefs` is undefined.

- [ ] **Step 3: Implement the scanner**

Create `internal/taskgraph/scanner.go`:

```go
package taskgraph

import (
	"regexp"
)

// templateBlockRe matches Go template blocks: {{ ... }}
var templateBlockRe = regexp.MustCompile(`\{\{(.+?)\}\}`)

// varRefRe matches variable references within a template block: .IDENTIFIER
// It requires at least one character after the dot to avoid matching {{.}} (current context).
var varRefRe = regexp.MustCompile(`\.([A-Za-z_][A-Za-z0-9_]*)`)

// extractVarRefs extracts unique variable names referenced in Go template expressions
// within the given string. It finds all {{ ... }} blocks and extracts .IDENTIFIER
// patterns from each block.
func extractVarRefs(s string) []string {
	seen := make(map[string]bool)

	for _, blockMatch := range templateBlockRe.FindAllStringSubmatch(s, -1) {
		block := blockMatch[1]
		for _, varMatch := range varRefRe.FindAllStringSubmatch(block, -1) {
			name := varMatch[1]
			seen[name] = true
		}
	}

	result := make([]string, 0, len(seen))
	for name := range seen {
		result = append(result, name)
	}

	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/taskgraph/ -run "TestExtractVarRefs" -v`
Expected: All PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/taskgraph/scanner.go internal/taskgraph/scanner_test.go
git commit -m "feat: add template reference scanner for variable extraction"
```

---

### Task 3: Configuration Changes

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/graphviz.go`
- Modify: `internal/config/mermaid.go`
- Modify: `internal/cmd/cli.go`
- Modify: `internal/cmd/cli_test.go`

- [ ] **Step 1: Write failing test for IncludeGlobalVars CLI flag**

Add to `internal/cmd/cli_test.go`:

```go
func TestCreateConfig_IncludeGlobalVarsFlagSetsConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{IncludeGlobalVars: true}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.IncludeGlobalVars).To(BeTrue())
}

func TestCreateConfig_DefaultIncludeGlobalVarsIsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.IncludeGlobalVars).To(BeFalse())
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/cmd/ -run "TestCreateConfig_IncludeGlobalVars" -v`
Expected: FAIL — `IncludeGlobalVars` field not found.

- [ ] **Step 3: Add IncludeGlobalVars to Config**

In `internal/config/config.go`, add the field to `Config` struct after `AutoColor`:

```go
	// IncludeGlobalVars controls whether global Taskfile variables are included
	// as nodes in the generated graph, with edges to the tasks that reference them.
	IncludeGlobalVars bool `json:"includeGlobalVars,omitempty" yaml:"includeGlobalVars,omitempty"`
```

- [ ] **Step 4: Add VariableNodes and VariableEdges to Graphviz config**

In `internal/config/graphviz.go`, add to the `Graphviz` struct:

```go
	// VariableNodes is the presentation for global variable nodes
	VariableNodes *GraphvizNode `json:"variableNodes,omitempty" yaml:"variableNodes,omitempty"`

	// VariableEdges is the presentation for edges from variables to tasks
	VariableEdges *GraphvizEdge `json:"variableEdges,omitempty" yaml:"variableEdges,omitempty"`
```

- [ ] **Step 5: Add VariableNodes to Mermaid config**

In `internal/config/mermaid.go`, add to the `Mermaid` struct:

```go
	// VariableNodes holds style properties for variable nodes in the Mermaid output.
	VariableNodes *MermaidStyle `json:"variableNodes,omitempty" yaml:"variableNodes,omitempty"`

	// VariableEdges holds style properties for variable edges in the Mermaid output.
	VariableEdges *MermaidStyle `json:"variableEdges,omitempty" yaml:"variableEdges,omitempty"`
```

Also add the `MermaidStyle` type to `internal/config/mermaid.go`:

```go
// MermaidStyle holds CSS-like style properties for Mermaid classDef directives.
type MermaidStyle struct {
	// Fill is the background fill color.
	Fill string `json:"fill,omitempty" yaml:"fill,omitempty"`

	// Stroke is the border/line color.
	Stroke string `json:"stroke,omitempty" yaml:"stroke,omitempty"`

	// Color is the text color.
	Color string `json:"color,omitempty" yaml:"color,omitempty"`
}
```

- [ ] **Step 6: Add defaults in config.New()**

In `internal/config/config.go`, update `New()` to include defaults for variable styling. Add to the `Graphviz` initializer:

```go
			VariableNodes: &GraphvizNode{
				Color:     "#666666",
				FillColor: "#e8e8e8",
				Style:     "filled",
			},
			VariableEdges: &GraphvizEdge{
				Color: "green",
				Width: 1,
				Style: "dotted",
			},
```

- [ ] **Step 7: Add CLI flag and override**

In `internal/cmd/cli.go`, add the field to the `CLI` struct:

```go
	//nolint:revive // Intentionally long name for clarity in the CLI help.
	IncludeGlobalVars bool `help:"Include global variables as nodes in the graph, with edges to consuming tasks." long:"include-global-vars"`
```

Add to `applyConfigOverrides`:

```go
	if c.IncludeGlobalVars {
		cfg.IncludeGlobalVars = true
	}
```

- [ ] **Step 8: Run tests to verify they pass**

Run: `go test ./internal/cmd/ -v`
Expected: All tests PASS.

- [ ] **Step 9: Commit**

```bash
git add internal/config/config.go internal/config/graphviz.go internal/config/mermaid.go internal/cmd/cli.go internal/cmd/cli_test.go
git commit -m "feat: add IncludeGlobalVars config, CLI flag, and variable styling config"
```

---

### Task 4: Builder — Variable Nodes and Edges

**Files:**
- Modify: `internal/taskgraph/taskgraph.go`
- Create: `internal/taskgraph/testdata/global-vars-taskfile.yml`
- Create: `internal/taskgraph/testdata/global-vars-taskfile.golden` (via -update)
- Modify: `internal/taskgraph/testgraph_test.go`

- [ ] **Step 1: Create test taskfile with global variables**

Create `internal/taskgraph/testdata/global-vars-taskfile.yml`:

```yaml
version: '3'

vars:
  PACKAGE: github.com/example/project
  VERSION:
    sh: git describe --tags
  OUTPUT_DIR: ./build

tasks:
  build:
    desc: Build the project
    cmds:
      - go build -ldflags "{{.PACKAGE}}" -o {{.OUTPUT_DIR}}/app

  test:
    desc: Run tests
    cmds:
      - go test ./...

  release:
    desc: Create a release
    deps: [build, test]
    cmds:
      - echo "Releasing {{.VERSION}}"
```

- [ ] **Step 2: Write failing test**

Add a new test case to `internal/taskgraph/testgraph_test.go`. First, add the `"github.com/theunrepentantgeek/task-graph/internal/graph"` import, then add this test function:

```go
func TestTaskGraphBuilder_WithGlobalVars_Graphviz(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	taskfilePath := filepath.Join("testdata", "global-vars-taskfile.yml")

	tf, err := loader.Load(t.Context(), taskfilePath)
	g.Expect(err).NotTo(HaveOccurred())

	builder := New(tf)
	builder.IncludeGlobalVars = true
	gr := builder.Build()

	// Verify variable nodes exist
	varNode, ok := gr.Node("var:PACKAGE")
	g.Expect(ok).To(BeTrue())
	g.Expect(varNode.Kind).To(Equal(graph.NodeKindVariable))

	buf := bytes.Buffer{}
	cfg := config.New()

	err = graphviz.WriteTo(&buf, gr, cfg)
	g.Expect(err).NotTo(HaveOccurred())

	gg := goldie.New(t)
	g.Expect(gg.WithFixtureDir("testdata")).To(Succeed())

	gg.Assert(t, "global-vars-taskfile", buf.Bytes())
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/taskgraph/ -run "TestTaskGraphBuilder_WithGlobalVars" -v`
Expected: FAIL — `IncludeGlobalVars` field not found on Builder.

- [ ] **Step 4: Implement Builder changes**

Modify `internal/taskgraph/taskgraph.go`. Add `IncludeGlobalVars` field to `Builder`:

```go
type Builder struct {
	taskfile *ast.Taskfile

	// IncludeGlobalVars controls whether global variables are added as nodes
	// to the graph, with edges pointing to the tasks that reference them.
	IncludeGlobalVars bool
}
```

Add variable building to `Build()` method — append after the existing edge creation loop:

```go
	if b.IncludeGlobalVars {
		b.addGlobalVariables(g)
	}
```

Add new methods:

```go
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

	// Collect all strings from the task that might contain template references
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
```

Add `"fmt"` to the imports in `taskgraph.go`.

- [ ] **Step 5: Run test to verify it compiles and the golden file needs generating**

Run: `go test ./internal/taskgraph/ -run "TestTaskGraphBuilder_WithGlobalVars" -v`
Expected: FAIL — golden file doesn't exist yet.

- [ ] **Step 6: Generate golden file**

Run: `go test ./internal/taskgraph/ -run "TestTaskGraphBuilder_WithGlobalVars" -update`

- [ ] **Step 7: Inspect the golden file**

Read the generated `internal/taskgraph/testdata/global-vars-taskfile.golden` and verify:
- Variable nodes appear with `shape="record"` (they won't yet — this comes in Task 5; for now they'll have `Mrecord` like task nodes)
- Variable→task edges exist with the correct connections
- All three variables (`PACKAGE`, `VERSION`, `OUTPUT_DIR`) appear

- [ ] **Step 8: Verify existing tests still pass**

Run: `go test ./internal/taskgraph/ -v`
Expected: All PASS — existing golden files unchanged because existing tests don't set `IncludeGlobalVars`.

- [ ] **Step 9: Commit**

```bash
git add internal/taskgraph/taskgraph.go internal/taskgraph/testgraph_test.go internal/taskgraph/testdata/global-vars-taskfile.yml internal/taskgraph/testdata/global-vars-taskfile.golden
git commit -m "feat: builder creates variable nodes and edges from template references"
```

---

### Task 5: Graphviz Renderer — Variable Node Support

**Files:**
- Modify: `internal/graphviz/graphviz.go`
- Modify: `internal/graphviz/graphviz_test.go`
- Create: `internal/graphviz/testdata/sample_graph_with_variables.golden` (via -update)

- [ ] **Step 1: Write failing golden file test**

Add to `internal/graphviz/graphviz_test.go`:

```go
func TestWriteTo_WithVariableNodes_WritesVariableNodesWithRankSink(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildGraphWithVariables(t)

	cfg := config.New()
	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	gg := goldie.New(t)
	g.Expect(gg.WithFixtureDir("testdata")).To(gomega.Succeed())

	gg.Assert(t, "sample_graph_with_variables", buf.Bytes())
}
```

Add the helper function:

```go
func buildGraphWithVariables(t *testing.T) *graph.Graph {
	t.Helper()

	gr := graph.New()

	// Task nodes
	build := gr.AddNode("build")
	build.Description = "Build the project"
	test := gr.AddNode("test")

	build.AddEdge(test).SetClass("dep")

	// Variable nodes
	pkg := gr.AddNode("var:PACKAGE")
	pkg.Kind = graph.NodeKindVariable
	pkg.Label = "PACKAGE"
	pkg.Description = "github.com/example/project"

	ver := gr.AddNode("var:VERSION")
	ver.Kind = graph.NodeKindVariable
	ver.Label = "VERSION"
	ver.Description = "sh: git describe --tags"

	// Variable edges
	pkgEdge := pkg.AddEdge(build)
	pkgEdge.SetClass("var")

	verEdge := ver.AddEdge(build)
	verEdge.SetClass("var")

	return gr
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/graphviz/ -run "TestWriteTo_WithVariableNodes" -v`
Expected: FAIL — golden file doesn't exist.

- [ ] **Step 3: Implement variable node rendering in Graphviz**

Modify `internal/graphviz/graphviz.go`.

**Update `WriteTo`** to separate task and variable nodes, and add a `rank=sink` block for variables. Replace the current body starting at the `nodes` sort through to `iw.Add("}")`:

In `WriteTo`, after collecting and sorting `nodes`, split them:

```go
	var taskNodes []*graph.Node
	var varNodes []*graph.Node

	for _, n := range nodes {
		if n.Kind == graph.NodeKindVariable {
			varNodes = append(varNodes, n)
		} else {
			taskNodes = append(taskNodes, n)
		}
	}
```

Update the `nodeIDs` collection to use all nodes (keep existing logic), then change `writeAllNodesTo` to use `taskNodes`:

```go
	err := writeAllNodesTo(root, taskNodes, cfg, reg)
	if err != nil {
		return err
	}

	if len(varNodes) > 0 {
		err = writeVariableNodesTo(root, varNodes, cfg, reg)
		if err != nil {
			return err
		}
	}
```

Add the `writeVariableNodesTo` function:

```go
func writeVariableNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	// Write variable node definitions
	for _, node := range nodes {
		err := writeVariableNodeDefinitionTo(root, node, cfg, reg)
		if err != nil {
			return err
		}

		for _, edge := range node.Edges() {
			writeEdgeTo(root, edge, cfg, reg)
		}

		root.Add("")
	}

	// Add rank=sink to force variables to bottom
	sink := root.Add("{ rank=sink")
	for _, node := range nodes {
		sink.Addf("\"%s\"", reg.ID(node.ID()))
	}
	root.Add("}")

	return nil
}

func writeVariableNodeDefinitionTo(
	root *indentwriter.Line,
	node *graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) error {
	margin := min((len(node.Description)+20)/2, 40)

	rec := newRecord()
	rec.add(nodeLabel(node))
	rec.addWrapped(margin, node.Description)

	props := newNodeProperties()
	props.Addf("shape", "record")
	props.Add("label", rec.String())

	if cfg != nil && cfg.Graphviz != nil {
		props.AddAttributes(cfg.Graphviz.VariableNodes)
	}

	for _, rule := range cfg.NodeStyleRules {
		err := props.AddStyleRuleAttributes(node.ID(), rule)
		if err != nil {
			return err
		}
	}

	if props.ContainsKey("fillcolor") && !props.ContainsKey("style") {
		props.Add("style", "filled")
	}

	id := fmt.Sprintf("\"%s\"", reg.ID(node.ID()))
	props.WriteTo(id, root)

	return nil
}
```

**Update `writeEdgeTo`** to handle the `"var"` edge class. Add a case in the `switch` block:

```go
		case "var":
			props.AddAttributes(cfg.Graphviz.VariableEdges)
```

**Update `writeAllNodesTo` and `writeGroupedNodesTo`:** These already receive only task nodes from the split above, so they need no changes.

- [ ] **Step 4: Generate golden file**

Run: `go test ./internal/graphviz/ -run "TestWriteTo_WithVariableNodes" -update`

- [ ] **Step 5: Inspect the golden file**

Read `internal/graphviz/testdata/sample_graph_with_variables.golden` and verify:
- Task nodes use `shape="Mrecord"`
- Variable nodes use `shape="record"` with `fillcolor="#e8e8e8"` and `style="filled"`
- Variable edges have `color="green"`, `style="dotted"`
- A `{ rank=sink ... }` block contains the variable node IDs
- Variables appear after task nodes

- [ ] **Step 6: Run all graphviz tests**

Run: `go test ./internal/graphviz/ -v`
Expected: All tests PASS, existing golden files unchanged.

- [ ] **Step 7: Commit**

```bash
git add internal/graphviz/graphviz.go internal/graphviz/graphviz_test.go internal/graphviz/testdata/sample_graph_with_variables.golden
git commit -m "feat: graphviz renderer supports variable nodes with rank=sink layout"
```

---

### Task 6: Mermaid Renderer — Variable Node Support

**Files:**
- Modify: `internal/mermaid/mermaid.go`
- Modify: `internal/mermaid/mermaid_test.go`
- Create: `internal/mermaid/testdata/sample_graph_with_variables.golden` (via -update)

- [ ] **Step 1: Write failing golden file test**

Add to `internal/mermaid/mermaid_test.go`:

```go
func TestWriteTo_WithVariableNodes_WritesVariableNodesAfterTasks(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	buf := bytes.Buffer{}
	gr := buildGraphWithVariables(t)

	cfg := config.New()
	err := WriteTo(&buf, gr, cfg)

	g.Expect(err).NotTo(gomega.HaveOccurred())

	gg := goldie.New(t)
	g.Expect(gg.WithFixtureDir("testdata")).To(gomega.Succeed())

	gg.Assert(t, "sample_graph_with_variables", buf.Bytes())
}
```

Add the helper (same graph structure as the graphviz test):

```go
func buildGraphWithVariables(t *testing.T) *graph.Graph {
	t.Helper()

	gr := graph.New()

	// Task nodes
	build := gr.AddNode("build")
	build.Description = "Build the project"
	test := gr.AddNode("test")

	build.AddEdge(test).SetClass("dep")

	// Variable nodes
	pkg := gr.AddNode("var:PACKAGE")
	pkg.Kind = graph.NodeKindVariable
	pkg.Label = "PACKAGE"
	pkg.Description = "github.com/example/project"

	ver := gr.AddNode("var:VERSION")
	ver.Kind = graph.NodeKindVariable
	ver.Label = "VERSION"
	ver.Description = "sh: git describe --tags"

	// Variable edges (variable -> task in graph model)
	pkgEdge := pkg.AddEdge(build)
	pkgEdge.SetClass("var")

	verEdge := ver.AddEdge(build)
	verEdge.SetClass("var")

	return gr
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/mermaid/ -run "TestWriteTo_WithVariableNodes" -v`
Expected: FAIL — golden file doesn't exist.

- [ ] **Step 3: Implement variable node rendering in Mermaid**

Modify `internal/mermaid/mermaid.go`.

**Update `WriteTo`** to split task and variable nodes. After the existing `nodes` collection and sort:

```go
	var taskNodes []*graph.Node
	var varNodes []*graph.Node

	for _, n := range nodes {
		if n.Kind == graph.NodeKindVariable {
			varNodes = append(varNodes, n)
		} else {
			taskNodes = append(taskNodes, n)
		}
	}
```

Replace the node writing block to use `taskNodes`:

```go
	if cfg != nil && cfg.GroupByNamespace {
		writeGroupedNodesTo(root, taskNodes, reg)
	} else {
		writeNodesTo(root, taskNodes, reg)
	}

	if len(varNodes) > 0 {
		writeVariableNodesTo(root, varNodes, reg)
	}
```

Add new functions:

```go
func writeVariableNodesTo(
	root *indentwriter.Line,
	nodes []*graph.Node,
	reg *safe.Registry,
) {
	for _, node := range nodes {
		writeVariableNodeDefinitionTo(root, node, reg)
		writeVariableEdgesTo(root, node, reg)
		root.Add("")
	}
}

func writeVariableNodeDefinitionTo(
	root *indentwriter.Line,
	node *graph.Node,
	reg *safe.Registry,
) {
	label := safe.Label(variableDisplayLabel(node))
	root.Addf("%s(\"%s\")", reg.ID(node.ID()), label)
}

func writeVariableEdgesTo(
	root *indentwriter.Line,
	node *graph.Node,
	reg *safe.Registry,
) {
	for _, edge := range node.Edges() {
		// Reverse edge direction visually: write as task ==> variable
		// so Mermaid's layout pushes variables below tasks
		from := reg.ID(edge.To().ID())
		to := reg.ID(edge.From().ID())
		root.Addf("%s ==> %s", from, to)
	}
}

func variableDisplayLabel(node *graph.Node) string {
	label := nodeDisplayLabel(node)
	if node.Description != "" {
		return label + ": " + node.Description
	}

	return label
}
```

**Update `writeStyleRulesTo`** call to pass all nodes (not just taskNodes) so style rules can target variable nodes. In the `WriteTo` function, the call to `writeStyleRulesTo` should use `nodes` (all nodes):

```go
	err := writeStyleRulesTo(root, nodes, cfg, reg)
```

Add a default `classDef` for variable nodes. In `WriteTo`, after the variable nodes are written and before `writeStyleRulesTo`:

```go
	if len(varNodes) > 0 {
		writeVariableClassDef(root, varNodes, cfg, reg)
	}
```

```go
func writeVariableClassDef(
	root *indentwriter.Line,
	nodes []*graph.Node,
	cfg *config.Config,
	reg *safe.Registry,
) {
	var parts []string

	if cfg != nil && cfg.Mermaid != nil && cfg.Mermaid.VariableNodes != nil {
		vs := cfg.Mermaid.VariableNodes
		if vs.Fill != "" {
			parts = append(parts, "fill:"+vs.Fill)
		}
		if vs.Stroke != "" {
			parts = append(parts, "stroke:"+vs.Stroke)
		}
		if vs.Color != "" {
			parts = append(parts, "color:"+vs.Color)
		}
	}

	if len(parts) == 0 {
		// Default variable styling
		parts = append(parts, "fill:#e8e8e8", "stroke:#666")
	}

	classDef := strings.Join(parts, ",")

	var ids []string
	for _, n := range nodes {
		ids = append(ids, reg.ID(n.ID()))
	}

	sort.Strings(ids)
	root.Addf("classDef varStyle %s", classDef)
	root.Addf("class %s varStyle", strings.Join(ids, ","))
}
```

- [ ] **Step 4: Generate golden file**

Run: `go test ./internal/mermaid/ -run "TestWriteTo_WithVariableNodes" -update`

- [ ] **Step 5: Inspect the golden file**

Read `internal/mermaid/testdata/sample_graph_with_variables.golden` and verify:
- Task nodes use `["..."]` (rectangle)
- Variable nodes use `("...")` (stadium)
- Variable edges use `==>` (thick arrow), written as `task ==> variable` (reversed)
- A `classDef varStyle` and `class` directive appear
- Variables come after task nodes

- [ ] **Step 6: Run all mermaid tests**

Run: `go test ./internal/mermaid/ -v`
Expected: All tests PASS, existing golden files unchanged.

- [ ] **Step 7: Commit**

```bash
git add internal/mermaid/mermaid.go internal/mermaid/mermaid_test.go internal/mermaid/testdata/sample_graph_with_variables.golden
git commit -m "feat: mermaid renderer supports variable nodes with stadium shapes"
```

---

### Task 7: Wire Builder to CLI

**Files:**
- Modify: `internal/cmd/cli.go`

- [ ] **Step 1: Update CLI.Run to pass IncludeGlobalVars to Builder**

In `internal/cmd/cli.go`, change the builder creation in `Run()` from:

```go
	gr := taskgraph.New(tf).Build()
```

to:

```go
	builder := taskgraph.New(tf)
	builder.IncludeGlobalVars = flags.Config.IncludeGlobalVars
	gr := builder.Build()
```

- [ ] **Step 2: Run all tests**

Run: `go test ./...`
Expected: All tests PASS.

- [ ] **Step 3: Manual smoke test**

Run:
```bash
go build -o build/task-graph && ./build/task-graph --output /tmp/test.dot --include-global-vars samples/aso-taskfile.yml
cat /tmp/test.dot | head -20
cat /tmp/test.dot | grep "rank=sink" 
cat /tmp/test.dot | grep "var:"
```

Expected: Variable nodes appear with `var:` prefix, `rank=sink` block present, variable edges visible.

- [ ] **Step 4: Commit**

```bash
git add internal/cmd/cli.go
git commit -m "feat: wire --include-global-vars flag to builder"
```

---

### Task 8: Update Existing Golden Files and Final Verification

**Files:**
- Possibly update: `internal/taskgraph/testdata/*.golden` (only if existing tests broke)

- [ ] **Step 1: Run full test suite**

Run: `go test ./...`
Expected: All PASS. If any golden files broke, investigate — existing tests should be unaffected since `IncludeGlobalVars` defaults to `false`.

- [ ] **Step 2: Run linter**

Run: `task lint`
Expected: No errors. Fix any lint issues.

- [ ] **Step 3: Run full CI**

Run: `task ci`
Expected: All PASS.

- [ ] **Step 4: Final commit if any fixes were needed**

```bash
git add -A
git commit -m "chore: fix lint issues and update golden files"
```
