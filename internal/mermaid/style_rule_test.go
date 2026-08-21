package mermaid

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/task-graph/internal/config"
)

func TestBuildClassDef_WithFillColor_ReturnsFillPart(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rule := config.NodeStyleRule{Match: "alpha", FillColor: "lightyellow"}

	result := buildClassDef(rule)

	g.Expect(result).To(Equal("fill:lightyellow"))
}

func TestBuildClassDef_WithColor_ReturnsStrokePart(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rule := config.NodeStyleRule{Match: "alpha", Color: "red"}

	result := buildClassDef(rule)

	g.Expect(result).To(Equal("stroke:red"))
}

func TestBuildClassDef_WithFontColor_ReturnsColorPart(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rule := config.NodeStyleRule{Match: "alpha", FontColor: "blue"}

	result := buildClassDef(rule)

	g.Expect(result).To(Equal("color:blue"))
}

func TestBuildClassDef_WithDashedStyle_ReturnsStrokeDasharray(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rule := config.NodeStyleRule{Match: "alpha", Style: "dashed"}

	result := buildClassDef(rule)

	g.Expect(result).To(Equal("stroke-dasharray:5,5"))
}

func TestBuildClassDef_WithDottedStyle_ReturnsStrokeDasharrayShort(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rule := config.NodeStyleRule{Match: "alpha", Style: "dotted"}

	result := buildClassDef(rule)

	g.Expect(result).To(Equal("stroke-dasharray:2,2"))
}

func TestBuildClassDef_WithBoldStyle_ReturnsStrokeWidth(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rule := config.NodeStyleRule{Match: "alpha", Style: "bold"}

	result := buildClassDef(rule)

	g.Expect(result).To(Equal("stroke-width:3px"))
}

func TestBuildClassDef_WithFilledStyle_NoExtraProperties(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// "filled" style in graphviz enables the fill color to show; in Mermaid, fill is always
	// applied when FillColor is set, so "filled" alone produces no Mermaid CSS properties.
	rule := config.NodeStyleRule{Match: "alpha", Style: "filled"}

	result := buildClassDef(rule)

	g.Expect(result).To(BeEmpty())
}

func TestBuildClassDef_WithDashedStyleAndFillColor_ReturnsBothParts(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rule := config.NodeStyleRule{Match: "alpha", FillColor: "lightyellow", Style: "dashed"}

	result := buildClassDef(rule)

	g.Expect(result).To(Equal("fill:lightyellow,stroke-dasharray:5,5"))
}

func TestBuildClassDef_WithNoProperties_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	rule := config.NodeStyleRule{Match: "alpha"}

	result := buildClassDef(rule)

	g.Expect(result).To(BeEmpty())
}
