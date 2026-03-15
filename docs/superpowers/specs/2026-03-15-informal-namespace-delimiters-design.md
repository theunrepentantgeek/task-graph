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
| `build-bin:test` | `build-bin` | `build` | Formal takes precedence; hyphen is literal in node ID but informal splitting applies to extracted namespace |
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

// Depth returns the nesting depth of a namespace.
// Counts all delimiters within the namespace's tier:
//   Tier 1 (contains ":"): counts colons. "cmd:test" → 1
//   Tier 2 (no ":"): counts hyphens and dots. "build-bin" → 1, "build-bin.linux" → 2
// A top-level namespace (no internal delimiters) has depth 0.
// Note: Tier selection is based on the input string. "cmd:build-bin" has depth 1 (one colon; tier 1).
func Depth(ns string) int

// CompileMatchPattern converts a glob-style pattern (using * and ?) to a compiled regexp.
// Returns an error if the resulting regex is invalid.
func CompileMatchPattern(pattern string) (*regexp.Regexp, error)

// MatchPattern returns a glob-style pattern string matching all nodes in the given namespace.
// The returned pattern is intended for storage in NodeStyleRule.Match and will be compiled
// via CompileMatchPattern when applied.
// For namespaces containing ":" (clearly formal): returns "ns:*" (e.g., "cmd:test:*")
// For namespaces without ":" (ambiguous tier): returns "ns[-.:]*" (e.g., "build[-.:]*")
// The ambiguous case matches all delimiter types, handling mixed-tier graphs correctly.
func MatchPattern(ns string) string
```

### Rule matching: switch from `path.Match` to regex

**Current state:** `path.Match` is used in graphviz and mermaid to match `NodeStyleRule.Match` patterns against node IDs. Errors from `path.Match` are silently discarded.

**New approach:**

1. `CompileMatchPattern` converts glob patterns (`*`, `?`, `[...]`) to regex by walking the pattern character by character:
   - `*` → `.*`
   - `?` → `.`
   - `[...]` character classes → bracket structure preserved; characters inside the brackets are passed through literally (regex character classes have the same semantics as glob character classes for our use cases)
   - All other characters → `regexp.QuoteMeta(char)` (escapes regex metacharacters like `.`)
   - Anchor the result with `^...$`
   - Compile with `regexp.Compile`
   
   This handles both user-defined glob patterns (e.g., `build*`, `*test*`) and autocolor-generated patterns containing character classes (e.g., `build[-.:]*`).
2. Both graphviz and mermaid use `CompileMatchPattern` for rule matching
3. **Errors are propagated** — invalid patterns produce a clear user-visible error rather than silent failure
4. Backward compatible — existing patterns like `build*`, `*test*`, `cmd:*` convert correctly

### Consumer changes

#### autocolor

- Replace local `nodeNamespace()` / `parentNamespace()` / `sortNamespaces` with `namespace.Namespace()` / `namespace.Parent()` / `namespace.Depth()`
- Pattern generation changes from `ns + ":*"` to `namespace.MatchPattern(ns)`:
  - Formal namespace `cmd:test` → `cmd:test:*`
  - Ambiguous namespace `build` → `build[-.:]*` (matches `build:x`, `build-x`, `build.x`)
  - A single rule per namespace handles mixed-tier graphs correctly
- Delete local namespace helper functions

#### graphviz

- Replace local `nodeNamespace()` / `parentNamespace()` with `namespace.Namespace()` / `namespace.Parent()`
- Replace `strings.Count(a, ":")` in sort comparisons with `namespace.Depth(a)`
- Replace `path.Match` with `namespace.CompileMatchPattern` + `regexp.MatchString`
- Propagate match compilation errors — signature changes:
  - `AddStyleRuleAttributes(nodeID, rule)` → `AddStyleRuleAttributes(nodeID, rule) error`
  - `writeNodeDefinitionTo(root, node, cfg, reg)` → `writeNodeDefinitionTo(root, node, cfg, reg) error`
  - `writeNodeTo(root, node, cfg, reg)` → `writeNodeTo(root, node, cfg, reg) error`
  - `writeNodesTo(root, nodes, cfg, reg)` → `writeNodesTo(root, nodes, cfg, reg) error`
  - `writeGroupedNodesTo(root, nodes, cfg, reg)` → `writeGroupedNodesTo(root, nodes, cfg, reg) error`
- Delete local namespace helper functions

#### mermaid

- Identical error propagation approach as graphviz — signature changes:
  - `findMatchingNodeIDs(nodes, pattern, reg) []string` → `findMatchingNodeIDs(nodes, pattern, reg) ([]string, error)`
  - `writeStyleRuleTo(root, nodes, i, rule, reg)` → `writeStyleRuleTo(root, nodes, i, rule, reg) error`
  - `writeStyleRulesTo(root, nodes, cfg, reg)` → `writeStyleRulesTo(root, nodes, cfg, reg) error`
  - `writeGroupedNodesTo(root, nodes, reg)` → `writeGroupedNodesTo(root, nodes, reg) error`
- Replace `path.Match` in `findMatchingNodeIDs` with `namespace.CompileMatchPattern`
- Delete local namespace helper functions

### Pattern compilation strategy

Patterns are compiled on-demand during output generation. Each `CompileMatchPattern` call returns a `*regexp.Regexp` that is used immediately. No caching is needed — the number of rules is small (typically < 20) and compilation is fast. If profiling later shows this matters, caching can be added inside `CompileMatchPattern` without changing its API.

### Autocolor pattern usage

Autocolor uses `namespace.MatchPattern(ns)` to generate the `Match` field for each `NodeStyleRule`. Graphviz and mermaid use `namespace.CompileMatchPattern(rule.Match)` to compile any rule's pattern (whether user-defined or autocolor-generated) for matching against node IDs.

**End-to-end pattern flow:**

| Source | Pattern (glob) | CompileMatchPattern output (regex) | Matches |
|---|---|---|---|
| User config: `match: "build*"` | `build*` | `^build.*$` | `build`, `build-bin`, `buildall` |
| User config: `match: "*test*"` | `*test*` | `^.*test.*$` | `test`, `mytest`, `test-unit` |
| User config: `match: "cmd:*"` | `cmd:*` | `^cmd:.*$` | `cmd:build`, `cmd:test` |
| Autocolor formal: `MatchPattern("cmd:test")` | `cmd:test:*` | `^cmd:test:.*$` | `cmd:test:unit`, `cmd:test:e2e` |
| Autocolor ambiguous: `MatchPattern("build")` | `build[-.:]*` | `^build[-.:].*$` | `build-bin`, `build.image`, `build:x` |

### Edge cases

- Leading/trailing delimiters (e.g., `-build`, `build-`): treated literally — `Namespace("-build")` returns `""` (empty prefix before `-`), `Namespace("build-")` returns `"build"`. Unlikely in practice.
- Consecutive delimiters (e.g., `build--bin`): produce empty segments. `Namespace("build--bin")` returns `"build-"`. Unlikely in practice.
- These are degenerate inputs; no special handling needed.

### Documentation update

`.github/copilot-instructions.md` references `path.Match` wildcards (`*`, `?`) for style rules. Update to reflect regex-based matching.

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
- `TestCompileMatchPattern_SpecialChars_Escaped` — dots/parens in literal positions are escaped
- `TestCompileMatchPattern_CharacterClass_PassedThrough` — `build[-.:]*` → `^build[-.:].*$` (brackets and contents passed through literally)
- `TestCompileMatchPattern_InvalidPattern_ReturnsError`
- `TestMatchPattern_FormalNamespace_ReturnsColonPattern` — `cmd:test` → `cmd:test:*`
- `TestMatchPattern_AmbiguousNamespace_ReturnsAllDelimiterPattern` — `build` → `build[-.:]*`

### Updated golden tests

- New `namespace_graph.golden` files for graphviz and mermaid with informal namespace grouping (tasks like `build-bin`, `build-image`, `tidy.format`, `tidy.lint`)
- Existing golden files regenerated — output should be identical, confirming backward compatibility

### autocolor tests

- Existing tests updated to use the `namespace` package
- New tests: `TestGenerateRules_InformalHyphenNamespace_ReturnsRule`, `TestGenerateRules_MixedFormalAndInformal_ReturnsBothRules`

### Error handling tests

- Integration tests confirming invalid `match` pattern in config produces a user-visible error
