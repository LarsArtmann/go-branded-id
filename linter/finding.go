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

// findingForPrimitiveID converts one primitive-ID violation into a BD003
// finding. The fix is not machine-applicable (it rewrites the declaration and
// every call site), so the strategy is Suggest and the actionable payload
// travels as metadata: suggested type and brand names, the import path, and
// a complete type definition to paste.
func findingForPrimitiveID(v primitiveIDViolation) gofinding.Finding {
	suggestedType := suggestIDTypeName(v.Name, v.ParentName)
	suggestedBrand := suggestIDBrand(v.Name, v.ParentName)

	builder := gofinding.NewBuilder(
		RuleIDBD003,
		ToolName,
		createPrimitiveIDMessage(v.Name, v.CurrentType, suggestedType, v.ParentName, v.Kind),
		primitiveIDRank(v.Name, v.Kind).severity(),
		gofinding.Position{
			File:   gofinding.FilePath(v.File),
			Line:   v.Line,
			Column: v.Column,
			Offset: v.Offset,
		},
	).WithCategory(gofinding.CategoryTypeSafety).
		WithConfidence(gofinding.ConfidenceHigh).
		WithSuggestion("Replace with strong ID type " + suggestedType).
		WithSnippet(v.Name + " " + v.CurrentType).
		WithFixStrategy(gofinding.FixStrategySuggest).
		WithBeforeCode(buildCodeSnippet(v.Name, v.CurrentType, v.Kind, v.ParentName)).
		WithAfterCode(buildCodeSnippet(v.Name, suggestedType, v.Kind, v.ParentName)).
		WithMetadata(map[string]string{
			"name":              v.Name,
			"currentType":       v.CurrentType,
			"kindDisplay":       v.Kind.display(),
			"suggestedTypeName": suggestedType,
			"suggestedBrand":    suggestedBrand,
			"importPath":        primitiveIDImportPath,
			"typeDefinition": buildTypeDefinition(
				suggestedType, suggestedBrand, v.CurrentType,
			),
		})

	if v.DeclEndLine > v.Line {
		builder = builder.WithRange(gofinding.Range{
			Start: gofinding.Position{File: gofinding.FilePath(v.File), Line: v.Line},
			End:   gofinding.Position{File: gofinding.FilePath(v.File), Line: v.DeclEndLine},
		})
	}

	return builder.MustBuild()
}
