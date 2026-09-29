package linter

import (
	"context"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gofinding "github.com/larsartmann/go-finding"
)

const repairBaseSource = "package x\n\n" +
	"import id \"github.com/larsartmann/go-branded-id\"\n\n" +
	"type UserBrand struct{}\n\n" +
	"func use() { _ = id.ID[UserBrand, string]{} }\n"

// repairSource writes src into its own directory, repairs it, and returns
// the insert count and the repaired file content.
func repairSource(tb testing.TB, src string) (int, string) {
	tb.Helper()

	dir := tb.TempDir()
	writeTestFile(tb, dir, "src.go", src)

	inserted, err := RepairPath(dir)
	if err != nil {
		tb.Fatalf("RepairPath() error = %v", err)
	}

	content, readErr := os.ReadFile(filepath.Join(dir, "src.go"))
	if readErr != nil {
		tb.Fatalf("ReadFile() error = %v", readErr)
	}

	return inserted, string(content)
}

// assertGofmtStable fails when formatted differs from content, proving the
// repair output is gofmt-clean.
func assertGofmtStable(tb testing.TB, content string) {
	tb.Helper()

	formatted, err := format.Source([]byte(content))
	if err != nil {
		tb.Fatalf("format.Source() error = %v (content:\n%s)", err, content)
	}

	if string(formatted) != content {
		tb.Errorf("repaired file is not gofmt-stable:\ngot:\n%s\nwant:\n%s", content, formatted)
	}
}

func TestRepairPath_InsertsStub(t *testing.T) {
	t.Parallel()

	inserted, content := repairSource(t, repairBaseSource)

	if inserted != 1 {
		t.Fatalf("RepairPath() inserted %d, want 1", inserted)
	}

	wantStub := "func (UserBrand) Name() string { return \"User\" }"
	if !strings.Contains(content, wantStub) {
		t.Errorf("repaired content misses stub %q:\n%s", wantStub, content)
	}

	if want := "type UserBrand struct{}\n\n" + wantStub + "\n\nfunc use()"; !strings.Contains(content, want) {
		t.Errorf("repaired content =\n%s\nwant it to contain\n%s", content, want)
	}

	assertGofmtStable(t, content)
}

func TestRepairPath_IsIdempotent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filename := writeTestFile(t, dir, "src.go", repairBaseSource)

	first, err := RepairPath(dir)
	if err != nil {
		t.Fatalf("first RepairPath() error = %v", err)
	}

	if first != 1 {
		t.Fatalf("first RepairPath() inserted %d, want 1", first)
	}

	afterFirst, readErr := os.ReadFile(filename)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}

	second, err := RepairPath(dir)
	if err != nil {
		t.Fatalf("second RepairPath() error = %v", err)
	}

	if second != 0 {
		t.Errorf("second RepairPath() inserted %d, want 0 (idempotent)", second)
	}

	afterSecond, readErr := os.ReadFile(filename)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}

	if string(afterFirst) != string(afterSecond) {
		t.Error("second repair changed the file despite inserting nothing")
	}
}

// TestRepairPath_LeavesDetectionClean is the repair/detection alignment
// proof: after repair, a re-scan must find no BD001 finding.
func TestRepairPath_LeavesDetectionClean(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	writeTestFile(t, dir, "src.go", repairBaseSource)

	if _, err := RepairPath(dir); err != nil {
		t.Fatalf("RepairPath() error = %v", err)
	}

	findings, err := DetectPath(dir)
	if err != nil {
		t.Fatalf("DetectPath() error = %v", err)
	}

	assertFindings(t, findings, nil, nil)
}

