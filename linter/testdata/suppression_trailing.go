package testdata

import id "github.com/larsartmann/go-branded-id"

type TrailingBrand struct{} // brandid-lint:ignore(BD001) String() output is a routing key

func exampleTrailing() {
	_ = id.ID[TrailingBrand, string]{}
}
