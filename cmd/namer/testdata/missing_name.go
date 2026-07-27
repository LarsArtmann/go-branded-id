package testdata

// UserBrand is used with id.ID but has no Name() method.
// The namer tool should flag this as missing.
type UserBrand struct{}

func exampleMissing() {
	var _ = id.ID[UserBrand, string]{}
}
