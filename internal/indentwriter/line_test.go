package indentwriter

import (
	"errors"
	"io"
	"testing"

	"github.com/onsi/gomega"
)

// failAfterWriter is an io.Writer that succeeds for the first n bytes, then
// returns an error for every subsequent write.
type failAfterWriter struct {
	remaining int
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return 0, errors.New("write failed")
	}

	if len(p) <= w.remaining {
		w.remaining -= len(p)

		return len(p), nil
	}

	n := w.remaining
	w.remaining = 0

	return n, errors.New("write failed")
}

var _ io.Writer = &failAfterWriter{}

// WriteTo — error paths.

func TestIndentWriter_WriteTo_WriteFailsOnFirstLine_PropagatesError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	iw := New()
	iw.Add("hello")

	w := &failAfterWriter{remaining: 0} // fail immediately

	// Act
	_, err := iw.WriteTo(w, "  ")

	// Assert
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("write failed"))
}

func TestIndentWriter_WriteTo_WriteFailsOnNestedLine_PropagatesError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	iw := New()
	parent := iw.Add("parent")
	parent.Add("child")

	// Allow parent line + newline to succeed, fail on the indent for the child.
	// "parent\n" = 7 bytes; then the indent "  " is the next write.
	w := &failAfterWriter{remaining: 7}

	// Act
	_, err := iw.WriteTo(w, "  ")

	// Assert
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("write failed"))
}

// writeIndent — error path.

func TestLine_writeIndent_WriteFailsAtDepth_PropagatesError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	line := &Line{text: "text"}
	w := &failAfterWriter{remaining: 0} // fail on first indent write

	// Act
	_, err := line.writeIndent(w, "  ", 1)

	// Assert
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("write failed"))
}

// writeTo — error path when writing text fails after indent succeeds.

func TestLine_writeTo_WriteFailsAfterIndent_PropagatesError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange: level=1, indent="  " (2 bytes written for indent), then text write fails.
	line := &Line{text: "hello"}
	w := &failAfterWriter{remaining: 2} // indent succeeds, text write fails

	// Act
	_, err := line.writeTo(w, "  ", 1)

	// Assert
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("write failed"))
}
