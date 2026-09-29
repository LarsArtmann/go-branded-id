package linter

import (
	"go/ast"
	"go/token"
)

const (
	// methodNamePlaceholder marks a Name() method whose return value could
	// not be statically resolved from the AST.
	methodNamePlaceholder = "(method on T)"
	// stringTypeName is the Go AST identifier for the built-in string type.
	stringTypeName = "string"
)

// BrandDecl describes one brand type declaration discovered by the scanner.
type BrandDecl struct {
	TypeName string // Brand type name (e.g. "UserBrand")
	File     string // Source file as passed to the scanner
	Line     int    // 1-based line of the type declaration
	Column   int    // 1-based column of the type declaration
	Offset   int    // 0-based byte offset of the type declaration
	HasName  bool   // Whether a Name() string method exists
}

// brandDeclsFromFile returns every brand declaration in f. A brand is an
// empty struct that is used as the first type argument of id.ID[...] in the
// same file. Files without any id.ID usage yield nil.
func brandDeclsFromFile(f *ast.File, fset *token.FileSet) []BrandDecl {
	brandsUsedWithID := brandTypeArgsFromFile(f)
	if len(brandsUsedWithID) == 0 {
		return nil
	}

	hasName := collectNameMethods(f)

	return collectBrandDecls(f, fset, brandsUsedWithID, hasName)
}

// collectBrandDecls traverses all type declarations in f and returns the
// empty-struct brands used with id.ID[...], each with its source position.
func collectBrandDecls(
	f *ast.File,
	fset *token.FileSet,
	brandsUsedWithID map[string]bool,
	hasName map[string]string,
) []BrandDecl {
	var decls []BrandDecl

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
			decls = append(decls, BrandDecl{
				TypeName: ts.Name.Name,
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Offset:   pos.Offset,
				HasName: hasName[ts.Name.Name] != "" &&
					hasName[ts.Name.Name] != methodNamePlaceholder,
			})
		}
	}

	return decls
}

// isEmptyStructBrand reports whether ts declares an empty struct that is
// used with id.ID[...].
func isEmptyStructBrand(ts *ast.TypeSpec, brandsUsedWithID map[string]bool) bool {
	structType, ok := ts.Type.(*ast.StructType)
	if !ok || len(structType.Fields.List) != 0 {
		return false
	}

	return brandsUsedWithID[ts.Name.Name]
}

// brandTypeArgsFromFile walks f and returns the set of type names used as the
// first type argument of id.ID[Brand, Value] or ID[Brand] (both qualified and
// same-package forms).
func brandTypeArgsFromFile(f *ast.File) map[string]bool {
	brands := make(map[string]bool)

	ast.Inspect(f, func(n ast.Node) bool {
		// Handle ID[Brand, Value] (two type parameters) and ID[Brand] (one
		// parameter).
		var idxExpr ast.Expr

		switch idx := n.(type) {
		case *ast.IndexListExpr:
			idxExpr = idx.X
		case *ast.IndexExpr:
			idxExpr = idx.X
		default:
			return true
		}

		if !isIDSelector(idxExpr) {
			return true
		}

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

// isIDSelector reports whether expr is a selector with field name "ID" or an
// identifier named "ID".
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

// collectNameMethods returns a map of type names to their Name() string
// return values (or methodNamePlaceholder when the return value cannot be
// determined statically).
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

		result[typeName] = methodNamePlaceholder
	}

	return result
}

// isNameMethod reports whether fn is a Name() string method with a receiver,
// and returns the receiver's type name if so.
func isNameMethod(fn *ast.FuncDecl) (string, bool) {
	if fn.Name.Name != "Name" {
		return "", false
	}

	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return "", false
	}

	sig := fn.Type
	if sig.Params != nil && len(sig.Params.List) != 0 ||
		sig.Results == nil || len(sig.Results.List) != 1 {
		return "", false
	}

	if !isStringType(sig.Results.List[0].Type) {
		return "", false
	}

	recv := fn.Recv.List[0].Type

	typeName := typeNameFromExpr(recv)
	if typeName == "" {
		return "", false
	}

	return typeName, true
}

// parseNameReturnValue returns the string literal of a single-statement
// `return "..."` body, or "" when the value cannot be determined statically.
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

// isStringType reports whether the type expression resolves to string.
func isStringType(e ast.Expr) bool {
	if ident, ok := e.(*ast.Ident); ok {
		return ident.Name == stringTypeName
	}

	if star, ok := e.(*ast.StarExpr); ok {
		return isStringType(star.X)
	}

	return false
}
