package loader

import (
	"context"
	"path/filepath"

	"github.com/go-task/task/v3/taskfile"
	"github.com/go-task/task/v3/taskfile/ast"
	"github.com/rotisserie/eris"
)

// Load reads and parses the Taskfile at the given filename, resolving relative
// paths before delegating to the go-task reader. Returns the merged AST or an
// error if the file cannot be read, parsed, or merged.
func Load(
	ctx context.Context,
	filename string,
) (*ast.Taskfile, error) {
	resolvedPath := filename

	// Resolve relative paths up front so the taskfile reader can locate
	// the file regardless of the current working directory.
	if !filepath.IsAbs(resolvedPath) {
		var err error

		resolvedPath, err = filepath.Abs(filename)
		if err != nil {
			return nil, eris.Wrapf(err, "failed to resolve path: %s", filename)
		}
	}

	dir := filepath.Dir(resolvedPath)
	entrypoint := resolvedPath

	node, err := taskfile.NewRootNode(
		entrypoint, // Taskfile to load
		dir,        // Initial directory
		false,      // Insecure mode
		0,          // Task execution timeout
	)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to create root node for taskfile: %s", entrypoint)
	}

	reader := taskfile.NewReader()

	graph, err := reader.Read(ctx, node)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to read taskfile: %s", entrypoint)
	}

	result, err := graph.Merge()
	if err != nil {
		return nil, eris.Wrapf(err, "failed to merge taskfile graph: %s", entrypoint)
	}

	return result, nil
}
