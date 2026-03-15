package namespace

import (
	"regexp"
	"strings"

	"github.com/rotisserie/eris"
)

// formalDelimiter is the go-task namespace delimiter (tier 1).
const formalDelimiter = ":"

// informalDelimiters are the characters treated as namespace separators
// when no formal delimiter is present (tier 2).
var informalDelimiters = "-."

// Namespace returns the namespace portion of a node ID.
// Tier 1: if id contains ":", returns everything before the last ":".
// Tier 2: otherwise, returns everything before the last "-" or ".".
// Returns "" if no delimiter is found.
func Namespace(id string) string {
	if strings.Contains(id, formalDelimiter) {
		idx := strings.LastIndex(id, formalDelimiter)

		return id[:idx]
	}

	idx := strings.LastIndexAny(id, informalDelimiters)
	if idx < 0 {
		return ""
	}

	return id[:idx]
}

// Parent returns the parent of a namespace string.
// Uses the same two-tier logic as Namespace.
// Returns "" if the namespace has no parent.
func Parent(ns string) string {
	if strings.Contains(ns, formalDelimiter) {
		idx := strings.LastIndex(ns, formalDelimiter)

		return ns[:idx]
	}

	idx := strings.LastIndexAny(ns, informalDelimiters)
	if idx < 0 {
		return ""
	}

	return ns[:idx]
}

// Depth returns the nesting depth of a namespace.
// Counts all delimiters within the namespace's tier:
//
//	Tier 1 (contains ":"): counts colons.
//	Tier 2 (no ":"): counts hyphens and dots.
//
// A top-level namespace (no internal delimiters) has depth 0.
func Depth(ns string) int {
	if strings.Contains(ns, formalDelimiter) {
		return strings.Count(ns, formalDelimiter)
	}

	count := 0

	for _, c := range ns {
		if strings.ContainsRune(informalDelimiters, c) {
			count++
		}
	}

	return count
}

// CompileMatchPattern converts a glob-style pattern (using *, ?, and [...])
// to a compiled regexp. Returns an error if the resulting regex is invalid.
func CompileMatchPattern(pattern string) (*regexp.Regexp, error) {
	var b strings.Builder

	b.WriteString("^")

	inBracket := false

	for i := range len(pattern) {
		c := pattern[i]
		inBracket = convertGlobChar(&b, c, inBracket)
	}

	b.WriteString("$")

	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil, eris.Wrapf(err, "failed to compile pattern %q", pattern)
	}

	return re, nil
}

// convertGlobChar writes one glob character to the regex builder and returns
// the updated bracket state.
func convertGlobChar(b *strings.Builder, c byte, inBracket bool) bool {
	switch {
	case c == '[' && !inBracket:
		b.WriteByte(c)

		return true
	case c == ']' && inBracket:
		b.WriteByte(c)

		return false
	case inBracket:
		b.WriteByte(c)
	case c == '*':
		b.WriteString(".*")
	case c == '?':
		b.WriteByte('.')
	default:
		b.WriteString(regexp.QuoteMeta(string(c)))
	}

	return inBracket
}

// MatchPattern returns a glob-style pattern string matching all nodes
// in the given namespace. The returned pattern is intended for storage
// in NodeStyleRule.Match and will be compiled via CompileMatchPattern.
//
// For namespaces containing ":" (clearly formal): returns "ns:*"
// For namespaces without ":" (ambiguous tier): returns "ns[-.:]*".
func MatchPattern(ns string) string {
	if strings.Contains(ns, formalDelimiter) {
		return ns + ":*"
	}

	return ns + "[-.:]*"
}
