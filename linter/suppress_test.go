package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	gofinding "github.com/larsartmann/go-finding"
)

// writeTestFile writes src as a Go file into dir and returns its path.
func writeTestFile(tb testing.TB, dir, name, src string) string {
	tb.Helper()

	filename := filepath.Join(dir, name)

	if err := os.WriteFile(filename, []byte(src), 0o600); err != nil {
		tb.Fatalf("WriteFile(%s) error = %v", filename, err)
	}

	return filename
}

// detectSource writes src into its own directory and returns the findings of
// scanning that directory.
func detectSource(tb testing.TB, src string) []gofinding.Finding {
	tb.Helper()

	return detectFile(writeTestFile(tb, tb.TempDir(), "src.go", src))
}

// assertFindings asserts the exact multiset of rules plus one required
// message substring per finding.
func assertFindings(
	tb testing.TB,
	findings []gofinding.Finding,
	want []gofinding.RuleName,
	wantMessageParts []string,
) {
	tb.Helper()

	if len(findings) != len(want) {
		tb.Fatalf("findings = %v, want %v rules", findingsSummary(findings), want)
	}

	for i, rule := range want {
		if findings[i].Rule != rule {
			tb.Errorf("findings[%d].Rule = %q, want %q", i, findings[i].Rule, rule)
		}

		if len(wantMessageParts) > i && wantMessageParts[i] != "" &&
			!strings.Contains(findings[i].Message, wantMessageParts[i]) {
			tb.Errorf(
				"findings[%d].Message = %q, want it to contain %q",
				i, findings[i].Message, wantMessageParts[i],
			)
		}
	}
}

func findingsSummary(findings []gofinding.Finding) []string {
	rules := make([]string, 0, len(findings))

	for _, f := range findings {
		rules = append(rules, string(f.Rule)+": "+f.Message)
	}

	return rules
}

func TestSuppression_ValidPlacements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		src              string
		want             []gofinding.RuleName
		wantMessageParts []string
	}{
		{
			name: "directly above the declaration",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"// Marker is deliberate.\n" +
				"//brandid-lint:ignore(BD001) String() output is a storage key\n" +
				"type Marker struct{}\n\n" +
				"func use() { _ = id.ID[Marker, string]{} }\n",
		},
		{
			name: "trailing on the declaration line",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"type Marker struct{} //brandid-lint:ignore(BD001) String() output is a storage key\n\n" +
				"func use() { _ = id.ID[Marker, string]{} }\n",
		},
		{
			name: "with leading spaces after the slashes",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"//   brandid-lint:ignore(BD001) String() output is a storage key\n" +
				"type Marker struct{}\n\n" +
				"func use() { _ = id.ID[Marker, string]{} }\n",
		},
		{
			name: "one of several brands in the file",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"//brandid-lint:ignore(BD001) String() output is a storage key\n" +
				"type Marker struct{}\n\n" +
				"type OtherBrand struct{}\n\n" +
				"func use() { _ = id.ID[Marker, string]{}; _ = id.ID[OtherBrand, string]{} }\n",
			want:             []gofinding.RuleName{RuleIDBD001},
			wantMessageParts: []string{"OtherBrand"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings := detectSource(t, tt.src)
			assertFindings(t, findings, tt.want, tt.wantMessageParts)
		})
	}
}

