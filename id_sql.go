package id

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
)

// scanIntegerLikeID scans src into the appropriate integer type for the ID.
// It handles int64, int, and float64 source types by converting them to V.
func scanIntegerLikeID[B any, V comparable](
	id *ID[B, V],
	src any,
	targetTypeName string,
	fromInt64 func(int64) V,
	fromInt func(int) V,
	fromFloat func(float64) V,
) error {
	var zero V

	switch v := src.(type) {
	case int64:
		*id = ID[B, V]{value: fromInt64(v)}

		return nil
	case int:
		*id = ID[B, V]{value: fromInt(v)}

		return nil
	case float64:
		*id = ID[B, V]{value: fromFloat(v)}

		return nil
	default:
		return fmt.Errorf(
			"id: cannot scan %T into %s-based ID (targetType=%T)",
			src,
			targetTypeName,
			zero,
		)
	}
}

// scanIntegerID is a helper that reduces boilerplate for integer ID scanning.
// It wraps scanIntegerLikeID by deriving the int and float64 converters from the int64 converter.
func scanIntegerID[B any, V comparable](
	id *ID[B, V],
	src any,
	typeName string,
	fromInt64 func(int64) V,
) error {
	return scanIntegerLikeID(
		id,
		src,
		typeName,
		fromInt64,
		func(v int) V { return fromInt64(int64(v)) },
		func(v float64) V { return fromInt64(int64(v)) },
	)
}

// Scan implements sql.Scanner for database deserialization.
// Supports string, []byte, int64, int, float64, and nil sources based on the underlying value type V.
//
//nolint:cyclop,funlen // exhaustive type switch over numeric types
func (id *ID[B, V]) Scan(src any) error {
	if id == nil {
		return errors.New("id: scan: receiver is nil")
	}

	if src == nil {
		id.Reset()

		return nil
	}

	switch any(*new(V)).(type) {
	case string:
		switch v := src.(type) {
		case string:
			//nolint:forcetypeassert // outer type switch guarantees V is string
			*id = ID[B, V]{value: any(v).(V)}

			return nil
		case []byte:
			//nolint:forcetypeassert // outer type switch guarantees V is string
			*id = ID[B, V]{value: any(string(v)).(V)}

			return nil
		default:
			return fmt.Errorf("id: cannot scan %T into string-based ID (src=%T)", src, src)
		}

	case int:
		return scanIntegerID(
			id,
			src,
			"int",
			func(v int64) V { return any(int(v)).(V) }, //nolint:forcetypeassert // guaranteed by outer type switch
		)
	case int8:
		return scanIntegerID(
			id,
			src,
			"int8",
			func(v int64) V {
				return any(int8(v)).(V) //nolint:gosec,forcetypeassert // G115: SQL drivers return int64; guaranteed by type switch
			},
		)
	case int16:
		return scanIntegerID(
			id,
			src,
			"int16",
			func(v int64) V {
				return any(int16(v)).(V) //nolint:gosec,forcetypeassert // G115: SQL drivers return int64; guaranteed by type switch
			},
		)
	case int32:
		return scanIntegerID(
			id,
			src,
			"int32",
			func(v int64) V {
				return any(int32(v)).(V) //nolint:gosec,forcetypeassert // G115: SQL drivers return int64; guaranteed by type switch
			},
		)
	case int64:
		return scanIntegerLikeID(
			id,
			src,
			"int64",
			func(v int64) V { return any(v).(V) }, //nolint:forcetypeassert // guaranteed by type switch
			func(v int) V { return any(int64(v)).(V) },     //nolint:forcetypeassert // guaranteed by type switch
			func(v float64) V { return any(int64(v)).(V) }, //nolint:forcetypeassert // guaranteed by type switch
		)
	case uint:
		return scanIntegerID(
			id,
			src,
			"uint",
			func(v int64) V {
				return any(uint(v)).(V) //nolint:gosec,forcetypeassert // G115: SQL drivers return int64; guaranteed by type switch
			},
		)
	case uint8:
		return scanIntegerID(
			id,
			src,
			"uint8",
			func(v int64) V {
				return any(uint8(v)).(V) //nolint:gosec,forcetypeassert // G115: SQL drivers return int64; guaranteed by type switch
			},
		)
	case uint16:
		return scanIntegerID(
			id,
			src,
			"uint16",
			func(v int64) V {
				return any(uint16(v)).(V) //nolint:gosec,forcetypeassert // G115: SQL drivers return int64; guaranteed by type switch
			},
		)
	case uint32:
		return scanIntegerID(
			id,
			src,
			"uint32",
			func(v int64) V {
				return any(uint32(v)).(V) //nolint:gosec,forcetypeassert // G115: SQL drivers return int64; guaranteed by type switch
			},
		)
	case uint64:
		return scanIntegerID(
			id,
			src,
			"uint64",
			func(v int64) V {
				return any(uint64(v)).(V) //nolint:gosec,forcetypeassert // G115: SQL drivers return int64; guaranteed by type switch
			},
		)

	default:
		// Check if V implements TextUnmarshaler for text-based scanning.
		var zero V
		if unmarshaler, ok := any(&zero).(encoding.TextUnmarshaler); ok {
			var text []byte

			switch v := src.(type) {
			case string:
				text = []byte(v)
			case []byte:
				text = v
			default:
				return fmt.Errorf(
					"id: cannot scan %T into text-unmarshalable ID (targetType=%T)",
					src,
					zero,
				)
			}

			err := unmarshaler.UnmarshalText(text)
			if err != nil {
				return fmt.Errorf("id: cannot scan text into %T: %w", zero, err)
			}

			*id = ID[B, V]{value: zero}

			return nil
		}

		return fmt.Errorf("id: unsupported target type %T for SQL scanning (src=%T)", *new(V), src)
	}
}

