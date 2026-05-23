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
	_ = verbose // TODO: use for verbose output
	flag.Parse()

	if *write {
		*dryRun = false
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: namer [flags] <path>...")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Scans Go files for brand types missing Name() string method.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "A brand type is an empty struct used with id.ID[Brand, Value].")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	r := &Result{}
	for _, path := range args {
		if err := scanPath(path, r); err != nil {
			log.Printf("error scanning %s: %v", path, err)
		}
	}

	if len(r.brands) == 0 {
		fmt.Println("No brand types found. Nothing to do.")
		return
	}

	missingName := make([]BrandInfo, 0)
	for _, b := range r.brands {
		if !b.HasName {
			missingName = append(missingName, b)
		}
	}

	if len(missingName) == 0 {
		fmt.Printf("All %d brand types have Name() methods.\n", len(r.brands))
		return
	}

	fmt.Printf(
		"Found %d brand types, %d missing Name() method:\n\n",
		len(r.brands),
		len(missingName),
	)
	for _, b := range missingName {
		suggested := suggestName(b.TypeName)
		fmt.Printf("  %s:%d — %s\n    → func (%s) Name() string { return %q }\n",
			b.File, b.Line, b.TypeName, b.TypeName, suggested)
	}

	if !*dryRun {
		fmt.Println("\n[Note: AST-based file insertion not implemented — use gofmt-aware editor]")
	} else {
		fmt.Println("\n(dry-run — no files written)")
		fmt.Println("Run with -write to see what would be suggested.")
	}
}

// Result holds the analysis results.
type Result struct {
	brands []BrandInfo
}

// scanPath walks a path (file or directory) and scans for brand types.
func scanPath(path string, r *Result) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	if info.IsDir() {
		return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsRegular() && strings.HasSuffix(p, ".go") {
				if err := scanFile(p, r); err != nil {
					log.Printf("error in %s: %v", p, err)
				}
			}
			return nil
		})
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
	hasName := make(map[string]string) // typeName -> return value or "(method)" if unknown
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Name" {
			continue
		}
		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		sig := fn.Type
		if len(sig.Params.List) != 0 || len(sig.Results.List) != 1 {
			continue
		}
		if !isStringType(sig.Results.List[0].Type) {
			continue
		}
		recv := fn.Recv.List[0].Type
		typeName := receiverTypeName(recv)
		if typeName == "" {
			continue
		}

		// Try to extract the return value from the method body.
		if fn.Body != nil && len(fn.Body.List) == 1 {
			if ret, ok := fn.Body.List[0].(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
				if lit, ok := ret.Results[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					hasName[typeName] = lit.Value
					continue
				}
			}
		}

		if _, isStar := recv.(*ast.StarExpr); isStar {
			hasName[typeName] = "(method on *T)"
		} else {
			hasName[typeName] = "(method on T)"
		}
	}

	// Find type declarations that are empty structs and used with id.ID[...].
	for _, decl := range f.Decls {
		typeDecl, ok := decl.(*ast.GenDecl)
		if !ok || typeDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range typeDecl.Specs {
			spec := spec.(*ast.TypeSpec)

			// Only process empty structs that are used with id.ID[Brand, Value].
			if structType, ok := spec.Type.(*ast.StructType); ok {
				if len(structType.Fields.List) == 0 && brandsUsedWithID[spec.Name.Name] {
					pos := fset.Position(spec.Pos())
					nameValue := hasName[spec.Name.Name]
					r.brands = append(r.brands, BrandInfo{
						TypeName: spec.Name.Name,
						File:     pos.Filename,
						Line:     pos.Line,
						HasName: nameValue != "" && nameValue != "(method on T)" &&
							nameValue != "(method on *T)",
						NameValue: nameValue,
					})
				}
			}
		}
	}

	return nil
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
		switch x := idxExpr.(type) {
		case *ast.SelectorExpr:
			// pkg.ID[Brand, Value] — qualified import from any package.
			// Any selector where the field name is "ID" is treated as a potential brand usage.
			if x.Sel.Name != "ID" {
				return true
			}
		case *ast.Ident:
			// ID[Brand, Value] — same package usage.
			if x.Name != "ID" {
				return true
			}
		default:
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
