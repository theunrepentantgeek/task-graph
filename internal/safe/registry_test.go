package safe_test

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/theunrepentantgeek/task-graph/internal/safe"
)

// TestIsValid_AlreadySafeNames_ReturnsTrue tests whether IsValid correctly identifies already-safe identifiers.
func TestIsValid_AlreadySafeNames_ReturnsTrue(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		name string
	}{
		"simple":             {name: "build"},
		"underscore":         {name: "_build"},
		"hyphen-in-middle":   {name: "test-node"},
		"mixed":              {name: "cmd_build"},
		"uppercase":          {name: "Build"},
		"with_digits":        {name: "build2"},
		"leading_underscore": {name: "_2build"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := gomega.NewWithT(t)

			// Act
			result := safe.IsValid(c.name)

			// Assert
			g.Expect(result).To(gomega.BeTrue(), "expected %q to be valid", c.name)
		})
	}
}

// TestIsValid_UnsafeNames_ReturnsFalse tests that unsafe identifiers are rejected.
func TestIsValid_UnsafeNames_ReturnsFalse(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		name string
	}{
		"empty":         {name: ""},
		"colon":         {name: "cmd:build"},
		"slash":         {name: "doc/taskfile"},
		"leading_digit": {name: "2build"},
		"leading_hyphen": {name: "-build"},
		"space":         {name: "my task"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := gomega.NewWithT(t)

			// Act
			result := safe.IsValid(c.name)

			// Assert
			g.Expect(result).To(gomega.BeFalse(), "expected %q to be invalid", c.name)
		})
	}
}

// TestSanitize tests that Sanitize produces the expected safe identifier.
func TestSanitize_VariousInputs_ProducesExpectedOutput(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input    string
		expected string
	}{
		"already_safe":   {input: "build", expected: "build"},
		"colon":          {input: "cmd:build", expected: "cmd_build"},
		"slash":          {input: "doc/taskfile", expected: "doc_taskfile"},
		"leading_digit":  {input: "2build", expected: "_2build"},
		"leading_hyphen": {input: "-build", expected: "_build"},
		"empty":          {input: "", expected: "_"},
		"multiple_colons": {input: "cmd:test:unit", expected: "cmd_test_unit"},
		"space":          {input: "my task", expected: "my_task"},
		"hyphen_in_middle": {input: "test-node", expected: "test-node"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := gomega.NewWithT(t)

			// Act
			result := safe.Sanitize(c.input)

			// Assert
			g.Expect(result).To(gomega.Equal(c.expected))
		})
	}
}

// TestRegistry_ID_AlreadySafeName_ReturnsUnchanged tests that safe names pass through unchanged.
func TestRegistry_ID_AlreadySafeName_ReturnsUnchanged(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	reg := safe.NewRegistry()

	// Act
	result := reg.ID("build")

	// Assert
	g.Expect(result).To(gomega.Equal("build"))
}

// TestRegistry_ID_UnsafeName_ReturnsSanitized tests that unsafe names are sanitized.
func TestRegistry_ID_UnsafeName_ReturnsSanitized(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	reg := safe.NewRegistry()

	// Act
	result := reg.ID("cmd:build")

	// Assert
	g.Expect(result).To(gomega.Equal("cmd_build"))
}

// TestRegistry_ID_SameName_ReturnsSameResult tests that the same input always produces the same output.
func TestRegistry_ID_SameName_ReturnsSameResult(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	reg := safe.NewRegistry()

	// Act
	first := reg.ID("cmd:build")
	second := reg.ID("cmd:build")

	// Assert
	g.Expect(first).To(gomega.Equal(second))
}

// TestRegistry_ID_CollidingNames_ProducesUniqueIDs tests that colliding names get distinct safe IDs.
func TestRegistry_ID_CollidingNames_ProducesUniqueIDs(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	reg := safe.NewRegistry()

	// Act
	id1 := reg.ID("doc:taskfile")
	id2 := reg.ID("doc/taskfile")

	// Assert
	g.Expect(id1).NotTo(gomega.Equal(id2))
}

// TestRegistry_ID_SafeNameTakesPrecedence tests that already-safe names win collisions.
func TestRegistry_ID_SafeNameTakesPrecedence_WhenPrepared(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	reg := safe.NewRegistry()
	reg.Prepare([]string{"doc_taskfile", "doc:taskfile", "doc/taskfile"})

	// Act
	safe1 := reg.ID("doc_taskfile")
	colon := reg.ID("doc:taskfile")
	slash := reg.ID("doc/taskfile")

	// Assert
	g.Expect(safe1).To(gomega.Equal("doc_taskfile"), "already-safe name should be unchanged")
	g.Expect(colon).NotTo(gomega.Equal("doc_taskfile"), "colon variant should be disambiguated")
	g.Expect(slash).NotTo(gomega.Equal("doc_taskfile"), "slash variant should be disambiguated")
	g.Expect(colon).NotTo(gomega.Equal(slash), "two colliding variants must get distinct IDs")
}

// TestRegistry_ID_DisambiguationSuffix tests that disambiguation suffixes follow the expected pattern.
func TestRegistry_ID_DisambiguationSuffix_UsesLetterSuffix(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	reg := safe.NewRegistry()
	reg.Prepare([]string{"doc_taskfile"})

	// Act
	reg.ID("doc_taskfile")       // claims doc_taskfile
	colon := reg.ID("doc:taskfile") // collides -> gets _A
	slash := reg.ID("doc/taskfile") // collides -> gets _B

	// Assert
	g.Expect(colon).To(gomega.Equal("doc_taskfile_A"))
	g.Expect(slash).To(gomega.Equal("doc_taskfile_B"))
}

// TestRegistry_IDWithPrefix_AddsPrefixBeforeTransformation verifies IDWithPrefix behavior.
func TestRegistry_IDWithPrefix_AddsPrefixBeforeTransformation(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	reg := safe.NewRegistry()

	// Act
	result := reg.IDWithPrefix("cluster_", "cmd:test")

	// Assert
	g.Expect(result).To(gomega.Equal("cluster_cmd_test"))
}

// TestRegistry_IDWithPrefix_ReducesCollisions tests that prefixed IDs don't interfere with unprefixed.
func TestRegistry_IDWithPrefix_ReducesCollisions(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange
	reg := safe.NewRegistry()

	// Act
	nodeID := reg.ID("cmd:test")
	clusterID := reg.IDWithPrefix("cluster_", "cmd:test")

	// Assert
	g.Expect(nodeID).To(gomega.Equal("cmd_test"))
	g.Expect(clusterID).To(gomega.Equal("cluster_cmd_test"))
	g.Expect(nodeID).NotTo(gomega.Equal(clusterID))
}

// TestLabel_EscapesDoubleQuote tests that double quotes are HTML-entity-encoded.
func TestLabel_EscapesDoubleQuote(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Act
	result := safe.Label(`say "hello"`)

	// Assert
	g.Expect(result).To(gomega.Equal("say &quot;hello&quot;"))
}

// TestLabel_EscapesPipe tests that pipe characters are HTML-entity-encoded.
func TestLabel_EscapesPipe(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Act
	result := safe.Label("a|b")

	// Assert
	g.Expect(result).To(gomega.Equal("a&#124;b"))
}

// TestLabel_PlainText_ReturnsUnchanged tests that text without special chars is unchanged.
func TestLabel_PlainText_ReturnsUnchanged(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Act
	result := safe.Label("simple label")

	// Assert
	g.Expect(result).To(gomega.Equal("simple label"))
}
