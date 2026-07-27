package id

import (
	"encoding/json"
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
