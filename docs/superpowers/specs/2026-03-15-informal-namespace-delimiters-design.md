# Informal Namespace Delimiters

**Date:** 2026-03-15
**Status:** Draft

## Problem

task-graph currently recognizes only `:` as a namespace delimiter, matching go-task's formal namespace convention for imported taskfiles. However, users commonly create _informal_ namespaces using other characters:

- `build-bin`, `build-image` → `build` namespace using `-`
- `tidy.format`, `tidy.lint`, `tidy.mod` → `tidy` namespace using `.`

These informal namespaces are invisible to `GroupByNamespace` and `AutoColor`, producing flat ungrouped graphs where users expect structure.

## Design Decisions

### Delimiter set

Recognized delimiters: `:`, `-`, `.`

- `:` is the formal go-task namespace delimiter (tier 1)
- `-` and `.` are informal delimiters (tier 2)
- `_` is excluded — too commonly used as a word joiner in task names (e.g., `run_tests`)
- `/` is excluded — collides with path handling

### Two-tier precedence

1. **Tier 1:** If a node ID contains `:`, split only on `:`. Hyphens and dots within segments are literal characters.
2. **Tier 2:** If no `:` is present, split on `-` and `.` equivalently.

Examples:

| Node ID | Namespace | Parent | Notes |
|---|---|---|---|
| `cmd:test:unit` | `cmd:test` | `cmd` | Formal, nested |
| `cmd:build` | `cmd` | (none) | Formal |
| `build-bin` | `build` | (none) | Informal hyphen |
| `tidy.format` | `tidy` | (none) | Informal dot |
| `build-bin.linux` | `build-bin` | `build` | Informal, nested |
| `build-bin:test` | `build-bin` | (none) | Formal takes precedence; hyphen is literal |
| `deploy` | (none) | N/A | No delimiter |

### Activation

No new configuration. The informal delimiters are recognized by namespace-aware features that are already opt-in:

- `GroupByNamespace` — gates namespace subgraph clustering
- `AutoColor` — gates namespace-based color assignment

Users who haven't enabled these features see no behavior change.

### No user-configurable delimiter set

The set of delimiters is fixed in code. The practical set is small and there's no downside to supporting all of them.

## Architecture

### New package: `internal/namespace`

A dedicated package replaces the duplicated `nodeNamespace()` / `parentNamespace()` functions currently in autocolor, graphviz, and mermaid. Single source of truth for namespace parsing.

**Exported API:**

```go
// Namespace returns the namespace portion of a node ID.
// Tier 1: if id contains ":", returns everything before the last ":".
// Tier 2: otherwise, returns everything before the last "-" or ".".
// Returns "" if no delimiter is found.
func Namespace(id string) string

// Parent returns the parent of a namespace string.
// Uses the same two-tier logic as Namespace.
// Returns "" if the namespace has no parent.
func Parent(ns string) string

// Depth returns the nesting depth of a namespace (number of internal delimiters).
func Depth(ns string) int

// CompileMatchPattern converts a glob-style pattern (using * and ?) to a compiled regexp.
// Returns an error if the resulting regex is invalid.
func CompileMatchPattern(pattern string) (*regexp.Regexp, error)

// MatchPattern returns a regex pattern string matching all nodes in the given namespace.
// For formal namespaces (containing ":"): returns "ns:.*"
// For informal namespaces: returns "ns[-.].*"
func MatchPattern(ns string) string
```

### Rule matching: switch from `path.Match` to regex

**Current state:** `path.Match` is used in graphviz and mermaid to match `NodeStyleRule.Match` patterns against node IDs. Errors from `path.Match` are silently discarded.

**New approach:**

1. `CompileMatchPattern` converts glob patterns (`*`, `?`) to regex:
   - `regexp.QuoteMeta(pattern)` escapes all special chars
   - Replace `\*` with `.*` and `\?` with `.`
   - Anchor with `^...$`
