//go:build !goexperiment.jsonv2

package id

import (
	"encoding/json/v2"
	"fmt"
)

// MarshalJSON implements json.Marshaler for proper null handling.
// String-based IDs serialize as JSON strings, numeric IDs as JSON numbers.
// Zero values serialize to JSON null.
func (id ID[B, V]) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return []byte("null"), nil
	}

	b, err := json.Marshal(id.value)
	if err != nil {
		return nil, fmt.Errorf("id: marshal JSON: %w", err)
	}

	return b, nil
}

// UnmarshalJSON implements json.Unmarshaler for JSON deserialization.
// Supports null, strings, and numeric values based on the underlying type V.
func (id *ID[B, V]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		id.Reset()

		return nil
	}

	var zero V

	err := json.Unmarshal(data, &zero)
	if err != nil {
		return fmt.Errorf("id: cannot unmarshal %s into %T: %w", string(data), zero, err)
	}

	*id = ID[B, V]{value: zero}

	return nil
}

// Compile-time assertions that ID implements the json interfaces.
var (
	_ json.Marshaler   = ID[struct{}, string]{value: ""}
	_ json.Unmarshaler = (*ID[struct{}, string])(nil)
	_ json.Marshaler   = ID[struct{}, int64]{value: 0}
	_ json.Unmarshaler = (*ID[struct{}, int64])(nil)
	_ json.Marshaler   = ID[struct{}, int32]{value: 0}
	_ json.Unmarshaler = (*ID[struct{}, int32])(nil)
	_ json.Marshaler   = ID[struct{}, uint64]{value: 0}
	_ json.Unmarshaler = (*ID[struct{}, uint64])(nil)
)
