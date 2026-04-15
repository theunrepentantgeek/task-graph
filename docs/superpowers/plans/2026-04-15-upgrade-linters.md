# Upgrade golangci-lint and Adopt New Linters — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade golangci-lint from v2.6.2 to v2.11.4 and adopt high-value new linters.

**Architecture:** Bump the version in the custom build template and install script, rebuild, fix findings from existing linters, then trial and adopt new linters one at a time.

**Tech Stack:** golangci-lint v2.11.4, Go, custom golangci-lint build with nilaway plugin

---

### Task 1: Bump golangci-lint version

**Files:**
- Modify: `.devcontainer/.custom-gcl.template.yml`
- Modify: `.devcontainer/install-dependencies.sh:~178` (golangci-lint install version)

- [ ] **Step 1: Update the custom build template version**

In `.devcontainer/.custom-gcl.template.yml`, change line 1:

```yaml
version: v2.11.4
```

(was `v2.6.2`)

- [ ] **Step 2: Update the standard golangci-lint install version**

In `.devcontainer/install-dependencies.sh`, find the line:

```bash
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2
```

Change to:

```bash
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4
```

- [ ] **Step 3: Rebuild the custom binary**

```bash
cd /workspaces/task-graph
.devcontainer/install-dependencies.sh --skip-installed
```

This will skip tools that are already installed except golangci-lint-custom
(since the template changed). If it skips golangci-lint-custom too, force
rebuild:

```bash
rm -f /usr/local/bin/golangci-lint-custom
.devcontainer/install-dependencies.sh --skip-installed
```

- [ ] **Step 4: Verify the new version**

```bash
golangci-lint-custom version
```

Expected: version string containing `v2.11.4`

- [ ] **Step 5: Commit version bump**

```bash
git add .devcontainer/.custom-gcl.template.yml .devcontainer/install-dependencies.sh
git commit -m "chore: bump golangci-lint from v2.6.2 to v2.11.4"
```

---

### Task 2: Fix findings from existing linters after upgrade

**Files:**
- Modify: `.golangci.yml` (if new revive rules need disabling)
- Modify: various source files (to fix legitimate findings)

- [ ] **Step 1: Run the linter and capture findings**

```bash
task lint 2>&1 | head -200
```

Review each finding. Categorise as:
- **Legitimate fix** → fix the code
- **False positive / too noisy** → suppress with `nolint` + explanation, or
  disable the specific rule

- [ ] **Step 2: Fix legitimate findings in source files**

Apply fixes to each flagged file. For autofixable issues:

```bash
golangci-lint-custom run --fix --verbose
```

Review the autofixed changes before proceeding.

- [ ] **Step 3: Disable noisy new revive rules if needed**

If any new revive rules are too noisy, add them to the `revive.rules` section
in `.golangci.yml` following the existing pattern:

```yaml
    revive:
      rules:
        - name: <noisy-rule-name>
          disabled: true
```

- [ ] **Step 4: Verify clean lint**

```bash
task lint
```

Expected: no errors

- [ ] **Step 5: Run tests to confirm no regressions**

```bash
go test ./...
```

Expected: all tests pass. If golden files need updating:

```bash
go test ./... -update
```

Then review the diffs in `testdata/` directories before accepting.

- [ ] **Step 6: Commit fixes**

```bash
git add -A
git commit -m "fix: resolve lint findings from golangci-lint v2.11.4 upgrade"
```

---

### Task 3: Trial and adopt Tier 1 linters

**Files:**
- Modify: `.golangci.yml` (add to `enable` list)
- Modify: `internal/cmd/cli.go:158` (fix inline error handling)

Trial each Tier 1 linter. For each: add to config, run lint, assess, keep or
remove.

- [ ] **Step 1: Add `noinlineerr` to the enable list**

In `.golangci.yml`, add `noinlineerr` to the `linters.enable` list
(alphabetical order, between `nlreturn` and `noctx`):

```yaml
    - noinlineerr
```

- [ ] **Step 2: Run lint to find inline error patterns**

```bash
task lint
```

Expected: one finding in `internal/cmd/cli.go:158`:

```go
if err := c.loadConfigFile(cfg); err != nil {
```

- [ ] **Step 3: Fix the inline error pattern**

In `internal/cmd/cli.go`, change:

```go
	if c.Config != "" {
		if err := c.loadConfigFile(cfg); err != nil {
			return nil, err
		}
	}
```

To:

```go
	if c.Config != "" {
		err := c.loadConfigFile(cfg)
		if err != nil {
			return nil, err
		}
	}
```

- [ ] **Step 4: Add `gochecknoinits` to the enable list**

In `.golangci.yml`, add `gochecknoinits` to the `linters.enable` list
(alphabetical order, between `gocheckcompilerdirectives` and
`gochecksumtype`):

```yaml
    - gochecknoinits
```

- [ ] **Step 5: Run lint to verify gochecknoinits is clean**

```bash
task lint
```

Expected: no findings (the codebase has no `init()` functions).

- [ ] **Step 6: Add `forcetypeassert` to the enable list**

In `.golangci.yml`, add `forcetypeassert` to the `linters.enable` list
(alphabetical order, between `fatcontext` and `funlen`):

```yaml
    - forcetypeassert
```

- [ ] **Step 7: Run lint to verify forcetypeassert is clean**

```bash
task lint
```

Expected: no findings (the one type assertion in the codebase already uses
comma-ok form at `internal/graphviz/node_properties.go:59`).

- [ ] **Step 8: Add `sloglint` to the enable list**

