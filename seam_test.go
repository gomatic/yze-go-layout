package layout

// White-box tests for the file/directory predicates that decide what counts as
// a package. They gate every layout check that follows, so a mistake here does
// not produce a wrong finding — it produces no finding at all, silently, for a
// directory that violates the layout.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
)

// TestIsDirectChildMatchesExactlyOneSegmentBelowTheMarker names isDirectChild's
// claim. The three-tier layout imposes obligations on the <cmd> directory
// itself and on nothing beneath it, so "exactly one segment" is the whole rule:
// matching deeper would impose command-package obligations on ordinary helper
// packages, and matching the marker directory itself would judge a directory
// that holds no command at all.
func TestIsDirectChildMatchesExactlyOneSegmentBelowTheMarker(t *testing.T) {
	t.Parallel()
	const marker treeSegment = "/internal/app/commands/"

	for _, tc := range []struct {
		dir  pkgDir
		why  string
		want bool
	}{
		{dir: "m/internal/app/commands/greet", want: true, why: "exactly one segment below"},
		{dir: "/abs/m/internal/app/commands/serve", want: true, why: "an absolute path is fine"},
		{dir: "m/internal/app/commands/greet/internal", want: false, why: "a nested package is not the command"},
		{dir: "m/internal/app/commands/a/b/c", want: false, why: "however deep"},
		{dir: "m/internal/app/commands/", want: false, why: "the marker directory holds no command"},
		{dir: "m/internal/app/handlers/greet", want: false, why: "a different marker"},
		{dir: "", want: false, why: "an empty directory"},
	} {
		assert.Equal(t, tc.want, isDirectChild(tc.dir, marker), "isDirectChild(%q): %s", tc.dir, tc.why)
	}
}

// TestIsPackageSourceRequiresAParsingPackageClause names isPackageSource's
// claim. The check is what stops a malformed or non-Go file WEARING a .go
// suffix from being counted as a package — a generated stub, a template, or a
// half-written file. Counting one would make the analyzer report a package
// clause it never actually read.
func TestIsPackageSourceRequiresAParsingPackageClause(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	write := func(name, content string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
	write("good.go", "package p\n\nfunc F() {}\n")
	write("clauseonly.go", "package p\n")
	write("broken.go", "this is not go at all\n")
	write("empty.go", "")
	write("good_test.go", "package p\n")
	write("notes.txt", "package p\n")

	for _, tc := range []struct {
		name fileName
		why  string
		want bool
	}{
		{name: "good.go", want: true, why: "an ordinary source file"},
		{name: "clauseonly.go", want: true, why: "only the clause is parsed, so a bare clause counts"},
		{name: "broken.go", want: false, why: "a .go suffix does not make a file Go"},
		{name: "empty.go", want: false, why: "an empty file has no package clause"},
		{name: "good_test.go", want: false, why: "a test file is not package source"},
		{name: "notes.txt", want: false, why: "not a .go file at all"},
		{name: "absent.go", want: false, why: "a file that does not exist"},
	} {
		assert.Equal(t, tc.want, isPackageSource(pkgDir(dir), tc.name),
			"isPackageSource(%q): %s", tc.name, tc.why)
	}
}

// TestRunIsANoOpForAPassWithNoSyntaxFiles names run's guard. A directory whose
// only Go files are EXTERNAL test files — an examples-only package — reaches
// the analyzer as a base-package pass carrying no syntax at all, and both the
// package-directory lookup and the report anchor index pass.Files[0]. Without
// the guard that is an index-out-of-range panic inside the linter, on a
// perfectly legitimate directory layout.
//
// Returning early is also correct rather than merely safe: such a directory can
// never be a three-tier command or domain package, so there is nothing to
// report even if a file could be found to anchor it on.
func TestRunIsANoOpForAPassWithNoSyntaxFiles(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		result, err := run(&analysis.Pass{})

		assert.Nil(t, result)
		assert.NoError(t, err)
	}, "an empty file list must return, not index pass.Files[0]")
}

// TestPackageDirHandlesAFileNameWithNoDirectory covers packageDir's fallback.
// Every path the loader produces normally carries a separator, but a synthesized
// pass — which is exactly what the empty-files guard above exists for — can
// carry a bare name, and slicing to a negative index would panic. Returning the
// name itself is the only answer that keeps the segment logic meaningful.
func TestPackageDirHandlesAFileNameWithNoDirectory(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		file string
		want string
		why  string
	}{
		{file: "m/internal/app/commands/greet/command.go", want: "m/internal/app/commands/greet", why: "an ordinary path"},
		{file: "bare.go", want: "bare.go", why: "no separator: the name IS the answer, not a negative slice"},
	} {
		assert.Equal(t, tc.want, packageDir(passForFile(t, tc.file)), "packageDir(%q): %s", tc.file, tc.why)
	}
}

// passForFile builds the minimal pass packageDir reads: one syntax file whose
// FileSet entry carries the given name.
func passForFile(t *testing.T, name string) *analysis.Pass {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, "package p\n", parser.PackageClauseOnly)
	require.NoError(t, err)
	return &analysis.Pass{Fset: fset, Files: []*ast.File{file}}
}