func TestSuppression_BrokenDirectivesStayLoud(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		src              string
		want             []gofinding.RuleName
		wantMessageParts []string
	}{
		{
			name: "missing reason does not suppress",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"//brandid-lint:ignore(BD001)\ntype Marker struct{}\n\n" +
				"func use() { _ = id.ID[Marker, string]{} }\n",
			want:             []gofinding.RuleName{RuleIDBD001, RuleIDBD002},
			wantMessageParts: []string{"Marker", "missing reason"},
		},
		{
			name: "unknown rule does not suppress",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"//brandid-lint:ignore(BD002) reason\ntype Marker struct{}\n\n" +
				"func use() { _ = id.ID[Marker, string]{} }\n",
			want:             []gofinding.RuleName{RuleIDBD001, RuleIDBD002},
			wantMessageParts: []string{"Marker", "unknown rule"},
		},
		{
			name: "missing parentheses do not suppress",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"//brandid-lint:ignore BD001 reason\ntype Marker struct{}\n\n" +
				"func use() { _ = id.ID[Marker, string]{} }\n",
			want:             []gofinding.RuleName{RuleIDBD001, RuleIDBD002},
			wantMessageParts: []string{"Marker", "must be written as"},
		},
		{
			name: "empty rule list does not suppress",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"//brandid-lint:ignore() reason\ntype Marker struct{}\n\n" +
				"func use() { _ = id.ID[Marker, string]{} }\n",
			want:             []gofinding.RuleName{RuleIDBD001, RuleIDBD002},
			wantMessageParts: []string{"Marker", "missing rule ID"},
		},
		{
			name: "stale directive on a brand with Name()",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"//brandid-lint:ignore(BD001) reason\ntype Marker struct{}\n\n" +
				"func (Marker) Name() string { return \"Marker\" }\n\n" +
				"func use() { _ = id.ID[Marker, string]{} }\n",
			want:             []gofinding.RuleName{RuleIDBD002},
			wantMessageParts: []string{"is stale"},
		},
		{
			name: "duplicate directive is redundant",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"//brandid-lint:ignore(BD001) reason one\n" +
				"//brandid-lint:ignore(BD001) reason two\n" +
				"type Marker struct{}\n\n" +
				"func use() { _ = id.ID[Marker, string]{} }\n",
			want:             []gofinding.RuleName{RuleIDBD002},
			wantMessageParts: []string{"is redundant"},
		},
		{
			name: "directive above a non-brand suppresses nothing",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"//brandid-lint:ignore(BD001) reason\n" +
				"const notABrand = 1\n\ntype Marker struct{}\n\n" +
				"func use() { _ = id.ID[Marker, string]{} }\n",
			want:             []gofinding.RuleName{RuleIDBD001, RuleIDBD002},
			wantMessageParts: []string{"Marker", "suppresses nothing"},
		},
		{
			name: "directive not last in its group does not associate",
			src: "package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"//brandid-lint:ignore(BD001) reason\n" +
				"// continuation of the doc comment\n" +
				"type Marker struct{}\n\n" +
				"func use() { _ = id.ID[Marker, string]{} }\n",
			want:             []gofinding.RuleName{RuleIDBD001, RuleIDBD002},
			wantMessageParts: []string{"Marker", "suppresses nothing"},
		},
		{
			name: "directive in a file without id.ID usage is unplaced",
			src: "package x\n\n" +
				"//brandid-lint:ignore(BD001) reason\n" +
				"type Marker struct{}\n\n" +
				"func use() { _ = Marker{} }\n",
			want:             []gofinding.RuleName{RuleIDBD002},
			wantMessageParts: []string{"suppresses nothing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings := detectSource(t, tt.src)
			assertFindings(t, findings, tt.want, tt.wantMessageParts)
		})
	}
}

func TestSuppression_BD002FindingShape(t *testing.T) {
	t.Parallel()

	findings := detectSource(t,
		"package x\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n"+
			"//brandid-lint:ignore(BD001)\ntype Marker struct{}\n\n"+
			"func use() { _ = id.ID[Marker, string]{} }\n",
	)

	assertFindings(t, findings, []gofinding.RuleName{RuleIDBD001, RuleIDBD002}, nil)

	directive := findings[1]
	if directive.ToolName != ToolName {
		t.Errorf("ToolName = %q, want %q", directive.ToolName, ToolName)
	}

	if directive.Severity != gofinding.SeverityWarning {
		t.Errorf("Severity = %q, want %q", directive.Severity, gofinding.SeverityWarning)
	}

	if directive.Category != gofinding.CategoryConfiguration {
		t.Errorf("Category = %q, want %q", directive.Category, gofinding.CategoryConfiguration)
	}

	if directive.Confidence != gofinding.ConfidenceHigh {
		t.Errorf("Confidence = %v, want %v", directive.Confidence, gofinding.ConfidenceHigh)
	}

	// The BD002 finding is positioned at the directive comment (line 6), not
	// at the brand declaration (line 7).
	if directive.Position.Line != 6 {
		t.Errorf("Position.Line = %d, want 6 (the directive line)", directive.Position.Line)
	}

	if want := "//brandid-lint:ignore(BD001)"; directive.Snippet != want {
		t.Errorf("Snippet = %q, want %q", directive.Snippet, want)
	}
}
