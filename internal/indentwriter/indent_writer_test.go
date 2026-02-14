package indentwriter

import (
	"bytes"
	"testing"

	"github.com/onsi/gomega"
	"github.com/sebdah/goldie/v2"
)

// WriteTo
func TestIndentWriter_WriteTo_EmptyWriter_WritesBlankLine(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	iw := New()
	var buf bytes.Buffer

	// Act
	bytesWritten, err := iw.WriteTo(&buf, "  ")

	// Assert
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bytesWritten).To(gomega.Equal(int64(buf.Len())))

	gg := goldie.New(t)
	gg.WithFixtureDir("testdata")
	gg.Assert(t, "empty_writer", buf.Bytes())
}

func TestIndentWriter_WriteTo_NestedLines_WritesIndentedStructure(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	iw := New()
	firstLine := iw.Add("line 1")
	firstLine.Add("child 1")
	firstLine.Addf("child %d", 2)
	iw.Add("line 2")

	var buf bytes.Buffer

	// Act
	bytesWritten, err := iw.WriteTo(&buf, "  ")

	// Assert
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bytesWritten).To(gomega.Equal(int64(buf.Len())))

	gg := goldie.New(t)
	gg.WithFixtureDir("testdata")
	gg.Assert(t, "nested_lines", buf.Bytes())
}
