// Package layout provides a go/analysis analyzer enforcing the cross-package
// correspondence of the gomatic three-tier CLI layout: every
// internal/app/commands/<cmd> package has a matching internal/domain/<cmd>
// package, and vice versa. Each package checks its own counterpart on the
// filesystem, so the two directions are reported without duplication.
package layout

import (
	"path/filepath"
	"strings"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
)

const (
	commandSegment = "/internal/app/commands/"
	domainSegment  = "/internal/domain/"
)

// Analyzer reports command or domain packages whose counterpart is missing.
var Analyzer = &analysis.Analyzer{
	Name: "layout",
	Doc:  "reports three-tier packages whose corresponding command or domain package is missing",
	Run:  run,
}

// Registration declares this analyzer to the yze framework.
var Registration = goyze.Registration{
	Name:       "layout",
	Categories: []goyze.Category{"structure"},
	URL:        "https://docs.gomatic.dev/yze/layout",
	Analyzer:   Analyzer,
}

// run reports when a command or domain package has no counterpart package.
//
// A package whose only Go files are external test files (an examples-only
// directory) is delivered by the driver as a base-package pass with no syntax
// files. There is no file to locate the package directory or anchor a report
// on, and such a directory is never a three-tier command or domain package, so
// the pass is a no-op rather than an index-out-of-range panic on pass.Files[0].
func run(pass *analysis.Pass) (any, error) {
	if len(pass.Files) == 0 {
		return nil, nil
	}
	counterpart, message, ok := counterpartOf(pkgDir(packageDir(pass)))
	if ok && !hasPackage(counterpart) {
		pass.Reportf(pass.Files[0].Name.Pos(), "%s", message)
	}
	return nil, nil
}

// packageDir returns the filesystem directory of the analyzed package,
// normalized to forward slashes (filepath.ToSlash, a no-op on POSIX) so the
// "/"-based segment logic also works on Windows file names.
func packageDir(pass *analysis.Pass) string {
	name := filepath.ToSlash(pass.Fset.File(pass.Files[0].Pos()).Name())
	idx := strings.LastIndex(name, "/")
	if idx < 0 {
		return name
	}
	return name[:idx]
}

// pkgDir is the filesystem directory of an analyzed package.
type pkgDir string

// counterpartOf returns the directory that must exist for a command or domain
// package, the diagnostic to emit if it is missing, and whether dir is a
// three-tier package at all.
//
// The correspondence requirement applies only at the direct <cmd> segment:
// packages nested deeper than one segment below commands/ or domain/ belong to
// their top-level <cmd> and carry no counterpart requirement of their own, so
// they are skipped and the check holds solely at the <cmd> pair.
func counterpartOf(dir pkgDir) (pkgDir, string, bool) {
	if isDirectChild(dir, commandSegment) {
		return pkgDir(strings.Replace(string(dir), commandSegment, domainSegment, 1)),
			"command package has no corresponding internal/domain package", true
	}
	if isDirectChild(dir, domainSegment) {
		return pkgDir(strings.Replace(string(dir), domainSegment, commandSegment, 1)),
			"domain package has no corresponding internal/app/commands package", true
	}
	return "", "", false
}

// isDirectChild reports whether dir is exactly one path segment below segment,
// i.e. the <cmd> directory itself rather than a nested subpackage of it.
func isDirectChild(dir pkgDir, segment string) bool {
	_, rest, found := strings.Cut(string(dir), segment)
	return found && rest != "" && !strings.Contains(rest, "/")
}
