package autocolor

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/task-graph/internal/graph"
)

func TestGenerateRules_NoNamespaces_ReturnsEmptyRules(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("build")
	gr.AddNode("test")
	gr.AddNode("lint")

	// Act
	rules := GenerateRules(gr)

	// Assert
	g.Expect(rules).To(BeEmpty())
}

func TestGenerateRules_SingleNamespace_ReturnsOneRule(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("build")
	gr.AddNode("cmd:build")
	gr.AddNode("cmd:test")

	// Act
	rules := GenerateRules(gr)

	// Assert
	g.Expect(rules).To(HaveLen(1))
	g.Expect(rules[0].Match).To(Equal("cmd:*"))
	g.Expect(rules[0].FillColor).To(Equal(Palette[0]))
	g.Expect(rules[0].Style).To(Equal("filled"))
}

func TestGenerateRules_MultipleNamespaces_AssignsColorsAlphabetically(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("controllers:build")
	gr.AddNode("cmd:build")

	// Act
	rules := GenerateRules(gr)

	// Assert
	// Alphabetical: "cmd" before "controllers"
	g.Expect(rules).To(HaveLen(2))

	cmd := rules[0]
	g.Expect(cmd.Match).To(Equal("cmd:*"))
	g.Expect(cmd.FillColor).To(Equal(Palette[0]))

	controllers := rules[1]
	g.Expect(controllers.Match).To(Equal("controllers:*"))
	g.Expect(controllers.FillColor).To(Equal(Palette[1]))
}

func TestGenerateRules_NestedNamespaces_GeneratesRulesForAll(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("cmd:build")
	gr.AddNode("cmd:test:unit")
	gr.AddNode("cmd:test:golden")

	// Act
	rules := GenerateRules(gr)

	// Assert
	// Expect rules for "cmd" (depth 0) and "cmd:test" (depth 1)
	g.Expect(rules).To(HaveLen(2))

	// "cmd" comes first (shallower), then "cmd:test"
	cmd := rules[0]
	g.Expect(cmd.Match).To(Equal("cmd:*"))
	g.Expect(cmd.FillColor).To(Equal(Palette[0]))

	cmdTest := rules[1]
	g.Expect(cmdTest.Match).To(Equal("cmd:test:*"))
	g.Expect(cmdTest.FillColor).To(Equal(Palette[1]))
}

func TestGenerateRules_MoreNamespacesThanPalette_CyclesColors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange: create more namespaces than palette entries
	gr := graph.New()
	for i := range len(Palette) + 1 {
		gr.AddNode(string(rune('a'+i)) + ":task")
	}

	// Act
	rules := GenerateRules(gr)

	// Assert
	g.Expect(rules).To(HaveLen(len(Palette) + 1))

	// The last rule should cycle back to the first palette color
	g.Expect(rules[len(Palette)].FillColor).To(Equal(Palette[0]))
}

func TestGenerateRules_EmptyGraph_ReturnsEmptyRules(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()

	// Act
	rules := GenerateRules(gr)

	// Assert
	g.Expect(rules).To(BeEmpty())
}

func TestGenerateRules_AllRulesHaveFilledStyle(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("ns1:task")
	gr.AddNode("ns2:task")

	// Act
	rules := GenerateRules(gr)

	// Assert
	for _, rule := range rules {
		g.Expect(rule.Style).To(Equal("filled"))
	}
}

func TestGenerateRules_MixedTopLevelAndNested_ShallowerFirst(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("build")           // no namespace
	gr.AddNode("cmd:build")       // namespace: cmd
	gr.AddNode("cmd:test:unit")   // namespaces: cmd, cmd:test
	gr.AddNode("cmd:test:golden") // namespaces: cmd, cmd:test

	// Act
	rules := GenerateRules(gr)

	// Assert
	// Should have rules for "cmd" (depth 0) and "cmd:test" (depth 1)
	g.Expect(rules).To(HaveLen(2))

	matches := make([]string, len(rules))
	for i, r := range rules {
		matches[i] = r.Match
	}

	g.Expect(matches).To(Equal([]string{"cmd:*", "cmd:test:*"}))
}

func TestGenerateRules_SingleNamespace_IncludesExpectedFieldValues(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("cmd:build")

	// Act
	rules := GenerateRules(gr)

	// Assert
	g.Expect(rules).To(HaveLen(1))

	rule := rules[0]
	g.Expect(rule.Match).To(Equal("cmd:*"))
	g.Expect(rule.FillColor).NotTo(BeEmpty())
	g.Expect(rule.Style).To(Equal("filled"))
	g.Expect(rule.Color).To(BeEmpty())
	g.Expect(rule.FontColor).To(BeEmpty())
}
