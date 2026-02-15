# Development guidelines for go-vcr-tidy

# Abstractions

TBC

## Coding style

Always include a type cast assertion in the form `var _ theInterface = &mystruct{}` when a struct needs to implement a given interface. Group them using `var ( ... )` if there are multiple interfaces to assert.

## Testing

We use [gomega](https://github.com/onsi/gomega) for unit test assertions, and [goldie](github.com/sebdah/goldie/v2) for 
golden tests, where required.

For golden tests managed by goldie, refresh fixtures with `go test ./... -update` and review the resulting diffs before committing.

Test cases are ordered in each test file, with later tests able to assume that system properties asserted by earlier 
tests are held. This helps to narrow the focus of each test. As a direct corollary of this, when diagnosing test 
failures, the earliest failing test in a file is a good place to start.

Table tests use `cases := map[string]struct{...}` to capture test cases, with the name of the test as the map key. Test 
case iteration uses `for name, c := range cases`.

* All tests are marked with `t.Parallel()` unless the test cannot run in parallel.
* Helper methods are always marked with `t.Helper()`.
* Only use a test package (e.g. with the suffix _test) if needed to avoid circular imports.
* Whenever you have a set of tests with similar structure, prefer to create a table test to avoid duplication of code.
* Use Roy Osherove's naming style for tests: Test<SubjectUnderTest>_<Scenario>_<Expecation>
* Tests for a given method should be grouped together with a leading comment (see monitor_deletion_test.go for this style)
* In most cases, the phases of a unit test should be marked with comments for Arrange/Act/Assert
  * if a test doesn't match this structure, this may indicate the test is doing too much and needs to be split into separate test cases

## Linting

**Always use `task lint` to run the linter.** Never run `golangci-lint` directly, as the project uses a custom build 
with `nilaway` integration that requires special configuration.

