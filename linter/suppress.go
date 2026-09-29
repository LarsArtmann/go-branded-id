package linter

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	gofinding "github.com/larsartmann/go-finding"
)

// directivePrefix namespaces the in-source suppression directive by the
// stable tool name (never the module path), so a mistyped directive cannot
// masquerade as a foreign tool's comment. Valid form, exactly one rule and a
// mandatory reason:
//
//	//brandid-lint:ignore(BD001) <reason>
//
// The directive may sit on the brand declaration line (trailing) or as the
// last line of the comment group directly above it. Only line comments carry
// directives; block comments are ignored.
const directivePrefix = "brandid-lint:ignore"

// suppressionDirective is one parsed directive comment. Problem is empty
// only for well-formed directives: correct grammar, the suppressible rule
// BD001, and a non-empty reason.
type suppressionDirective struct {
	Comment     string // Raw comment text (including "//"), reused as the finding snippet
	Rule        string
	Reason      string
	File        string
	Line        int  // 1-based line of the comment
	Column      int  // 1-based column of the comment
	Offset      int  // 0-based byte offset of the comment
	LastInGroup bool // comment is the final line of its comment group
	Problem     string
}

// wellFormed reports whether the directive can suppress a finding.
func (d suppressionDirective) wellFormed() bool {
	return d.Problem == ""
}

// suppressionDirectivesFromFile returns every brandid-lint directive in f
// with its source position, in source order.
func suppressionDirectivesFromFile(f *ast.File, fset *token.FileSet) []suppressionDirective {
	lastInGroup := make(map[token.Pos]bool)

	for _, group := range f.Comments {
		last := group.List[len(group.List)-1]
		lastInGroup[last.Slash] = true
	}

	var directives []suppressionDirective

	for _, group := range f.Comments {
		for _, comment := range group.List {
			parsed, ok := parseSuppressionDirective(comment)
			if !ok {
				continue
			}

			parsed.LastInGroup = lastInGroup[comment.Slash]

			pos := fset.Position(comment.Slash)
			parsed.File = pos.Filename
			parsed.Line = pos.Line
			parsed.Column = pos.Column
			parsed.Offset = pos.Offset

			directives = append(directives, parsed)
		}
	}

	return directives
}

// parseSuppressionDirective parses one comment. ok is false when the comment
// is not a brandid-lint directive at all (ordinary comment, not an error).
// A comment that starts with the directive prefix but violates the grammar
// returns ok=true with Problem set, so a no-op directive stays loud instead
// of silently suppressing nothing.
func parseSuppressionDirective(comment *ast.Comment) (suppressionDirective, bool) {
	if !strings.HasPrefix(comment.Text, "//") {
		//nolint:exhaustruct // zero value: not a directive, caller ignores it
		return suppressionDirective{}, false
	}

	body := strings.TrimLeft(strings.TrimPrefix(comment.Text, "//"), " \t")
	if !strings.HasPrefix(body, directivePrefix) {
		//nolint:exhaustruct // zero value: not a directive, caller ignores it
		return suppressionDirective{}, false
	}

	//nolint:exhaustruct // position fields are filled by suppressionDirectivesFromFile
	directive := suppressionDirective{Comment: comment.Text}

	rest := body[len(directivePrefix):]

	openIdx := strings.Index(rest, "(")
	closeIdx := strings.Index(rest, ")")

	switch {
	case openIdx != 0 || closeIdx < 0:
		directive.Problem = fmt.Sprintf(
			"must be written as //%s(%s) <reason>", directivePrefix, RuleIDBD001,
		)
	default:
		directive.Rule = rest[1:closeIdx]
		directive.Reason = strings.TrimSpace(rest[closeIdx+1:])

		switch {
		case directive.Rule == "":
			directive.Problem = "missing rule ID"
		case directive.Rule != string(RuleIDBD001):
			directive.Problem = fmt.Sprintf(
				"unknown rule %q: only %s is suppressible", directive.Rule, RuleIDBD001,
			)
		case directive.Reason == "":
			directive.Problem = "missing reason: document why this brand deliberately skips Name()"
		}
	}

	return directive, true
}

// applySuppressions pairs directives with brand declarations and decides
// which BD001 findings are suppressed. It returns the byte offsets of
// suppressed brand declarations plus one BD002 finding per broken directive:
// malformed syntax, a directive on a brand that already has Name() (stale),
// a duplicate suppression, or one that annotates no brand declaration at
// all. Every broken directive stays loud — a silent no-op directive is worse
// than none.
func applySuppressions(
	decls []BrandDecl,
	directives []suppressionDirective,
) (map[int]bool, []gofinding.Finding) {
	declByLine := make(map[int]int, len(decls))

	for i, decl := range decls {
		if _, taken := declByLine[decl.Line]; !taken {
			declByLine[decl.Line] = i
		}
	}

	dirsByDecl := make(map[int][]suppressionDirective)
	var unplaced []suppressionDirective

	for _, directive := range directives {
		switch idx, ok := declByLine[directive.Line]; {
		case ok: // trailing comment on the declaration line
			dirsByDecl[idx] = append(dirsByDecl[idx], directive)
		default:
			aboveIdx, aboveOK := declByLine[directive.Line+1]
			if directive.LastInGroup && aboveOK {
				// last line of the comment group directly above the declaration
				dirsByDecl[aboveIdx] = append(dirsByDecl[aboveIdx], directive)

				continue
			}

			unplaced = append(unplaced, directive)
		}
	}

	suppressed := make(map[int]bool)
	var findings []gofinding.Finding

	for i, decl := range decls {
		for _, directive := range dirsByDecl[i] {
			if decl.HasName {
				findings = append(findings, findingForDirective(
					directive,
					fmt.Sprintf("is stale: brand %s already has a Name() method", decl.TypeName),
				))

				continue
			}

			switch {
			case !directive.wellFormed():
				findings = append(findings, findingForDirective(directive, directive.Problem))
			case suppressed[decl.Offset]:
				findings = append(findings, findingForDirective(
					directive,
					fmt.Sprintf(
						"is redundant: brand %s is already suppressed by another directive",
						decl.TypeName,
					),
				))
			default:
				suppressed[decl.Offset] = true
			}
		}
	}

	for _, directive := range unplaced {
		problem := "suppresses nothing: place it on the brand declaration line " +
			"or as the last line of the comment group directly above"
		if directive.Problem != "" {
			problem = directive.Problem
		}

		findings = append(findings, findingForDirective(directive, problem))
	}

	return suppressed, findings
}
