package linter

import (
	"go/ast"
	"go/token"
	"strings"

	gofinding "github.com/larsartmann/go-finding"
)

// primitiveLocation describes where a primitive-typed ID was found. The
// string values are stable metadata (kindDisplay) carried on BD003 findings.
type primitiveLocation string

const (
	locationParam  primitiveLocation = "function-param"
	locationField  primitiveLocation = "struct-field"
	locationReturn primitiveLocation = "return-param"
)

// display returns the short human-readable label used in finding metadata.
func (l primitiveLocation) display() string {
	switch l {
	case locationParam:
		return "fn param"
	case locationField:
		return "struct field"
	case locationReturn:
		return "return param"
	default:
		return string(l)
	}
}

// primitiveIDViolation is one BD003 fact: a primitive-typed parameter, named
// return, or struct field whose name marks it as an identifier.
type primitiveIDViolation struct {
	Name        string // Parameter or field name (e.g. "userId")
	CurrentType string // Exact primitive spelling (e.g. "string", "int64")
	ParentName  string // Enclosing function or type name, "" when none
	Kind        primitiveLocation
	File        string
	Line        int // 1-based line of the field declaration
	Column      int // 1-based column of the field declaration
	Offset      int // 0-based byte offset of the field declaration
	DeclLine    int // 1-based line of the enclosing func or type declaration
	DeclEndLine int // 1-based last line of the field declaration
}

// primitiveTypeOf reports whether expr is a plain identifier naming a
// primitive Go type, returning its exact spelling. Everything else —
// pointers, slices, named types, qualified types — is not primitive here:
// detection is purely syntactic, like the BD001 scan.
func primitiveTypeOf(expr ast.Expr) (string, bool) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return "", false
	}

	switch ident.Name {
	case "string", "bool",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		return ident.Name, true
	default:
		return "", false
	}
}

// primitiveScanExcludedPaths contains path substrings excluded from the BD003
// scan: UI/rendering, output/formatting, and CLI wiring code where primitive
// IDs are presentation concerns, plus test and tooling code. Ported from the
// strong-id exclusion table so parity with the old findings holds.
//
//nolint:gochecknoglobals // Immutable lookup table.
var primitiveScanExcludedPaths = []string{
	"ui/",
	"internal/output/",
	"cmd/",
	"testhelpers/",
	"testutil/",
	"docs/examples/",
	"tools/",
}

// primitiveScanExcludedSuffixes are file-name suffixes excluded from the
// BD003 scan; _test.go files are not production code.
//
//nolint:gochecknoglobals // Immutable lookup table.
var primitiveScanExcludedSuffixes = []string{
	"_test.go",
	"_test_helpers.go",
	"_test_utils.go",
	"test_fixtures.go",
}

// excludedFromPrimitiveScan reports whether filename is excluded from BD003
// analysis. The exclusions apply to BD003 only: BD001 keeps scanning every
// file it scans today.
func excludedFromPrimitiveScan(filename string) bool {
	for _, pattern := range primitiveScanExcludedPaths {
		if strings.Contains(filename, pattern) {
			return true
		}
	}

	for _, suffix := range primitiveScanExcludedSuffixes {
		if strings.HasSuffix(filename, suffix) {
			return true
		}
	}

	return false
}

// primitiveIDViolations returns every primitive-typed ID violation in f, in
// declaration order. The traversal mirrors the strong-id analyzer this rule
// was ported from: one AST walk over function declarations (parameters and
// named returns; String/Error method declarations are skipped) and type
// declarations (struct fields, including function-local type declarations).
func primitiveIDViolations(f *ast.File, fset *token.FileSet, filename string) []primitiveIDViolation {
	if f == nil || fset == nil || excludedFromPrimitiveScan(filename) {
		return nil
	}

	var violations []primitiveIDViolation

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			violations = append(violations, violationsFromFuncDecl(node, fset)...)
		case *ast.GenDecl:
			if node.Tok != token.TYPE {
				return true
			}

			for _, spec := range node.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				violations = append(violations, violationsFromTypeSpec(typeSpec, fset)...)
			}
		}

		return true
	})

	return violations
}

