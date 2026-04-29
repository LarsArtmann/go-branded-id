package id

import (
	"encoding"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

// Byte sizes for binary marshaling of integer types.
const (
	byteSizeInt16 = 2 // size of int16 and uint16 in bytes
	byteSizeInt32 = 4 // size of int32 and uint32 in bytes
	byteSizeInt64 = 8 // size of int64, uint, and uint64 in bytes
)

// validateSize checks if data has at least the expected size and returns an error if not.
func validateSize(data []byte, want int, typeName string, zero any) error {
	if len(data) < want {
		return fmt.Errorf(
			"id: insufficient data for %s: got %d bytes, want %d (data=%x, targetType=%T)",
			typeName,
			len(data),
			want,
			data,
			zero,
		)
	}

	return nil
}

// readUnsigned reads an unsigned integer from data and assigns it to id.
func (id *ID[B, V]) readUnsigned(
	data []byte,
	byteSize int,
	typeName string,
	readFunc func([]byte) uint64,
	convertFunc func(uint64) V,
) error {
	if err := validateSize(data, byteSize, typeName, *id); err != nil {
		return err
	}

	*id = ID[B, V]{value: convertFunc(readFunc(data))}

	return nil
}

// readBinary reads a fixed-size integer from data and converts it to V.
// I is the raw type read from bytes (uint16, uint32, or uint64).
func readBinary[V, I any](
	data []byte,
	typeName string,
	readFunc func([]byte) I,
	convertFunc func(I) V,
	byteSize int,
) (V, error) {
	var zero V

	if len(data) < byteSize {
		return zero, fmt.Errorf(
			"id: insufficient data for %s: got %d bytes, want %d",
			typeName,
			len(data),
			byteSize,
		)
	}

	return convertFunc(readFunc(data)), nil
}

// readByte reads a single byte and assigns it to id.
func (id *ID[B, V]) readByte(data []byte, typeName string, convertFunc func(byte) V) error {
	err := validateSize(data, 1, typeName, *id)
	if err != nil {
		return err
	}

	*id = ID[B, V]{value: convertFunc(data[0])}

	return nil
}

// readUint16 reads a uint16 from data using LittleEndian.
func readUint16(data []byte) uint16 {
	return binary.LittleEndian.Uint16(data)
}

// readUint32 reads a uint32 from data using LittleEndian.
func readUint32(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data)
}

// readUint64 reads a uint64 from data using LittleEndian.
func readUint64(data []byte) uint64 {
	return binary.LittleEndian.Uint64(data)
}

