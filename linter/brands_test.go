package linter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestSuggestName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		brandName string
		want      string
	}{
		{name: "strips Brand suffix", brandName: "UserBrand", want: "User"},
		{name: "strips ID suffix", brandName: "SessionID", want: "Session"},
		{name: "strips stacked suffixes", brandName: "OrderIDBrand", want: "Order"},
		{name: "strips leading T prefix", brandName: "TProductBrand", want: "Product"},
		{name: "keeps real word starting with T", brandName: "TenantBrand", want: "Tenant"},
		{name: "strips suffix leaving lone prefix intact", brandName: "TBrand", want: "T"},
		{name: "falls back to input when empty", brandName: "Brand", want: "Brand"},
		{name: "returns plain name unchanged", brandName: "Account", want: "Account"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := suggestName(tt.brandName); got != tt.want {
				t.Errorf("suggestName(%q) = %q, want %q", tt.brandName, got, tt.want)
			}
		})
	}
}

func TestIsIDSelector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want bool
	}{
		{name: "qualified id.ID", src: "id.ID", want: true},
		{name: "same-package ID", src: "ID", want: true},
		{name: "wrong selector", src: "foo.Bar", want: false},
		{name: "wrong identifier", src: "Type", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			expr := parseExpr(t, tt.src)
			if got := isIDSelector(expr); got != tt.want {
				t.Errorf("isIDSelector(%s) = %v, want %v", tt.src, got, tt.want)
			}
		})
	}
}

func TestBrandTypeArgsFromFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want []string
	}{
		{
			name: "qualified two type parameters",
			src:  "package p\n\nfunc f() { _ = id.ID[UserBrand, string]{} }\n",
			want: []string{"UserBrand"},
		},
		{
			name: "same-package single type parameter",
			src:  "package p\n\nvar _ ID[OrderBrand]\n",
			want: []string{"OrderBrand"},
		},
		{
			name: "pointer type argument",
			src:  "package p\n\nfunc f() { _ = id.ID[*PtrBrand, string]{} }\n",
			want: []string{"PtrBrand"},
		},
		{
			name: "unrelated generic usage",
			src:  "package p\n\nfunc f() { _ = other.Type[NotBrand, string]{} }\n",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := parseSource(t, token.NewFileSet(), tt.src)
			got := brandTypeArgsFromFile(f)

			if len(got) != len(tt.want) {
				t.Fatalf("brandTypeArgsFromFile() found %d brands, want %d", len(got), len(tt.want))
			}

			for _, brand := range tt.want {
				if !got[brand] {
					t.Errorf("brandTypeArgsFromFile() missing brand %q", brand)
				}
			}
		})
	}
}

func TestCollectNameMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		src         string
		typeName    string
		wantFound   bool
		wantLiteral string
	}{
		{
			name: "value receiver with literal",
			src: "package p\n\ntype A struct{}\n\n" +
				"func (A) Name() string { return \"Alpha\" }\n",
			typeName:    "A",
			wantFound:   true,
			wantLiteral: "\"Alpha\"",
		},
		{
			name: "pointer receiver with literal",
			src: "package p\n\ntype B struct{}\n\n" +
				"func (*B) Name() string { return \"Beta\" }\n",
			typeName:    "B",
			wantFound:   true,
			wantLiteral: "\"Beta\"",
		},
		{
			name:        "non-literal return yields placeholder",
			src:         "package p\n\nfunc (C) Name() string { return variable }\n",
			typeName:    "C",
			wantFound:   true,
			wantLiteral: methodNamePlaceholder,
		},
		{
			name:      "non-string return is ignored",
			src:       "package p\n\nfunc (D) Name() int { return 1 }\n",
			typeName:  "D",
			wantFound: false,
		},
		{
			name:      "method without receiver is ignored",
			src:       "package p\n\nfunc Name() string { return \"E\" }\n",
			typeName:  "E",
			wantFound: false,
		},
		{
			name:      "method with parameters is ignored",
			src:       "package p\n\nfunc (F) Name(x int) string { return \"F\" }\n",
			typeName:  "F",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := parseSource(t, token.NewFileSet(), tt.src)
			got := collectNameMethods(f)

			literal, found := got[tt.typeName]
			if found != tt.wantFound {
				t.Fatalf("collectNameMethods() found = %v, want %v", found, tt.wantFound)
			}

			if found && literal != tt.wantLiteral {
				t.Errorf("collectNameMethods() = %q, want %q", literal, tt.wantLiteral)
			}
		})
	}
}

func TestBrandDeclsFromFile(t *testing.T) {
	t.Parallel()

	src := "package p\n\nimport id \"example.com/id\"\n\n" +
		"// Named has a Name method.\n" +
		"type Named struct{}\n\n" +
		"func (Named) Name() string { return \"Named\" }\n\n" +
		"// Unnamed has no Name method.\n" +
		"type Unnamed struct{}\n\n" +
		"type NotUsed struct{}\n\n" +
		"func f() {\n" +
		"\t_ = id.ID[Named, string]{}\n" +
		"\t_ = id.ID[Unnamed, string]{}\n" +
		"}\n"

	fset := token.NewFileSet()
	f := parseSource(t, fset, src)
	decls := brandDeclsFromFile(f, fset)

	if len(decls) != 2 {
		t.Fatalf("brandDeclsFromFile() found %d brands, want 2", len(decls))
	}

	if decls[0].TypeName != "Named" || !decls[0].HasName {
		t.Errorf("decls[0] = %+v, want Named with HasName=true", decls[0])
	}

	if decls[1].TypeName != "Unnamed" || decls[1].HasName {
		t.Errorf("decls[1] = %+v, want Unnamed with HasName=false", decls[1])
	}
}

func TestBrandDeclsFromFile_NoIDUsage(t *testing.T) {
	t.Parallel()

	src := "package p\n\ntype Lonely struct{}\n"
	f := parseSource(t, token.NewFileSet(), src)

	if decls := brandDeclsFromFile(f, token.NewFileSet()); decls != nil {
		t.Errorf("brandDeclsFromFile() = %+v, want nil for files without id.ID usage", decls)
	}
}

// parseSource parses src into an *ast.File, failing the test on parse errors.
// Positions must be resolved with the same fset.
func parseSource(tb testing.TB, fset *token.FileSet, src string) *ast.File {
	tb.Helper()

	f, err := parser.ParseFile(fset, "src.go", src, parser.ParseComments)
	if err != nil {
		tb.Fatalf("ParseFile() error = %v", err)
	}

	return f
}

// parseExpr parses src into an ast.Expr, failing the test on parse errors.
func parseExpr(tb testing.TB, src string) ast.Expr {
	tb.Helper()

	expr, err := parser.ParseExpr(src)
	if err != nil {
		tb.Fatalf("ParseExpr(%q) error = %v", src, err)
	}

	return expr
}
