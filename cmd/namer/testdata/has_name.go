package testdata

import id "github.com/larsartmann/go-branded-id"

// OrderBrand has a Name() method, so the namer tool should NOT flag it.
type OrderBrand struct{}

func (OrderBrand) Name() string { return "Order" }

func exampleHasName() {
	_ = id.ID[OrderBrand, int64]{}
}
