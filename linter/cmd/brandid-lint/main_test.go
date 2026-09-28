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
