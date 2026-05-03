package taskgraph

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestExtractVarRefs_SimpleReference_ReturnsVarName(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("{{.FOO}}")

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("FOO"))
}

func TestExtractVarRefs_PipedExpression_ReturnsVarName(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("{{.FOO | lowercase}}")

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("FOO"))
}

func TestExtractVarRefs_MultipleVariables_ReturnsAllVarNames(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs(`{{printf "%s/%s" .FOO .BAR}}`)

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("FOO", "BAR"))
}

func TestExtractVarRefs_Conditional_ReturnsVarName(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("{{if .FOO}}yes{{end}}")

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("FOO"))
}

func TestExtractVarRefs_PlainString_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("just a plain string")

	// Assert
	g.Expect(refs).To(gomega.BeEmpty())
}

func TestExtractVarRefs_EmptyString_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("")

	// Assert
	g.Expect(refs).To(gomega.BeEmpty())
}

func TestExtractVarRefs_CurrentContext_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("{{.}}")

	// Assert
	g.Expect(refs).To(gomega.BeEmpty())
}

func TestExtractVarRefs_MultipleTemplateBlocks_ReturnsAllVarNames(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("go build -ldflags {{.LDFLAGS}} -o {{.OUTPUT}}")

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("LDFLAGS", "OUTPUT"))
}

func TestExtractVarRefs_DuplicateReferences_ReturnsUniqueNames(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Arrange & Act
	refs := extractVarRefs("{{.FOO}} and {{.FOO}}")

	// Assert
	g.Expect(refs).To(gomega.ConsistOf("FOO"))
}
