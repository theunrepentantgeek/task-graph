package dot

import (
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
	if _, err := findDotOnPath(); err != nil {
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

func TestFindExecutable_EmptyDotPathNoDot_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// If dot is already on PATH, we can't test this case.
	if _, err := findDotOnPath(); err == nil {
		t.Skip("dot is on PATH, cannot test missing dot")
	}

	// Act
	result, err := FindExecutable("")

	// Assert
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("dot executable not found on PATH")))
	g.Expect(result).To(BeEmpty())
}

// findDotOnPath is a helper for tests that need to check if dot is on PATH.
func findDotOnPath() (string, error) {
	return FindExecutable("")
}
