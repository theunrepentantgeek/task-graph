# How We Do Things Here

This is a living guide to the conventions, patterns, and practices we follow.
It is written for both human developers and AI assistants, and serves as the primary reference when questions arise about how code should be structured, tested, or reviewed.
The principles here are deliberately portable — most of them apply equally well to any Go project with similar goals.

## The Principles That Shape Our Code

These four principles shape every decision we make.
They are not aspirational; they describe how the codebase actually works today.

1. **Small, focused packages** — each package has one clear responsibility and can be understood
   in isolation. A good package can answer three questions: what does it do, how do you use it,
   and what does it depend on?

2. **Explicit over implicit** — dependencies are passed as function parameters or via context
   structs. There are no hidden globals, no dependency injection frameworks, and no service
   locators. Objects are created at the entry point and threaded through the call stack, so the
   flow of data is always visible.

3. **Wrap errors with context** — every error carries enough information to diagnose without
   reaching for a debugger. Use `eris.Wrap`, `eris.Wrapf`, or `eris.New`; never bare
   `fmt.Errorf`. When something fails, the error message should tell you what happened, where it
   happened, and ideally why.

4. **Strict linting as a safety net** — the linter configuration is deliberately strict. It
   catches real bugs (nilaway, errcheck) and enforces consistency (funlen, line length, import
   order). Treat linter warnings as errors, not suggestions. If a rule feels wrong for a specific
   case, suppress it with a targeted comment and move on.

## What We Build With

Each library earns its place by solving a specific problem well.

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

## How Code Is Organised

All application code lives under `internal/`. Each package has a single, well-defined
responsibility — if you find yourself reaching across multiple concerns in one package, it is
probably time to split it.

Constructor functions (`New()`) create types with sensible defaults. Dependencies are passed as
function parameters or through context structs, keeping the wiring explicit and easy to trace.

We do not use interface-based dependency injection frameworks; all wiring happens explicitly at
the entry point. This means you can trace the construction of any object by starting at `main()`
and following the calls. It takes a little more typing, but the clarity is worth it — you never
have to wonder where something came from.

## Conventions Worth Knowing

These conventions keep the codebase consistent across contributors. They are enforced by the
linter where possible, and by code review where not.

- **Error handling**: Always use `eris.Wrap`, `eris.Wrapf`, or `eris.New`. Never use
  `fmt.Errorf` — it discards stack trace information that eris preserves.
- **Interface assertions**: Place `var _ TheInterface = &MyStruct{}` at package level to verify
  interface compliance at compile time. Group multiple assertions with `var ( ... )`.
- **Formatting**: All code is formatted with `gofumpt`, which is stricter than `gofmt`.
- **Import order** (enforced by `gci`): standard → dot imports → alias imports → default →
  local module.
- **Function length**: Keep functions under 60 lines, excluding comments. If a function outgrows
  this limit, look for a natural seam to extract a helper.
- **Line length**: 120 characters maximum.

Interface assertion example:

```go
var _ io.Reader = &MyReader{}

// or when implementing multiple interfaces:
var (
	_ io.Reader = &MyBuffer{}
	_ io.Writer = &MyBuffer{}
)
```

## How We Write Tests

We take testing seriously, not because we enjoy writing tests, but because well-structured tests
are the fastest way to build confidence in a change.

- **Assertion library**: Use gomega (`Expect`, fluent matchers). Never use testify.
- **Golden file tests**: Use goldie for verifying file or output content. Refresh fixtures with
  `go test ./... -update` and review the resulting diffs before committing.
- **Test naming**: Follow Roy Osherove style — `Test<Subject>_<Scenario>_<Expectation>`. The
  name should read like a sentence describing what is being verified.
- **Parallelism**: Mark all tests with `t.Parallel()` unless the test genuinely cannot run
  concurrently.
- **Helpers**: Mark helper functions with `t.Helper()` so failure messages point to the right
  line in the calling test, not the helper itself.
- **Structure**: Use `// Arrange`, `// Act`, `// Assert` comments to delineate phases. If a test
  does not fit this shape, it may be doing too much and should be split.
- **Table tests**: Use `cases := map[string]struct{...}{...}` with
  `for name, c := range cases`. Each sub-test gets its own `t.Parallel()` call.
- **Test ordering**: Earlier tests in a file assert foundational properties that later tests may
  rely on. When diagnosing failures, start at the first failing test — it usually points to the
  root cause.
- **Test grouping**: Tests for a given method should be grouped together with a leading comment.
- **Prefer table tests**: When multiple tests share similar structure, prefer a table test to
  avoid duplication.
- **Test packages**: Only use a `_test` package suffix if needed to break circular imports.

Here is a complete table test showing the naming convention, map pattern, gomega usage, and
Arrange/Act/Assert comments:

```go
func TestSanitise_VariousInputs_ProducesExpectedOutput(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input    string
		expected string
	}{
		"already clean": {
			input:    "hello",
			expected: "hello",
		},
		"leading spaces removed": {
			input:    "  hello",
			expected: "hello",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := gomega.NewWithT(t)

			// Arrange
			s := NewSanitiser()

			// Act
			result := s.Sanitise(c.input)

			// Assert
			g.Expect(result).To(gomega.Equal(c.expected))
		})
	}
}
```

## Building, Testing, and Linting

All build, test, and lint operations go through `Taskfile.yml`. This ensures everyone — humans
and CI alike — runs exactly the same commands with exactly the same flags.

- **Build**: `task build` compiles the project.
- **Test**: `go test ./...` runs the full test suite. For a quick check during development, you
  can scope to a single package.
- **Lint**: Always run `task lint`. Never run `golangci-lint` directly — the project uses a
  custom build that includes nilaway, and running the stock binary will miss checks or produce
  incorrect results.
- **Tidy**: `task tidy` runs `gofumpt`, `go mod tidy`, and `golangci-lint --fix` in one step.
  Run this before committing to catch formatting and import issues early.
- **CI**: `task ci` runs build, tests, lint, and SBOM generation. If CI passes, you are good to
  merge.
- **Golden files**: Refresh with `task update-golden-files` or `go test ./... -update`. Always
  review the resulting diffs before committing — golden file changes should be intentional, not
  accidental.

## Staying Focused and Disciplined

Good habits compound over time. These are the ones that matter most here.

- Stay focused on the current task. It is tempting to fix every issue you encounter, but
  scattered changes make PRs harder to review and riskier to revert.
- If you spot an unrelated issue, add it to `TODO.md` at the repo root for later action. This
  ensures good ideas are not lost, without derailing the work in progress.
- Do not fix pre-existing issues in the same PR unless they are tightly coupled to your changes.
  A separate PR for a separate concern keeps the history clean and makes reverts safe.
- Keep PRs focused: one concern per PR makes review easier and reverts safer.

## Working with AI Assistants

AI-generated code meets the same standards as human-written code. There are no separate rules, no
relaxed expectations, and no shortcuts. If the linter rejects it, fix it. If the tests fail,
investigate.

This file — `DEVELOPMENT.md` — is the primary convention reference for both humans and AI
agents. Project-specific details such as repository layout, configuration schema, and CI pipeline
configuration live in `.github/copilot-instructions.md`.
