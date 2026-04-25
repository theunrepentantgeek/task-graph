package dot

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

// TestFindExecutable_EmptyDotPath_FindsOnPath tests that FindExecutable
// locates the dot executable on the system PATH when dotPath is empty.
func TestFindExecutable_EmptyDotPath_FindsOnPath(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// This test requires dot to be available on PATH.
	// If dot is not installed, we skip the test.
	_, err := findDotOnPath()
	if err != nil {
		t.Skip("dot not found on PATH, skipping")
	}

	// Act
	result, err := FindExecutable("")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).NotTo(BeEmpty())
}

func TestFindExecutable_DirectoryWithoutDot_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange: create an empty directory
	dir := t.TempDir()

	// Act
	result, err := FindExecutable(dir)

	// Assert
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("dot executable not found in directory")))
	g.Expect(result).To(BeEmpty())
}

func TestFindExecutable_DirectoryWithDot_ReturnsDotPath(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange: create a directory containing a file named "dot"
	dir := t.TempDir()
	dotPath := filepath.Join(dir, "dot")
	err := os.WriteFile(dotPath, []byte("fake dot"), 0o600)
	g.Expect(err).NotTo(HaveOccurred())

	// Act
	result, err := FindExecutable(dir)

	// Assert
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(dotPath))
}

func TestFindExecutable_NonExistentPath_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	result, err := FindExecutable("/this/path/does/not/exist")

	// Assert
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("dotPath not found")))
	g.Expect(result).To(BeEmpty())
}

func TestFindExecutable_SpecificFilePath_ReturnsPath(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange: create a temp file to serve as a fake executable
	dir := t.TempDir()
	fakeDot := filepath.Join(dir, "dot")
	err := os.WriteFile(fakeDot, []byte("fake dot"), 0o600)
	g.Expect(err).NotTo(HaveOccurred())

	// Act: pass the direct file path (not a directory)
	result, err := FindExecutable(fakeDot)

	// Assert: the path is returned as-is without lookup
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(fakeDot))
}

func TestFindExecutable_EmptyDotPathNoDot_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// If dot is already on PATH, we can't test this case.
	_, err := findDotOnPath()
	if err == nil {
		t.Skip("dot is on PATH, cannot test missing dot")
	}

	// Act
	result, err := FindExecutable("")

	// Assert
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("dot executable not found on PATH")))
	g.Expect(result).To(BeEmpty())
}

func TestRenderImage_NonExistentExecutable_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act: pass a non-existent executable path
	err := RenderImage(context.Background(), "/nonexistent/dot", "input.dot", "output.png", "png")

	// Assert
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("dot command failed")))
}

func TestRenderImage_ExecutableFails_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange: use /bin/false which always exits with an error
	falseExe := "/bin/false"

	_, statErr := os.Stat(falseExe)
	if statErr != nil {
		t.Skip("/bin/false not available, skipping")
	}

	// Act
	err := RenderImage(context.Background(), falseExe, "input.dot", "output.png", "png")

	// Assert
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("dot command failed")))
}

// findDotOnPath is a helper for tests that need to check if dot is on PATH.
func findDotOnPath() (string, error) {
	return FindExecutable("")
}
