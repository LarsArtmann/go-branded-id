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
// string method. Files that fail to parse are skipped: a syntax error is not
// a BD001 finding, and the Go toolchain reports it far more precisely.
func DetectPath(path string) ([]gofinding.Finding, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}

	if !info.IsDir() {
		return detectFile(path)
	}

	var findings []gofinding.Finding

	walkErr := filepath.WalkDir(path, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk %s: %w", p, err)
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

		fileFindings, fileErr := detectFile(p)
		if fileErr != nil {
			return fileErr
		}

		findings = append(findings, fileFindings...)

		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk %s: %w", path, walkErr)
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

// detectFile scans one Go file. Files that fail to parse yield no findings
// and no error (see DetectPath).
func detectFile(filename string) ([]gofinding.Finding, error) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, nil //nolint:nilerr,nilnil // unparseable files are skipped by design
	}

	var findings []gofinding.Finding

	for _, decl := range brandDeclsFromFile(f, fset) {
		if decl.HasName {
			continue
		}

		findings = append(findings, findingForBrand(decl))
	}

	return findings, nil
}
