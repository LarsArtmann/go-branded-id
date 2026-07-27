package testdata

import id "github.com/larsartmann/go-branded-id"

// ProductBrand has Name() — should not be flagged.
type ProductBrand struct{}

func (ProductBrand) Name() string { return "Product" }

// TenantBrand has no Name() — should be flagged.
type TenantBrand struct{}

// SessionBrand has no Name() — should be flagged.
type SessionBrand struct{}

// NotABrand is an empty struct not used with id.ID — should be ignored.
type NotABrand struct{}

func exampleMixed() {
	_ = id.ID[ProductBrand, string]{}
	_ = id.ID[TenantBrand, string]{}
	_ = id.ID[SessionBrand, int]{}
	_ = NotABrand{}
}