// MarshalBinary implements encoding.BinaryMarshaler for binary encoding.
//
//nolint:cyclop,funlen // exhaustive type switch over numeric types
func (id ID[B, V]) MarshalBinary() ([]byte, error) {
	if id.IsZero() {
		return nil, nil
	}

	switch v := any(id.value).(type) {
	case string:
		return []byte(v), nil
	case int:
		b := make([]byte, byteSizeInt64)
		//nolint:gosec // G115: int to uint64 is safe for binary serialization
		binary.LittleEndian.PutUint64(b, uint64(v))

		return b, nil
	case int8:
		return []byte{byte(v)}, nil //nolint:gosec // G115: int8 to byte is safe for serialization
	case int16:
		b := make([]byte, byteSizeInt16)
		//nolint:gosec // G115: int16 to uint16 is safe for binary serialization
		binary.LittleEndian.PutUint16(b, uint16(v))

		return b, nil
	case int32:
		b := make([]byte, byteSizeInt32)
		//nolint:gosec // G115: int32 to uint32 is safe for binary serialization
		binary.LittleEndian.PutUint32(b, uint32(v))

		return b, nil
	case int64:
		b := make([]byte, byteSizeInt64)
		//nolint:gosec // G115: int64 to uint64 is safe for binary serialization
		binary.LittleEndian.PutUint64(b, uint64(v))

		return b, nil
	case uint:
		b := make([]byte, byteSizeInt64)

		binary.LittleEndian.PutUint64(b, uint64(v))

		return b, nil
	case uint8:
		return []byte{v}, nil
	case uint16:
		b := make([]byte, byteSizeInt16)
		binary.LittleEndian.PutUint16(b, v)

		return b, nil
	case uint32:
		b := make([]byte, byteSizeInt32)
		binary.LittleEndian.PutUint32(b, v)

		return b, nil
	case uint64:
		b := make([]byte, byteSizeInt64)
		binary.LittleEndian.PutUint64(b, v)

		return b, nil
	default:
		if marshaler, ok := any(id.value).(encoding.BinaryMarshaler); ok {
			data, err := marshaler.MarshalBinary()
			if err != nil {
				return nil, fmt.Errorf("id: binary marshal %T: %w", id.value, err)
			}

			return data, nil
		}

		return nil, fmt.Errorf("id: unsupported type %T for binary marshaling", id.value)
	}
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler for binary decoding.
//
//nolint:cyclop,funlen // exhaustive type switch over numeric types
func (id *ID[B, V]) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		id.Reset()

		return nil
	}

	var zero V
	switch any(zero).(type) {
	case string:
		//nolint:forcetypeassert // type switch guarantees V is string
		*id = ID[B, V]{value: any(string(data)).(V)}

		return nil
	case int:
		n, err := readBinary(
			data,
			"int",
			readUint64,
			func(n uint64) V {
				return any(int(n)).(V) //nolint:gosec,forcetypeassert // G115: uint64 to int for binary deserialization; guaranteed by type switch
			},
			byteSizeInt64,
		)
		if err != nil {
			return err
		}

		*id = ID[B, V]{value: n}

		return nil
	case int8:
		return id.readByte(
			data,
			"int8",
			func(b byte) V { //nolint:gosec,forcetypeassert // G115: byte to int8 is safe for deserialization; guaranteed by type switch
				return any(int8(b)).(V)
			},
		)
	case int16:
		n, err := readBinary(
			data,
			"int16",
			readUint16,
			func(n uint16) V {
				return any(int16(n)).(V) //nolint:gosec,forcetypeassert // G115: uint16 to int16 for binary deserialization; guaranteed by type switch
			},
			byteSizeInt16,
		)
		if err != nil {
			return err
		}

		*id = ID[B, V]{value: n}

		return nil
	case int32:
		n, err := readBinary(
			data,
			"int32",
			readUint32,
			func(n uint32) V {
				return any(int32(n)).(V) //nolint:gosec,forcetypeassert // G115: uint32 to int32 for binary deserialization; guaranteed by type switch
			},
			byteSizeInt32,
		)
		if err != nil {
			return err
		}

		*id = ID[B, V]{value: n}

		return nil
	case int64:
		n, err := readBinary(
			data,
			"int64",
			readUint64,
			func(n uint64) V {
				return any(int64(n)).(V) //nolint:gosec,forcetypeassert // G115: uint64 to int64 for binary deserialization; guaranteed by type switch
			},
			byteSizeInt64,
		)
		if err != nil {
			return err
		}

		*id = ID[B, V]{value: n}

		return nil
	case uint:
		n, err := readBinary(
			data,
			"uint",
			readUint64,
			func(n uint64) V {
				return any(uint(n)).(V) //nolint:forcetypeassert // guaranteed by type switch
			},
			byteSizeInt64,
		)
		if err != nil {
			return err
		}

		*id = ID[B, V]{value: n}

		return nil
	case uint8:
		return id.readByte(
			data,
			"uint8",
			func(b byte) V {
				return any(b).(V) //nolint:forcetypeassert // guaranteed by outer type switch
			},
		)
	case uint16:
		return id.readUnsigned(
			data,
			byteSizeInt16,
			"uint16",
			func(d []byte) uint64 { return uint64(readUint16(d)) },
			func(n uint64) V { //nolint:gosec // G115: controlled conversion for binary deserialization
				return any(uint16(n)).(V)
			},
		)
	case uint32:
		return id.readUnsigned(
			data,
			byteSizeInt32,
			"uint32",
			func(d []byte) uint64 { return uint64(readUint32(d)) },
			func(n uint64) V { //nolint:gosec // G115: controlled conversion for binary deserialization
				return any(uint32(n)).(V)
			},
		)
	case uint64:
		return id.readUnsigned(
			data,
			byteSizeInt64,
			"uint64",
			readUint64,
			func(n uint64) V { //nolint:forcetypeassert // guaranteed by outer type switch
				return any(n).(V)
			},
		)
	default:
		var zero V
		if unmarshaler, ok := any(&zero).(encoding.BinaryUnmarshaler); ok {
			err := unmarshaler.UnmarshalBinary(data)
			if err != nil {
				return fmt.Errorf("id: cannot unmarshal binary into %T: %w", zero, err)
			}

			*id = ID[B, V]{value: zero}

			return nil
		}

		return fmt.Errorf("id: unsupported type %T for binary unmarshaling (data=%x)", zero, data)
	}
}

// GobEncode implements gob.GobEncoder for Go-specific encoding.
func (id ID[B, V]) GobEncode() ([]byte, error) {
	return id.MarshalBinary()
}

// GobDecode implements gob.GobDecoder for Go-specific decoding.
func (id *ID[B, V]) GobDecode(data []byte) error {
	return id.UnmarshalBinary(data)
}

// Compile-time interface assertions for binary encoding.
var (
	_ encoding.BinaryMarshaler   = ID[struct{}, string]{value: ""}
	_ encoding.BinaryUnmarshaler = (*ID[struct{}, string])(nil)
	_ encoding.BinaryMarshaler   = ID[struct{}, int64]{value: 0}
	_ encoding.BinaryUnmarshaler = (*ID[struct{}, int64])(nil)
	_ gob.GobEncoder             = ID[struct{}, string]{value: ""}
	_ gob.GobDecoder             = (*ID[struct{}, string])(nil)
)
