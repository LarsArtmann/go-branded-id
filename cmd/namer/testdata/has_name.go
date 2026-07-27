package testdata

// OrderBrand has a Name() method, so the namer tool should NOT flag it.
type OrderBrand struct{}

func (OrderBrand) Name() string { return "Order" }

func exampleHasName() {
	var _ = id.ID[OrderBrand, int64]{}
}
