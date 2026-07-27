package testdata

// ConfigBrand is an empty struct but NOT used with id.ID.
// The namer tool should ignore it entirely.
type ConfigBrand struct{}

func exampleNoID() {
	_ = ConfigBrand{}
}
