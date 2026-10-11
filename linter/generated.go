package linter

import (
	"go/ast"
	"regexp"
)

// generatedFilePattern matches the ecosystem-standard generated-code header
// line as it survives CommentGroup.Text processing (the leading "//" and one
// space are stripped): "Code generated <by tool>. DO NOT EDIT." — emitted by
// sqlc, templ, protoc-gen-go, mockgen, stringer, and the toolchain's own
// cgo with minor punctuation variations the surrounding wildcards absorb.
var generatedFilePattern = regexp.MustCompile(`(?m)^Code generated .* DO NOT EDIT\.$`)

// isGenerated reports whether f carries the standard generated-file header
// comment before the package clause.
//
// Generated files are never linted and never repaired: their contents are
// machine output, so findings there are unactionable (the fix belongs in the
// generator's input) and inserted Name() stubs would be overwritten by the
// next generation run. This restores the excludeGenerated default of the
// rule's branching-flow ancestor, which skipped such files since its first
// release; the parity corpus contained no generated files, so the gap only
// surfaced on real sqlc/templ output (CV, 2026-10-11).
func isGenerated(f *ast.File) bool {
	for _, group := range f.Comments {
		if group.Pos() >= f.Package {
			break
		}

		if generatedFilePattern.MatchString(group.Text()) {
			return true
		}
	}

	return false
}
