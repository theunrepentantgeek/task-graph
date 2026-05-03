package namespace

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestNamespace_VariousInputs_ReturnsExpectedNamespace(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input    string
		expected string
	}{
		"formal delimiter":           {input: "cmd:build", expected: "cmd"},
		"nested formal delimiter":    {input: "cmd:test:unit", expected: "cmd:test"},
		"informal hyphen":            {input: "build-bin", expected: "build"},
		"informal dot":               {input: "tidy.format", expected: "tidy"},
		"mixed informal":             {input: "build-bin.linux", expected: "build-bin"},
		"formal takes precedence":    {input: "build-bin:test", expected: "build-bin"},
		"no delimiter returns empty": {input: "deploy", expected: ""},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			// Act
			ns := Namespace(c.input)

			// Assert
			g.Expect(ns).To(Equal(c.expected))
		})
	}
}

func TestParent_VariousInputs_ReturnsExpectedParent(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input    string
		expected string
	}{
		"formal nested":           {input: "cmd:test", expected: "cmd"},
		"informal nested":         {input: "build-bin", expected: "build"},
		"top level returns empty": {input: "cmd", expected: ""},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			// Act
			p := Parent(c.input)

			// Assert
			g.Expect(p).To(Equal(c.expected))
		})
	}
}

func TestDepth_VariousNamespaces_ReturnsCorrectDepth(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		ns       string
		expected int
	}{
		"top-level formal":   {ns: "cmd", expected: 0},
		"nested formal":      {ns: "cmd:test", expected: 1},
		"deep formal":        {ns: "a:b:c", expected: 2},
		"top-level informal": {ns: "build", expected: 0},
		"nested informal":    {ns: "build-bin", expected: 1},
		"deep informal":      {ns: "build-bin.linux", expected: 2},
		"mixed formal":       {ns: "cmd:build-bin", expected: 1},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			// Act
			d := Depth(c.ns)

			// Assert
			g.Expect(d).To(Equal(c.expected))
		})
	}
}

func TestCompileMatchPattern_VariousPatterns_ConvertsToExpectedRegex(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		pattern  string
		expected string
	}{
		"glob star":     {pattern: "build*", expected: "^build.*$"},
		"glob question": {pattern: "build?", expected: "^build.$"},
		"special chars": {pattern: "a.b(c)", expected: `^a\.b\(c\)$`},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			// Act
			re, err := CompileMatchPattern(c.pattern)

			// Assert
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(re.String()).To(Equal(c.expected))
		})
	}
}

func TestCompileMatchPattern_CharacterClass_PassedThrough(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	re, err := CompileMatchPattern("build[-.:]*")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(re.String()).To(Equal(`^build[-.:].*$`))
	g.Expect(re.MatchString("build-bin")).To(BeTrue())
	g.Expect(re.MatchString("build.image")).To(BeTrue())
	g.Expect(re.MatchString("build:x")).To(BeTrue())
	g.Expect(re.MatchString("buildall")).To(BeFalse())
}

func TestCompileMatchPattern_InvalidPattern_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act — unclosed bracket
	_, err := CompileMatchPattern("[unclosed")

	// Assert
	g.Expect(err).To(HaveOccurred())
}

func TestCompileMatchPattern_ExistingPatterns_BackwardCompatible(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		pattern string
		match   string
		noMatch string
	}{
		"prefix glob": {
			pattern: "build*",
			match:   "build-bin",
			noMatch: "test",
		},
		"contains glob": {
			pattern: "*test*",
			match:   "mytest-unit",
			noMatch: "build",
		},
		"formal namespace glob": {
			pattern: "cmd:*",
			match:   "cmd:build",
			noMatch: "controllers:build",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			// Act
			re, err := CompileMatchPattern(c.pattern)

			// Assert
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(re.MatchString(c.match)).To(BeTrue())
			g.Expect(re.MatchString(c.noMatch)).To(BeFalse())
		})
	}
}

func TestMatchPattern_FormalNamespace_ReturnsColonPattern(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	p := MatchPattern("cmd:test")

	// Assert
	g.Expect(p).To(Equal("cmd:test:*"))
}

func TestMatchPattern_AmbiguousNamespace_ReturnsAllDelimiterPattern(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	p := MatchPattern("build")

	// Assert
	g.Expect(p).To(Equal("build[-.:]*"))
}

// BenchmarkCompileMatchPattern measures allocations for a typical style-rule pattern.
func BenchmarkCompileMatchPattern(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_, _ = CompileMatchPattern("build:*")
	}
}
