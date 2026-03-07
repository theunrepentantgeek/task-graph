package safe

import "strings"

// Label returns text with HTML entity encoding for characters that have special
// meaning in Mermaid flowchart syntax.
func Label(text string) string {
	text = strings.ReplaceAll(text, `"`, "&quot;")
	text = strings.ReplaceAll(text, `|`, "&#124;")

	return text
}
