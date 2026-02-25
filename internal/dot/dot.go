package dot

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rotisserie/eris"
)

// FindExecutable returns the path to the dot executable.
// If dotPath is the full path to an executable, it is returned directly.
// If dotPath is a directory, the dot executable is looked up within that directory.
// If dotPath is empty, dot is looked up on the system PATH.
func FindExecutable(dotPath string) (string, error) {
	if dotPath == "" {
		path, err := exec.LookPath("dot")
		if err != nil {
			return "", eris.Wrap(err, "dot executable not found on PATH")
		}

		return path, nil
	}

	info, err := os.Stat(dotPath)
	if err != nil {
		return "", eris.Wrapf(err, "dotPath not found: %s", dotPath)
	}

	if info.IsDir() {
		candidate := filepath.Join(dotPath, "dot")
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, nil
		}

		// Try with .exe extension on Windows
		candidateExe := candidate + ".exe"
		if _, statErr := os.Stat(candidateExe); statErr == nil {
			return candidateExe, nil
		}

		return "", eris.Errorf("dot executable not found in directory: %s", dotPath)
	}

	return dotPath, nil
}

// RenderImage runs the dot executable to render dotFile to imageFile using the given fileType.
// The fileType is passed to dot as -T<fileType> (e.g. "png", "svg").
func RenderImage(ctx context.Context, dotExecutable, dotFile, imageFile, fileType string) error {
	cmd := exec.CommandContext(ctx, dotExecutable, "-T"+fileType, dotFile, "-o", imageFile) //nolint:gosec // dotExecutable is resolved from a trusted config path or system PATH
	output, err := cmd.CombinedOutput()
	if err != nil {
		return eris.Wrapf(err, "dot command failed: %s", string(output))
	}

	return nil
}
