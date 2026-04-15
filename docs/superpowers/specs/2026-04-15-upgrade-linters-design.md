# Upgrade golangci-lint and Adopt New Linters

**Date**: 2026-04-15
**Status**: Draft

## Summary

Upgrade the custom golangci-lint build from v2.6.2 to v2.11.4 and adopt
high-value new linters that have been added since v2.6.2. Each candidate linter
is trial-run against the codebase; only those with a good signal-to-noise ratio
are included.

## Context

The project uses a custom golangci-lint build (with nilaway plugin) pinned at
v2.6.2. The latest stable release is v2.11.4. Since v2.6.2, golangci-lint has
added several new linters and existing linters have gained significant new rules
and features.

The project already enables 74 linters with a strict configuration. The goal is
to stay current, catch more bugs, and enforce more consistency — while keeping
false positives manageable. A handful of `nolint` suppressions are acceptable
where the linter's overall value justifies them.

## Version Upgrade: v2.6.2 → v2.11.4

### File changes

- `.devcontainer/.custom-gcl.template.yml`: bump `version` from `v2.6.2` to
  `v2.11.4`
- Rebuild the custom binary via `install-dependencies.sh`

### Existing linter changes surfaced by the upgrade

Already-enabled linters have gained new rules and features that may produce new
findings on existing code:

| Linter | Key changes since v2.6.2 |
|--------|--------------------------|
| gosec | ~17 new rules (G113, G116–G123, G408, G602, G701–G707) |
| revive | New rules: `package-naming`, `epoch-naming`, `use-slices-sort`, `enforce-switch-style`, `forbidden-call-in-wg-go`, `unnecessary-if`, `inefficient-map-lookup`, `time-date`, `unnecessary-format`, `use-fmt-print` |
| staticcheck | 0.6.1 → 0.7.0 |
| modernize | New analyzers: `stringscut`, `unsafefuncs` |
| wsl_v5 | 5.3.0 → 5.6.0 (new `after-block` rule) |
| errcheck | Excludes `crypto/rand.Read` by default |

Since revive is configured with `enable-all-rules: true`, new revive rules are
automatically enabled. If any new rule is too noisy, add a targeted
`disabled: true` entry — same pattern as existing `add-constant`,
`function-length`, and `unhandled-error` exclusions.

Findings from existing linters are fixed (or suppressed with explanation) as
part of the upgrade.

## New Linters to Adopt

Each candidate is trial-run against the codebase. Include only if the
signal-to-noise ratio is good.

### Tier 1 — High confidence

**`noinlineerr`** (added v2.2.0) — Enforces split error handling style. Forbids
`if err := f(); err != nil {` in favour of separate assignment and check.
Autofix available. Matches the project's preferred style.

**`gochecknoinits`** (pre-existing, not yet enabled) — Forbids `init()`
functions. The codebase has zero today; this prevents them from creeping in.
Zero-noise linter.

**`forcetypeassert`** (pre-existing, not yet enabled) — Flags bare type
assertions (`x.(T)`) that don't use the comma-ok form. Real bug finder — an
unchecked type assertion is a panic waiting to happen.

**`sloglint`** (pre-existing, not yet enabled) — Enforces consistent `log/slog`
usage patterns (key naming style, argument structure). The project uses slog for
CLI logging.

### Tier 2 — Medium confidence (trial and decide)

**`iotamixing`** (added v2.5.0) — Flags const blocks that mix iota with
non-iota values. Fast, low noise.

**`funcorder`** (added v2.1.0) — Enforces ordering of constructors,
exported/unexported methods. May need config tuning. Adopt if it aligns with
existing patterns; skip if it requires major reshuffling.

**`forbidigo`** (pre-existing, not yet enabled) — Can be configured to forbid
`fmt.Errorf`, upgrading the eris convention from documentation to enforcement.

**`embeddedstructfieldcheck`** (added v2.2.0) — Embedded types at top of struct
with a blank line separator. Adopt if the codebase already follows this pattern.

### Tier 3 — Low confidence

**`nonamedreturns`** (pre-existing, not yet enabled) — Forbids named returns.
Trial and adopt only if the codebase doesn't currently use them.

### Not adopting

| Linter | Reason |
|--------|--------|
| `unqueryvet` | No SQL in the project |
| `arangolint` | No ArangoDB |
| `canonicalheader` | No net/http usage |
| `nosprintfhostport` | No URL construction |
| `musttag` | Already disabled as "extremely slow" |
| `mnd` | Magic number detection is typically very noisy |
| `exhaustruct` | Too noisy — requires initializing every struct field |
| `ireturn` | Opinionated; would need many suppressions |
| `testpackage` | Project deliberately uses same-package tests |
| `lll` | Already covered by revive's `line-length-limit` |
| `gochecknoglobals` | Existing globals are all effectively constants/caches |
| `containedctx` | No `context.Context` stored in structs |
| `varnamelen` | Settings configured but not enabled — likely evaluated and skipped previously |
| `decorder` | Low value for this codebase size |

## Implementation Strategy

### Trial process for each candidate linter

1. Temporarily add the linter to the `enable` list in `.golangci.yml`
2. Run `task lint`
3. Assess findings:
   - **All clean or ≤3 legitimate findings** → adopt, fix findings
   - **Findings that are all autofixable** → adopt, apply autofix, review result
   - **Many findings but all legitimate** → adopt, fix as part of the PR
   - **Noisy / many false positives** → skip, don't include in final config
4. Move on to next candidate

### Order of operations

1. Bump version in `.custom-gcl.template.yml` → rebuild custom binary → run
   `task lint` → fix findings from existing linters
2. Trial Tier 1 linters one at a time (noinlineerr, gochecknoinits,
   forcetypeassert, sloglint)
3. Trial Tier 2 linters one at a time (iotamixing, funcorder, forbidigo,
   embeddedstructfieldcheck)
4. Trial Tier 3 if time permits (nonamedreturns)
5. Run full CI (`task ci`) to confirm everything passes
6. Update golden files if output format changed (`go test ./... -update`,
   review diffs)

### Files touched

- `.devcontainer/.custom-gcl.template.yml` — version bump
- `.golangci.yml` — add adopted linters to `enable` list, add settings entries
  as needed, disable noisy new revive rules if necessary
- Source files — fix legitimate findings, add `nolint` directives with
  explanations where suppression is warranted

### Risk and mitigation

**Large diff from upgrade**: The v2.11.x upgrade may surface many new findings
from existing linters, making the PR larger than expected. Mitigation: split
into two commits within the same PR (upgrade + fix existing, then add new
linters) to make review easier.

**Nilaway plugin compatibility**: The custom build uses nilaway as a module
plugin. The v2.11.4 custom build framework may require a compatible nilaway
version. If the build fails, check for nilaway updates or pin a compatible
version in `.custom-gcl.template.yml`.
