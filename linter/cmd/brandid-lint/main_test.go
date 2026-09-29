package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cleanDir returns a directory whose only Go file declares a brand that HAS a
// Name() method, so scanning it yields no findings.
func cleanDir(tb testing.TB) string {
	tb.Helper()

	dir := tb.TempDir()
	src := "package clean\n\nimport id \"github.com/larsartmann/go-branded-id\"\n\n" +
		"type OkBrand struct{}\n\n" +
		"func (OkBrand) Name() string { return \"Ok\" }\n\n" +
		"func use() { _ = id.ID[OkBrand, string]{} }\n"

	if err := os.WriteFile(filepath.Join(dir, "ok.go"), []byte(src), 0o600); err != nil {
		tb.Fatalf("WriteFile() error = %v", err)
	}

	return dir
}

func TestRun_FindingsExitOne(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{"-format", "text", "../../testdata"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("run() exit code = %d, want 1", code)
	}

	if !strings.Contains(stdout.String(), "UserBrand") {
		t.Errorf("stdout = %q, want it to mention UserBrand", stdout.String())
	}
}

func TestRun_CleanExitZero(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{cleanDir(t)}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run() exit code = %d, want 0", code)
	}
}

func TestRun_SarifFormat(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{"-format", "sarif", "../../testdata/missing_name.go"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("run() exit code = %d, want 1", code)
	}

	if !strings.Contains(stdout.String(), "\"runs\"") {
		t.Errorf("stdout = %q, want SARIF with runs", stdout.String())
	}
}

func TestRun_UnknownFormatExitTwo(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{"-format", "yaml", "../../testdata/missing_name.go"}, &stdout, &stderr)

	if code != 2 {
		t.Errorf("run() exit code = %d, want 2", code)
	}
}

func TestRun_NoPathsExitTwo(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	if code := run(nil, &stdout, &stderr); code != 2 {
		t.Errorf("run() exit code = %d, want 2", code)
	}

	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("stderr = %q, want usage text", stderr.String())
	}
}

func TestRun_MissingPathExitTwo(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{filepath.Join(t.TempDir(), "absent.go")}, &stdout, &stderr)

	if code != 2 {
		t.Errorf("run() exit code = %d, want 2", code)
	}
}

func TestRun_FixAppliesRepairs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := "package fixme\n\n" +
		"import id \"github.com/larsartmann/go-branded-id\"\n\n" +
		"type FixBrand struct{}\n\n" +
		"func use() { _ = id.ID[FixBrand, string]{} }\n"

	filename := filepath.Join(dir, "fix.go")
	if err := os.WriteFile(filename, []byte(src), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout, stderr bytes.Buffer

	code := run([]string{"-fix", dir}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run() exit code = %d, want 0 (repair clears all findings)", code)
	}

	if !strings.Contains(stderr.String(), "inserted 1") {
		t.Errorf("stderr = %q, want an insertion summary", stderr.String())
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if want := "func (FixBrand) Name() string { return \"Fix\" }"; !strings.Contains(
		string(content),
		want,
	) {
		t.Errorf("repaired file misses %q:\n%s", want, content)
	}
}

func TestRun_FixLeavesBrokenDirectivesLoud(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := "package fixme\n\n" +
		"import id \"github.com/larsartmann/go-branded-id\"\n\n" +
		"//brandid-lint:ignore(BD001)\ntype NoReasonBrand struct{}\n\n" +
		"func use() { _ = id.ID[NoReasonBrand, string]{} }\n"

	filename := filepath.Join(dir, "noreason.go")
	if err := os.WriteFile(filename, []byte(src), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout, stderr bytes.Buffer

	code := run([]string{"-fix", dir}, &stdout, &stderr)

	// The missing-reason directive yields BD002 after repair; findings
	// remain, so the exit code stays 1.
	if code != 1 {
		t.Errorf("run() exit code = %d, want 1 (BD002 remains after repair)", code)
	}

	if !strings.Contains(stdout.String(), "BD002") &&
		!strings.Contains(stdout.String(), "missing reason") {
		t.Errorf("stdout = %q, want it to report the broken directive", stdout.String())
	}
}
