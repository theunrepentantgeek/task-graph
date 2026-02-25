package cmd

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/task-graph/internal/config"
)

func TestCreateConfig_DefaultsWhenNoFileProvided(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg).To(Equal(config.New()))
}

func TestCreateConfig_LoadsYAMLConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{
		Config: filepath.Join("testdata", "config.yaml"),
	}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.Graphviz.Font).To(Equal("Fira Code"))
	g.Expect(cfg.Graphviz.FontSize).To(Equal(12))
	g.Expect(cfg.Graphviz.CallEdges.Color).To(Equal("red"))
	g.Expect(cfg.Graphviz.TaskNodes.Color).To(Equal("black"))
}

func TestCreateConfig_LoadsJSONConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{
		Config: filepath.Join("testdata", "config.json"),
	}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.Graphviz.Font).To(Equal("JetBrains Mono"))
	g.Expect(cfg.Graphviz.DependencyEdges.Color).To(Equal("green"))
	// Ensure defaults remain when not overridden.
	g.Expect(cfg.Graphviz.CallEdges.Style).To(Equal("dashed"))
}

func TestCreateConfig_MissingFileReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{Config: "does-not-exist.yml"}

	cfg, err := cli.CreateConfig()

	g.Expect(cfg).To(BeNil())
	g.Expect(err).To(MatchError(ContainSubstring("failed to read config file")))
}
