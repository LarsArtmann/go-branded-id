package testdata

import id "github.com/larsartmann/go-branded-id"

// EventBrand is a marker brand whose String() output is a data key.
// brandid-lint:ignore(BD001) String() is used directly as the stream name
type EventBrand struct{}

func exampleSuppressed() {
	_ = id.ID[EventBrand, string]{}
}
