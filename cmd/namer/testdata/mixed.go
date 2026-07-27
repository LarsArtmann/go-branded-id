package testdata

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
	var _ = id.ID[ProductBrand, string]{}
	var _ = id.ID[TenantBrand, string]{}
	var _ = id.ID[SessionBrand, int]{}
	_ = NotABrand{}
}
