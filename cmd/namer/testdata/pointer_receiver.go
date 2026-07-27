package testdata

import id "github.com/larsartmann/go-branded-id"

// PointerBrand has a Name() method with a pointer receiver.
// The namer tool should detect this and mark HasName=true.
type PointerBrand struct{}

func (*PointerBrand) Name() string { return "Pointer" }

func examplePointerReceiver() {
	_ = id.ID[PointerBrand, string]{}
}
