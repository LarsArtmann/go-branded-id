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
) error {
	err := validateSize(data, byteSize, typeName, *id)
	if err != nil {
		return err
	}

	n := readFunc(data)
	//nolint:forcetypeassert // readUnsigned is only called for uint types where V matches uint64
	*id = ID[B, V]{value: any(n).(V)}

	return nil
}

// readSigned reads a signed integer from data.
// IntType is the unsigned type used to read the bytes (uint16, uint32, or uint64).
func readSigned[V, IntType any](
	data []byte,
	typeName string,
	readFunc func([]byte) IntType,
	convertFunc func(IntType) V,
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

	//nolint:forcetypeassert // readByte is only called for byte-sized types where V matches
	*id = ID[B, V]{value: any(convertFunc(data[0])).(V)}

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
		n, err := readSigned(
			data,
			"int",
			readUint64,
			func(n uint64) V {
				return any(int(n)).(V)
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
			func(b byte) V { //nolint:gosec // G115: byte to int8 is safe for deserialization
				return any(int8(b)).(V)
			},
		)
	case int16:
		n, err := readSigned(
			data,
			"int16",
			readUint16,
			func(n uint16) V {
				return any(int16(n)).(V)
			},
			byteSizeInt16,
		)
		if err != nil {
			return err
		}

		*id = ID[B, V]{value: n}

		return nil
	case int32:
		n, err := readSigned(
			data,
			"int32",
			readUint32,
			func(n uint32) V {
				return any(int32(n)).(V)
			},
			byteSizeInt32,
		)
		if err != nil {
			return err
		}

		*id = ID[B, V]{value: n}

		return nil
	case int64:
		n, err := readSigned(
			data,
			"int64",
			readUint64,
			func(n uint64) V {
				return any(int64(n)).(V)
			},
			byteSizeInt64,
		)
		if err != nil {
			return err
		}

		*id = ID[B, V]{value: n}

		return nil
	case uint:
		n, err := readSigned(
			data,
			"uint",
			readUint64,
			func(n uint64) V {
				return any(uint(n)).(V)
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
				return any(b).(V)
			},
		)
	case uint16:
		return id.readUnsigned(
			data,
			byteSizeInt16,
			"uint16",
			func(d []byte) uint64 { return uint64(readUint16(d)) },
		)
	case uint32:
		return id.readUnsigned(
			data,
			byteSizeInt32,
			"uint32",
			func(d []byte) uint64 { return uint64(readUint32(d)) },
		)
	case uint64:
		return id.readUnsigned(data, byteSizeInt64, "uint64", readUint64)
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
