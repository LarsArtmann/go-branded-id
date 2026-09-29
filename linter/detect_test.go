package linter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	gofinding "github.com/larsartmann/go-finding"
)

// assertBD001 asserts the identity, classification, and fix fields of one
// BD001 finding.
func assertBD001(
	tb testing.TB,
	f gofinding.Finding,
	wantFile string,
	wantLine int,
	wantType string,
) {
	tb.Helper()

	if f.Rule != RuleIDBD001 {
		tb.Errorf("Rule = %q, want %q", f.Rule, RuleIDBD001)
	}

	if f.ToolName != ToolName {
		tb.Errorf("ToolName = %q, want %q", f.ToolName, ToolName)
	}

	if f.Severity != gofinding.SeverityWarning {
		tb.Errorf("Severity = %q, want %q", f.Severity, gofinding.SeverityWarning)
	}

	if f.Category != gofinding.CategoryNaming {
		tb.Errorf("Category = %q, want %q", f.Category, gofinding.CategoryNaming)
	}

	if f.Confidence != gofinding.ConfidenceHigh {
		tb.Errorf("Confidence = %v, want %v", f.Confidence, gofinding.ConfidenceHigh)
	}

	if f.Position.File != gofinding.FilePath(wantFile) {
		tb.Errorf("Position.File = %q, want %q", f.Position.File, wantFile)
	}

	if f.Position.Line != wantLine {
		tb.Errorf("Position.Line = %d, want %d", f.Position.Line, wantLine)
	}

	wantSuggestion := "func (" + wantType + ") Name() string { return \"" +
		suggestName(wantType) + "\" }"
	if f.Suggestion != wantSuggestion {
		tb.Errorf("Suggestion = %q, want %q", f.Suggestion, wantSuggestion)
	}

	if f.FixStrategy != gofinding.FixStrategyDirect {
		tb.Errorf("FixStrategy = %q, want %q", f.FixStrategy, gofinding.FixStrategyDirect)
	}

	wantBefore := "type " + wantType + " struct{}"
	if f.BeforeCode != wantBefore {
		tb.Errorf("BeforeCode = %q, want %q", f.BeforeCode, wantBefore)
	}

	wantAfter := wantBefore + "\n\n" + wantSuggestion
	if f.AfterCode != wantAfter {
		tb.Errorf("AfterCode = %q, want %q", f.AfterCode, wantAfter)
	}

	wantSnippet := "type " + wantType + " struct{}"
	if f.Snippet != wantSnippet {
		tb.Errorf("Snippet = %q, want %q", f.Snippet, wantSnippet)
	}
}

