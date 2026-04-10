# Global Variables in Task Graph

**Date:** 2026-04-10
**Status:** Design
**Branch:** `feature/global-variables`

## Summary

Add a `--include-global-vars` CLI flag and matching `includeGlobalVars` configuration option that includes global Taskfile variables in the generated graph as a distinct node type. Variable nodes appear below task nodes, with edges linking each variable to the tasks that consume it. Each variable node displays both its name and value (or shell expression for dynamic variables).

## Requirements

- Global variables from `Taskfile.Vars` are rendered as graph nodes when `--include-global-vars` is specified.
- Each variable node shows the variable name and its value. For shell variables (`sh:`), the shell command is displayed.
- Variable-to-task edges are created by scanning task template references (`{{.VAR_NAME}}`), including piped expressions like `{{.FOO | lowercase}}` and multi-variable expressions like `{{printf "%s" .FOO .BAR}}`.
- Variable nodes appear below task nodes in the rendered graph.
- Variable nodes and edges have distinct default styling, configurable independently from task nodes.
- The feature is opt-in; disabled by default.
- Both Graphviz (dot) and Mermaid output are supported.
- Task-local variables are out of scope.

## Approach

Graph-level variable nodes: add variable nodes directly to the existing `graph.Graph` with a new node `Kind` field to distinguish them from task nodes. The `taskgraph.Builder` scans task fields for template references and creates edges from variable nodes to consuming tasks. Both renderers check node kind and apply appropriate styling and layout.

## Design

### Graph Model

The `graph.Node` struct gains a `Kind` field:

```go
type NodeKind string

const (
    NodeKindTask     NodeKind = "task"
    NodeKindVariable NodeKind = "variable"
)
```

- `Kind` defaults to `NodeKindTask` for backwards compatibility.
- Variable nodes use `Label` for the variable name and `Description` for the value or shell expression.
- For shell variables, the description is the shell command prefixed with `sh:` (e.g. `sh: scripts/build_version.py v2`).

Edge `Class` gains a new value `"var"` for variable→task edges, alongside the existing `"dep"` and `"call"`.

Variable node IDs are prefixed with `var:` to distinguish them from task IDs (e.g. `var:PACKAGE`). This enables targeting them with `NodeStyleRules` patterns like `var:*`.

### Template Reference Scanning

The `taskgraph.Builder` scans task fields to detect which tasks reference which global variables:

1. Iterate over each task's template-capable string fields: `Cmds` (shell commands and task call names), `Dir`, `Env`, `Vars` (task-local values), `Label`, `Sources`, `Generates`, `Status`, `Preconditions`, and `Deps` (task names).
2. Find all `{{ ... }}` template blocks in each field.
3. Within each block, extract every `.IDENTIFIER` occurrence using the pattern `\.([A-Za-z_][A-Za-z0-9_]*)`.
4. Match extracted names against the known global variable keys from `Taskfile.Vars`.
5. Create edges from the variable node to each consuming task node, with class `"var"`.

**Edge direction:** Variable → task. This matches data flow (variables feed into tasks) and produces upward-pointing arrows when variables are laid out below tasks.

**Known limitations:**
- Variable-to-variable template references (within `sh:` expressions) are not scanned.
- Variables referenced via `env:` at the task level or through included taskfile variable forwarding are not detected.
- Template expressions using Go template functions that indirectly reference variables are not detected.

These are acceptable simplifications; the scanner catches the most common usage pattern (direct `{{.VAR}}` references in task fields).

### CLI and Configuration

**CLI flag:** `--include-global-vars` (bool) on the `CLI` struct, following the pattern of `--group-by-namespace` and `--auto-color`.

**Config field:** `IncludeGlobalVars bool` on the `Config` struct with JSON/YAML tags. Three-layer override: defaults (`false`) → config file → CLI flag.

**Graphviz config additions** on the `Graphviz` struct:
- `VariableNodes *GraphvizNode` — default: `box` shape, fill `#e8e8e8` (light gray), style `filled`.
- `VariableEdges *GraphvizEdge` — default: color `green`, style `dotted`, width `1`.

**Mermaid config additions** on the `Mermaid` struct:
- `VariableNodes` — styling for the variable `classDef` (fill, stroke, etc.).
- `VariableEdges` — edge styling (limited by Mermaid; mainly connector style).

### Graphviz Rendering

