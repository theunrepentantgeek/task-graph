package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"gopkg.in/yaml.v3"

	"github.com/theunrepentantgeek/task-graph/internal/config"
	"github.com/theunrepentantgeek/task-graph/internal/graph"
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

func TestCreateConfig_AutoColorFlagSetsConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{AutoColor: true}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.AutoColor).To(BeTrue())
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

	cli := CLI{
		Highlight: "build",
	}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.NodeStyleRules).To(HaveLen(1))

	rule := cfg.NodeStyleRules[0]
	g.Expect(rule.Match).To(Equal("build"))
	g.Expect(rule.FillColor).To(Equal("yellow"))
	g.Expect(rule.Style).To(Equal("filled"))
}

func TestCreateConfig_HighlightFlagWithCommaSeparatorAddsMultipleStyleRules(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{Highlight: "build,doc"}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.NodeStyleRules).To(HaveLen(2))
	g.Expect(cfg.NodeStyleRules[0].Match).To(Equal("build"))
	g.Expect(cfg.NodeStyleRules[1].Match).To(Equal("doc"))
}

func TestCreateConfig_HighlightFlagWithSemicolonSeparatorAddsMultipleStyleRules(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{Highlight: "build;doc"}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.NodeStyleRules).To(HaveLen(2))
	g.Expect(cfg.NodeStyleRules[0].Match).To(Equal("build"))
	g.Expect(cfg.NodeStyleRules[1].Match).To(Equal("doc"))
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
	g.Expect(cfg.NodeStyleRules).To(HaveLen(1))
	g.Expect(cfg.NodeStyleRules[0].FillColor).To(Equal("lightblue"))
}

func TestCreateConfig_HighlightFlagWithGlobPattern(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{Highlight: "cmd:*"}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.NodeStyleRules).To(HaveLen(1))
	g.Expect(cfg.NodeStyleRules[0].Match).To(Equal("cmd:*"))
}

func TestCreateConfig_IncludeGlobalVarsFlagSetsConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{IncludeGlobalVars: true}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.IncludeGlobalVars).To(BeTrue())
}

func TestCreateConfig_DefaultIncludeGlobalVarsIsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.IncludeGlobalVars).To(BeFalse())
}

// TestApplyAutoColor

func TestApplyAutoColor_WhenDisabled_LeavesRulesUnchanged(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	cfg := config.New()
	cfg.AutoColor = false
	cfg.NodeStyleRules = []config.NodeStyleRule{
		{Match: "user:*", FillColor: "red", Style: "filled"},
	}

	gr := graph.New()
	gr.AddNode("cmd:build")

	// Act
	applyAutoColor(cfg, gr)

	// Assert
	g.Expect(cfg.NodeStyleRules).To(HaveLen(1))
	g.Expect(cfg.NodeStyleRules[0].Match).To(Equal("user:*"))
}

func TestApplyAutoColor_WithNamespacedGraph_GeneratesRulesForNamespaces(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	cfg := config.New()
	cfg.AutoColor = true

	gr := graph.New()
	gr.AddNode("cmd:build")
	gr.AddNode("cmd:test")
	gr.AddNode("controllers:deploy")

	// Act
	applyAutoColor(cfg, gr)

	// Assert
	matches := make([]string, len(cfg.NodeStyleRules))
	for i, r := range cfg.NodeStyleRules {
		matches[i] = r.Match
	}

	g.Expect(matches).To(ContainElements("cmd[-.:]*", "controllers[-.:]*"))
}

func TestApplyAutoColor_AutoRulesPrependedBeforeUserRules(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	cfg := config.New()
	cfg.AutoColor = true
	cfg.NodeStyleRules = []config.NodeStyleRule{
		{Match: "cmd:build", FillColor: "gold", Style: "filled"},
	}

	gr := graph.New()
	gr.AddNode("cmd:build")
	gr.AddNode("cmd:test")

	// Act
	applyAutoColor(cfg, gr)

	// Assert: auto-generated rules come first, user-defined rule comes last
	g.Expect(cfg.NodeStyleRules).To(HaveLen(3))

	exactRule := cfg.NodeStyleRules[0]
	g.Expect(exactRule.Match).To(Equal("cmd"))

	childrenRule := cfg.NodeStyleRules[1]
	g.Expect(childrenRule.Match).To(Equal("cmd[-.:]*"))

	userRule := cfg.NodeStyleRules[2]
	g.Expect(userRule.Match).To(Equal("cmd:build"))
	g.Expect(userRule.FillColor).To(Equal("gold"))
}

// TestExportConfigToFile

func TestExportConfigToFile_NoExportPathDoesNothing(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{}
	cfg := config.New()

	err := cli.ExportConfigToFile(cfg)

	g.Expect(err).NotTo(HaveOccurred())
}

func TestExportConfigToFile_WritesYAMLFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.yaml")

	cli := CLI{ExportConfig: outFile}
	cfg := config.New()

	err := cli.ExportConfigToFile(cfg)

	g.Expect(err).NotTo(HaveOccurred())

	raw, readErr := os.ReadFile(outFile)
	g.Expect(readErr).NotTo(HaveOccurred())

	var loaded config.Config
	g.Expect(yaml.Unmarshal(raw, &loaded)).To(Succeed())
	g.Expect(loaded.Graphviz.Font).To(Equal(cfg.Graphviz.Font))
	g.Expect(loaded.Graphviz.FontSize).To(Equal(cfg.Graphviz.FontSize))
}

