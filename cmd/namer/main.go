// Package main provides a codemod tool that scans Go files for brand types
// missing the Name() string method and optionally generates stubs.
//
// A brand type is identified by:
//   - Being an empty struct
//   - Being used as a type argument in id.ID[Brand, Value] patterns
//
// Brand types that are empty structs but NOT used with id.ID[...] are not flagged.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// BrandInfo represents a discovered brand type.
type BrandInfo struct {
	TypeName  string // Brand type name (e.g., "UserBrand")
	File      string // Source file
	Line      int    // Line number of type declaration
	HasName   bool   // Whether Name() string method exists
	NameValue string // Value returned by Name() if present
}

func main() {
	dryRun := flag.Bool("dry-run", true, "Print fixes without writing files")
	write := flag.Bool("write", false, "Write Name() stubs to files (disables dry-run)")
	verbose := flag.Bool("v", false, "Verbose output")
	_ = verbose
	flag.Parse()

	if *write {
		*dryRun = false
	}

	args := flag.Args()
	if len(args) < 1 {
		_, _ = fmt.Fprintln(os.Stderr, "Usage: namer [flags] <path>...")
		_, _ = fmt.Fprintln(os.Stderr, "")
		_, _ = fmt.Fprintln(os.Stderr,
			"Scans Go files for brand types missing Name() string method.")
		_, _ = fmt.Fprintln(os.Stderr, "")
		_, _ = fmt.Fprintln(os.Stderr,
			"A brand type is an empty struct used with id.ID[Brand, Value].")
		_, _ = fmt.Fprintln(os.Stderr, "")
		_, _ = fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	r := &Result{
		brands: []BrandInfo{},
	}
	for _, path := range args {
		if err := scanPath(path, r); err != nil {
			log.Printf("error scanning %s: %v", path, err)
		}
	}

	printResults(os.Stdout, r.brands, *dryRun)
}

func printResults(w *os.File, brands []BrandInfo, dryRun bool) {
	var b strings.Builder
	missingName := filterMissing(brands)
	if len(brands) == 0 {
		b.WriteString("No brand types found. Nothing to do.\n")
		_, _ = w.WriteString(b.String())
		return
	}
	if len(missingName) == 0 {
		fmt.Fprintf(&b, "All %d brand types have Name() methods.\n", len(brands))
		_, _ = w.WriteString(b.String())
		return
	}
	fmt.Fprintf(&b,
		"Found %d brand types, %d missing Name() method:\n\n",
		len(brands), len(missingName))
	for _, item := range missingName {
		suggested := suggestName(item.TypeName)
		fmt.Fprintf(&b,
			"  %s:%d — %s\n    → func (%s) Name() string { return %q }\n",
			item.File, item.Line, item.TypeName, item.TypeName, suggested)
	}
	if !dryRun {
		b.WriteString(
			"\n[Note: AST-based file insertion not implemented — use gofmt-aware editor]\n",
		)
	} else {
		b.WriteString("\n(dry-run — no files written)\n")
		b.WriteString("Run with -write to see what would be suggested.\n")
	}
	_, _ = w.WriteString(b.String())
}

func filterMissing(brands []BrandInfo) []BrandInfo {
	result := make([]BrandInfo, 0, len(brands))
	for _, b := range brands {
		if !b.HasName {
			result = append(result, b)
		}
	}
	return result
}

// Result holds the analysis results.
type Result struct {
	brands []BrandInfo
}

// walkFn returns a filepath.WalkFunc that scans each .go file.
func walkFn(_ string, r *Result) filepath.WalkFunc {
	return func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk %s: %w", p, err)
		}
		if info.Mode().IsRegular() && strings.HasSuffix(p, ".go") {
			if err := scanFile(p, r); err != nil {
				log.Printf("error in %s: %v", p, err)
			}
		}
		return nil
	}
}

// scanPath walks a path (file or directory) and scans for brand types.
func scanPath(path string, r *Result) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	if info.IsDir() {
		if err := filepath.Walk(path, walkFn(path, r)); err != nil {
			return fmt.Errorf("walk %s: %w", path, err)
		}
		return nil
	}

	return scanFile(path, r)
}

// scanFile parses a single Go file and extracts brand type information.
func scanFile(filename string, r *Result) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	// Collect brands used with id.ID[Brand, Value] in this file.
	brandsUsedWithID := brandTypeArgsFromFile(f)

	if len(brandsUsedWithID) == 0 {
		return nil // No id.ID usage found, skip file.
	}

	// Collect all type names that have a Name() string method.
	hasName := collectNameMethods(f)

	// Find type declarations that are empty structs and used with id.ID[...].
	collectBrandTypes(f, fset, brandsUsedWithID, hasName, r)

	return nil
}

// isIDSelector returns true if expr is a selector with field name "ID" or
// an identifier named "ID".
func isIDSelector(expr ast.Expr) bool {
	switch x := expr.(type) {
	case *ast.SelectorExpr:
		return x.Sel.Name == "ID"
	case *ast.Ident:
		return x.Name == "ID"
	default:
		return false
	}
}

