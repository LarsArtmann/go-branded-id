package linter

import (
	"fmt"

	gofinding "github.com/larsartmann/go-finding"
)

// nameStub returns the Name() method text for a brand type, without leading
// newlines. It is the single source of the suggestion, the display codes, and
// the repair insertion text.
func nameStub(typeName string) string {
	return fmt.Sprintf("func (%s) Name() string { return %q }", typeName, suggestName(typeName))
}

// findingForBrand converts an unnamed brand declaration into a BD001 finding
// positioned at the type declaration, where the fix (adding Name) lands. The
// fix is machine-applicable: BeforeCode/AfterCode carry the stub insertion
// that Repair applies, keyed to the declaration this finding anchors.
func findingForBrand(b BrandDecl) gofinding.Finding {
	stub := nameStub(b.TypeName)
	before := fmt.Sprintf("type %s struct{}", b.TypeName)

	return gofinding.NewBuilder(
		RuleIDBD001,
		ToolName,
		fmt.Sprintf(
			"brand type %s is used with id.ID but has no Name() string method",
			b.TypeName,
		),
		gofinding.SeverityWarning,
		gofinding.Position{
			File:   gofinding.FilePath(b.File),
			Line:   b.Line,
			Column: b.Column,
			Offset: b.Offset,
		},
	).WithCategory(gofinding.CategoryNaming).
		WithConfidence(gofinding.ConfidenceHigh).
		WithSuggestion(stub).
		WithSnippet(before).
		WithFixStrategy(gofinding.FixStrategyDirect).
		WithBeforeCode(before).
		WithAfterCode(before + "\n\n" + stub).
		MustBuild()
}

// findingForDirective turns a broken suppression directive into a BD002
// finding positioned at the directive comment itself, so the fix (correcting
// or removing the comment) lands where the user is looking.
func findingForDirective(directive suppressionDirective, problem string) gofinding.Finding {
	return gofinding.NewBuilder(
		RuleIDBD002,
		ToolName,
		fmt.Sprintf("//%s directive %s", directivePrefix, problem),
		gofinding.SeverityWarning,
		gofinding.Position{
			File:   gofinding.FilePath(directive.File),
			Line:   directive.Line,
			Column: directive.Column,
			Offset: directive.Offset,
		},
	).WithCategory(gofinding.CategoryConfiguration).
		WithConfidence(gofinding.ConfidenceHigh).
		WithSnippet(directive.Comment).
		MustBuild()
}
