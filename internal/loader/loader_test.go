package loader

import (
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
