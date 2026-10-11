package linter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	gofinding "github.com/larsartmann/go-finding"
)

// typeExprFor parses "package p\n\nvar x <src>\n" and returns the type
// expression of x, so type-table tests run against real AST nodes.
func typeExprFor(tb testing.TB, src string) ast.Expr {
	tb.Helper()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "src.go", "package p\n\nvar x "+src+"\n", 0)
	if err != nil {
		tb.Fatalf("ParseFile(%s) error = %v", src, err)
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			return valueSpec.Type
		}
	}

	tb.Fatalf("no var declaration found for %q", src)

	return nil
}

// detectSourceRules scans in-memory source with an explicit rule set.
func detectSourceRules(tb testing.TB, src string, rules RuleSet) []gofinding.Finding {
	tb.Helper()

	return detectFile(writeTestFile(tb, tb.TempDir(), src), rules)
}

// bd003Only is the rule set that selects just the primitive-ID rule.
func bd003Only() RuleSet {
	return RuleSet{BD003: true}
}

func TestDetectPrimitiveIDs_Fixtures(t *testing.T) {
	t.Parallel()

	findings, err := DetectPathWithRules(filepath.Join("testdata", "bd003"), bd003Only())
	if err != nil {
		t.Fatalf("DetectPathWithRules() error = %v", err)
	}

	type wantFinding struct {
		file string
		line int
		name string // "" for the BD002 directive finding
		rule gofinding.RuleName
	}

	join := func(parts ...string) string { return filepath.Join(parts...) }

	want := []wantFinding{
		// cmdwire/main.go is excluded (cmd/ path), fields.go first in walk order.
		{file: join("testdata", "bd003", "fields.go"), line: 9, name: "OrderID", rule: RuleIDBD003},
		{file: join("testdata", "bd003", "fields.go"), line: 10, name: "userID", rule: RuleIDBD003},
		{file: join("testdata", "bd003", "fields.go"), line: 17, name: "ownerID", rule: RuleIDBD003},
		{file: join("testdata", "bd003", "fields.go"), line: 17, name: "tenantID", rule: RuleIDBD003},
		{file: join("testdata", "bd003", "fields.go"), line: 23, name: "sessionTokenID", rule: RuleIDBD003},
		{file: join("testdata", "bd003", "params.go"), line: 8, name: "userId", rule: RuleIDBD003},
		{file: join("testdata", "bd003", "params.go"), line: 11, name: "id", rule: RuleIDBD003},
		{file: join("testdata", "bd003", "params.go"), line: 14, name: "PID", rule: RuleIDBD003},
		{file: join("testdata", "bd003", "returns.go"), line: 6, name: "orderId", rule: RuleIDBD003},
		{file: join("testdata", "bd003", "returns.go"), line: 9, name: "userId", rule: RuleIDBD003},
		{file: join("testdata", "bd003", "suppressed.go"), line: 27, name: "", rule: RuleIDBD002},
	}

	if len(findings) != len(want) {
		t.Fatalf("DetectPathWithRules() found %d findings, want %d:\n%s",
			len(findings), len(want), findingsSummary(findings))
	}

	for i, w := range want {
		f := findings[i]

		if f.Rule != w.rule {
			t.Errorf("findings[%d].Rule = %q, want %q", i, f.Rule, w.rule)
		}

		if f.Position.File != gofinding.FilePath(w.file) {
			t.Errorf("findings[%d].File = %q, want %q", i, f.Position.File, w.file)
		}

		if f.Position.Line != w.line {
			t.Errorf("findings[%d].Line = %d, want %d", i, f.Position.Line, w.line)
		}

		if w.name != "" && !strings.Contains(f.Message, "'"+w.name+"'") {
			t.Errorf("findings[%d].Message = %q, want it to mention %q", i, f.Message, w.name)
		}
	}
}

func TestDetectPrimitiveIDs_DefaultRuleSetIsSilent(t *testing.T) {
	t.Parallel()

	t.Run("fixtures contribute nothing by default", func(t *testing.T) {
		t.Parallel()

		findings, err := DetectPath(filepath.Join("testdata", "bd003"))
		if err != nil {
			t.Fatalf("DetectPath() error = %v", err)
		}

		if len(findings) != 0 {
			t.Errorf("DetectPath() found %d findings, want 0 (BD003 is opt-in):\n%s",
				len(findings), findingsSummary(findings))
		}
	})

	t.Run("primitive IDs in fresh source stay silent by default", func(t *testing.T) {
		t.Parallel()

		findings := detectSource(t, "package x\n\nfunc getUser(userId string) {}\n")
		if len(findings) != 0 {
			t.Errorf("default rules found %d findings, want 0", len(findings))
		}
	})
}

