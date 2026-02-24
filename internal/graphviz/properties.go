package graphviz

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/theunrepentantgeek/task-graph/internal/indentwriter"
)

// properties represents the properties of a node or edge in the Graphviz output.
type properties map[string]string

func newProperties() properties {
	return make(map[string]string)
}

// Add adds a property with the given key and value to the properties map.
func (p properties) Add(
	key string,
	value string,
) {
	p[key] = value
}

// Addf adds a property with the given key and value to the properties map.
// The value is formatted using fmt.Sprintf.
func (p properties) Addf(
	key string,
	format string,
	args ...any,
) {
	p.Add(key, fmt.Sprintf(format, args...))
}

// AddWrapped adds a property with the given key and value to the properties map.
// The value is wrapped to the specified width using indentwriter.WordWrap.
func (p properties) AddWrapped(
	key string,
	width int,
	value string,
) {
	lines := indentwriter.WordWrap(value, width)
	p[key] = strings.Join(lines, "\\n")
}

// AddWrappedf adds a property with the given key and value to the properties map.
// The value is formatted using fmt.Sprintf and then wrapped to the specified width using indentwriter.WordWrap.
func (p properties) AddWrappedf(
	key string,
	width int,
	format string,
	args ...any,
) {
	p.AddWrapped(key, width, fmt.Sprintf(format, args...))
}

// WriteTo writes the properties to the given indentwriter.Line in the format expected by Graphviz.
func (p properties) WriteTo(
	label string,
	root *indentwriter.Line,
) {
	if len(p) == 0 {
		// No properties to write, so just write the node label.
		root.Add(label)

		return
	}

	nested := root.Addf("%s [", label)

	keys := slices.Collect(maps.Keys(p))
	sort.Strings(keys)

	for _, key := range keys {
		value := p[key]
		nested.Addf("%s=\"%s\"", key, value)
	}

	root.Add("]")
}
