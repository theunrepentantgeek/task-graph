package indentwriter

import (
	"fmt"
	"io"
)

// IndentWriter is a utility for writing indented text structures. It allows you to
// build a tree of lines, where each line can have nested lines. When writing to
// an io.Writer, the structure is output with the appropriate indentation for each
// level of nesting.
type IndentWriter struct {
	lines []*Line
}

// New creates a new instance of IndentWriter.
func New() *IndentWriter {
	return &IndentWriter{}
}

// WriteTo writes the indented structure to the provided io.Writer.
// The indent parameter specifies the string to use for each level of
// indentation. It returns the number of bytes written and any error
// that occurred.
func (iw *IndentWriter) WriteTo(
	w io.Writer,
	indent string,
) (n int64, err error) {
	var bytesWritten int64

	for _, line := range iw.lines {
		n, err := line.writeTo(w, indent, 0)
		if err != nil {
			return bytesWritten, err
		}

		bytesWritten += n
	}

	return bytesWritten, nil
}

// Add adds a new line with the provided text
// is the content of the line to be added.
// It returns a pointer to the newly added line.
func (iw *IndentWriter) Add(text string) *Line {
	result := &Line{
		text: text,
	}

	iw.lines = append(iw.lines, result)

	return result
}

// Addf adds a new line with the provided formatted text
// It uses fmt.Sprintf to format the text with the provided arguments.
// It returns a pointer to the newly added line.
func (iw *IndentWriter) Addf(format string, args ...any) *Line {
	return iw.Add(fmt.Sprintf(format, args...))
}
