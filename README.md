# task-graph

Document the build processes you've built with [task](https://taskfile.dev/) by rendering them as easy to read graphs.

For example, here's the graph of the tasks for `task-graph` itself:

![Tasks of task-graph](https://github.com/theunrepentantgeek/task-graph/blob/main/docs/taskfile.png)

**Documentation** - generate a graph for reference, helping users of your project understand what tasks are available.

**Troubleshooting** - a graph of dependencies can be invaluable when debugging problems with your taskfile.

## Project Status

Beta - works on all the Taskfiles I've tried, but needs to be exercised in the "real world".

## Installation

Install from source:

```
go install github.com/theunrepentantgeek/task-graph
```

## Usage

Generate a DOT file using default options:

``` bash
task-graph Taskfile.yml --output taskfile.dot
```

Render that DOT file as a PNG:

``` bash
dot taskfile.dot -Tpng -o taskfile.png 
```

### Full command-line options

``` bash
Usage: task-graph --output=STRING <taskfile> [flags]

Arguments:
  <taskfile>    Path to the taskfile to process.

Flags:
  -h, --help                    Show context-sensitive help.
  -o, --output=STRING           Path to the output file.
  -c, --config=STRING           Path to a config file (YAML or JSON).
      --group-by-namespace      Group tasks in the same namespace together in the output.
      --auto-color              Automatically color nodes by namespace using a built-in palette.
      --colorblind-mode         Use an accessibility-optimised colour palette (Okabe-Ito) for --auto-color instead of the
                                default palette.
      --graph-type=STRING       Type of graph to generate (dot or mermaid). Defaults to dot.
      --highlight=STRING        Highlight specific tasks in the graph. Accepts task names or glob patterns, separated by
                                commas or semicolons.
      --render-image=STRING     Render the graph as an image using graphviz dot. Specify the file type (e.g. png, svg).
      --export-config=STRING    Export the effective configuration to a file (YAML or JSON based on file extension).
      --focus=STRING            Show only tasks matching the given patterns together with all their transitive
                                dependencies and dependents. Accepts task names or glob patterns, separated by commas
                                or semicolons.
      --verbose                 Enable verbose logging.
```

## Samples

### go-vcr-tidy

![go-vcr-tidy](https://github.com/theunrepentantgeek/task-graph/blob/main/samples/go-vcr-tidy.png)

### task

![task](https://github.com/theunrepentantgeek/task-graph/blob/main/samples/task-taskfile.png)
