// Package id provides branded, strongly-typed identifiers that prevent mixing
// different entity IDs at compile time using phantom types.
//
// # Type Safety
//
// The ID type uses phantom typing (brand types) to create distinct identifier
// types that cannot be accidentally mixed:
//
//	type UserBrand struct{}
//	type OrderBrand struct{}
//	type UserID = ID[UserBrand, string]
//	type OrderID = ID[OrderBrand, int64]
//
//	func ProcessUser(id UserID) { ... }
//	func ProcessOrder(id OrderID) { ... }
//
//	userID := NewID[UserBrand]("user-123")
//	orderID := NewID[OrderBrand, int64](456)
//
//	ProcessUser(userID)   // OK
//	ProcessUser(orderID)  // Compile error: type mismatch
//
// # Serialization
//
// ID supports multiple serialization formats:
//   - JSON: string-based IDs serialize as strings, numeric IDs as numbers
//   - Text (XML/TOML): string-based IDs only
//   - SQL: string, int64, int32, uint64 types supported
//   - Binary: efficient binary representation
//   - Gob: Go-specific encoding
package id

import (
	"cmp"
	"encoding"
	"errors"
	"fmt"
	"strconv"
)

// ErrNotOrdered is returned when Compare is called on an ID with a non-ordered value type.
var ErrNotOrdered = errors.New("id: Compare requires an ordered type (int, uint, or string)")

// ID is a branded, strongly-typed identifier that prevents mixing different entity IDs.
// B is the brand (phantom type for distinctness), V is the value type.
//
// The zero value represents an unset/empty ID and serializes to null in JSON.
type ID[B any, V comparable] struct{ value V }

// NewID creates a new branded identifier from the given value.
func NewID[B any, V comparable](v V) ID[B, V] { return ID[B, V]{value: v} }

// Get returns the underlying value.
func (id ID[B, V]) Get() V { return id.value }

// IsZero returns true if the ID has its zero value.
func (id ID[B, V]) IsZero() bool {
	var zero V

	return id.value == zero
}

// Reset sets the ID to its zero value.
func (id *ID[B, V]) Reset() {
	var zero V

	*id = ID[B, V]{value: zero}
}

// Equal returns true if this ID equals the other ID.
// Both IDs must have the same brand and value type.
func (id ID[B, V]) Equal(other ID[B, V]) bool {
	return id.value == other.value
}

// Compare returns -1 if id < other, 0 if equal, 1 if id > other.
// Returns ErrNotOrdered if V is not an ordered type.
//
//nolint:cyclop // exhaustive type switch over ordered types
func (id ID[B, V]) Compare(other ID[B, V]) (int, error) {
	switch a := any(id.value).(type) {
	case int:
		//nolint:forcetypeassert // V is same type for both id and other
		return cmp.Compare(a, any(other.value).(int)), nil
	case int8:
		//nolint:forcetypeassert // V is same type for both id and other
		return cmp.Compare(a, any(other.value).(int8)), nil
	case int16:
		//nolint:forcetypeassert // V is same type for both id and other
		return cmp.Compare(a, any(other.value).(int16)), nil
	case int32:
		//nolint:forcetypeassert // V is same type for both id and other
		return cmp.Compare(a, any(other.value).(int32)), nil
	case int64:
		//nolint:forcetypeassert // V is same type for both id and other
		return cmp.Compare(a, any(other.value).(int64)), nil
	case uint:
		//nolint:forcetypeassert // V is same type for both id and other
		return cmp.Compare(a, any(other.value).(uint)), nil
	case uint8:
		//nolint:forcetypeassert // V is same type for both id and other
		return cmp.Compare(a, any(other.value).(uint8)), nil
	case uint16:
		//nolint:forcetypeassert // V is same type for both id and other
		return cmp.Compare(a, any(other.value).(uint16)), nil
	case uint32:
		//nolint:forcetypeassert // V is same type for both id and other
		return cmp.Compare(a, any(other.value).(uint32)), nil
	case uint64:
		//nolint:forcetypeassert // V is same type for both id and other
		return cmp.Compare(a, any(other.value).(uint64)), nil
	case string:
		//nolint:forcetypeassert // V is same type for both id and other
		return cmp.Compare(a, any(other.value).(string)), nil
	default:
		return 0, ErrNotOrdered
	}
}

// Or returns the ID if not zero, otherwise returns the provided default.
func (id ID[B, V]) Or(defaultValue ID[B, V]) ID[B, V] {
	if id.IsZero() {
		return defaultValue
	}

	return id
}

// String returns a string representation of the value.
//
//nolint:cyclop // exhaustive type switch over numeric types
func (id ID[B, V]) String() string {
	switch v := any(id.value).(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		if marshaler, ok := any(id.value).(encoding.TextMarshaler); ok {
			text, err := marshaler.MarshalText()
			if err != nil {
				return fmt.Sprintf("id:%v", id.value)
			}

			return string(text)
		}

		return fmt.Sprintf("%v", id.value)
	}
}

// GoString implements fmt.GoStringer for debugging.
func (id ID[B, V]) GoString() string { return id.String() }

// Format implements fmt.Formatter for custom formatting.
// Supports %s (string), %d (decimal), %v (default), %#v (GoString), %q (quoted).
func (id ID[B, V]) Format(f fmt.State, verb rune) {
	switch verb {
	case 's':
		_, _ = fmt.Fprint(f, id.String())
	case 'd':
		switch v := any(id.value).(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			_, _ = fmt.Fprintf(f, "%d", v)
		default:
			_, _ = fmt.Fprintf(f, "%%!d(type=%T)", id.value)
		}
	case 'q':
		_, _ = fmt.Fprintf(f, "%q", id.String())
	case 'v':
		if f.Flag('#') {
			_, _ = fmt.Fprintf(f, "id(%s)", id.String())
		} else {
			_, _ = fmt.Fprint(f, id.String())
		}
	default:
		_, _ = fmt.Fprintf(f, "%%!%c(type=%T)", verb, id.value)
	}
}
