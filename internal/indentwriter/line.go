package indentwriter

import (
	"fmt"
	"io"

	"github.com/rotisserie/eris"
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
	bytesWritten, err := l.writeIndent(w, indent, level)
	if err != nil {
		return bytesWritten, eris.Wrapf(err, "failed to write indent for level %d", level)
	}

	// Write the text for the current line
	n, err := w.Write([]byte(l.text + "\n"))
	if err != nil {
		return bytesWritten, eris.Wrapf(err, "failed to write line: %s", l.text)
	}

	bytesWritten += int64(n)

	// Write the nested lines with increased indentation
	for _, nestedLine := range l.nested {
		n, err := nestedLine.writeTo(w, indent, level+1)
		if err != nil {
			return bytesWritten, eris.Wrapf(err, "failed to write nested line: %s", nestedLine.text)
		}

		bytesWritten += n
	}

	return bytesWritten, nil
}

func (*Line) writeIndent(
	w io.Writer,
	indent string,
	level int,
) (int64, error) {
	var bytesWritten int64

	// Write the indent for the current level
	for range level {
		n, err := w.Write([]byte(indent))
		if err != nil {
			return bytesWritten, eris.Wrapf(err, "failed to write indent for level %d", level)
		}

		bytesWritten += int64(n)
	}

	return bytesWritten, nil
}
