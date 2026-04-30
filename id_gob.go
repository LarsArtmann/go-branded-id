package id

import (
	"encoding/gob"
)

// GobEncode implements gob.GobEncoder for Go-specific encoding.
func (id ID[B, V]) GobEncode() ([]byte, error) {
	return id.MarshalBinary()
}

// GobDecode implements gob.GobDecoder for Go-specific decoding.
func (id *ID[B, V]) GobDecode(data []byte) error {
	return id.UnmarshalBinary(data)
}

// Compile-time interface assertions for Gob encoding.
var (
	_ gob.GobEncoder = ID[struct{}, string]{value: ""}
	_ gob.GobDecoder = (*ID[struct{}, string])(nil)
	_ gob.GobEncoder = ID[struct{}, int64]{value: 0}
	_ gob.GobDecoder = (*ID[struct{}, int64])(nil)
)
