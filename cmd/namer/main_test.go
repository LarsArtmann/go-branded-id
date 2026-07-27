package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// parseSource parses a Go source string and returns the *ast.File for testing.
func parseSource(t *testing.T, src string) *ast.File {
	t.Helper()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	return f
}

// capturePrint calls printResults and captures its output via an os.Pipe.
// This tests the real printResults function without duplicating its logic.
func capturePrint(t *testing.T, brands []BrandInfo, dryRun bool) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}

	printResults(w, brands, dryRun)

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe writer: %v", err)
	}

	var buf bytes.Buffer

	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read pipe: %v", err)
	}

	return buf.String()
}

func TestSuggestName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		brandName string
		want      string
	}{
		{"strips Brand suffix", "UserBrand", "User"},
		{"strips ID suffix", "OrderID", "Order"},
		{"strips T prefix", "TProduct", "Product"},
		{"strips Brand and ID suffixes", "EventBrandID", "Event"},
		{"strips T prefix and Brand suffix", "TCategoryBrand", "Category"},
		{"returns original when result empty", "Brand", "Brand"},
		{"returns original when only suffix", "ID", "ID"},
		{"preserves T when part of word", "Tenant", "Tenant"},
		{"returns original when no suffixes", "Product", "Product"},
		{"handles empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := suggestName(tt.brandName)
			if got != tt.want {
				t.Errorf("suggestName(%q) = %q, want %q", tt.brandName, got, tt.want)
			}
		})
	}
}

func TestFilterMissing(t *testing.T) {
	t.Parallel()

	t.Run("returns only brands without Name", func(t *testing.T) {
		t.Parallel()

		brands := []BrandInfo{
			{TypeName: "With", HasName: true},
			{TypeName: "Without1", HasName: false},
			{TypeName: "Without2", HasName: false},
		}

		missing := filterMissing(brands)
		if len(missing) != 2 {
			t.Fatalf("expected 2 missing, got %d", len(missing))
		}

		if missing[0].TypeName != "Without1" || missing[1].TypeName != "Without2" {
			t.Errorf("unexpected brands: %v", missing)
		}
	})

	t.Run("returns empty when all have Name", func(t *testing.T) {
		t.Parallel()

		brands := []BrandInfo{
			{TypeName: "A", HasName: true},
			{TypeName: "B", HasName: true},
		}

		if missing := filterMissing(brands); len(missing) != 0 {
			t.Errorf("expected 0 missing, got %d", len(missing))
		}
	})

	t.Run("returns all when none have Name", func(t *testing.T) {
		t.Parallel()

		brands := []BrandInfo{
			{TypeName: "A", HasName: false},
			{TypeName: "B", HasName: false},
		}

		if missing := filterMissing(brands); len(missing) != 2 {
			t.Errorf("expected 2 missing, got %d", len(missing))
		}
	})

	t.Run("returns empty for nil slice", func(t *testing.T) {
		t.Parallel()

		if missing := filterMissing(nil); len(missing) != 0 {
			t.Errorf("expected 0 missing for nil, got %d", len(missing))
		}
	})
}

func TestIsIDSelector(t *testing.T) {
	t.Parallel()

	t.Run("recognizes qualified id.ID selector", func(t *testing.T) {
		t.Parallel()

		expr := &ast.SelectorExpr{
			X:   &ast.Ident{Name: "id"},
			Sel: &ast.Ident{Name: "ID"},
		}

		if !isIDSelector(expr) {
			t.Error("expected id.ID selector to be recognized")
		}
	})

	t.Run("recognizes bare ID identifier", func(t *testing.T) {
		t.Parallel()

		expr := &ast.Ident{Name: "ID"}

		if !isIDSelector(expr) {
			t.Error("expected bare ID identifier to be recognized")
		}
	})

	t.Run("rejects non-ID selector", func(t *testing.T) {
		t.Parallel()

		expr := &ast.SelectorExpr{
			X:   &ast.Ident{Name: "id"},
			Sel: &ast.Ident{Name: "Foo"},
		}

		if isIDSelector(expr) {
			t.Error("expected non-ID selector to be rejected")
		}
	})

	t.Run("rejects non-ID identifier", func(t *testing.T) {
		t.Parallel()

		expr := &ast.Ident{Name: "Foo"}

		if isIDSelector(expr) {
			t.Error("expected non-ID identifier to be rejected")
		}
	})
}

