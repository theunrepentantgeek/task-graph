package safe

import (
	"fmt"
	"strings"
)

// Registry maintains a cache of safe identifiers, ensuring that each unique input
// produces a unique output, with already-safe names taking precedence.
type Registry struct {
	claimed map[string]string // safeID -> original name that claimed it
	results map[string]string // original name -> assigned safe ID
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		claimed: make(map[string]string),
		results: make(map[string]string),
	}
}

// Prepare pre-registers all names that are already valid safe identifiers so they
// take precedence over names that require sanitization. Must be called with all
// names before calling ID() to guarantee deterministic results.
func (r *Registry) Prepare(names []string) {
	for _, name := range names {
		if IsValid(name) {
			if _, ok := r.results[name]; !ok {
				r.claimed[name] = name
				r.results[name] = name
			}
		}
	}
}

// ID returns a safe identifier for the given name. If two different names produce
// the same sanitized identifier, a unique letter suffix is appended to disambiguate.
func (r *Registry) ID(name string) string {
	if result, ok := r.results[name]; ok {
		return result
	}

	base := Sanitize(name)
	result := r.claim(name, base)
	r.results[name] = result

	return result
}

// IDWithPrefix returns a safe identifier with the given prefix prepended before
// transformation, reducing name collisions between different identifier namespaces.
func (r *Registry) IDWithPrefix(prefix, name string) string {
	return r.ID(prefix + name)
}

// claim assigns base (or a disambiguated variant) to the given original name.
func (r *Registry) claim(original, base string) string {
	if claimer, ok := r.claimed[base]; !ok || claimer == original {
		r.claimed[base] = original
		return base
	}

	for _, suffix := range knownSuffixes {
		candidate := base + "_" + suffix
		if claimer, ok := r.claimed[candidate]; !ok || claimer == original {
			r.claimed[candidate] = original
			return candidate
		}
	}

	// Fall back to numeric suffixes for extreme collision scenarios
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if claimer, ok := r.claimed[candidate]; !ok || claimer == original {
			r.claimed[candidate] = original
			return candidate
		}
	}
}

// IsValid returns true if name is already a valid safe identifier requiring no transformation.
func IsValid(name string) bool {
	if len(name) == 0 {
		return false
	}

	runes := []rune(name)
	if !isValidStart(runes[0]) {
		return false
	}

	for _, ch := range runes[1:] {
		if !isValidMiddle(ch) {
			return false
		}
	}

	return true
}

// Sanitize converts name to a safe identifier by replacing invalid characters with underscores.
// A leading digit is preserved by prepending an underscore.
func Sanitize(name string) string {
	if name == "" {
		return "_"
	}

	runes := []rune(name)

	var b strings.Builder

	first := runes[0]

	switch {
	case isValidStart(first):
		b.WriteRune(first)
	case isDigit(first):
		// Keep digit but prepend underscore so the result starts with a valid char
		b.WriteRune('_')
		b.WriteRune(first)
	default:
		b.WriteRune('_')
	}

	for _, ch := range runes[1:] {
		if isValidMiddle(ch) {
			b.WriteRune(ch)
		} else {
			b.WriteRune('_')
		}
	}

	return b.String()
}

func isValidStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isValidMiddle(ch rune) bool {
	return isValidStart(ch) || isDigit(ch) || ch == '-'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// knownSuffixes contains the letter-based disambiguation suffixes in order: A–Z, AA–ZZ.
var knownSuffixes = buildSuffixes()

func buildSuffixes() []string {
	suffixes := make([]string, 0, 26+26*26)

	for c := 'A'; c <= 'Z'; c++ {
		suffixes = append(suffixes, string(c))
	}

	for c1 := 'A'; c1 <= 'Z'; c1++ {
		for c2 := 'A'; c2 <= 'Z'; c2++ {
			suffixes = append(suffixes, string([]rune{c1, c2}))
		}
	}

	return suffixes
}
