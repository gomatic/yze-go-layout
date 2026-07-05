package layout

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// hasPackage reports whether a directory is a real Go package. It is a
// package-level seam, defaulting to the filesystem implementation, so tests
// drive the existence decision with a fake and the analysis stays hermetic
// (no direct OS reach-through, per the gomatic dependency-injection standard).
var hasPackage = osHasPackage

// osHasPackage reports whether dir contains at least one non-test Go source
// file whose package clause parses, i.e. is a real Go package rather than a
// bare directory or one holding only Go-suffixed files that are not Go source.
// A counterpart directory without such a file does not satisfy the layout, so
// neither existence alone nor a mere .go suffix is enough.
func osHasPackage(dir pkgDir) bool {
	entries, err := os.ReadDir(string(dir))
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && isPackageSource(dir, fileName(entry.Name())) {
			return true
		}
	}
	return false
}

// fileName is a bare file name within a package directory.
type fileName string

// isPackageSource reports whether name is a non-test Go source file in dir
// whose package clause actually parses (parser.PackageClauseOnly), so a
// malformed or non-Go file wearing a .go suffix never counts as a package.
func isPackageSource(dir pkgDir, name fileName) bool {
	if !isGoSource(name) {
		return false
	}
	path := filepath.Join(string(dir), string(name))
	_, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.PackageClauseOnly)
	return err == nil
}

// isGoSource reports whether name is a non-test Go source filename.
func isGoSource(name fileName) bool {
	return strings.HasSuffix(string(name), ".go") && !strings.HasSuffix(string(name), "_test.go")
}
