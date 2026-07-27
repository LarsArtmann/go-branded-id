package main

import (
	"bytes"
	"go/ast"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_NoArgs(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}

	code := run([]string{}, w, w)
	_ = w.Close()

	if code != 1 {
		t.Errorf("expected exit code 1 for no args, got %d", code)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read pipe: %v", err)
	}

	if !strings.Contains(buf.String(), "Usage:") {
		t.Errorf("expected usage message, got: %s", buf.String())
	}
}

func TestRun_ScanTestData(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}

	code := run([]string{filepath.Join("testdata", "missing_name.go")}, w, w)
	_ = w.Close()

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read pipe: %v", err)
	}

	if !strings.Contains(buf.String(), "UserBrand") {
		t.Errorf("expected UserBrand in output, got: %s", buf.String())
	}
}

func TestIsIDSelector_DefaultCase(t *testing.T) {
	t.Parallel()

	t.Run("returns false for binary expression", func(t *testing.T) {
		t.Parallel()

		expr := &ast.BinaryExpr{
			X:  &ast.Ident{Name: "a"},
			Op: token.ADD,
			Y:  &ast.Ident{Name: "b"},
		}

		if isIDSelector(expr) {
			t.Error("expected false for *ast.BinaryExpr")
		}
	})

	t.Run("returns false for call expression", func(t *testing.T) {
		t.Parallel()

		expr := &ast.CallExpr{
			Fun:  &ast.Ident{Name: "foo"},
			Args: []ast.Expr{},
		}

		if isIDSelector(expr) {
			t.Error("expected false for *ast.CallExpr")
		}
	})

	t.Run("returns false for basic literal", func(t *testing.T) {
		t.Parallel()

		expr := &ast.BasicLit{Kind: token.STRING, Value: `"hello"`}

		if isIDSelector(expr) {
			t.Error("expected false for *ast.BasicLit")
		}
	})
}

func TestIsEmptyStructBrand_NonStructType(t *testing.T) {
	t.Parallel()

	brandsUsedWithID := map[string]bool{"FooBrand": true}

	tests := []struct {
		name string
		typ  ast.Expr
		want bool
	}{
		{"interface type returns false", &ast.InterfaceType{Methods: &ast.FieldList{}}, false},
		{"non-empty struct returns false", &ast.StructType{
			Fields: &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{{Name: "x"}},
				Type:  &ast.Ident{Name: "int"},
			}}},
		}, false},
		{"ident type returns false", &ast.Ident{Name: "int"}, false},
		{"empty struct NOT in brandsUsedWithID returns false", &ast.StructType{Fields: &ast.FieldList{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &ast.TypeSpec{Name: &ast.Ident{Name: "SomeBrand"}, Type: tt.typ}

			got := isEmptyStructBrand(ts, brandsUsedWithID)
			if got {
				t.Errorf("expected false, got true")
			}
		})
	}

	t.Run("empty struct IN brandsUsedWithID returns true", func(t *testing.T) {
		t.Parallel()

		ts := &ast.TypeSpec{
			Name: &ast.Ident{Name: "FooBrand"},
			Type: &ast.StructType{Fields: &ast.FieldList{}},
		}

		if !isEmptyStructBrand(ts, brandsUsedWithID) {
			t.Error("expected true for empty struct used with id.ID")
		}
	})
}

func TestScanFile_PointerReceiver(t *testing.T) {
	t.Parallel()

	result := &Result{brands: []BrandInfo{}}

	err := scanFile(filepath.Join("testdata", "pointer_receiver.go"), result)
	if err != nil {
		t.Fatalf("scanFile failed: %v", err)
	}

	if len(result.brands) != 1 {
		t.Fatalf("expected 1 brand, got %d", len(result.brands))
	}

	brand := result.brands[0]
	if brand.TypeName != "PointerBrand" {
		t.Errorf("expected PointerBrand, got %q", brand.TypeName)
	}

	if !brand.HasName {
		t.Error("expected HasName=true for pointer receiver")
	}
}

func TestWalkFn_ErrorPath(t *testing.T) {
	t.Parallel()

	result := &Result{brands: []BrandInfo{}}

	wf := walkFn("nonexistent", result)

	err := wf("/nonexistent/path/file.go", nil, assertWalkError("simulated walk error"))
	if err == nil {
		t.Error("expected error from walkFn when err input is non-nil")
	}

	if !strings.Contains(err.Error(), "simulated walk error") {
		t.Errorf("expected wrapped error to contain original, got: %v", err)
	}
}

type assertWalkError string

func (e assertWalkError) Error() string { return string(e) }

func FuzzScanFile(f *testing.F) {
	seedInputs := []string{
		"package test",
		"package test\ntype Foo struct{}\n",
		"invalid Go code {{{}}}",
		"package test\nvar _ id.ID[FooBrand, string]\ntype FooBrand struct{}\n",
		"package test\nfunc main() { x := 42 _ = x }",
		"",
		"package test\n// comment only\n",
		"!!!@#$%^&*()",
	}

	for _, seed := range seedInputs {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		tmpFile, err := os.CreateTemp(t.TempDir(), "fuzz-*.go")
		if err != nil {
			t.Skip("could not create temp file")
		}

		defer func() { _ = os.Remove(tmpFile.Name()) }()

		if _, err := tmpFile.WriteString(input); err != nil {
			_ = tmpFile.Close()

			t.Skip("could not write temp file")
		}

		_ = tmpFile.Close()

		result := &Result{brands: []BrandInfo{}}

		// scanFile must never panic on arbitrary input.
		_ = scanFile(tmpFile.Name(), result)
	})
}
