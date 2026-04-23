package loader

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		taskfile      string
		expectedTasks int
		expectedError string
	}{
		"go-vcr-tidy": {
			taskfile:      "go-vcr-tidy-taskfile.yml",
			expectedTasks: 14,
		},
		"crddoc-vcr-tidy": {
			taskfile:      "crddoc-taskfile.yml",
			expectedTasks: 19,
		},
		"aso": {
			taskfile:      "aso-taskfile.yml",
			expectedTasks: 118,
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			fn := filepath.Join("testdata", c.taskfile)

			tf, err := Load(t.Context(), fn)
			if c.expectedError == "" {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(tf.Tasks.Len()).To(Equal(c.expectedTasks))
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(c.expectedError)))
			}
		})
	}
}

func TestLoad_AbsolutePath_Succeeds(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange: use an absolute path to the testdata file
	abs, err := filepath.Abs(filepath.Join("testdata", "go-vcr-tidy-taskfile.yml"))
	g.Expect(err).NotTo(HaveOccurred())

	// Act
	tf, err := Load(t.Context(), abs)

	// Assert: absolute path is handled the same as relative
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(tf.Tasks.Len()).To(Equal(14))
}

func TestLoad_NonExistentFile_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	tf, err := Load(t.Context(), "testdata/does-not-exist.yml")

	// Assert
	g.Expect(err).To(HaveOccurred())
	g.Expect(tf).To(BeNil())
}

func TestLoad_InvalidYAML_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange: write a temp file with invalid YAML content
	dir := t.TempDir()
	badFile := filepath.Join(dir, "Taskfile.yml")
	err := os.WriteFile(badFile, []byte("not: valid: yaml: content: :::"), 0o644)
	g.Expect(err).NotTo(HaveOccurred())

	// Act
	tf, err := Load(t.Context(), badFile)

	// Assert
	g.Expect(err).To(HaveOccurred())
	g.Expect(tf).To(BeNil())
}

