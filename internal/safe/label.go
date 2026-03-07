package safe

import "strings"

// Label returns text with HTML entity encoding for characters that have special
// meaning in Mermaid flowchart syntax. Graphviz record labels use a different
// encoding strategy (backslash escaping, handled by record.String()).
func Label(text string) string {
	text = strings.ReplaceAll(text, `"`, "&quot;")
	text = strings.ReplaceAll(text, `|`, "&#124;")

	return text
}
