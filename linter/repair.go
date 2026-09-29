package linter

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	gofinding "github.com/larsartmann/go-finding"
)

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
// unsuppressed brand declaration flagged by BD001, one stub per finding. It
// returns the number of applied insertions. Suppressed brands are never
// touched — a documented exception must stay an exception — and repair is
// idempotent: brands that already have Name() produce no finding, so a
// second run inserts nothing.
func RepairPath(path string) (int, error) {
	findings, err := DetectPath(path)
	if err != nil {
		return 0, fmt.Errorf("detect %s: %w", path, err)
	}

	byFile := make(map[gofinding.FilePath][]gofinding.Finding)

	for _, f := range findings {
		if f.Rule != RuleIDBD001 || len(f.Edits) == 0 {
			continue
		}

		byFile[f.Position.File] = append(byFile[f.Position.File], f)
	}

	total := 0

	for file, fileFindings := range byFile {
		applied, err := applyNameStubs(file, fileFindings)
		if err != nil {
			return total, fmt.Errorf("repair %s: %w", file, err)
		}

		total += applied
	}

	return total, nil
}

// nameStubEdit is one pending byte-level insertion.
type nameStubEdit struct {
	offset int
	line   int
	text   string
}

// applyNameStubs inserts one stub per finding into a single file. Insertions
// are applied in memory, back to front, so no edit shifts another; for
// grouped type declarations (identical offsets) the later declaration is
// written first, keeping stub order aligned with declaration order.
func applyNameStubs(file gofinding.FilePath, findings []gofinding.Finding) (int, error) {
	content, err := os.ReadFile(string(file))
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	edits := make([]nameStubEdit, 0, len(findings))

	for _, f := range findings {
		offset := f.Edits[0].Start.Offset
		if offset < 0 || offset > len(content) {
			return 0, fmt.Errorf(
				"finding %s: insertion offset %d outside file bounds [0, %d]",
				f.ID, offset, len(content),
			)
		}

		edits = append(edits, nameStubEdit{
			offset: offset,
			line:   f.Position.Line,
			text:   completeInsertion(string(content), offset, f.Edits[0].NewText),
		})
	}

	slices.SortFunc(edits, func(a, b nameStubEdit) int {
		if c := cmp.Compare(b.offset, a.offset); c != 0 {
			return c
		}

		return cmp.Compare(b.line, a.line)
	})

	var out strings.Builder

	cursor := len(content)

	for _, edit := range edits {
		out.Write(content[edit.offset:cursor])
		out.WriteString(edit.text)

		cursor = edit.offset
	}

	out.Write(content[0:cursor])

	mode := os.FileMode(0o644)
	if info, statErr := os.Stat(string(file)); statErr == nil {
		mode = info.Mode().Perm()
	}

	if err := os.WriteFile(string(file), []byte(out.String()), mode); err != nil {
		return 0, fmt.Errorf("write file: %w", err)
	}

	return len(edits), nil
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