// violationsFromFuncDecl collects violations from one function declaration's
// parameters and named return values. Methods named String or Error are
// skipped: they implement common interfaces whose signatures must stay
// primitive.
func violationsFromFuncDecl(fn *ast.FuncDecl, fset *token.FileSet) []primitiveIDViolation {
	if fn.Type == nil || fn.Name == nil {
		return nil
	}

	if fn.Name.Name == "String" || fn.Name.Name == "Error" {
		return nil
	}

	var violations []primitiveIDViolation

	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			violations = append(violations,
				violationsFromField(field, fn.Name.Name, locationParam, fset.Position(fn.Pos()).Line, fset)...)
		}
	}

	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			if len(field.Names) == 0 {
				continue
			}

			violations = append(violations,
				violationsFromField(field, fn.Name.Name, locationReturn, fset.Position(fn.Pos()).Line, fset)...)
		}
	}

	return violations
}

// violationsFromTypeSpec collects violations from one struct type declaration.
func violationsFromTypeSpec(ts *ast.TypeSpec, fset *token.FileSet) []primitiveIDViolation {
	structType, ok := ts.Type.(*ast.StructType)
	if !ok || structType.Fields == nil {
		return nil
	}

	var violations []primitiveIDViolation

	for _, field := range structType.Fields.List {
		violations = append(violations,
			violationsFromField(field, ts.Name.Name, locationField, fset.Position(ts.Pos()).Line, fset)...)
	}

	return violations
}

// violationsFromField flags every name of one field whose type is primitive
// and whose name marks it as an identifier. Fields without names (unnamed
// parameters or returns) cannot carry an ID name and are skipped.
func violationsFromField(
	field *ast.Field,
	parentName string,
	kind primitiveLocation,
	declLine int,
	fset *token.FileSet,
) []primitiveIDViolation {
	typeStr, isPrimitive := primitiveTypeOf(field.Type)
	if !isPrimitive {
		return nil
	}

	var violations []primitiveIDViolation

	for _, name := range field.Names {
		if !containsIDPattern(name.Name) {
			continue
		}

		pos := fset.Position(field.Pos())

		violations = append(violations, primitiveIDViolation{
			Name:        name.Name,
			CurrentType: typeStr,
			ParentName:  parentName,
			Kind:        kind,
			File:        pos.Filename,
			Line:        pos.Line,
			Column:      pos.Column,
			Offset:      pos.Offset,
			DeclLine:    declLine,
			DeclEndLine: fset.Position(field.End()).Line,
		})
	}

	return violations
}

// suppressPrimitiveIDs pairs well-formed ignore(BD003) directives with
// primitive-ID violations. A directive suppresses the violations it covers:
// its own line (trailing comment), the line directly below when it is the
// last line of a comment group, or the enclosing function/type declaration
// (its line or its doc comment). Every directive must suppress at least one
// not-yet-suppressed violation; otherwise it is a BD002 finding, because a
// silent no-op directive is worse than none.
func suppressPrimitiveIDs(
	violations []primitiveIDViolation,
	directives []suppressionDirective,
) (kept []primitiveIDViolation, findings []gofinding.Finding) {
	suppressed := make(map[int]bool, len(violations))

	for _, directive := range directives {
		coversAny, suppressesNew := false, false

		for i, v := range violations {
			if !directiveCoversViolation(directive, v) {
				continue
			}

			coversAny = true

			if !suppressed[i] {
				suppressed[i] = true
				suppressesNew = true
			}
		}

		if suppressesNew {
			continue
		}

		problem := "suppresses nothing: place it on the primitive-ID declaration line, " +
			"the line directly above it, or the enclosing function or type declaration"

		if coversAny {
			problem = "is redundant: the finding(s) it covers are already suppressed by another directive"
		}

		findings = append(findings, findingForDirective(directive, problem))
	}

	for i, v := range violations {
		if suppressed[i] {
			continue
		}

		kept = append(kept, v)
	}

	return kept, findings
}

// directiveCoversViolation reports whether one ignore(BD003) directive
// covers one violation: a trailing comment on the violation's line, the last
// line of the comment group directly above it, or a directive attached to
// the enclosing declaration (trailing or doc comment).
func directiveCoversViolation(directive suppressionDirective, v primitiveIDViolation) bool {
	if directive.Line == v.Line {
		return true
	}

	if directive.LastInGroup {
		return directive.Line+1 == v.Line || directive.Line+1 == v.DeclLine
	}

	return directive.Line == v.DeclLine
}
