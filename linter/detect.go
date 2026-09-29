package linter

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"

	gofinding "github.com/larsartmann/go-finding"
)

// Detect runs the BD001 rule on the working directory carried by ctx (see
// gofinding.WorkingDirFromContext), falling back to the process working
// directory. This is the entry point the BuildFlow provider and any
// go-finding pipeline consumer call.
func Detect(ctx context.Context) ([]gofinding.Finding, error) {
	root := gofinding.WorkingDirFromContext(ctx)
	if root == "" {
		root = "."
	}

	return DetectPath(root)
}

// DetectPath scans path (a single Go file or a directory tree) and returns
// one BD001 finding per brand type that is used with id.ID but has no Name()
// string method, plus one BD002 finding per broken suppression directive
// (see suppress.go). Files that fail to parse are skipped: a syntax error is
// not a brandid-lint finding, and the Go toolchain reports it far more
// precisely.
func DetectPath(path string) ([]gofinding.Finding, error) {
	var findings []gofinding.Finding

	walkErr := walkGoFiles(path, func(filename string) error {
		findings = append(findings, detectFile(filename)...)

		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	return findings, nil
}

// skipDir reports whether a directory must never be scanned.
func skipDir(name string) bool {
	switch name {
	case ".git", "vendor", "node_modules":
		return true
	default:
		return false
	}
}

// walkGoFiles walks path (a single Go file or a directory tree), skipping
// the directories .git, vendor, and node_modules and every non-Go file, and
// calls visit for each Go file in deterministic lexicographic order. The
// first error returned by visit aborts the walk.
func walkGoFiles(path string, visit func(filename string) error) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	if !info.IsDir() {
		if filepath.Ext(path) == ".go" {
			if err := visit(path); err != nil {
				return fmt.Errorf("visit %s: %w", path, err)
			}
		}

		return nil
	}

	walkErr := filepath.WalkDir(path, func(p string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk %s: %w", p, walkErr)
		}

		if entry.IsDir() {
			if skipDir(entry.Name()) && p != path {
				return filepath.SkipDir
			}

			return nil
		}

		if filepath.Ext(p) != ".go" {
			return nil
		}

		return visit(p)
	})
	if walkErr != nil {
		return fmt.Errorf("walk %s: %w", path, walkErr)
	}

	return nil
}

// fileScan is the per-file scan result shared by detection and repair, so
// both see exactly the same brand declarations and suppressions.
type fileScan struct {
	decls      []BrandDecl
	suppressed map[int]bool
	directives []gofinding.Finding
}

// scanFile parses one Go file and pairs its brand declarations with their
// suppression directives. Unparseable files yield an empty scan: a syntax
// error is not a brandid-lint finding (see DetectPath).
func scanFile(filename string) fileScan {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		//nolint:exhaustruct_v5 // zero value: an unparseable file has no scan result
		return fileScan{}
	}

	decls := brandDeclsFromFile(f, fset)
	suppressed, directiveFindings := applySuppressions(
		decls,
		suppressionDirectivesFromFile(f, fset),
	)

	return fileScan{decls: decls, suppressed: suppressed, directives: directiveFindings}
}

// detectFile returns the findings of one Go file: one BD001 per unsuppressed
// unnamed brand declaration, plus the file's BD002 directive findings.
func detectFile(filename string) []gofinding.Finding {
	scan := scanFile(filename)

	var findings []gofinding.Finding

	for _, decl := range scan.decls {
		if decl.HasName || scan.suppressed[decl.Offset] {
			continue
		}

		findings = append(findings, findingForBrand(decl))
	}

	return append(findings, scan.directives...)
}
