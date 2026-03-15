package namespace

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestNamespace_FormalDelimiter_ReturnsPrefix(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	ns := Namespace("cmd:build")

	// Assert
	g.Expect(ns).To(Equal("cmd"))
}

func TestNamespace_NestedFormalDelimiter_ReturnsFullPrefix(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	ns := Namespace("cmd:test:unit")

	// Assert
	g.Expect(ns).To(Equal("cmd:test"))
}

func TestNamespace_InformalHyphen_ReturnsPrefix(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	ns := Namespace("build-bin")

	// Assert
	g.Expect(ns).To(Equal("build"))
}

func TestNamespace_InformalDot_ReturnsPrefix(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	ns := Namespace("tidy.format")

	// Assert
	g.Expect(ns).To(Equal("tidy"))
}

func TestNamespace_MixedInformal_ReturnsPrefix(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	ns := Namespace("build-bin.linux")

	// Assert
	g.Expect(ns).To(Equal("build-bin"))
}

func TestNamespace_FormalTakesPrecedence_IgnoresInformal(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	ns := Namespace("build-bin:test")

	// Assert
	g.Expect(ns).To(Equal("build-bin"))
}

func TestNamespace_NoDelimiter_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	ns := Namespace("deploy")

	// Assert
	g.Expect(ns).To(BeEmpty())
}

func TestParent_FormalNested_ReturnsParent(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	p := Parent("cmd:test")

	// Assert
	g.Expect(p).To(Equal("cmd"))
}

func TestParent_InformalNested_ReturnsParent(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	p := Parent("build-bin")

	// Assert
	g.Expect(p).To(Equal("build"))
}

func TestParent_TopLevel_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	p := Parent("cmd")

	// Assert
	g.Expect(p).To(BeEmpty())
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

func TestCompileMatchPattern_GlobStar_ConvertsToRegex(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	re, err := CompileMatchPattern("build*")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(re.String()).To(Equal("^build.*$"))
}

func TestCompileMatchPattern_GlobQuestion_ConvertsToRegex(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	re, err := CompileMatchPattern("build?")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(re.String()).To(Equal("^build.$"))
}

func TestCompileMatchPattern_SpecialChars_Escaped(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Act
	re, err := CompileMatchPattern("a.b(c)")

	// Assert
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(re.String()).To(Equal(`^a\.b\(c\)$`))
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
