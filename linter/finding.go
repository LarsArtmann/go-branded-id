package linter

import (
	"fmt"

	gofinding "github.com/larsartmann/go-finding"
)

// findingForBrand converts an unnamed brand declaration into a BD001 finding
// positioned at the type declaration, where the fix (adding Name) lands.
func findingForBrand(b BrandDecl) gofinding.Finding {
	suggestion := fmt.Sprintf(
		"func (%s) Name() string { return %q }",
		b.TypeName, suggestName(b.TypeName),
	)

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
		WithSuggestion(suggestion).
		WithSnippet(fmt.Sprintf("type %s struct{}", b.TypeName)).
		MustBuild()
}
