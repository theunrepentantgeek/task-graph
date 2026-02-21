package graphviz

import (
	"fmt"
	"strings"

	"github.com/theunrepentantgeek/task-graph/internal/indentwriter"
)

// record represents a record-shaped node in the Graphviz output.
type record struct {
	parts []string
}

// newRecord creates a new record with no parts.
func newRecord() *record {
	return &record{}
}

// add adds a part to the record.
func (r *record) add(text string) {
	if len(text) > 0 {
		r.parts = append(r.parts, text)
	}
}

// addf adds a formatted part to the record.
func (r *record) addf(format string, args ...any) {
	r.add(fmt.Sprintf(format, args...))
}

// addWrapped adds a part to the record, wrapping the text to the specified width.
func (r *record) addWrapped(width int, text string) {
	lines := indentwriter.WordWrap(text, width)
	r.add(strings.Join(lines, "\\n"))
}

// addWrappedf adds a formatted part to the record, wrapping the text to the specified width.
func (r *record) addWrappedf(width int, format string, args ...any) {
	r.addWrapped(width, fmt.Sprintf(format, args...))
}

var escapings = map[string]string{
	`{`: `\{`,
	`}`: `\}`,
	`"`: `\"`,
}

// String returns the string representation of the record, which is the parts joined by " | ".
func (r *record) String() string {
	if len(r.parts) == 1 {
		return r.parts[0]
	}

	content := strings.Join(r.parts, " | ")
	for s, r := range escapings {
		content = strings.ReplaceAll(content, s, r)
	}

	return fmt.Sprintf("{%s}", content)
}