func TestRepairPath_SkipsSuppressedAndNamedBrands(t *testing.T) {
	t.Parallel()

	src := "package x\n\n" +
		"import id \"github.com/larsartmann/go-branded-id\"\n\n" +
		"//brandid-lint:ignore(BD001) String() output is a data key\n" +
		"type Marker struct{}\n\n" +
		"type NamedBrand struct{}\n\n" +
		"func (NamedBrand) Name() string { return \"Named\" }\n\n" +
		"type PlainBrand struct{}\n\n" +
		"func use() {\n" +
		"	_ = id.ID[Marker, string]{}\n" +
		"	_ = id.ID[NamedBrand, string]{}\n" +
		"	_ = id.ID[PlainBrand, string]{}\n" +
		"}\n"

	inserted, content := repairSource(t, src)

	if inserted != 1 {
		t.Fatalf("RepairPath() inserted %d, want 1 (only PlainBrand)", inserted)
	}

	if !strings.Contains(content, "func (PlainBrand) Name() string") {
		t.Error("repaired content misses the PlainBrand stub")
	}

	for _, unwanted := range []string{"func (Marker) Name()", "func (NamedBrand) Name()"} {
		if strings.Contains(content, unwanted) {
			t.Errorf("repaired content must not contain %q", unwanted)
		}
	}

	assertGofmtStable(t, content)
}

func TestRepairPath_GroupedTypeBlock(t *testing.T) {
	t.Parallel()

	src := "package x\n\n" +
		"import id \"github.com/larsartmann/go-branded-id\"\n\n" +
		"type (\n" +
		"	AlphaBrand struct{}\n" +
		"	BetaBrand struct{}\n" +
		")\n\n" +
		"func use() {\n" +
		"	_ = id.ID[AlphaBrand, string]{}\n" +
		"	_ = id.ID[BetaBrand, string]{}\n" +
		"}\n"

	inserted, content := repairSource(t, src)

	if inserted != 2 {
		t.Fatalf("RepairPath() inserted %d, want 2", inserted)
	}

	alpha := strings.Index(content, "func (AlphaBrand) Name()")
	beta := strings.Index(content, "func (BetaBrand) Name()")

	if alpha < 0 || beta < 0 {
		t.Fatalf("repaired content misses a stub:\n%s", content)
	}

	if alpha > beta {
		t.Errorf("stub order does not match declaration order:\n%s", content)
	}

	assertGofmtStable(t, content)
}

func TestRepairPath_EdgeWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
	}{
		{
			name: "declaration at EOF without trailing newline",
			src: "package x\n\n" +
				"import id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"type UserBrand struct{}",
		},
		{
			name: "blank line already follows the declaration",
			src: "package x\n\n" +
				"import id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"type UserBrand struct{}\n\n\n" +
				"func use() { _ = id.ID[UserBrand, string]{} }\n",
		},
		{
			name: "two adjacent brands",
			src: "package x\n\n" +
				"import id \"github.com/larsartmann/go-branded-id\"\n\n" +
				"type UserBrand struct{}\n" +
				"type OtherBrand struct{}\n" +
				"func use() { _ = id.ID[UserBrand, string]{}; _ = id.ID[OtherBrand, string]{} }\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inserted, content := repairSource(t, tt.src)

			if inserted < 1 {
				t.Fatalf("RepairPath() inserted %d, want >= 1", inserted)
			}

			if !strings.HasSuffix(content, "\n") {
				t.Errorf("repaired file must end with exactly one newline:\n%q", content)
			}

			assertGofmtStable(t, content)
		})
	}
}

func TestRepairPath_MissingPath(t *testing.T) {
	t.Parallel()

	if _, err := RepairPath(filepath.Join(t.TempDir(), "does_not_exist")); err == nil {
		t.Error("RepairPath() error = nil, want stat failure")
	}
}

func TestRepair_WorkingDirFromContext(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeTestFile(t, dir, "src.go", repairBaseSource)

	inserted, err := Repair(gofinding.WithWorkingDir(context.Background(), dir))
	if err != nil {
		t.Fatalf("Repair() error = %v", err)
	}

	if inserted != 1 {
		t.Fatalf("Repair() inserted %d, want 1", inserted)
	}
}
