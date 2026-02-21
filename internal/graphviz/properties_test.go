package graphviz

import (
	"bytes"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/task-graph/internal/indentwriter"
)

func TestNewProperties_EmptyMap_ReturnsEmptyMap(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	p := newProperties()

	// Assert
	g.Expect(p).NotTo(BeNil())
	g.Expect(p).To(BeEmpty())
}

func TestAdd_NewKey_AddsValue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newProperties()

	// Act
	p.Add("color", "red")

	// Assert
	g.Expect(p).To(HaveKeyWithValue("color", "red"))
}

func TestAdd_ExistingKey_OverwritesPreviousValue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newProperties()
	p["color"] = "red"

	// Act
	p.Add("color", "blue")

	// Assert
	g.Expect(p).To(HaveKeyWithValue("color", "blue"))
}

func TestAddf_NewKey_AddsFormattedValue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newProperties()

	// Act
	p.Addf("color", "%s-%d", "blue", 7)

	// Assert
	g.Expect(p).To(HaveKeyWithValue("color", "blue-7"))
}

func TestAddf_ExistingKey_OverwritesPreviousValue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newProperties()
	p["color"] = "red"

	// Act
	p.Addf("color", "%s", "blue")

	// Assert
	g.Expect(p).To(HaveKeyWithValue("color", "blue"))
}

func TestAddWrapped_TextExceedsWidth_WrapsWithEmbeddedNewlines(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newProperties()

	// Act
	p.AddWrapped("description", 4, "abc def ghi")

	// Assert
	g.Expect(p).To(HaveKeyWithValue("description", "abc \\ndef \\nghi"))
}

func TestAddWrapped_EmptyValue_StoresEmptyString(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newProperties()

	// Act
	p.AddWrapped("note", 0, "")

	// Assert
	g.Expect(p).To(HaveKeyWithValue("note", ""))
}

func TestAddWrappedf_TextExceedsWidth_WrapsWithEmbeddedNewlines(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newProperties()

	// Act
	p.AddWrappedf("description", 4, "%s", "abc def ghi")

	// Assert
	g.Expect(p).To(HaveKeyWithValue("description", "abc \\ndef \\nghi"))
}

func TestAddWrappedf_TextWithinWidth_StoresSingleLine(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	p := newProperties()

	// Act
	p.AddWrappedf("summary", 40, "%s", "short text")

	// Assert
	g.Expect(p).To(HaveKeyWithValue("summary", "short text"))
}

func TestWriteTo_UnsortedKeys_WritesKeysInAscendingOrder(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	props := properties{
		"beta":  "two",
		"alpha": "one",
		"gamma": "three",
	}

	iw := indentwriter.New()
	root := iw.Add("root")

	// Act
	props.WriteTo("node", root)

	buf := bytes.Buffer{}
	_, err := iw.WriteTo(&buf, "  ")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(buf.String()).
		To(Equal("root\n  \"node\" [\n    alpha=\"one\"\n    beta=\"two\"\n    gamma=\"three\"\n  ]\n"))
}

func TestWriteTo_NoProperties_WritesLabelWithEmptyBlock(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	props := newProperties()

	iw := indentwriter.New()
	root := iw.Add("root")

	// Act
	props.WriteTo("node", root)

	buf := bytes.Buffer{}
	_, err := iw.WriteTo(&buf, "  ")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(buf.String()).To(Equal("root\n  \"node\"\n"))
}
