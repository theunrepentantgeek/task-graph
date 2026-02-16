package loader

import (
	"context"
	"path/filepath"

	"github.com/go-task/task/v3/taskfile"
	"github.com/go-task/task/v3/taskfile/ast"
)

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
			return nil, err
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
		return nil, err
	}

	reader := taskfile.NewReader()

	graph, err := reader.Read(ctx, node)
	if err != nil {
		return nil, err
	}

	result, err := graph.Merge()
	if err != nil {
		return nil, err
	}

	return result, nil
}
