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

func TestCreateConfig_GroupByNamespaceFlagSetsConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{GroupByNamespace: true}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.GroupByNamespace).To(BeTrue())
}

func TestCreateConfig_GroupByNamespaceFlagOverridesConfigFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{
		Config:           filepath.Join("testdata", "config.yaml"),
		GroupByNamespace: true,
	}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.GroupByNamespace).To(BeTrue())
}

func TestCreateConfig_HighlightFlagAddsSingleStyleRule(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{Highlight: "build"}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.Graphviz.StyleRules).To(HaveLen(1))
	g.Expect(cfg.Graphviz.StyleRules[0].Match).To(Equal("build"))
	g.Expect(cfg.Graphviz.StyleRules[0].FillColor).To(Equal("yellow"))
	g.Expect(cfg.Graphviz.StyleRules[0].Style).To(Equal("filled"))
}

func TestCreateConfig_HighlightFlagWithCommaSeparatorAddsMultipleStyleRules(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{Highlight: "build,doc"}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.Graphviz.StyleRules).To(HaveLen(2))
	g.Expect(cfg.Graphviz.StyleRules[0].Match).To(Equal("build"))
	g.Expect(cfg.Graphviz.StyleRules[1].Match).To(Equal("doc"))
}

func TestCreateConfig_HighlightFlagWithSemicolonSeparatorAddsMultipleStyleRules(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{Highlight: "build;doc"}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.Graphviz.StyleRules).To(HaveLen(2))
	g.Expect(cfg.Graphviz.StyleRules[0].Match).To(Equal("build"))
	g.Expect(cfg.Graphviz.StyleRules[1].Match).To(Equal("doc"))
}

func TestCreateConfig_HighlightFlagUsesConfiguredHighlightColor(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{
		Config:    filepath.Join("testdata", "highlight_color.yaml"),
		Highlight: "build",
	}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.Graphviz.StyleRules).To(HaveLen(1))
	g.Expect(cfg.Graphviz.StyleRules[0].FillColor).To(Equal("lightblue"))
}

func TestCreateConfig_HighlightFlagWithGlobPattern(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{Highlight: "cmd:*"}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.Graphviz.StyleRules).To(HaveLen(1))
	g.Expect(cfg.Graphviz.StyleRules[0].Match).To(Equal("cmd:*"))
}