func TestExportConfigToFile_WritesJSONFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.json")

	cli := CLI{ExportConfig: outFile}
	cfg := config.New()

	err := cli.ExportConfigToFile(cfg)

	g.Expect(err).NotTo(HaveOccurred())

	raw, readErr := os.ReadFile(outFile)
	g.Expect(readErr).NotTo(HaveOccurred())

	var loaded config.Config
	g.Expect(json.Unmarshal(raw, &loaded)).To(Succeed())
	g.Expect(loaded.Graphviz.Font).To(Equal(cfg.Graphviz.Font))
	g.Expect(loaded.Graphviz.FontSize).To(Equal(cfg.Graphviz.FontSize))
}

func TestExportConfigToFile_WritesYMLFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.yml")

	cli := CLI{ExportConfig: outFile}
	cfg := config.New()

	err := cli.ExportConfigToFile(cfg)

	g.Expect(err).NotTo(HaveOccurred())

	raw, readErr := os.ReadFile(outFile)
	g.Expect(readErr).NotTo(HaveOccurred())

	var loaded config.Config
	g.Expect(yaml.Unmarshal(raw, &loaded)).To(Succeed())
	g.Expect(loaded.Graphviz.Font).To(Equal(cfg.Graphviz.Font))
	g.Expect(loaded.Graphviz.FontSize).To(Equal(cfg.Graphviz.FontSize))
}

func TestExportConfigToFile_InvalidPathReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{ExportConfig: "/nonexistent-dir/out.yaml"}
	cfg := config.New()

	err := cli.ExportConfigToFile(cfg)

	g.Expect(err).To(MatchError(ContainSubstring("failed to write config file")))
}

func TestExportConfigToFile_UnknownExtensionReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	dir := t.TempDir()
	cli := CLI{ExportConfig: filepath.Join(dir, "out.txt")}
	cfg := config.New()

	err := cli.ExportConfigToFile(cfg)

	g.Expect(err).To(MatchError(ContainSubstring("unsupported file extension")))
}

// TestCreateConfig_GraphTypeFlagSetsConfig

func TestCreateConfig_GraphTypeFlagSetsConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{GraphType: graphTypeMermaid}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.GraphType).To(Equal(graphTypeMermaid))
}

// TestLoadConfigFile

func TestLoadConfigFile_UnknownExtensionFallsBackToYAML(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{
		Config: filepath.Join("testdata", "config.conf"),
	}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.Graphviz.Font).To(Equal("Courier New"))
	g.Expect(cfg.Graphviz.FontSize).To(Equal(11))
}

// TestResolveGraphType

func TestResolveGraphType_DefaultsToDot(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{}
	flags := &Flags{Config: config.New()}

	result := cli.resolveGraphType(flags)

	g.Expect(result).To(Equal(graphTypeDot))
}

func TestResolveGraphType_CLIFlagTakesPrecedence(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{GraphType: graphTypeMermaid}
	cfg := config.New()
	cfg.GraphType = graphTypeDot
	flags := &Flags{Config: cfg}

	result := cli.resolveGraphType(flags)

	g.Expect(result).To(Equal(graphTypeMermaid))
}

func TestResolveGraphType_FallsBackToConfigWhenFlagEmpty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{}
	cfg := config.New()
	cfg.GraphType = graphTypeMermaid
	flags := &Flags{Config: cfg}

	result := cli.resolveGraphType(flags)

	g.Expect(result).To(Equal(graphTypeMermaid))
}

// TestApplyFocus

func TestApplyFocus_NoMatchingPatterns_ReturnsSameGraph(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("build")
	gr.AddNode("test")

	// Act
	result, err := applyFocus(gr, "deploy")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(gr))
}

func TestApplyFocus_ExactMatch_ReturnsMatchedNode(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("build")
	gr.AddNode("test")
	gr.AddNode("deploy")

	// Act
	result, err := applyFocus(gr, "build")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())

	ids := collectNodeIDs(result)
	g.Expect(ids).To(ConsistOf("build"))
}

func TestApplyFocus_GlobPattern_MatchesMultipleNodes(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("cmd:build")
	gr.AddNode("cmd:test")
	gr.AddNode("api:serve")

	// Act
	result, err := applyFocus(gr, "cmd:*")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())

	ids := collectNodeIDs(result)
	g.Expect(ids).To(ConsistOf("cmd:build", "cmd:test"))
}

func TestApplyFocus_IncludesTransitiveDependencies(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	compile := gr.AddNode("compile")
	testNode := gr.AddNode("test")
	deploy := gr.AddNode("deploy")

	compile.AddEdge(testNode)
	testNode.AddEdge(deploy)

	// Act — focus on "test"; should include "compile" (dependency) and "deploy" (dependent)
	result, err := applyFocus(gr, "test")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())

	ids := collectNodeIDs(result)
	g.Expect(ids).To(ConsistOf("compile", "test", "deploy"))
}

func TestApplyFocus_MultiplePatternsSeparatedByComma(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("build")
	gr.AddNode("test")
	gr.AddNode("deploy")

	// Act
	result, err := applyFocus(gr, "build,test")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())

	ids := collectNodeIDs(result)
	g.Expect(ids).To(ConsistOf("build", "test"))
}

// collectNodeIDs returns all node IDs from a graph as a slice.
func collectNodeIDs(gr *graph.Graph) []string {
	var ids []string

	for n := range gr.Nodes() {
		ids = append(ids, n.ID())
	}

	return ids
}
