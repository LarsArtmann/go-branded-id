package linter

import (
	"context"
	"fmt"
	"go/ast"
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

// DetectPath scans path (a single Go file or a directory tree) with the
// default rule set and returns one BD001 finding per brand type that is used
// with id.ID but has no Name() string method, plus one BD002 finding per
// broken suppression directive (see suppress.go). It is byte-identical to
// DetectPathWithRules(path, DefaultRuleSet()). Files that fail to parse are
// skipped: a syntax error is not a brandid-lint finding, and the Go
// toolchain reports it far more precisely.
func DetectPath(path string) ([]gofinding.Finding, error) {
	return DetectPathWithRules(path, DefaultRuleSet())
}

// DetectPathWithRules scans path with an explicitly selected rule set.
// Selecting BD003 adds primitive-ID findings and polices ignore(BD003)
// directives; deselecting BD001 drops brand findings and the BD002s that
// police BD001 directives. Malformed directives stay BD002 findings in every
// mode. See rules.go for the rule-set contract.
func DetectPathWithRules(path string, rules RuleSet) ([]gofinding.Finding, error) {
	var findings []gofinding.Finding

	walkErr := walkGoFiles(path, func(filename string) error {
		findings = append(findings, detectFile(filename, rules)...)

		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	return findings, nil
}

// skipDir reports whether a directory must never be scanned. testdata is
// skipped by ecosystem convention (go vet, staticcheck, and golangci-lint do
// the same): fixture files hold intentionally incomplete brands, and repair
// passes were rewriting them — this linter's own testdata was corrupted
// twice that way before the skip existed. A path passed as the scan root is
// still visited, so the fixtures stay directly testable.
func skipDir(name string) bool {
	switch name {
	case ".git", "vendor", "node_modules", "testdata":
		return true
	default:
		return false
	}
}

// walkGoFiles walks path (a single Go file or a directory tree), skipping
// the directories reported by skipDir and every non-Go file, and calls
// visit for each Go file in deterministic lexicographic order. The first
// error returned by visit aborts the walk.
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
// both see exactly the same brand declarations and suppressions. The parsed
// file and its FileSet ride along for the opt-in BD003 scan, which reuses
// the same parse instead of re-reading the file.
type fileScan struct {
	decls           []BrandDecl
	suppressed      map[int]bool
	bd001Directives []gofinding.Finding      // BD002s policing BD001 directives
	brokenDirectives []gofinding.Finding     // BD002s for malformed directives
	bd003Directives []suppressionDirective   // well-formed ignore(BD003) directives
	file            *ast.File
	fset            *token.FileSet
	filename        string
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
	outcome := applySuppressions(decls, suppressionDirectivesFromFile(f, fset))

	return fileScan{
		decls:            decls,
		suppressed:       outcome.suppressed,
		bd001Directives:  outcome.bd001Findings,
		brokenDirectives: outcome.brokenFindings,
		bd003Directives:  outcome.bd003Directives,
		file:             f,
		fset:             fset,
		filename:         filename,
	}
}

// detectFile returns the findings of one Go file for the selected rules:
// BD001 findings plus their BD002 directive findings when BD001 is active,
// malformed-directive BD002s in every mode, and BD003 findings plus their
// BD002 directive findings when BD003 is active.
func detectFile(filename string, rules RuleSet) []gofinding.Finding {
	scan := scanFile(filename)

	var findings []gofinding.Finding

	if rules.BD001 {
		for _, decl := range scan.decls {
			if decl.HasName || scan.suppressed[decl.Offset] {
				continue
			}

			findings = append(findings, findingForBrand(decl))
		}

		findings = append(findings, scan.bd001Directives...)
	}

	findings = append(findings, scan.brokenDirectives...)

	if rules.BD003 {
		violations := primitiveIDViolations(scan.file, scan.fset, scan.filename)
		kept, directiveFindings := suppressPrimitiveIDs(violations, scan.bd003Directives)

		for _, violation := range kept {
			findings = append(findings, findingForPrimitiveID(violation))
		}

		findings = append(findings, directiveFindings...)
	}

	return findings
}
