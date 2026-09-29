package linter

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	gofinding "github.com/larsartmann/go-finding"
)

// defaultFileMode is the fallback permission for files whose mode cannot be
// stat'ed; repaired files keep their existing permissions otherwise.
const defaultFileMode = 0o644

// errOffsetOutOfBounds is wrapped when a declaration's insertion point no
// longer lies inside the file content read from disk.
var errOffsetOutOfBounds = errors.New("insertion offset outside file bounds")

// Repair runs the BD001 repair on the working directory carried by ctx (see
// gofinding.WorkingDirFromContext), falling back to the process working
// directory. This is the entry point the BuildFlow provider calls.
func Repair(ctx context.Context) (int, error) {
	root := gofinding.WorkingDirFromContext(ctx)
	if root == "" {
		root = "."
	}

	return RepairPath(root)
}

// RepairPath scans path and inserts the suggested Name() stub after every
// unsuppressed brand declaration flagged by BD001, one stub per brand. It
// returns the number of applied insertions. Suppressed brands are never
// touched — a documented exception must stay an exception — and repair is
// idempotent: brands that already have Name() produce no finding, so a
// second run inserts nothing.
func RepairPath(path string) (int, error) {
	inserted := 0

	walkErr := walkGoFiles(path, func(filename string) error {
		count, repairErr := repairFile(filename)
		if repairErr != nil {
			return fmt.Errorf("repair %s: %w", filename, repairErr)
		}

		inserted += count

		return nil
	})
	if walkErr != nil {
		return inserted, walkErr
	}

	return inserted, nil
}

// repairFile inserts Name() stubs into one file for every unsuppressed,
// still-unnamed brand declaration. It returns the number of insertions.
func repairFile(filename string) (int, error) {
	scan := scanFile(filename)

	content, err := os.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	edits := make([]nameStubEdit, 0, len(scan.decls))

	for _, decl := range scan.decls {
		if decl.HasName || scan.suppressed[decl.Offset] {
			continue
		}

		if decl.DeclEnd < 0 || decl.DeclEnd > len(content) {
			return 0, fmt.Errorf(
				"brand %s: %w [%d, %d]",
				decl.TypeName, errOffsetOutOfBounds, decl.DeclEnd, len(content),
			)
		}

		edits = append(edits, nameStubEdit{
			offset: decl.DeclEnd,
			line:   decl.Line,
			text: completeInsertion(
				string(content),
				decl.DeclEnd,
				"\n\n"+nameStub(decl.TypeName),
			),
		})
	}

	if len(edits) == 0 {
		return 0, nil
	}

	mode := os.FileMode(defaultFileMode)
	if info, statErr := os.Stat(filename); statErr == nil {
		mode = info.Mode().Perm()
	}

	//nolint:gosec // G703: filename originates from WalkDir over the caller-provided scan root, the same paths Detect reads.
	if err := os.WriteFile(filename, spliceInsertions(content, edits), mode); err != nil {
		return 0, fmt.Errorf("write file: %w", err)
	}

	return len(edits), nil
}

// nameStubEdit is one pending byte-level insertion.
type nameStubEdit struct {
	offset int
	line   int
	text   string
}

// spliceInsertions applies pure insertions to content in ascending offset
// order; for grouped type declarations (identical offsets) the earlier
// declaration is written first, keeping stub order aligned with declaration
// order.
func spliceInsertions(content []byte, edits []nameStubEdit) []byte {
	slices.SortFunc(edits, func(a, b nameStubEdit) int {
		if c := cmp.Compare(a.offset, b.offset); c != 0 {
			return c
		}

		return cmp.Compare(a.line, b.line)
	})

	out := make([]byte, 0, len(content)+len(edits)*64)
	prev := 0

	for _, edit := range edits {
		out = append(out, content[prev:edit.offset]...)
		out = append(out, edit.text...)

		prev = edit.offset
	}

	out = append(out, content[prev:]...)

	return out
}

// completeInsertion adjusts an insertion's trailing newline so the repaired
// file stays gofmt-stable: a blank line separates the new method from a
// following declaration, the file keeps exactly one trailing newline, and an
// already-blank separator is not doubled.
func completeInsertion(content string, offset int, insertion string) string {
	rest := content[offset:]

	switch {
	case rest == "":
		return insertion + "\n"
	case strings.HasPrefix(rest, "\n\n"):
		return insertion
	case strings.HasPrefix(rest, "\n"):
		return insertion + "\n"
	default:
		return insertion
	}
}
