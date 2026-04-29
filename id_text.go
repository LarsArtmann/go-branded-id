package id

import (
	"encoding"
	"errors"
	"fmt"
	"strconv"
)

const (
	parseBaseDecimal = 10
	parseBitSize64   = 64
)

// MarshalText implements encoding.TextMarshaler for text-based encoding (e.g., XML, TOML).
// For JSON, prefer the json.Marshaler implementation which handles null properly.
func (id ID[B, V]) MarshalText() ([]byte, error) {
	if id.IsZero() {
		return nil, nil
	}

	return []byte(id.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for text-based decoding (e.g., XML, TOML).
// Note: This only works for string-based IDs. For other types, use JSON.
func (id *ID[B, V]) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		id.Reset()

		return nil
	}

	var zero V
	switch any(zero).(type) {
	case string:
		v, ok := any(string(data)).(V)
		if !ok {
			return errors.New("id: internal error: type assertion failed for string")
		}

		*id = ID[B, V]{value: v}

		return nil
	case int:
		n, err := strconv.Atoi(string(data))
		if err != nil {
			return fmt.Errorf("id: cannot parse %q as int: %w", data, err)
		}

		v, ok := any(n).(V)
		if !ok {
			return errors.New("id: internal error: type assertion failed for int")
		}

		*id = ID[B, V]{value: v}

		return nil
	case int64:
		return parseIntegerID(id, data, func(s string, base, bits int) (signedInt, error) {
			n, err := strconv.ParseInt(s, base, bits)

			return signedInt(n), err
		}, "int64")
	case uint64:
		return parseIntegerID(id, data, func(s string, base, bits int) (unsignedInt, error) {
			n, err := strconv.ParseUint(s, base, bits)

			return unsignedInt(n), err
		}, "uint64")
	default:
		var zero V
		if unmarshaler, ok := any(&zero).(encoding.TextUnmarshaler); ok {
			err := unmarshaler.UnmarshalText(data)
			if err != nil {
				return fmt.Errorf("id: cannot unmarshal text into %T: %w", zero, err)
			}

			*id = ID[B, V]{value: zero}

			return nil
		}

		return fmt.Errorf(
			"id: cannot unmarshal text into %T (only string and numeric IDs supported, got data=%q)",
			zero,
			string(data),
		)
	}
}

type (
	signedInt   int64
	unsignedInt uint64
)

func parseIntegerID[B any, V comparable, I signedInt | unsignedInt](
	id *ID[B, V],
	data []byte,
	parse func(string, int, int) (I, error),
	typeName string,
) error {
	n, err := parse(string(data), parseBaseDecimal, parseBitSize64)
	if err != nil {
		return fmt.Errorf("id: cannot parse %q as %s: %w", data, typeName, err)
	}

	var v V

	switch any(n).(type) {
	case signedInt:
		v = any(int64(n)).(V)
	case unsignedInt:
		v = any(uint64(n)).(V)
	}

	*id = ID[B, V]{value: v}

	return nil
}

// Compile-time interface assertions for text marshaling.
var (
	_ encoding.TextMarshaler   = ID[struct{}, string]{value: ""}
	_ encoding.TextUnmarshaler = (*ID[struct{}, string])(nil)
	_ encoding.TextMarshaler   = ID[struct{}, int64]{value: 0}
	_ encoding.TextUnmarshaler = (*ID[struct{}, int64])(nil)
)
