package taskgraph

import (
	"bytes"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sebdah/goldie/v2"

	"github.com/theunrepentantgeek/task-graph/internal/graphviz"
	"github.com/theunrepentantgeek/task-graph/internal/loader"
)

func TestTaskGraphBuilder_Graphviz(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		taskfile   string
		goldenName string
	}{
		"go-vcr-tidy": {
			taskfile:   "go-vcr-tidy-taskfile.yml",
			goldenName: "go-vcr-tidy-taskfile",
		},
		"crddoc": {
			taskfile:   "crddoc-taskfile.yml",
			goldenName: "crddoc-taskfile",
		},
		"aso": {
			taskfile:   "aso-taskfile.yml",
			goldenName: "aso-taskfile",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			taskfilePath := filepath.Join("testdata", c.taskfile)

			tf, err := loader.Load(t.Context(), taskfilePath)
			g.Expect(err).NotTo(HaveOccurred())

			gr := New(tf).Build()

			buf := bytes.Buffer{}
			err = graphviz.WriteTo(&buf, gr)
			g.Expect(err).NotTo(HaveOccurred())

			gg := goldie.New(t)
			gg.WithFixtureDir("testdata")
			gg.Assert(t, c.goldenName, buf.Bytes())
		})
	}
}