2. Both graphviz and mermaid use `CompileMatchPattern` for rule matching
3. **Errors are propagated** — invalid patterns produce a clear user-visible error rather than silent failure
4. Backward compatible — existing patterns like `build*`, `*test*`, `cmd:*` convert correctly

### Consumer changes

#### autocolor

- Replace local `nodeNamespace()` / `parentNamespace()` / `sortNamespaces` with `namespace.Namespace()` / `namespace.Parent()` / `namespace.Depth()`
- Pattern generation changes from `ns + ":*"` to `namespace.MatchPattern(ns)`:
  - Formal namespace `cmd` → `cmd:.*`
  - Informal namespace `build` → `build[-.].*`
- Delete local namespace helper functions

#### graphviz

- Replace local `nodeNamespace()` / `parentNamespace()` with `namespace.Namespace()` / `namespace.Parent()`
- Replace `strings.Count(a, ":")` in sort comparisons with `namespace.Depth(a)`
- Replace `path.Match` with `namespace.CompileMatchPattern` + `regexp.MatchString`
- Propagate match compilation errors from `AddStyleRuleAttributes` through `WriteTo`
- Delete local namespace helper functions

#### mermaid

- Identical changes as graphviz
- Replace `path.Match` in `findMatchingNodeIDs` with `namespace.CompileMatchPattern`
- Propagate errors through `WriteTo`
- Delete local namespace helper functions

### What doesn't change

- `config.Config` — no new fields
- `config.NodeStyleRule` — struct unchanged; `Match` field now holds regex-compatible patterns (backward compatible)
- `safe.Registry` — untouched
- `graph.Graph` / `graph.Node` — untouched
- User-facing config YAML — existing `match: "build*"` patterns continue to work

## Testing

### namespace package (new)

Core parsing logic:

- `TestNamespace_FormalDelimiter_ReturnsPrefix` — `cmd:build` → `cmd`
- `TestNamespace_NestedFormalDelimiter_ReturnsFullPrefix` — `cmd:test:unit` → `cmd:test`
- `TestNamespace_InformalHyphen_ReturnsPrefix` — `build-bin` → `build`
- `TestNamespace_InformalDot_ReturnsPrefix` — `tidy.format` → `tidy`
- `TestNamespace_MixedInformal_ReturnsPrefix` — `build-bin.linux` → `build-bin`
- `TestNamespace_FormalTakesPrecedence_IgnoresInformal` — `build-bin:test` → `build-bin`
- `TestNamespace_NoDelimiter_ReturnsEmpty` — `deploy` → `""`
- `TestParent_FormalNested_ReturnsParent` — `cmd:test` → `cmd`
- `TestParent_InformalNested_ReturnsParent` — `build-bin` → `build`
- `TestParent_TopLevel_ReturnsEmpty` — `cmd` → `""`
- `TestDepth_VariousNamespaces_ReturnsCorrectDepth`
- `TestCompileMatchPattern_GlobStar_ConvertsToRegex` — `build*` → `^build.*$`
- `TestCompileMatchPattern_GlobQuestion_ConvertsToRegex` — `build?` → `^build.$`
- `TestCompileMatchPattern_SpecialChars_Escaped` — dots/parens are escaped
- `TestCompileMatchPattern_InvalidRegex_ReturnsError`
- `TestMatchPattern_FormalNamespace_ReturnsColonPattern` — `cmd` → `cmd:.*`
- `TestMatchPattern_InformalNamespace_ReturnsBracketPattern` — `build` → `build[-.].*`

### Updated golden tests

- New `namespace_graph.golden` files for graphviz and mermaid with informal namespace grouping (tasks like `build-bin`, `build-image`, `tidy.format`, `tidy.lint`)
- Existing golden files regenerated — output should be identical, confirming backward compatibility

### autocolor tests

- Existing tests updated to use the `namespace` package
- New tests: `TestGenerateRules_InformalHyphenNamespace_ReturnsRule`, `TestGenerateRules_MixedFormalAndInformal_ReturnsBothRules`

### Error handling tests

- Integration tests confirming invalid `match` pattern in config produces a user-visible error
