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
	Line
}

// Line represents a single line of text, which can have nested lines. 
type Line struct {
	text string
	nested []*Line
}

// New creates a new instance of IndentWriter 
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
	return iw.writeTo(w, indent, 0)
}

// Add adds a new line with the provided text as a nested line under 
// the current line.
// text is the content of the line to be added. 
// It returns a pointer to the newly added line.
func (l *Line) Add(text string) *Line {
	result := &Line{
		text: text,
	}
	
	l.nested = append(l.nested, result)
	
	return result
}

// Addf adds a new line with the provided formatted text as a nested line under
// the current line. 
// It uses fmt.Sprintf to format the text with the provided arguments.
// It returns a pointer to the newly added line.
func (l *Line) Addf(format string, args ...any) *Line {
	return l.Add(fmt.Sprintf(format, args...))
}

// writeTo is a helper method that recursively writes the line and its nested lines
// to the provided io.Writer with the appropriate indentation.
func (l *Line) writeTo(
	w io.Writer, 
	indent string,
	level int,
) (int64, error) {
	var bytesWritten int64 = 0

	// Write the indent for the current level
	for range level {
		if n, err := w.Write([]byte(indent)); err != nil {
			return bytesWritten, err
		} else {
			bytesWritten += int64(n)
		}
	}

	// Write the text for the current line
	if n, err := w.Write([]byte(l.text + "\n")); err != nil {
		return bytesWritten, err
	} else {
		bytesWritten += int64(n)
	}

	// Write the nested lines with increased indentation
	for _, nestedLine := range l.nested {
		if n, err := nestedLine.writeTo(w, indent, level+1); err != nil {
			return bytesWritten, err
		} else {
			bytesWritten += n
		}
	}

	return bytesWritten, nil
}
