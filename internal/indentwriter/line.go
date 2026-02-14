package indentwriter

import (
	"fmt"
	"io"
)

// Line represents a single line of text, which can have nested lines.
type Line struct {
	text   string
	nested []*Line
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

	if level > 0 {
		// Write the indent for the current level
		for range level {
			if n, err := w.Write([]byte(indent)); err != nil {
				return bytesWritten, err
			} else {
				bytesWritten += int64(n)
			}
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