func TestDetectPath_MissingNameFile(t *testing.T) {
	t.Parallel()

	findings, err := DetectPath(filepath.Join("testdata", "missing_name.go"))
	if err != nil {
		t.Fatalf("DetectPath() error = %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("DetectPath() found %d findings, want 1", len(findings))
	}

	f := findings[0]
	if f.Position.Column != 6 {
		t.Errorf("Position.Column = %d, want 6", f.Position.Column)
	}

	if f.Position.Offset < 0 {
		t.Errorf("Position.Offset = %d, want >= 0", f.Position.Offset)
	}

	wantMessage := "brand type UserBrand is used with id.ID but has no Name() string method"
	if f.Message != wantMessage {
		t.Errorf("Message = %q, want %q", f.Message, wantMessage)
	}

	assertBD001(t, f, filepath.Join("testdata", "missing_name.go"), 7, "UserBrand")
}

func TestDetectPath_Directory(t *testing.T) {
	t.Parallel()

	findings, err := DetectPath("testdata")
	if err != nil {
		t.Fatalf("DetectPath() error = %v", err)
	}

	// WalkDir order is lexicographic; suppression_valid.go and
	// suppression_trailing.go contribute nothing (valid suppressions).
	want := []gofinding.RuleName{
		RuleIDBD001, // missing_name.go: UserBrand
		RuleIDBD001, // mixed.go: TenantBrand
		RuleIDBD001, // mixed.go: SessionBrand
		RuleIDBD001, // suppression_invalid.go: WaitBrand (directive without reason)
		RuleIDBD002, // suppression_invalid.go: missing reason
		RuleIDBD002, // suppression_stale.go: stale directive on named brand
	}
	if len(findings) != len(want) {
		t.Fatalf("DetectPath() found %d findings, want %d", len(findings), len(want))
	}

	for i, rule := range want {
		if findings[i].Rule != rule {
			t.Errorf("findings[%d].Rule = %q, want %q", i, findings[i].Rule, rule)
		}
	}

	if snippet := findings[0].Snippet; snippet != "type UserBrand struct{}" {
		t.Errorf("findings[0].Snippet = %q, want %q", snippet, "type UserBrand struct{}")
	}
}

func TestDetectPath_CleanFiles(t *testing.T) {
	t.Parallel()

	cleanFiles := []string{"has_name.go", "no_id_usage.go", "pointer_receiver.go"}

	for _, name := range cleanFiles {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			findings, err := DetectPath(filepath.Join("testdata", name))
			if err != nil {
				t.Fatalf("DetectPath() error = %v", err)
			}

			if len(findings) != 0 {
				t.Errorf("DetectPath() found %d findings, want 0", len(findings))
			}
		})
	}
}

func TestDetectPath_MissingPath(t *testing.T) {
	t.Parallel()

	if _, err := DetectPath(filepath.Join("testdata", "does_not_exist.go")); err == nil {
		t.Error("DetectPath() error = nil, want stat failure")
	}
}

func TestDetectPath_SkipsUnparseableFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	broken := filepath.Join(dir, "broken.go")

	if err := os.WriteFile(broken, []byte("this is not go source"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	findings, err := DetectPath(dir)
	if err != nil {
		t.Fatalf("DetectPath() error = %v, want nil (broken files are skipped)", err)
	}

	if len(findings) != 0 {
		t.Errorf("DetectPath() found %d findings, want 0", len(findings))
	}
}

func TestDetectPath_SkipsIgnoredDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	for _, sub := range []string{"vendor", ".git", "node_modules"} {
		nested := filepath.Join(dir, sub, "nested")

		if err := os.MkdirAll(nested, 0o750); err != nil {
			t.Fatalf("MkdirAll(%s) error = %v", nested, err)
		}

		src := "package nested\n\ntype BadBrand struct{}\n"
		if err := os.WriteFile(filepath.Join(nested, "bad.go"), []byte(src), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	rootSrc := "package root\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
		"type RootBrand struct{}\n\nfunc use() { _ = id.ID[RootBrand, string]{} }\n"
	if err := os.WriteFile(filepath.Join(dir, "root.go"), []byte(rootSrc), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	findings, err := DetectPath(dir)
	if err != nil {
		t.Fatalf("DetectPath() error = %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("DetectPath() found %d findings, want 1 (ignored dirs skipped)", len(findings))
	}

	assertBD001(t, findings[0], filepath.Join(dir, "root.go"), 5, "RootBrand")
}

func TestDetect_UsesWorkingDirFromContext(t *testing.T) {
	t.Parallel()

	findings, err := Detect(gofinding.WithWorkingDir(context.Background(), "testdata"))
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if len(findings) != 6 {
		t.Fatalf("Detect() found %d findings, want 6", len(findings))
	}
}

func TestDetect_FallsBackToProcessWorkingDirectory(t *testing.T) {
	t.Parallel()

	findings, err := Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	// The process working directory is the linter package itself, whose
	// non-test files declare no brands; testdata is excluded from scanning
	// only in vendor/.git/node_modules terms, so its fixtures ARE scanned
	// here and must yield the same six findings (4 BD001 + 2 BD002).
	if len(findings) != 6 {
		t.Fatalf("Detect() found %d findings, want 6", len(findings))
	}
}
