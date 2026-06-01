package safe

import "strings"

// labelReplacer replaces characters that have special meaning in Mermaid flowchart syntax.
// Using strings.NewReplacer processes all substitutions in a single pass.
var labelReplacer = strings.NewReplacer(
	`"`, "&quot;",
	`|`, "&#124;",
)

// Label returns text with HTML entity encoding for characters that have special
// meaning in Mermaid flowchart syntax. Graphviz record labels use a different
// encoding strategy (backslash escaping, handled by record.String()).
func Label(text string) string {
	return labelReplacer.Replace(text)
}
