package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/go-branded-id/linter"
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

//nolint:paralleltest // t.Chdir locks the process working directory
func TestRun_NoPathsDefaultsToWorkingDir(t *testing.T) {
	t.Chdir(cleanDir(t))

	var stdout, stderr bytes.Buffer

	if code := run(nil, &stdout, &stderr); code != 0 {
		t.Errorf("run() exit code = %d, want 0 (clean working directory)", code)
	}

	if stderr.String() != "" {
		t.Errorf("stderr = %q, want it empty instead of a usage lecture", stderr.String())
	}
}

func TestRun_RulesDefaultExcludesBD003(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	// The BD003 fixtures are full of primitive IDs; the default rule set
	// must not report a single one.
	code := run([]string{filepath.Join("..", "..", "testdata", "bd003")}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run() exit code = %d, want 0 (BD003 is opt-in)", code)
	}
}

func TestRun_RulesBD003ReportsPrimitiveIDs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{
		"-rules", "BD003",
		filepath.Join("..", "..", "testdata", "bd003", "params.go"),
	}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("run() exit code = %d, want 1", code)
	}

	if !strings.Contains(stdout.String(), "userId") {
		t.Errorf("stdout = %q, want it to mention userId", stdout.String())
	}
}

func TestRun_UnknownRuleExitTwo(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{"-rules", "BD999", cleanDir(t)}, &stdout, &stderr)

	if code != 2 {
		t.Errorf("run() exit code = %d, want 2", code)
	}

	if !strings.Contains(stderr.String(), "unknown rule") {
		t.Errorf("stderr = %q, want unknown-rule error", stderr.String())
	}
}

func TestRun_BD002AcceptedAsNoop(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{"-rules", "BD001,BD002,BD003", cleanDir(t)}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run() exit code = %d, want 0 (BD002 rides along, it is not an error)", code)
	}
}

func TestRun_AllRulesAlias(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{
		"-rules", "all",
		filepath.Join("..", "..", "testdata", "bd003", "params.go"),
	}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("run() exit code = %d, want 1 (all includes BD003)", code)
	}

	if !strings.Contains(stdout.String(), "userId") {
		t.Errorf("stdout = %q, want it to mention userId", stdout.String())
	}
}

func TestRun_Version(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{"-version"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run() exit code = %d, want 0", code)
	}

	want := linter.ToolName + " " + linter.Version + "\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRun_FlagsAfterPaths(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	// "."-style paths first used to end flag parsing, so "-rules" became a
	// stat target; the permuted parse must accept the same invocation shape.
	code := run([]string{cleanDir(t), "-rules", "all", "-format", "text"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run() exit code = %d, want 0 (flags after paths are flags)", code)
	}

	if stderr.String() != "" {
		t.Errorf("stderr = %q, want it empty", stderr.String())
	}
}

func TestRun_PatternArguments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{filepath.ToSlash(cleanDir(t)) + "/..."}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run() exit code = %d, want 0 (./... expands to the tree root)", code)
	}
}

//nolint:paralleltest // t.Chdir locks the process working directory
func TestRun_FixWithoutPathsDefaultsToWorkingDir(t *testing.T) {
	t.Chdir(cleanDir(t))

	var stdout, stderr bytes.Buffer

	if code := run([]string{"-fix"}, &stdout, &stderr); code != 0 {
		t.Errorf("run() exit code = %d, want 0", code)
	}
}

func TestRun_FixWithoutBD001ExitTwo(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{"-rules", "BD003", "-fix", cleanDir(t)}, &stdout, &stderr)

	if code != 2 {
		t.Errorf("run() exit code = %d, want 2", code)
	}

	if !strings.Contains(stderr.String(), "BD001") {
		t.Errorf("stderr = %q, want the BD001 repair hint", stderr.String())
	}
}

func TestRun_MissingPathExitTwo(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := run([]string{filepath.Join(t.TempDir(), "absent.go")}, &stdout, &stderr)

	if code != 2 {
		t.Errorf("run() exit code = %d, want 2", code)
	}

	// The error names the path exactly once: the phase label plus the raw
	// *fs.PathError, never the triple-wrapped stat chains of older versions.
	if got := stderr.String(); strings.Count(got, "stat ") != 1 {
		t.Errorf("stderr = %q, want the stat error exactly once", got)
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
