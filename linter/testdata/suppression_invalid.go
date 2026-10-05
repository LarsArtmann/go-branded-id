package testdata

import id "github.com/larsartmann/go-branded-id"

// WaitBrand carries a directive without a reason, so it must stay flagged.
// brandid-lint:ignore(BD001)
type WaitBrand struct{}

func exampleInvalidDirective() {
	_ = id.ID[WaitBrand, string]{}
}
