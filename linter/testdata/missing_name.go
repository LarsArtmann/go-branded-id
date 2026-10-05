package testdata

import id "github.com/larsartmann/go-branded-id"

// UserBrand is used with id.ID but has no Name() method.
// The brandid-lint should flag this as missing.
type UserBrand struct{}

func exampleMissing() {
	_ = id.ID[UserBrand, string]{}
}