// collectBrandTypes traverses all type declarations in f and appends
// empty-struct brands used with id.ID[...] to r.brands.
func collectBrandTypes(
	f *ast.File,
	fset *token.FileSet,
	brandsUsedWithID map[string]bool,
	hasName map[string]string,
	r *Result,
) {
	for _, decl := range f.Decls {
		typeDecl, ok := decl.(*ast.GenDecl)
		if !ok || typeDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range typeDecl.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || !isEmptyStructBrand(ts, brandsUsedWithID) {
				continue
			}
			pos := fset.Position(ts.Pos())
			nameValue := hasName[ts.Name.Name]
			r.brands = append(r.brands, BrandInfo{
				TypeName: ts.Name.Name,
				File:     pos.Filename,
				Line:     pos.Line,
				HasName: nameValue != "" && nameValue != "(method on T)" &&
					nameValue != "(method on *T)",
				NameValue: nameValue,
			})
		}
	}
}

func isEmptyStructBrand(ts *ast.TypeSpec, brandsUsedWithID map[string]bool) bool {
	structType, ok := ts.Type.(*ast.StructType)
	if !ok || len(structType.Fields.List) != 0 {
		return false
	}
	return brandsUsedWithID[ts.Name.Name]
}

// brandTypeArgsFromFile does a full traversal to find id.ID[Brand, Value] and
// ID[Brand, Value] usages (both qualified and same-package).
func brandTypeArgsFromFile(f *ast.File) map[string]bool {
	brands := make(map[string]bool)

	ast.Inspect(f, func(n ast.Node) bool {
		// Handle ID[Brand, Value] (two type parameters) and ID[Brand] (one parameter).
		var idxExpr ast.Expr

		switch idx := n.(type) {
		case *ast.IndexListExpr:
			// ID[Brand, Value] — two or more type parameters.
			idxExpr = idx.X
		case *ast.IndexExpr:
			// ID[Brand] — single type parameter.
			idxExpr = idx.X
		default:
			return true
		}

		// Check if X is id.ID or just ID (same package).
		if !isIDSelector(idxExpr) {
			return true
		}

		// Get the Brand type argument (first type parameter).
		var idxNode ast.Expr
		switch idx := n.(type) {
		case *ast.IndexListExpr:
			if len(idx.Indices) < 1 {
				return true
			}
			idxNode = idx.Indices[0]
		case *ast.IndexExpr:
			idxNode = idx.Index
		default:
			return true
		}

		brandName := typeNameFromExpr(idxNode)
		if brandName != "" {
			brands[brandName] = true
		}

		return true
	})

	return brands
}

// typeNameFromExpr extracts a type name from an ast.Expr.
func typeNameFromExpr(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.StarExpr:
		return typeNameFromExpr(v.X)
	default:
		return ""
	}
}

// receiverTypeName extracts the type name from a receiver expression.
func receiverTypeName(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.StarExpr:
		return typeNameFromExpr(v.X)
	default:
		return ""
	}
}

// isStringType returns true if the type expression resolves to string.
func isStringType(e ast.Expr) bool {
	if ident, ok := e.(*ast.Ident); ok {
		return ident.Name == "string"
	}
	if star, ok := e.(*ast.StarExpr); ok {
		return isStringType(star.X)
	}
	return false
}

// collectNameMethods returns a map of type names to their Name() string return
// values (or "(method)" placeholders when the return value can't be determined).
func collectNameMethods(f *ast.File) map[string]string {
	result := make(map[string]string)
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		typeName, isNameFn := isNameMethod(fn)
		if !isNameFn {
			continue
		}
		if strVal := parseNameReturnValue(fn); strVal != "" {
			result[typeName] = strVal
			continue
		}
		result[typeName] = "(method on T)"
	}
	return result
}

func isNameMethod(fn *ast.FuncDecl) (string, bool) {
	if fn.Name.Name != "Name" {
		return "", false
	}
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return "", false
	}
	sig := fn.Type
	if len(sig.Params.List) != 0 || len(sig.Results.List) != 1 {
		return "", false
	}
	if !isStringType(sig.Results.List[0].Type) {
		return "", false
	}
	recv := fn.Recv.List[0].Type
	typeName := receiverTypeName(recv)
	if typeName == "" {
		return "", false
	}
	if _, isStar := recv.(*ast.StarExpr); isStar {
		return typeName, true
	}
	return typeName, true
}

func parseNameReturnValue(fn *ast.FuncDecl) string {
	if fn.Body == nil || len(fn.Body.List) != 1 {
		return ""
	}
	ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return ""
	}
	lit, ok := ret.Results[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	return lit.Value
}

// suggestName suggests a Name() return value based on the brand type name.
// Strips common suffixes.
func suggestName(brandName string) string {
	name := brandName

	name = strings.TrimSuffix(name, "Brand")
	name = strings.TrimSuffix(name, "ID")
	name = strings.TrimPrefix(name, "T")

	if name == "" {
		return brandName
	}
	return name
}