func TestPrimitiveID_FindingShape(t *testing.T) {
	t.Parallel()

	findings := detectSourceRules(t, "package x\n\nfunc getUser(userId string) {}\n", bd003Only())
	if len(findings) != 1 {
		t.Fatalf("detectSourceRules() found %d findings, want 1", len(findings))
	}

	f := findings[0]

	if f.Rule != RuleIDBD003 {
		t.Errorf("Rule = %q, want %q", f.Rule, RuleIDBD003)
	}

	if f.ToolName != ToolName {
		t.Errorf("ToolName = %q, want %q", f.ToolName, ToolName)
	}

	// "user" is a high-priority domain in a function signature: rank high,
	// which maps to error severity (the strong-id conversion table).
	if f.Severity != gofinding.SeverityError {
		t.Errorf("Severity = %q, want %q", f.Severity, gofinding.SeverityError)
	}

	if f.Category != gofinding.CategoryTypeSafety {
		t.Errorf("Category = %q, want %q", f.Category, gofinding.CategoryTypeSafety)
	}

	if f.Confidence != gofinding.ConfidenceHigh {
		t.Errorf("Confidence = %v, want %v", f.Confidence, gofinding.ConfidenceHigh)
	}

	if f.FixStrategy != gofinding.FixStrategySuggest {
		t.Errorf("FixStrategy = %q, want %q", f.FixStrategy, gofinding.FixStrategySuggest)
	}

	if want := "userId string"; f.Snippet != want {
		t.Errorf("Snippet = %q, want %q", f.Snippet, want)
	}

	if want := "Replace with strong ID type UserID"; f.Suggestion != want {
		t.Errorf("Suggestion = %q, want %q", f.Suggestion, want)
	}

	if want := "func getUser(userId string)"; f.BeforeCode != want {
		t.Errorf("BeforeCode = %q, want %q", f.BeforeCode, want)
	}

	if want := "func getUser(userId UserID)"; f.AfterCode != want {
		t.Errorf("AfterCode = %q, want %q", f.AfterCode, want)
	}

	wantMessage := "Replace parameter 'userId' (string) with strong ID type UserID in getUser"
	if f.Message != wantMessage {
		t.Errorf("Message = %q, want %q", f.Message, wantMessage)
	}

	for key, want := range map[string]string{
		"name":              "userId",
		"currentType":       "string",
		"kindDisplay":       "fn param",
		"suggestedTypeName": "UserID",
		"suggestedBrand":    "UserBrand",
		"importPath":        primitiveIDImportPath,
		"typeDefinition":    "type UserID = id.ID[UserBrand, string]\n\ntype UserBrand struct{}",
	} {
		if got := f.Metadata[key]; got != want {
			t.Errorf("Metadata[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestPrimitiveID_ExactTypeSpelling(t *testing.T) {
	t.Parallel()

	findings := detectSourceRules(t, "package x\n\nfunc load(largeID int64) {}\n", bd003Only())
	if len(findings) != 1 {
		t.Fatalf("detectSourceRules() found %d findings, want 1", len(findings))
	}

	if got := findings[0].Metadata["currentType"]; got != "int64" {
		t.Errorf("currentType = %q, want exact spelling %q", got, "int64")
	}
}

func TestPrimitiveID_Suppression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		src              string
		want             []gofinding.RuleName
		wantMessageParts []string
	}{
		{
			name: "trailing directive on the param line",
			src: "package x\n\n" +
				"func keep(userId string) {} //brandid-lint:ignore(BD003) display passthrough\n",
		},
		{
			name: "directive directly above the param line",
			src: "package x\n\nfunc keep(\n" +
				"//brandid-lint:ignore(BD003) external API contract\n" +
				"userId string) {}\n",
		},
		{
			name: "doc directive covers the whole function",
			src: "package x\n\n" +
				"// Doc text.\n" +
				"//brandid-lint:ignore(BD003) legacy wire format\n" +
				"func keep(orderId string) (userId string) { return }\n",
		},
		{
			name: "doc directive covers all struct fields",
			src: "package x\n\n" +
				"//brandid-lint:ignore(BD003) protobuf-generated\n" +
				"type Kept struct {\n\tproductId string\n\torderId int\n}\n",
		},
		{
			name: "directive above a non-ID function suppresses nothing",
			src: "package x\n\n" +
				"//brandid-lint:ignore(BD003) leftover\n" +
				"func clean(count int) {}\n",
			want:             []gofinding.RuleName{RuleIDBD002},
			wantMessageParts: []string{"suppresses nothing"},
		},
		{
			name: "second directive over the same param is redundant",
			src: "package x\n\n" +
				"//brandid-lint:ignore(BD003) reason above\n" +
				"func keep(userId string) {} //brandid-lint:ignore(BD003) reason trailing\n",
			want:             []gofinding.RuleName{RuleIDBD002},
			wantMessageParts: []string{"is redundant"},
		},
		{
			name: "missing reason stays a broken directive",
			src: "package x\n\n" +
				"//brandid-lint:ignore(BD003)\nfunc keep(userId string) {}\n",
			want:             []gofinding.RuleName{RuleIDBD003, RuleIDBD002},
			wantMessageParts: []string{"userId", "missing reason"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings := detectSourceRules(t, tt.src, bd003Only())
			assertFindings(t, findings, tt.want, tt.wantMessageParts)
		})
	}
}

func TestPrimitiveID_DirectivesSilentWhenRuleInactive(t *testing.T) {
	t.Parallel()

	t.Run("well-formed directive is not policed in default mode", func(t *testing.T) {
		t.Parallel()

		findings := detectSource(t, "package x\n\n"+
			"//brandid-lint:ignore(BD003) leftover\nfunc clean(count int) {}\n")
		if len(findings) != 0 {
			t.Errorf("default rules found %d findings, want 0 (BD003 inactive):\n%s",
				len(findings), findingsSummary(findings))
		}
	})

	t.Run("malformed directive stays loud in default mode", func(t *testing.T) {
		t.Parallel()

		findings := detectSource(t, "package x\n\n"+
			"//brandid-lint:ignore(BD003)\nfunc clean(count int) {}\n")
		assertFindings(t, findings, []gofinding.RuleName{RuleIDBD002}, []string{"missing reason"})
	})
}

func TestPrimitiveID_BD003OnlyModeDropsBD001(t *testing.T) {
	t.Parallel()

	src := "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
		"type UnnamedBrand struct{}\n\n" +
		"func use() { _ = id.ID[UnnamedBrand, string]{} }\n\n" +
		"func keep(userId string) {}\n"

	t.Run("BD003-only sees the primitive ID only", func(t *testing.T) {
		t.Parallel()

		findings := detectSourceRules(t, src, bd003Only())
		assertFindings(t, findings, []gofinding.RuleName{RuleIDBD003}, []string{"userId"})
	})

	t.Run("all rules see both", func(t *testing.T) {
		t.Parallel()

		findings := detectSourceRules(t, src, AllRules())
		assertFindings(
			t,
			findings,
			[]gofinding.RuleName{RuleIDBD001, RuleIDBD003},
			[]string{"UnnamedBrand", "userId"},
		)
	})
}

func TestExcludedFromPrimitiveScan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		filename string
		want     bool
	}{
		{filename: "cmd/server/main.go", want: true},
		{filename: "/repo/internal/ui/render.go", want: true},
		{filename: "/repo/pkg/user/handler.go", want: false},
		{filename: "/repo/pkg/user/handler_test.go", want: true},
		{filename: "/repo/tools/gen/main.go", want: true},
		{filename: "/repo/testhelpers/db.go", want: true},
		{filename: "main.go", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()

			if got := excludedFromPrimitiveScan(tt.filename); got != tt.want {
				t.Errorf("excludedFromPrimitiveScan(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestPrimitiveTypeOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		src       string
		wantType  string
		wantPrim  bool
	}{
		{src: "string", wantType: "string", wantPrim: true},
		{src: "int64", wantType: "int64", wantPrim: true},
		{src: "uint", wantType: "uint", wantPrim: true},
		{src: "bool", wantType: "bool", wantPrim: true},
		{src: "uintptr", wantType: "uintptr", wantPrim: true},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			t.Parallel()

			expr := typeExprFor(t, tt.src)
			gotType, gotPrim := primitiveTypeOf(expr)

			if gotType != tt.wantType || gotPrim != tt.wantPrim {
				t.Errorf("primitiveTypeOf(%s) = (%q, %v), want (%q, %v)",
					tt.src, gotType, gotPrim, tt.wantType, tt.wantPrim)
			}
		})
	}

	t.Run("non-identifier types are not primitive", func(t *testing.T) {
		t.Parallel()

		for _, src := range []string{"*string", "[]string", "map[string]int", "UserID", "id.ID"} {
			expr := typeExprFor(t, src)
			if _, ok := primitiveTypeOf(expr); ok {
				t.Errorf("primitiveTypeOf(%s) = true, want false", src)
			}
		}
	})
}
