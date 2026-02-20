package graphviz

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewRecord_NoParts_ReturnsEmptyRecord(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	r := newRecord()

	// Assert
	g.Expect(r).NotTo(BeNil())
	g.Expect(r.parts).To(BeEmpty())
}

func TestAdd_SinglePart_AppendsPart(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	r := newRecord()

	// Act
	r.add("alpha")

	// Assert
	g.Expect(r.parts).To(Equal([]string{"alpha"}))
}

func TestAdd_MultipleParts_PreservesOrder(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	r := newRecord()

	// Act
	r.add("alpha")
	r.add("beta")
	r.add("gamma")

	// Assert
	g.Expect(r.parts).To(Equal([]string{"alpha", "beta", "gamma"}))
}

func TestAddf_FormatString_AppendsFormattedPart(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	r := newRecord()

	// Act
	r.addf("%s-%d", "blue", 7)

	// Assert
	g.Expect(r.parts).To(Equal([]string{"blue-7"}))
}

func TestAddWrapped_TextExceedsWidth_InsertsEscapedNewlines(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	r := newRecord()

	// Act
	r.addWrapped(4, "abc def ghi")

	// Assert
	g.Expect(r.parts).To(Equal([]string{"abc \\ndef \\nghi"}))
}

func TestAddWrappedf_FormatAndWrap_InsertsEscapedNewlines(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	r := newRecord()

	// Act
	r.addWrappedf(4, "%s", "abc def ghi")

	// Assert
	g.Expect(r.parts).To(Equal([]string{"abc \\ndef \\nghi"}))
}

func TestString_WithParts_JoinsWithSeparatorsAndBraces(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	r := newRecord()
	r.add("alpha")
	r.add("beta")

	// Act
	result := r.String()

	// Assert
	g.Expect(result).To(Equal("{alpha | beta}"))
}

func TestString_NoParts_ReturnsEmptyBraces(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	r := newRecord()

	// Act
	result := r.String()

	// Assert
	g.Expect(result).To(Equal("{}"))
}