In `.golangci.yml`, add `sloglint` to the `linters.enable` list (alphabetical
order, between `rowserrcheck` and `spancheck`):

```yaml
    - sloglint
```

- [ ] **Step 9: Run lint to assess sloglint findings**

```bash
task lint
```

Review any findings. The project uses `log/slog` in `internal/cmd/cli.go` and
`internal/cmd/context.go`. If findings are legitimate, fix them. If sloglint
is too noisy for this codebase's usage pattern, remove it from the enable list.

- [ ] **Step 10: Verify all tests still pass**

```bash
go test ./...
```

- [ ] **Step 11: Commit Tier 1 linters**

```bash
git add -A
git commit -m "chore: adopt noinlineerr, gochecknoinits, forcetypeassert, sloglint linters"
```

Adjust the commit message to list only the linters that were actually adopted
(drop any that were removed during trial).

---

### Task 4: Trial and adopt Tier 2 linters

**Files:**
- Modify: `.golangci.yml` (add to `enable` list, add settings)

Trial each Tier 2 linter. For each: add to config, run lint, assess, keep or
remove.

- [ ] **Step 1: Add `iotamixing` to the enable list**

In `.golangci.yml`, add `iotamixing` to the `linters.enable` list
(alphabetical order, between `intrange` and `loggercheck`):

```yaml
    - iotamixing
```

- [ ] **Step 2: Run lint to verify iotamixing is clean**

```bash
task lint
```

Expected: no findings (the codebase has no iota usage).

- [ ] **Step 3: Add `funcorder` to the enable list**

In `.golangci.yml`, add `funcorder` to the `linters.enable` list
(alphabetical order, between `fatcontext`/`forcetypeassert` and `funlen`):

```yaml
    - funcorder
```

- [ ] **Step 4: Run lint to assess funcorder findings**

```bash
task lint
```

If many files need reordering, the linter doesn't match existing conventions —
remove it. If only a few files need adjustments or it's clean, keep it.

- [ ] **Step 5: Fix funcorder findings or remove if too noisy**

If keeping: reorder functions in flagged files to match the expected order
(constructors first, then exported methods, then unexported methods).

If removing: delete `funcorder` from the enable list.

- [ ] **Step 6: Add `forbidigo` with `fmt.Errorf` rule**

In `.golangci.yml`, add `forbidigo` to the `linters.enable` list
(alphabetical order, between `fatcontext`/`forcetypeassert`/`funcorder` and
`funlen`):

```yaml
    - forbidigo
```

And add settings:

```yaml
    forbidigo:
      forbid:
        - pattern: 'fmt\.Errorf'
          msg: "use eris.Wrap, eris.Wrapf, or eris.New instead of fmt.Errorf"
```

- [ ] **Step 7: Run lint to verify forbidigo is clean**

```bash
task lint
```

Expected: no findings (the codebase does not use `fmt.Errorf`). This linter
serves as prevention — enforcing the eris convention going forward.

- [ ] **Step 8: Add `embeddedstructfieldcheck` to the enable list**

In `.golangci.yml`, add `embeddedstructfieldcheck` to the `linters.enable`
list (alphabetical order, between `dupword` and `errchkjson`):

```yaml
    - embeddedstructfieldcheck
```

- [ ] **Step 9: Run lint to assess embeddedstructfieldcheck findings**

```bash
task lint
```

Three structs have embedded types:
- `internal/graphviz/node_properties.go`: `nodeProperties` embeds `properties`
- `internal/graphviz/edge_properties.go`: `edgeProperties` embeds `properties`
- `internal/graph/node.go`: `Node` embeds `NodeID`

If the linter flags them, check whether the embedded field is already at the
top with a blank line. Fix if easy; remove linter if it causes too many changes
or conflicts with existing style.

- [ ] **Step 10: Verify all tests still pass**

```bash
go test ./...
```

- [ ] **Step 11: Commit Tier 2 linters**

```bash
git add -A
git commit -m "chore: adopt iotamixing, funcorder, forbidigo, embeddedstructfieldcheck linters"
```

Adjust the commit message to list only the linters that were actually adopted.

---

### Task 5: Trial Tier 3 linter

**Files:**
- Modify: `.golangci.yml` (add to `enable` list)

- [ ] **Step 1: Add `nonamedreturns` to the enable list**

In `.golangci.yml`, add `nonamedreturns` to the `linters.enable` list
(alphabetical order, between `nolintlint` and `paralleltest`):

```yaml
    - nonamedreturns
```

- [ ] **Step 2: Run lint to verify nonamedreturns is clean**

```bash
task lint
```

Expected: no findings (the codebase has no named returns).

- [ ] **Step 3: Commit if adopted**

```bash
git add .golangci.yml
git commit -m "chore: adopt nonamedreturns linter"
```

---

### Task 6: Final verification

- [ ] **Step 1: Run full CI**

```bash
task ci
```

Expected: build, tests, lint, and SBOM generation all pass.

- [ ] **Step 2: Update golden files if needed**

If any test failures are due to golden file changes:

```bash
go test ./... -update
```

Review the diffs in `testdata/` directories. Only commit if the changes are
expected (e.g., output format changes from linter upgrades don't affect golden
files in this project, but test changes from code fixes might).

```bash
git add -A
git commit -m "test: update golden files after linter upgrade"
```

- [ ] **Step 3: Run full CI one more time**

```bash
task ci
```

Expected: all green.

- [ ] **Step 4: Review the full diff**

```bash
git --no-pager log --oneline main..HEAD
git --no-pager diff main..HEAD --stat
```

Confirm the change is reasonable in scope. If the diff is very large, consider
whether the commits are well-structured for review.
