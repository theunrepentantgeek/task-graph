# Copilot Instructions for task-graph

## Repository Overview

`task-graph` is a Go command-line tool that reads a [Taskfile](https://taskfile.dev) (a `Taskfile.yml`) and generates a Graphviz `.dot` file representing the dependency graph of tasks. It can optionally render the graph as an image using the `dot` executable.

- **Language**: Go (requires Go 1.22+)
- **Entry point**: `main.go`
- **Key packages**: `github.com/alecthomas/kong` (CLI), `github.com/go-task/task/v3` (Taskfile parsing), `gopkg.in/yaml.v3` (YAML config), `github.com/rotisserie/eris` (error wrapping)
- **Test libraries**: `github.com/onsi/gomega` (assertions), `github.com/sebdah/goldie/v2` (golden file tests)
- **Build tool**: [Task](https://taskfile.dev) (`task` command)

## Project Layout

```
main.go                        # Entry point; wires CLI via kong
internal/
  cmd/                         # CLI command structs and Run() method
  config/                      # Config structs (Config, Graphviz, GraphvizNode, etc.)
  dot/                         # dot executable discovery and image rendering
  graph/                       # Core graph data structures
  graphviz/                    # .dot file generation from graph
  indentwriter/                # Indented writer utility
  loader/                      # Taskfile loading via go-task library
  taskgraph/                   # Building the task graph from a loaded Taskfile
.github/
  workflows/
    pr-validation.yml          # CI: runs `task ci` inside devcontainer
    copilot-setup-steps.yml    # Sets up the environment for Copilot coding agent
    codeql.yml                 # CodeQL security analysis
.devcontainer/
  install-dependencies.sh      # Installs all tools (golangci-lint-custom, gofumpt, task, etc.)
  .custom-gcl.template.yml     # Template for the custom golangci-lint build
Taskfile.yml                   # Project task definitions (build, test, lint, etc.)
.golangci.yml                  # golangci-lint v2 configuration (strict; many linters enabled)
DEVELOPMENT.md                 # Developer guidelines
samples/                       # Sample Taskfiles and generated .dot/.png files
docs/                          # Generated documentation (taskfile.dot, taskfile.png)
tools/                         # Local tool binaries (installed by install-dependencies.sh)
```

## Build, Test, and Lint

### Environment Setup (required once)

Tools are installed locally into the `./tools` directory. To install them:

```bash
.devcontainer/install-dependencies.sh --skip-installed
export PATH="$PATH:$(realpath ./tools)"
```

The `PATH` must include `./tools` for all task commands to work.

### Build

```bash
task build
# or directly:
go build -o build/task-graph
```

### Run Tests

```bash
go test ./...
```

To update golden test fixtures after changing output:

```bash
go test ./... -update
# or via task:
task update-golden-files
```

### Lint

**Always use `task lint` to run the linter.** Do NOT run `golangci-lint` directly — the project uses a custom-built `golangci-lint-custom` binary (with `nilaway` integration) that requires the custom build process in `install-dependencies.sh`.

```bash
task lint
```

### Full CI

```bash
task ci
```

This runs build, tests, lint, and SBOM generation.

### Tidy

```bash
task tidy   # runs gofumpt, go mod tidy, and golangci-lint --fix
```

## Coding Conventions

- **Error wrapping**: Always use `eris.Wrap`, `eris.Wrapf`, or `eris.New` (from `github.com/rotisserie/eris`); never `fmt.Errorf`.
- **Interface assertions**: Always add `var _ TheInterface = &MyStruct{}` when a struct implements an interface. Group multiple assertions in `var ( ... )`.
- **Test naming**: Use Roy Osherove style: `Test<Subject>_<Scenario>_<Expectation>`.
- **Test structure**: Mark tests with `t.Parallel()` unless not possible. Mark helpers with `t.Helper()`. Use Arrange/Act/Assert comments.
- **Table tests**: Use `cases := map[string]struct{...}{...}` with `for name, c := range cases`.
- **Test ordering**: In a test file, earlier tests assert foundational properties that later tests may rely on.
- **Goldie golden tests**: Use for verifying file/output content; refresh with `go test ./... -update`.
- **Gomega**: Use `gomega` (`Expect`, `Ω`, `Eventually`, etc.) for assertions; never `testify`.
- **Formatting**: Code is formatted with `gofumpt` (stricter than `gofmt`).
- **Import order** (enforced by `gci`): standard → dot imports → alias imports → default → local module.
- **Function length**: Keep functions under 60 lines (excluding comments).
- **Line length**: Maximum 120 characters.
- **TODO tracking**: If you notice issues unrelated to your current task, add them to `TODO.md` at the repo root rather than fixing them immediately.

## Configuration

Config is loaded from a YAML or JSON file passed via `--config`. The `Config` struct (in `internal/config/config.go`) supports Graphviz styling:

- `graphviz.taskNodes`: Default node presentation (`color`, `fillColor`, `style`, `fontColor`)
- `graphviz.styleRules[]`: Pattern-matched style overrides using `path.Match` wildcards (`*`, `?`)
- `graphviz.dependencyEdges`, `graphviz.callEdges`: Edge styling
- `graphviz.font`, `graphviz.fontSize`: Label font settings

## CI / PR Validation

The CI pipeline (`.github/workflows/pr-validation.yml`) runs `task ci` inside the devcontainer image. PRs must pass this check. The devcontainer image is built and cached in GitHub Container Registry (`ghcr.io/theunrepentantgeek/task-graph-devcontainer`).

The Copilot setup steps (`.github/workflows/copilot-setup-steps.yml`) install dependencies via `install-dependencies.sh --skip-installed` and add `./tools` to `PATH`.
