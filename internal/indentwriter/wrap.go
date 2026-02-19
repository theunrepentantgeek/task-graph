package indentwriter

import "strings"

/*
 * Borrowed under MIT license from
 * https://github.com/Azure/azure-service-operator/blob/main/v2/tools/generator/internal/astbuilder/comments.go
 */

// WordWrap applies word wrapping to the specified string, returning a slice containing the lines.
func WordWrap(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	expectedLines := len(text)/width + 1
	result := make([]string, 0, expectedLines)

	if width == 0 && text == "" {
		return []string{}
	}

	start := 0
	for start < len(text) {
		finish := findBreakPoint(text, start, width)
		result = append(result, text[start:finish+1])
		start = finish + 1
	}

	return result
}

// findBreakPoint finds the character at which to break two lines
// Returned index points to the last character that should be included on the line
// If breaking at a space, this will give a trailing space, but allows for
// breaking at other points too as no characters will be omitted.
func findBreakPoint(line string, start int, width int) int {
	limit := start + width + 1
	if limit >= len(line) {
		return len(line) - 1
	}

	// Look for a word break within the line
	index := strings.LastIndex(line[start:limit], " ")
	if index >= 0 {
		return start + index
	}

	// Line contains continuous text, we don't want to break it in two, so find the end of it
	index = strings.Index(line[limit:], " ")
	if index >= 0 {
		return limit + index
	}

	return len(line) - 1
}
