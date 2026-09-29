package testdata

import id "github.com/larsartmann/go-branded-id"

// DoneBrand has a Name() method, so the directive below suppresses nothing.
// brandid-lint:ignore(BD001) added later, directive never removed
type DoneBrand struct{}

func (DoneBrand) Name() string { return "Done" }

func exampleStaleDirective() {
	_ = id.ID[DoneBrand, string]{}
}