func TestTypeNameExtraction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{"ident extracts name", &ast.Ident{Name: "UserBrand"}, "UserBrand"},
		{
			"star expr extracts inner name",
			&ast.StarExpr{X: &ast.Ident{Name: "UserBrand"}},
			"UserBrand",
		},
		{"unsupported returns empty", &ast.BasicLit{Kind: token.STRING, Value: "\"foo\""}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/typeNameFromExpr", func(t *testing.T) {
			t.Parallel()

			if got := typeNameFromExpr(tt.expr); got != tt.want {
				t.Errorf("typeNameFromExpr() = %q, want %q", got, tt.want)
			}
		})

		t.Run(tt.name+"/receiverTypeName", func(t *testing.T) {
			t.Parallel()

			if got := receiverTypeName(tt.expr); got != tt.want {
				t.Errorf("receiverTypeName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsStringType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr ast.Expr
		want bool
	}{
		{"string ident", &ast.Ident{Name: "string"}, true},
		{"pointer to string", &ast.StarExpr{X: &ast.Ident{Name: "string"}}, true},
		{"non-string ident", &ast.Ident{Name: "int"}, false},
		{"int literal", &ast.BasicLit{Kind: token.INT, Value: "42"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isStringType(tt.expr); got != tt.want {
				t.Errorf("isStringType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectNameMethods(t *testing.T) {
	t.Parallel()

	t.Run("finds Name method with string return value", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Name() string { return "FooValue" }
`
		f := parseSource(t, src)
		methods := collectNameMethods(f)

		val, ok := methods["Foo"]
		if !ok {
			t.Fatal("expected Foo to have Name method recorded")
		}

		if val != "\"FooValue\"" {
			t.Errorf("expected return value \"FooValue\", got %q", val)
		}
	})

	t.Run("records method placeholder when return not literal", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Name() string { return someFunc() }
`
		f := parseSource(t, src)
		methods := collectNameMethods(f)

		val, ok := methods["Foo"]
		if !ok {
			t.Fatal("expected Foo to have Name method recorded")
		}

		if val != "(method on T)" {
			t.Errorf("expected placeholder, got %q", val)
		}
	})

	t.Run("ignores non-Name methods", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Other() string { return "no" }
`
		f := parseSource(t, src)
		methods := collectNameMethods(f)

		if _, ok := methods["Foo"]; ok {
			t.Error("expected Foo to not be recorded (method is not Name)")
		}
	})

	t.Run("ignores Name with wrong signature", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Name() int { return 42 }
`
		f := parseSource(t, src)
		methods := collectNameMethods(f)

		if _, ok := methods["Foo"]; ok {
			t.Error("expected Foo to not be recorded (returns int, not string)")
		}
	})

	t.Run("handles pointer receiver", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (*Foo) Name() string { return "Ptr" }
`
		f := parseSource(t, src)
		methods := collectNameMethods(f)

		val, ok := methods["Foo"]
		if !ok {
			t.Fatal("expected Foo to have Name method recorded")
		}

		if val != "\"Ptr\"" {
			t.Errorf("expected return value \"Ptr\", got %q", val)
		}
	})
}

func TestIsNameMethod(t *testing.T) {
	t.Parallel()

	t.Run("valid Name method on value receiver", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Name() string { return "x" }
`
		f := parseSource(t, src)

		fn := findFuncDecl(t, f, "Name")

		typeName, ok := isNameMethod(fn)
		if !ok {
			t.Fatal("expected isNameMethod to return true")
		}

		if typeName != "Foo" {
			t.Errorf("expected type name 'Foo', got %q", typeName)
		}
	})

	t.Run("rejects function with wrong name", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Other() string { return "x" }
`
		f := parseSource(t, src)

		fn := findFuncDecl(t, f, "Other")
		if _, ok := isNameMethod(fn); ok {
			t.Error("expected isNameMethod to return false for non-Name function")
		}
	})

	t.Run("rejects function without receiver", func(t *testing.T) {
		t.Parallel()

		src := `package test
func Name() string { return "x" }
`
		f := parseSource(t, src)

		fn := findFuncDecl(t, f, "Name")
		if _, ok := isNameMethod(fn); ok {
			t.Error("expected isNameMethod to return false for no receiver")
		}
	})

	t.Run("rejects Name with parameters", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Name(extra string) string { return "x" }
`
		f := parseSource(t, src)

		fn := findFuncDecl(t, f, "Name")
		if _, ok := isNameMethod(fn); ok {
			t.Error("expected isNameMethod to return false for method with parameters")
		}
	})

	t.Run("rejects Name with no return values", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Name() {}
`
		f := parseSource(t, src)

		fn := findFuncDecl(t, f, "Name")
		if _, ok := isNameMethod(fn); ok {
			t.Error("expected isNameMethod to return false for void method")
		}
	})

	t.Run("rejects Name returning non-string", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Name() int { return 42 }
`
		f := parseSource(t, src)

		fn := findFuncDecl(t, f, "Name")
		if _, ok := isNameMethod(fn); ok {
			t.Error("expected isNameMethod to return false for non-string return")
		}
	})
}

func TestParseNameReturnValue(t *testing.T) {
	t.Parallel()

	t.Run("extracts string literal return value", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Name() string { return "UserValue" }
`
		f := parseSource(t, src)
		fn := findFuncDecl(t, f, "Name")

		if got := parseNameReturnValue(fn); got != "\"UserValue\"" {
			t.Errorf("expected \"UserValue\", got %q", got)
		}
	})

	t.Run("returns empty for non-literal return", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Name() string { return someVar }
`
		f := parseSource(t, src)
		fn := findFuncDecl(t, f, "Name")

		if got := parseNameReturnValue(fn); got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})

	t.Run("returns empty for multiple statements", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func (Foo) Name() string { x := "a"; return x }
`
		f := parseSource(t, src)
		fn := findFuncDecl(t, f, "Name")

		if got := parseNameReturnValue(fn); got != "" {
			t.Errorf("expected empty for multi-statement body, got %q", got)
		}
	})
}

func TestBrandTypeArgsFromFile(t *testing.T) {
	t.Parallel()

	t.Run("finds qualified id.ID[Brand, Value] usage", func(t *testing.T) {
		t.Parallel()

		src := `package test
type UserBrand struct{}
type OrderBrand struct{}
func f() {
	var _ id.ID[UserBrand, string]
	var _ id.ID[OrderBrand, int64]
}`
		f := parseSource(t, src)
		brands := brandTypeArgsFromFile(f)

		if !brands["UserBrand"] {
			t.Error("expected UserBrand to be found")
		}

		if !brands["OrderBrand"] {
			t.Error("expected OrderBrand to be found")
		}

		if len(brands) != 2 {
			t.Errorf("expected 2 brands, got %d", len(brands))
		}
	})

	t.Run("finds bare ID[Brand, Value] usage", func(t *testing.T) {
		t.Parallel()

		src := `package test
type CatBrand struct{}
func f() {
	var _ ID[CatBrand, string]
}`
		f := parseSource(t, src)
		brands := brandTypeArgsFromFile(f)

		if !brands["CatBrand"] {
			t.Error("expected CatBrand to be found with bare ID")
		}
	})

	t.Run("finds single type arg id.ID[Brand]", func(t *testing.T) {
		t.Parallel()

		src := `package test
type SoloBrand struct{}
func f() {
	var _ id.ID[SoloBrand]
}`
		f := parseSource(t, src)
		brands := brandTypeArgsFromFile(f)

		if !brands["SoloBrand"] {
			t.Error("expected SoloBrand to be found")
		}
	})

	t.Run("returns empty when no id.ID usage", func(t *testing.T) {
		t.Parallel()

		src := `package test
type Foo struct{}
func f() { _ = Foo{} }
`
		f := parseSource(t, src)
		brands := brandTypeArgsFromFile(f)

		if len(brands) != 0 {
			t.Errorf("expected 0 brands, got %d: %v", len(brands), brands)
		}
	})
}

func TestScanFile_MissingName(t *testing.T) {
	t.Parallel()

	r := &Result{brands: []BrandInfo{}}

	err := scanFile(filepath.Join("testdata", "missing_name.go"), r)
	if err != nil {
		t.Fatalf("scanFile failed: %v", err)
	}

	if len(r.brands) != 1 {
		t.Fatalf("expected 1 brand, got %d", len(r.brands))
	}

	brand := r.brands[0]
	if brand.TypeName != "UserBrand" {
		t.Errorf("expected UserBrand, got %q", brand.TypeName)
	}

	if brand.HasName {
		t.Error("expected HasName to be false")
	}
}

func TestScanFile_HasName(t *testing.T) {
	t.Parallel()

	r := &Result{brands: []BrandInfo{}}

	err := scanFile(filepath.Join("testdata", "has_name.go"), r)
	if err != nil {
		t.Fatalf("scanFile failed: %v", err)
	}

	if len(r.brands) != 1 {
		t.Fatalf("expected 1 brand, got %d", len(r.brands))
	}

	brand := r.brands[0]
	if brand.TypeName != "OrderBrand" {
		t.Errorf("expected OrderBrand, got %q", brand.TypeName)
	}

	if !brand.HasName {
		t.Error("expected HasName to be true")
	}

	if brand.NameValue != "\"Order\"" {
		t.Errorf("expected NameValue \"Order\", got %q", brand.NameValue)
	}
}

func TestScanFile_NoIDUsage(t *testing.T) {
	t.Parallel()

	r := &Result{brands: []BrandInfo{}}

	err := scanFile(filepath.Join("testdata", "no_id_usage.go"), r)
	if err != nil {
		t.Fatalf("scanFile failed: %v", err)
	}

	if len(r.brands) != 0 {
		t.Errorf("expected 0 brands (empty struct not used with id.ID), got %d", len(r.brands))
	}
}

func TestScanFile_Mixed(t *testing.T) {
	t.Parallel()

	r := &Result{brands: []BrandInfo{}}

	err := scanFile(filepath.Join("testdata", "mixed.go"), r)
	if err != nil {
		t.Fatalf("scanFile failed: %v", err)
	}

	if len(r.brands) != 3 {
		t.Fatalf(
			"expected 3 brands (Product, Tenant, Session — not NotABrand), got %d",
			len(r.brands),
		)
	}

	byType := make(map[string]BrandInfo, len(r.brands))
	for _, b := range r.brands {
		byType[b.TypeName] = b
	}

	if !byType["ProductBrand"].HasName {
		t.Error("ProductBrand should have Name")
	}

	if byType["TenantBrand"].HasName {
		t.Error("TenantBrand should NOT have Name")
	}

	if byType["SessionBrand"].HasName {
		t.Error("SessionBrand should NOT have Name")
	}

	if _, ok := byType["NotABrand"]; ok {
		t.Error("NotABrand should not be included (not used with id.ID)")
	}
}

func TestScanPath_Directory(t *testing.T) {
	t.Parallel()

	r := &Result{brands: []BrandInfo{}}

	err := scanPath("testdata", r)
	if err != nil {
		t.Fatalf("scanPath failed: %v", err)
	}

	if len(r.brands) == 0 {
		t.Fatal("expected to find brands across testdata directory")
	}

	types := make(map[string]bool, len(r.brands))
	for _, b := range r.brands {
		types[b.TypeName] = true
	}

	for _, expected := range []string{"UserBrand", "OrderBrand", "ProductBrand", "TenantBrand", "SessionBrand"} {
		if !types[expected] {
			t.Errorf("expected to find %s in scan results", expected)
		}
	}
}

func TestScanPath_SingleFile(t *testing.T) {
	t.Parallel()

	r := &Result{brands: []BrandInfo{}}

	err := scanPath(filepath.Join("testdata", "missing_name.go"), r)
	if err != nil {
		t.Fatalf("scanPath failed: %v", err)
	}

	if len(r.brands) != 1 {
		t.Fatalf("expected 1 brand from single file, got %d", len(r.brands))
	}
}

func TestScanPath_NonExistent(t *testing.T) {
	t.Parallel()

	r := &Result{brands: []BrandInfo{}}

	err := scanPath("testdata/nonexistent.go", r)
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestPrintResults_NoBrands(t *testing.T) {
	t.Parallel()

	output := capturePrint(t, nil, true)

	if !strings.Contains(output, "No brand types found") {
		t.Errorf("expected 'No brand types found' message, got: %q", output)
	}
}

func TestPrintResults_AllHaveName(t *testing.T) {
	t.Parallel()

	brands := []BrandInfo{
		{TypeName: "Foo", HasName: true, NameValue: "\"Foo\""},
		{TypeName: "Bar", HasName: true, NameValue: "\"Bar\""},
	}

	output := capturePrint(t, brands, true)

	if !strings.Contains(output, "All 2 brand types have Name()") {
		t.Errorf("expected 'All 2 brand types have Name()' in output, got: %s", output)
	}
}

func TestPrintResults_MissingName(t *testing.T) {
	t.Parallel()

	brands := []BrandInfo{
		{TypeName: "UserBrand", HasName: false, File: "main.go", Line: 10},
		{TypeName: "OrderBrand", HasName: true, NameValue: "\"Order\""},
		{TypeName: "TenantBrand", HasName: false, File: "main.go", Line: 20},
	}

	output := capturePrint(t, brands, true)

	if !strings.Contains(output, "Found 3 brand types") {
		t.Errorf("expected 'Found 3 brand types' in output, got: %s", output)
	}

	if !strings.Contains(output, "2 missing Name()") {
		t.Errorf("expected '2 missing Name()' in output, got: %s", output)
	}

	if !strings.Contains(output, "UserBrand") {
		t.Errorf("expected 'UserBrand' in output, got: %s", output)
	}

	if !strings.Contains(output, "TenantBrand") {
		t.Errorf("expected 'TenantBrand' in output, got: %s", output)
	}

	if strings.Contains(output, "func (OrderBrand)") {
		t.Errorf("OrderBrand has Name() — should NOT appear as a fix suggestion, got: %s", output)
	}

	if !strings.Contains(output, "(dry-run") {
		t.Errorf("expected '(dry-run' marker in dry-run output, got: %s", output)
	}
}

func TestPrintResults_WriteMode(t *testing.T) {
	t.Parallel()

	brands := []BrandInfo{
		{TypeName: "UserBrand", HasName: false, File: "main.go", Line: 10},
	}

	output := capturePrint(t, brands, false)

	if strings.Contains(output, "(dry-run") {
		t.Errorf("dry-run marker should not appear in write mode, got: %s", output)
	}

	if !strings.Contains(output, "AST-based file insertion not implemented") {
		t.Errorf("expected write mode note, got: %s", output)
	}
}

func TestPrintResults_SuggestNameIntegration(t *testing.T) {
	t.Parallel()

	brands := []BrandInfo{
		{TypeName: "SessionBrand", HasName: false, File: "main.go", Line: 5},
	}

	output := capturePrint(t, brands, true)

	expected := `func (SessionBrand) Name() string { return "Session" }`
	if !strings.Contains(output, expected) {
		t.Errorf(
			"expected suggested Name() stub in output:\n%s\nwant to contain: %s",
			output,
			expected,
		)
	}
}

// findFuncDecl finds the first FuncDecl with the given name in f.
func findFuncDecl(t *testing.T, f *ast.File, name string) *ast.FuncDecl {
	t.Helper()

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if fn.Name.Name == name {
			return fn
		}
	}

	t.Fatalf("function %q not found in source", name)

	return nil
}
