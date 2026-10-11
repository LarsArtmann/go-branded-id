package testdata

import id "github.com/larsartmann/go-branded-id"

// brandNameConst is the package-level constant the brand's Name() returns.
const brandNameConst = "Widget"

// WidgetBrand has a Name() method that returns a constant instead of a
// string literal, so the brandid-lint should NOT flag it (issue #10).
type WidgetBrand struct{}

func (WidgetBrand) Name() string { return brandNameConst }

func exampleConstName() {
	_ = id.ID[WidgetBrand, string]{}
}
