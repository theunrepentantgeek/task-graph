package taskgraph

import (
	"regexp"
)

// templateBlockRe matches Go template blocks: {{ ... }}
var templateBlockRe = regexp.MustCompile(`\{\{(.+?)\}\}`)

// varRefRe matches variable references within a template block: .IDENTIFIER
// It requires at least one character after the dot to avoid matching {{.}} (current context).
var varRefRe = regexp.MustCompile(`\.([A-Za-z_][A-Za-z0-9_]*)`)

// extractVarRefs extracts unique variable names referenced in Go template expressions
// within the given string. It finds all {{ ... }} blocks and extracts .IDENTIFIER
// patterns from each block.
func extractVarRefs(s string) []string {
	seen := make(map[string]bool)

	for _, blockMatch := range templateBlockRe.FindAllStringSubmatch(s, -1) {
		block := blockMatch[1]
		for _, varMatch := range varRefRe.FindAllStringSubmatch(block, -1) {
			name := varMatch[1]
			seen[name] = true
		}
	}

	result := make([]string, 0, len(seen))
	for name := range seen {
		result = append(result, name)
	}

	return result
}
