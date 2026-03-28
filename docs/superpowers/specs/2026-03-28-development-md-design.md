# Design: DEVELOPMENT.md — How We Do Things Here

## Problem

The existing `DEVELOPMENT.md` is incomplete (references a different project name, has "TBC" placeholders) and duplicates content scattered across `.github/copilot-instructions.md`.
We need a single, authoritative source for development conventions that is generic enough to copy across projects, while also eliminating duplication with the Copilot instructions file.

## Approach

Replace the current `DEVELOPMENT.md` with a philosophy-first conventions document.
Update `.github/copilot-instructions.md` to reference `DEVELOPMENT.md` for conventions rather than duplicating them.

## Writing Style

Follow the voice and grammar guidelines in `.github/writing-style.md`:

- Target experienced developers; do not over-explain basics
- Friendly but direct — like explaining to a new team member
- British English spelling (colour, behaviour, favour)
- Explain *why*, not just *what*
- Conversational but professional register

## Deliverables

### 1. `DEVELOPMENT.md` (replace existing)

Nine sections, each scaled to complexity:

#### §1 Introduction

One paragraph establishing this as a living guide to "how we do things here."
Addresses both human developers and AI assistants.
Notes it is written to be portable across projects with similar conventions.

#### §2 Design Philosophy

Four principles, each with a brief rationale:

1. **Small, focused packages** — each package has one clear responsibility and can be understood in isolation.
   A good package can answer three questions: what does it do, how do you use it, and what does it depend on?
2. **Explicit over implicit** — dependencies are passed as function parameters or via context structs.
   No hidden globals, no DI frameworks, no service locators.
   Objects are created at the entry point and threaded through the call stack.
3. **Wrap errors with context** — every error carries enough information to diagnose without reaching for a debugger.
   Use `eris.Wrap`, `eris.Wrapf`, or `eris.New`; never bare `fmt.Errorf`.
4. **Strict linting as a safety net** — the linter configuration is deliberately strict.
   It catches real bugs (nilaway, errcheck) and enforces consistency (funlen, line length, import order).
   Treat linter warnings as errors, not suggestions.

#### §3 Tech Stack and Libraries

A table or list naming each key library with a one-line rationale:

| Library | Purpose |
|---------|---------|
| Go | Primary language |
| kong | Struct-tag driven CLI parsing with minimal boilerplate |
| eris | Error wrapping with stack traces; preferred over `fmt.Errorf` |
| gomega | Fluent test assertions; preferred over testify |
| goldie | Golden file testing for output verification |
| gofumpt | Stricter `gofmt` for consistent formatting |
| golangci-lint | Broad linter set with custom build (includes nilaway) |
| Task | Build automation via `Taskfile.yml` |

#### §4 Code Organisation

- All application code lives under `internal/`
- Each package has a single, well-defined responsibility
- Constructor functions (`New()`) create types with sensible defaults
- Dependencies are passed as function parameters or through context structs
- No interface-based DI frameworks; wiring happens explicitly at the entry point

#### §5 Coding Conventions

- **Error handling**: Always `eris.Wrap`/`eris.Wrapf`/`eris.New`, never `fmt.Errorf`
- **Interface assertions**: `var _ TheInterface = &MyStruct{}` at package level; group with `var ( ... )` when multiple
- **Formatting**: `gofumpt` (stricter than `gofmt`)
- **Import order** (enforced by `gci`): standard → dot imports → alias imports → default → local module
- **Function length**: Under 60 lines, excluding comments
- **Line length**: 120 characters maximum

Include a brief code example for the interface assertion pattern.

#### §6 Testing

- **Assertion library**: gomega (`Expect`, fluent matchers); never testify
- **Golden file tests**: goldie for verifying file/output content; refresh with `go test ./... -update`
- **Test naming**: Roy Osherove style — `Test<Subject>_<Scenario>_<Expectation>`
- **Parallelism**: Mark all tests with `t.Parallel()` unless the test genuinely cannot run concurrently
- **Helpers**: Mark helper functions with `t.Helper()` so failure messages point to the right line
- **Structure**: Use `// Arrange`, `// Act`, `// Assert` comments; if a test does not fit this shape, it may be doing too much
- **Table tests**: `cases := map[string]struct{...}{...}` with `for name, c := range cases`; each sub-test gets `t.Parallel()` too
- **Test ordering**: Earlier tests in a file assert foundational properties that later tests may rely on. When diagnosing failures, start at the first failing test.
- **Test grouping**: Tests for a given method should be grouped together with a leading comment
- **Prefer table tests**: When multiple tests share similar structure, prefer a table test to avoid duplication
- **Test packages**: Only use a `_test` package suffix if needed to break circular imports

Include a brief table test example showing the naming, map pattern, and Arrange/Act/Assert comments.

#### §7 Build and CI

- **Task runner**: All build, test, and lint operations go through `Taskfile.yml`
- **Lint**: Always `task lint`; never run `golangci-lint` directly (the project uses a custom build with nilaway)
- **Tidy**: `task tidy` runs `gofumpt`, `go mod tidy`, and `golangci-lint --fix`
- **CI**: `task ci` runs build, tests, lint, and SBOM generation
- **Golden files**: Refresh with `task update-golden-files` or `go test ./... -update`; review diffs before committing

#### §8 Developer Discipline

- Stay focused on the current task
- If you spot unrelated issues, add them to `TODO.md` at the repo root for later action
- Do not fix pre-existing issues in the same PR unless they are tightly coupled to your changes
- Keep PRs focused: one concern per PR makes review easier and reverts safer

#### §9 Working with AI Assistants

- AI-generated code meets the same standards as human-written code
- `DEVELOPMENT.md` is the primary convention reference for both humans and AI agents
- Project-specific details (repo layout, configuration schema, CI pipeline) live in `.github/copilot-instructions.md`

### 2. `.github/copilot-instructions.md` (update)

**Remove** the duplicated "Coding Conventions" section.
**Replace** with a short reference:

```markdown
## Coding Conventions

See `DEVELOPMENT.md` in the repository root for all coding conventions, testing practices,
and design philosophy. The same standards apply to AI-generated code.
```

**Keep** all project-specific sections:
- Repository Overview
- Project Layout
- Build, Test, and Lint (commands are project-specific)
- Configuration
- CI / PR Validation

## Out of Scope

- Changes to `.golangci.yml` or `Taskfile.yml`
- Changes to any Go source code
- Changes to `.github/writing-style.md`
- Creating new tooling or scripts
