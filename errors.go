package id

import "errors"

// Sentinel errors returned by ID operations. Match with errors.Is to branch on
// error category; call sites wrap them with fmt.Errorf("%w: ...", sentinel, ...)
// to attach runtime context (type names, byte counts, etc.) without losing the
// sentinel identity.
var (
	// ErrInvalidID is returned when ID validation fails (zero value).
	ErrInvalidID = errors.New("id: invalid")

	// ErrNotOrdered is returned when Compare is called on an ID with a non-ordered value type.
	ErrNotOrdered = errors.New("id: Compare requires an ordered type (int, uint, or string)")

	// ErrUnsupportedType is returned when a serialization format does not support
	// the ID's value type V (e.g., a struct passed to binary marshaling).
	ErrUnsupportedType = errors.New("id: unsupported type")

	// ErrCannotScan is returned when a SQL source value cannot be scanned into the
	// ID because the source type does not match the target value type.
	ErrCannotScan = errors.New("id: cannot scan")

	// ErrInsufficientData is returned when binary data is too short to decode a value.
	ErrInsufficientData = errors.New("id: insufficient data")

	// ErrInternal is returned for unreachable internal errors, such as a type
	// assertion that the type switch should have guaranteed.
	ErrInternal = errors.New("id: internal error")

	// ErrNilReceiver is returned when a method is called on a nil pointer receiver.
	ErrNilReceiver = errors.New("id: nil receiver")

	// ErrMarshal is returned when marshaling an ID's value to a serialization format
	// fails because the value type's own marshaler (encoding.TextMarshaler,
	// encoding.BinaryMarshaler, json.Marshaler) returned an error.
	ErrMarshal = errors.New("id: marshal failed")

	// ErrUnmarshal is returned when unmarshaling data into an ID fails because the
	// value type's own unmarshaler returned an error, or because the text/JSON input
	// could not be parsed as the target value type.
	ErrUnmarshal = errors.New("id: unmarshal failed")
)
