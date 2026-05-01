package autocolor

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/task-graph/internal/graph"
	"github.com/theunrepentantgeek/task-graph/internal/namespace"
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

func TestGenerateRules_SingleNamespace_ReturnsTwoRules(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("build")
	gr.AddNode("cmd:build")
	gr.AddNode("cmd:test")

	// Act
	rules := GenerateRules(gr)

	// Assert: exact match + children match
	g.Expect(rules).To(HaveLen(2))
	g.Expect(rules[0].Match).To(Equal("cmd"))
	g.Expect(rules[0].FillColor).To(Equal(palette[0]))
	g.Expect(rules[0].Style).To(Equal("filled"))
	g.Expect(rules[1].Match).To(Equal("cmd[-.:]*"))
	g.Expect(rules[1].FillColor).To(Equal(palette[0]))
	g.Expect(rules[1].Style).To(Equal("filled"))
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
	// Alphabetical: "cmd" before "controllers", two rules per namespace
	g.Expect(rules).To(HaveLen(4))

	g.Expect(rules[0].Match).To(Equal("cmd"))
	g.Expect(rules[0].FillColor).To(Equal(palette[0]))
	g.Expect(rules[1].Match).To(Equal("cmd[-.:]*"))
	g.Expect(rules[1].FillColor).To(Equal(palette[0]))

	g.Expect(rules[2].Match).To(Equal("controllers"))
	g.Expect(rules[2].FillColor).To(Equal(palette[1]))
	g.Expect(rules[3].Match).To(Equal("controllers[-.:]*"))
	g.Expect(rules[3].FillColor).To(Equal(palette[1]))
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
	// Expect two rules per namespace: "cmd" (depth 0) and "cmd:test" (depth 1)
	g.Expect(rules).To(HaveLen(4))

	// "cmd" comes first (shallower), then "cmd:test"
	g.Expect(rules[0].Match).To(Equal("cmd"))
	g.Expect(rules[0].FillColor).To(Equal(palette[0]))
	g.Expect(rules[1].Match).To(Equal("cmd[-.:]*"))
	g.Expect(rules[1].FillColor).To(Equal(palette[0]))

	g.Expect(rules[2].Match).To(Equal("cmd:test"))
	g.Expect(rules[2].FillColor).To(Equal(palette[1]))
	g.Expect(rules[3].Match).To(Equal("cmd:test:*"))
	g.Expect(rules[3].FillColor).To(Equal(palette[1]))
}

func TestGenerateRules_MoreNamespacesThanPalette_CyclesColors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange: create more namespaces than palette entries
	gr := graph.New()
	for i := range len(palette) + 1 {
		gr.AddNode(string(rune('a'+i)) + ":task")
	}

	// Act
	rules := GenerateRules(gr)

	// Assert: two rules per namespace
	g.Expect(rules).To(HaveLen((len(palette) + 1) * 2))

	// The last namespace's rules should cycle back to the first palette color
	g.Expect(rules[len(palette)*2].FillColor).To(Equal(palette[0]))
	g.Expect(rules[len(palette)*2+1].FillColor).To(Equal(palette[0]))
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
	// Should have two rules per namespace: "cmd" (depth 0) and "cmd:test" (depth 1)
	g.Expect(rules).To(HaveLen(4))

	matches := make([]string, len(rules))
	for i, r := range rules {
		matches[i] = r.Match
	}

	g.Expect(matches).To(Equal([]string{"cmd", "cmd[-.:]*", "cmd:test", "cmd:test:*"}))
}

func TestGenerateRules_SingleNamespace_IncludesExpectedFieldValues(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("cmd:build")

	// Act
	rules := GenerateRules(gr)

	// Assert: two rules per namespace
	g.Expect(rules).To(HaveLen(2))

	// Check exact-match rule
	rule := rules[0]
	g.Expect(rule.Match).To(Equal("cmd"))
	g.Expect(rule.FillColor).NotTo(BeEmpty())
	g.Expect(rule.Style).To(Equal("filled"))
	g.Expect(rule.Color).To(BeEmpty())
	g.Expect(rule.FontColor).To(BeEmpty())

	// Check children-match rule
	rule = rules[1]
	g.Expect(rule.Match).To(Equal("cmd[-.:]*"))
	g.Expect(rule.FillColor).NotTo(BeEmpty())
	g.Expect(rule.Style).To(Equal("filled"))
	g.Expect(rule.Color).To(BeEmpty())
	g.Expect(rule.FontColor).To(BeEmpty())
}

func TestGenerateRules_TopLevelTaskMatchesNamespace_GetsColored(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange: tidy is a top-level task; tidy:gofumpt and tidy:lint are subtasks
	gr := graph.New()
	gr.AddNode("tidy")
	gr.AddNode("tidy:gofumpt")
	gr.AddNode("tidy:lint")

	// Act
	rules := GenerateRules(gr)

	// Assert: at least one rule should match "tidy" itself
	matched := false

	for _, rule := range rules {
		re, err := namespace.CompileMatchPattern(rule.Match)
		g.Expect(err).NotTo(HaveOccurred())

		if re.MatchString("tidy") {
			matched = true

			break
		}
	}

	g.Expect(matched).To(BeTrue(), "expected a rule matching the top-level task 'tidy'")
}

// TestGenerateRulesWithPalette

func TestGenerateRulesWithPalette_EmptyPalette_ReturnsNil(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("cmd:build")

	// Act
	rules := GenerateRulesWithPalette(gr, nil)

	// Assert
	g.Expect(rules).To(BeNil())
}

func TestGenerateRulesWithPalette_CustomPalette_UsesSuppliedColors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("cmd:build")

	customPalette := []string{"#FF0000", "#00FF00"}

	// Act
	rules := GenerateRulesWithPalette(gr, customPalette)

	// Assert: both rules use the first custom color
	g.Expect(rules).To(HaveLen(2))
	g.Expect(rules[0].FillColor).To(Equal("#FF0000"))
	g.Expect(rules[1].FillColor).To(Equal("#FF0000"))
}

func TestGenerateRulesWithPalette_ColorblindPalette_UsesHexColors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	gr := graph.New()
	gr.AddNode("ns:task")

	// Act
	rules := GenerateRulesWithPalette(gr, ColorblindPalette)

	// Assert: Okabe-Ito palette entries are hex strings
	g.Expect(rules).NotTo(BeEmpty())

	for _, r := range rules {
		g.Expect(r.FillColor).To(HavePrefix("#"), "expected hex color from Okabe-Ito palette")
	}
}