**Node shape:** Variable nodes use `box` shape instead of `Mrecord`. The record label shows `name | value`, e.g.:
- `"PACKAGE | github.com/Azure/azure-service-operator/v2"`
- `"VERSION | sh: scripts/build_version.py v2"`

Long values are wrapped using the existing `indentwriter.Wrap` logic.

**Layout:** A `{ rank=sink; var1; var2; ... }` block forces all variable nodes to the bottom of the graph. Combined with variable→task edge direction, arrows naturally point upward.

**Edge rendering:** Variable edges use the `"var"` class. The renderer applies `cfg.Graphviz.VariableEdges` styling.

**Namespace grouping:** Variable nodes are global, not namespaced. When `--group-by-namespace` is active, variable nodes render outside all namespace subgraphs, in the `rank=sink` block.

**Style rules:** Existing `NodeStyleRules` apply to variable nodes via the `var:` prefixed IDs. Users can target them with patterns like `var:*` or `var:VERSION`.

### Mermaid Rendering

**Node shape:** Stadium shape `(" ")` for variable nodes, visually distinct from task rectangles `[" "]`.

**Node label:** Shows name and value: `var_PACKAGE("PACKAGE: github.com/Azure/...v2")`. For shell variables: `var_VERSION("VERSION: sh: scripts/build_version.py v2")`.

**Layout:** In Mermaid `flowchart TD`, variable node definitions are written after all task nodes. Edges go from task to variable (task → variable, reversed from Graphviz) so that Mermaid's layout algorithm pushes variables below tasks. Arrows point downward — acceptable tradeoff per requirements since positioning below is the priority.

**Edge style:** Variable edges use `==>` (thick arrow), distinct from `-->` (dependency) and `-.->` (call). A `classDef` applies the configured color (default green).

**Style rules:** Variable nodes get their own `classDef` block. `NodeStyleRules` pattern matching applies using the `var:` prefixed IDs.

### Builder Changes

The `taskgraph.Builder.Build()` method gains an optional `IncludeGlobalVars` parameter (or reads it from a config/options struct passed at construction). When enabled:

1. After creating task nodes and edges (existing logic), iterate over `Taskfile.Vars.All()`.
2. For each global variable, create a node with `Kind = NodeKindVariable`, ID `var:<name>`, label = variable name, description = value or `sh: <command>`.
3. For each task, scan its template-capable fields for references to global variables.
4. Create edges from variable nodes to referencing task nodes with class `"var"`.

When `IncludeGlobalVars` is false (default), the builder behaves exactly as today.

## Testing

All tests follow existing conventions: Roy Osherove naming, gomega assertions, `t.Parallel()`, table-driven where appropriate, `// Arrange / Act / Assert` comments.

### Template Scanner

Table-driven tests covering:
- Simple reference: `{{.FOO}}` → extracts `FOO`
- Piped expression: `{{.FOO | lowercase}}` → extracts `FOO`
- Multiple variables: `{{printf "%s" .FOO .BAR}}` → extracts `FOO`, `BAR`
- Conditionals: `{{if .FOO}}...{{end}}` → extracts `FOO`
- No matches: plain strings, empty strings
- Non-variable dots: `{{.}}` (current context) is not a variable reference

### Graph Model

- `Node.Kind` defaults to `NodeKindTask`
- Variable nodes and task nodes coexist in the graph
- `"var"` edge class works alongside `"dep"` and `"call"`

### Builder

- Golden file tests with taskfiles containing global variables
- Verify variable nodes appear with correct labels and descriptions
- Verify variable→task edges exist for tasks that reference variables
- Verify unreferenced variables still appear as disconnected nodes
- Verify existing golden files unchanged when `IncludeGlobalVars` is false

### Graphviz Renderer

Golden file tests for:
- Variable nodes with default styling and `rank=sink` layout
- Variable edges with `"var"` class styling
- Namespace grouping with variables outside subgraphs
- Style rules applied to variable nodes via `var:*` patterns

### Mermaid Renderer

Golden file tests for:
- Variable node stadium shapes
- Variable edge thick arrows and `classDef` rules
- Variables positioned after task nodes in output

### Configuration

- `IncludeGlobalVars` defaults to `false`
- Loads correctly from YAML and JSON config files
- CLI `--include-global-vars` flag overrides config file
- `VariableNodes` and `VariableEdges` config fields serialize/deserialize correctly
- Export config includes new fields

### CLI Integration

- `--include-global-vars` flag wired through `applyConfigOverrides()`
- End-to-end: sample taskfile with variables produces expected output when flag is set
