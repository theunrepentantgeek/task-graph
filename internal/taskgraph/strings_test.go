package taskgraph

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-task/task/v3/taskfile/ast"
)

func ptr[T any](v T) *T {
	return &v
}

// collectEnvStrings tests

func TestCollectEnvStrings_NilEnv_ReturnsOriginalSlice(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	result := collectEnvStrings([]string{"existing"}, nil)

	g.Expect(result).To(Equal([]string{"existing"}))
}

func TestCollectEnvStrings_EnvWithValue_AppendsStringValue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	env := ast.NewVars(
		&ast.VarElement{Key: "FOO", Value: ast.Var{Value: "bar"}},
	)

	result := collectEnvStrings(nil, env)

	g.Expect(result).To(ConsistOf("bar"))
}

func TestCollectEnvStrings_EnvWithShCommand_IsIgnored(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	env := ast.NewVars(
		&ast.VarElement{Key: "FOO", Value: ast.Var{Sh: ptr("echo hi")}},
	)

	result := collectEnvStrings(nil, env)

	g.Expect(result).To(BeEmpty())
}

func TestCollectEnvStrings_EnvWithNilValue_IsIgnored(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	env := ast.NewVars(
		&ast.VarElement{Key: "FOO", Value: ast.Var{Value: nil}},
	)

	result := collectEnvStrings(nil, env)

	g.Expect(result).To(BeEmpty())
}

func TestCollectEnvStrings_MultipleEnvVars_AppendsAllValues(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	env := ast.NewVars(
		&ast.VarElement{Key: "A", Value: ast.Var{Value: "alpha"}},
		&ast.VarElement{Key: "B", Value: ast.Var{Value: "beta"}},
	)

	result := collectEnvStrings(nil, env)

	g.Expect(result).To(ConsistOf("alpha", "beta"))
}

// collectVarStrings tests

func TestCollectVarStrings_NilVars_ReturnsOriginalSlice(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	result := collectVarStrings([]string{"existing"}, nil)

	g.Expect(result).To(Equal([]string{"existing"}))
}

func TestCollectVarStrings_VarWithValue_AppendsStringValue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	vars := ast.NewVars(
		&ast.VarElement{Key: "FOO", Value: ast.Var{Value: "bar"}},
	)

	result := collectVarStrings(nil, vars)

	g.Expect(result).To(ConsistOf("bar"))
}

func TestCollectVarStrings_VarWithShCommand_AppendsShString(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	vars := ast.NewVars(
		&ast.VarElement{Key: "FOO", Value: ast.Var{Sh: ptr("echo hi")}},
	)

	result := collectVarStrings(nil, vars)

	g.Expect(result).To(ConsistOf("echo hi"))
}

func TestCollectVarStrings_VarWithValueAndSh_AppendsBoth(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	vars := ast.NewVars(
		&ast.VarElement{Key: "FOO", Value: ast.Var{Value: "static", Sh: ptr("echo dynamic")}},
	)

	result := collectVarStrings(nil, vars)

	g.Expect(result).To(ConsistOf("static", "echo dynamic"))
}

func TestCollectVarStrings_VarWithOnlyRef_IsIgnored(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	vars := ast.NewVars(
		&ast.VarElement{Key: "FOO", Value: ast.Var{Ref: "SOME_OTHER_VAR"}},
	)

	result := collectVarStrings(nil, vars)

	g.Expect(result).To(BeEmpty())
}

func TestCollectVarStrings_MultipleVars_AppendsAll(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	vars := ast.NewVars(
		&ast.VarElement{Key: "A", Value: ast.Var{Value: "alpha"}},
		&ast.VarElement{Key: "B", Value: ast.Var{Sh: ptr("echo beta")}},
	)

	result := collectVarStrings(nil, vars)

	g.Expect(result).To(ConsistOf("alpha", "echo beta"))
}

// varDescription tests

func TestVarDescription_WithShCommand_ReturnsPrefixedSh(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	v := ast.Var{Sh: ptr("echo hello")}

	g.Expect(varDescription(v)).To(Equal("sh: echo hello"))
}

func TestVarDescription_WithEmptyShCommand_FallsThrough(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Empty Sh should not be returned; Value is checked next.
	v := ast.Var{Sh: ptr(""), Value: "fallback"}

	g.Expect(varDescription(v)).To(Equal("fallback"))
}

func TestVarDescription_WithValue_ReturnsFormattedValue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	v := ast.Var{Value: "hello"}

	g.Expect(varDescription(v)).To(Equal("hello"))
}

func TestVarDescription_WithRef_ReturnsPrefixedRef(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	v := ast.Var{Ref: "OTHER_VAR"}

	g.Expect(varDescription(v)).To(Equal("ref: OTHER_VAR"))
}

func TestVarDescription_EmptyVar_ReturnsEmptyString(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	v := ast.Var{}

	g.Expect(varDescription(v)).To(BeEmpty())
}

// appendNonEmpty tests

func TestAppendNonEmpty_AllNonEmpty_AppendsAll(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	result := appendNonEmpty(nil, "a", "b", "c")

	g.Expect(result).To(Equal([]string{"a", "b", "c"}))
}

func TestAppendNonEmpty_SomeEmpty_SkipsEmpties(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	result := appendNonEmpty(nil, "a", "", "c")

	g.Expect(result).To(Equal([]string{"a", "c"}))
}

func TestAppendNonEmpty_AllEmpty_ReturnsNil(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	result := appendNonEmpty(nil, "", "")

	g.Expect(result).To(BeNil())
}

func TestAppendNonEmpty_AppendsToExistingSlice(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	result := appendNonEmpty([]string{"existing"}, "", "new")

	g.Expect(result).To(Equal([]string{"existing", "new"}))
}