// Value implements driver.Valuer for database serialization.
// Returns nil for zero values, otherwise the underlying value.
//
//nolint:cyclop // exhaustive type switch over numeric types
func (id ID[B, V]) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, nil //nolint:nilnil // zero value maps to SQL NULL
	}

	switch v := any(id.value).(type) {
	case string:
		return v, nil
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil //nolint:gosec // G115: uint to int64 for SQL value
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil //nolint:gosec // G115: uint64 to int64 for SQL value
	default:
		// Check if V implements TextMarshaler for text-based value conversion.
		if marshaler, ok := any(id.value).(encoding.TextMarshaler); ok {
			text, err := marshaler.MarshalText()
			if err != nil {
				return nil, fmt.Errorf(
					"id: cannot marshal %T to text for SQL value: %w",
					id.value,
					err,
				)
			}

			return string(text), nil
		}

		return nil, fmt.Errorf("id: unsupported type %T for SQL value", id.value)
	}
}

// Compile-time interface assertions.
var (
	_ fmt.Stringer     = ID[struct{}, string]{value: ""}
	_ fmt.GoStringer   = ID[struct{}, string]{value: ""}
	_ fmt.Formatter    = ID[struct{}, string]{value: ""}
	_ json.Marshaler   = ID[struct{}, string]{value: ""}
	_ json.Unmarshaler = (*ID[struct{}, string])(nil)
	_ json.Marshaler   = ID[struct{}, int64]{value: 0}
	_ json.Unmarshaler = (*ID[struct{}, int64])(nil)
	_ json.Marshaler   = ID[struct{}, int32]{value: 0}
	_ json.Unmarshaler = (*ID[struct{}, int32])(nil)
	_ json.Marshaler   = ID[struct{}, uint64]{value: 0}
	_ json.Unmarshaler = (*ID[struct{}, uint64])(nil)
	_ sql.Scanner      = (*ID[struct{}, string])(nil)
	_ sql.Scanner      = (*ID[struct{}, int64])(nil)
	_ sql.Scanner      = (*ID[struct{}, int32])(nil)
	_ sql.Scanner      = (*ID[struct{}, uint64])(nil)
	_ driver.Valuer    = ID[struct{}, string]{value: ""}
	_ driver.Valuer    = ID[struct{}, int64]{value: 0}
	_ driver.Valuer    = ID[struct{}, int32]{value: 0}
	_ driver.Valuer    = ID[struct{}, uint64]{value: 0}
)
