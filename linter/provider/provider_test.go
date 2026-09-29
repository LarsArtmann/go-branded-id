package provider

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/go-branded-id/linter"
	gofinding "github.com/larsartmann/go-finding"
	gofstoolsdk "github.com/larsartmann/go-finding/toolsdk"
)

// findProviderSpec returns the brandid-lint spec from the process-global
// registry, failing the test when it is absent or duplicated.
func findProviderSpec(tb testing.TB) gofstoolsdk.Spec {
	tb.Helper()

	var found []gofstoolsdk.Spec

	for _, spec := range gofstoolsdk.All() {
		if spec.Name == linter.ToolName {
			found = append(found, spec)
		}
	}

	if len(found) != 1 {
		tb.Fatalf(
			"toolsdk registry holds %d specs named %q, want exactly 1",
			len(found),
			linter.ToolName,
		)
	}

	return found[0]
}

func TestProviderRegistered(t *testing.T) {
	t.Parallel()

	spec := findProviderSpec(t)

	if spec.Description == "" {
		t.Error("Spec.Description is empty")
	}

	if spec.Trigger.Language != "go" {
		t.Errorf("Spec.Trigger.Language = %q, want %q", spec.Trigger.Language, "go")
	}

	if len(spec.Trigger.Files) == 0 {
		t.Error("Spec.Trigger.Files is empty")
	}

	if !spec.ModuleFanOut {
		t.Error("Spec.ModuleFanOut = false, want true (brands are a per-module concept)")
	}

	if spec.Detect == nil {
		t.Error("Spec.Detect is nil, want a finding.Detector")
	}

	if spec.Repair == nil {
		t.Error("Spec.Repair is nil, want a Repairer (AST-based Name() insertion)")
	}
}

func TestProviderDetect(t *testing.T) {
	t.Parallel()

	spec := findProviderSpec(t)

	findings, err := spec.Detect.Detect(
		gofinding.WithWorkingDir(context.Background(), "../testdata"))
	if err != nil {
		t.Fatalf("Spec.Detect.Detect() error = %v", err)
	}

	if len(findings) != 6 {
		t.Fatalf("Spec.Detect() found %d findings, want 6 (4 BD001 + 2 BD002)", len(findings))
	}

	for _, f := range findings {
		if f.ToolName != linter.ToolName {
			t.Errorf("finding.ToolName = %q, want %q", f.ToolName, linter.ToolName)
		}
	}
}

func TestProviderDetectorName(t *testing.T) {
	t.Parallel()

	if Provider.Detect == nil {
		t.Fatal("Provider.Detect is nil")
	}

	if got := Provider.Detect.Name(); got != linter.ToolName {
		t.Errorf("Provider.Detect.Name() = %q, want %q", got, linter.ToolName)
	}
}

func TestProviderRepair(t *testing.T) {
	t.Parallel()

	if Provider.Repair == nil {
		t.Fatal("Provider.Repair is nil, want a Repairer")
	}

	dir := t.TempDir()
	src := "package repairme\n\n" +
		"import id \"github.com/larsartmann/go-branded-id\"\n\n" +
		"type FixBrand struct{}\n\n" +
		"func use() { _ = id.ID[FixBrand, string]{} }\n"

	if err := os.WriteFile(filepath.Join(dir, "fix.go"), []byte(src), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := Provider.Repair.Repair(
		gofinding.WithWorkingDir(context.Background(), dir))
	if err != nil {
		t.Fatalf("Provider.Repair.Repair() error = %v", err)
	}

	if result.Description == "" {
		t.Error("RepairResult.Description is empty, want an insertion summary")
	}

	content, readErr := os.ReadFile(filepath.Join(dir, "fix.go"))
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}

	want := "func (FixBrand) Name() string { return \"Fix\" }"
	if !strings.Contains(string(content), want) {
		t.Errorf("repaired file misses %q:\n%s", want, content)
	}
}
