package id

import (
	"encoding/json"
	"math"
	"testing"
)

func TestCompareAllTypes(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()
		testCompareOrdered[IntBrand, int](t, 1, 2, -1)
		testCompareOrdered[IntBrand, int](t, 5, 5, 0)
		testCompareOrdered[IntBrand, int](t, 3, 1, 1)
	})

	t.Run("int8", func(t *testing.T) {
		t.Parallel()
		testCompareOrdered[Int8Brand, int8](t, 1, 2, -1)
		testCompareOrdered[Int8Brand, int8](t, 5, 5, 0)
		testCompareOrdered[Int8Brand, int8](t, 3, 1, 1)
	})

	t.Run("int16", func(t *testing.T) {
		t.Parallel()
		testCompareOrdered[Int16Brand, int16](t, 1, 2, -1)
		testCompareOrdered[Int16Brand, int16](t, 5, 5, 0)
		testCompareOrdered[Int16Brand, int16](t, 3, 1, 1)
	})

	t.Run("uint", func(t *testing.T) {
		t.Parallel()
		testCompareOrdered[UintBrand, uint](t, 1, 2, -1)
		testCompareOrdered[UintBrand, uint](t, 5, 5, 0)
		testCompareOrdered[UintBrand, uint](t, 3, 1, 1)
	})

	t.Run("uint8", func(t *testing.T) {
		t.Parallel()
		testCompareOrdered[Uint8Brand, uint8](t, 1, 2, -1)
		testCompareOrdered[Uint8Brand, uint8](t, 5, 5, 0)
		testCompareOrdered[Uint8Brand, uint8](t, 3, 1, 1)
	})

	t.Run("uint16", func(t *testing.T) {
		t.Parallel()
		testCompareOrdered[Uint16Brand, uint16](t, 1, 2, -1)
		testCompareOrdered[Uint16Brand, uint16](t, 5, 5, 0)
		testCompareOrdered[Uint16Brand, uint16](t, 3, 1, 1)
	})

	t.Run("uint32", func(t *testing.T) {
		t.Parallel()
		testCompareOrdered[Uint32Brand, uint32](t, 1, 2, -1)
		testCompareOrdered[Uint32Brand, uint32](t, 5, 5, 0)
		testCompareOrdered[Uint32Brand, uint32](t, 3, 1, 1)
	})
}

func testCompareOrdered[B any, V comparable](t *testing.T, a, b V, expected int) {
	t.Helper()

	idA := NewID[B, V](a)
	idB := NewID[B, V](b)

	result, err := idA.Compare(idB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Compare(%v, %v): expected %d, got %d", a, b, expected, result)
	}
}

func TestStringAllTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       any
		expected string
	}{
		{"string", NewID[StringBrand](testIDValue), testIDValue},
		{"int", NewID[IntBrand, int](42), "42"},
		{"int8", NewID[Int8Brand, int8](42), "42"},
		{"int16", NewID[Int16Brand, int16](42), "42"},
		{"int32", NewID[Int32Brand, int32](42), "42"},
		{"int64", NewID[Int64Brand, int64](42), "42"},
		{"uint", NewID[UintBrand, uint](42), "42"},
		{"uint8", NewID[Uint8Brand, uint8](42), "42"},
		{"uint16", NewID[Uint16Brand, uint16](42), "42"},
		{"uint32", NewID[Uint32Brand, uint32](42), "42"},
		{"uint64", NewID[Uint64Brand, uint64](42), "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assertCmpEqual(t, stringer(t, tt.id), tt.expected)
		})
	}
}

//nolint:cyclop // exhaustive type switch
func stringer(t *testing.T, v any) string {
	t.Helper()

	switch id := v.(type) {
	case ID[StringBrand, string]:
		return id.String()
	case ID[IntBrand, int]:
		return id.String()
	case ID[Int8Brand, int8]:
		return id.String()
	case ID[Int16Brand, int16]:
		return id.String()
	case ID[Int32Brand, int32]:
		return id.String()
	case ID[Int64Brand, int64]:
		return id.String()
	case ID[UintBrand, uint]:
		return id.String()
	case ID[Uint8Brand, uint8]:
		return id.String()
	case ID[Uint16Brand, uint16]:
		return id.String()
	case ID[Uint32Brand, uint32]:
		return id.String()
	case ID[Uint64Brand, uint64]:
		return id.String()
	default:
		t.Fatalf("unsupported type: %T", v)

		return ""
	}
}

func testBinaryRoundTrip[B any, V comparable](t *testing.T, value V) {
	t.Helper()

	testIDRoundTrip(t, value,
		func(id ID[B, V]) ([]byte, error) { return id.MarshalBinary() },
		func(id *ID[B, V], data []byte) error { return id.UnmarshalBinary(data) },
	)
}

func testJSONRoundTrip[B any, V comparable](t *testing.T, value V) {
	t.Helper()

	testIDRoundTrip(t, value,
		func(id ID[B, V]) ([]byte, error) { return json.Marshal(id) },
		func(id *ID[B, V], data []byte) error { return json.Unmarshal(data, id) },
	)
}

//nolint:funlen // table-driven test with multiple type sub-tests
func TestBinaryRoundTripAllTypes(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[IntBrand, int](t, 42)
	})

	t.Run("int8", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[Int8Brand, int8](t, 42)
		testBinaryRoundTrip[Int8Brand, int8](t, -128)
		testBinaryRoundTrip[Int8Brand, int8](t, 127)
	})

	t.Run("int16", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[Int16Brand, int16](t, 1024)
		testBinaryRoundTrip[Int16Brand, int16](t, -32768)
		testBinaryRoundTrip[Int16Brand, int16](t, 32767)
	})

	t.Run("int32", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[Int32Brand, int32](t, 100000)
	})

	t.Run("int64 negative", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[Int64Brand, int64](t, -42)
	})

	t.Run("int64 max", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[Int64Brand, int64](t, math.MaxInt64)
	})

	t.Run("uint", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[UintBrand, uint](t, 42)
	})

	t.Run("uint8", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[Uint8Brand, uint8](t, 42)
		testBinaryRoundTrip[Uint8Brand, uint8](t, 0)
		testBinaryRoundTrip[Uint8Brand, uint8](t, 255)
	})

	t.Run("uint16", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[Uint16Brand, uint16](t, 1024)
		testBinaryRoundTrip[Uint16Brand, uint16](t, 0)
		testBinaryRoundTrip[Uint16Brand, uint16](t, 65535)
	})

	t.Run("uint32", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[Uint32Brand, uint32](t, 100000)
		testBinaryRoundTrip[Uint32Brand, uint32](t, 0)
	})

	t.Run("uint64 max", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[Uint64Brand, uint64](t, math.MaxUint64)
	})
}

func TestBinaryZeroValueAllTypes(t *testing.T) {
	t.Parallel()

	testBinaryZeroRoundTrip[IntBrand, int](t)
	testBinaryZeroRoundTrip[Int8Brand, int8](t)
	testBinaryZeroRoundTrip[Int16Brand, int16](t)
	testBinaryZeroRoundTrip[Int32Brand, int32](t)
	testBinaryZeroRoundTrip[Int64Brand, int64](t)
	testBinaryZeroRoundTrip[UintBrand, uint](t)
	testBinaryZeroRoundTrip[Uint8Brand, uint8](t)
	testBinaryZeroRoundTrip[Uint16Brand, uint16](t)
	testBinaryZeroRoundTrip[Uint32Brand, uint32](t)
	testBinaryZeroRoundTrip[Uint64Brand, uint64](t)
	testBinaryZeroRoundTrip[StringBrand, string](t)
}

func testBinaryZeroRoundTrip[B any, V comparable](t *testing.T) {
	t.Helper()

	var original ID[B, V]

	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary zero failed for %T: %v", original, err)
	}

	if data != nil {
		t.Errorf("expected nil data for zero %T, got %v", original, data)
	}

	var restored ID[B, V]

	err = restored.UnmarshalBinary(nil)
	if err != nil {
		t.Fatalf("UnmarshalBinary nil failed for %T: %v", restored, err)
	}

	if !restored.IsZero() {
		t.Errorf("expected zero after unmarshal nil for %T", restored)
	}
}

func TestBinaryMarshalUnsupportedType(t *testing.T) {
	t.Parallel()

	type Custom struct{ X int }

	id := NewID[struct{}, Custom](Custom{X: 1})

	_, err := id.MarshalBinary()
	if err == nil {
		t.Error("expected error for unsupported binary type")
	}
}

func TestBinaryUnmarshalInsufficientData(t *testing.T) {
	t.Parallel()

	t.Run("int64 too short", func(t *testing.T) {
		t.Parallel()

		var id ID[Int64Brand, int64]

		err := id.UnmarshalBinary([]byte{1, 2, 3})
		if err == nil {
			t.Error("expected error for insufficient data")
		}
	})

	t.Run("int16 too short", func(t *testing.T) {
		t.Parallel()

		var id ID[Int16Brand, int16]

		err := id.UnmarshalBinary([]byte{1})
		if err == nil {
			t.Error("expected error for insufficient data")
		}
	})

	t.Run("int8 empty resets", func(t *testing.T) {
		t.Parallel()

		var id ID[Int8Brand, int8]

		err := id.UnmarshalBinary([]byte{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !id.IsZero() {
			t.Error("expected zero after unmarshaling empty data")
		}
	})

	t.Run("uint8 empty resets", func(t *testing.T) {
		t.Parallel()

		var id ID[Uint8Brand, uint8]

		err := id.UnmarshalBinary([]byte{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !id.IsZero() {
			t.Error("expected zero after unmarshaling empty data")
		}
	})
}

func TestJSONRoundTripAllTypes(t *testing.T) {
	t.Parallel()

	t.Run("int8", func(t *testing.T) {
		t.Parallel()
		testJSONRoundTrip[Int8Brand, int8](t, 42)
		testJSONRoundTrip[Int8Brand, int8](t, -128)
	})

	t.Run("int16", func(t *testing.T) {
		t.Parallel()
		testJSONRoundTrip[Int16Brand, int16](t, 1024)
	})

	t.Run("int", func(t *testing.T) {
		t.Parallel()
		testJSONRoundTrip[IntBrand, int](t, 42)
	})

	t.Run("uint", func(t *testing.T) {
		t.Parallel()
		testJSONRoundTrip[UintBrand, uint](t, 42)
	})

	t.Run("uint8", func(t *testing.T) {
		t.Parallel()
		testJSONRoundTrip[Uint8Brand, uint8](t, 255)
	})

	t.Run("uint16", func(t *testing.T) {
		t.Parallel()
		testJSONRoundTrip[Uint16Brand, uint16](t, 65535)
	})

	t.Run("uint32", func(t *testing.T) {
		t.Parallel()
		testJSONRoundTrip[Uint32Brand, uint32](t, 100000)
	})

	t.Run("negative int64", func(t *testing.T) {
		t.Parallel()
		testJSONRoundTrip[Int64Brand, int64](t, -42)
	})

	t.Run("max uint64", func(t *testing.T) {
		t.Parallel()
		testJSONRoundTrip[Uint64Brand, uint64](t, math.MaxUint64)
	})
}

func TestScanAllIntegerTypes(t *testing.T) {
	t.Parallel()

	t.Run("int from int64", func(t *testing.T) {
		t.Parallel()
		testScanRoundTrip[IntBrand, int](t, int64(42), 42)
	})

	t.Run("int8 from int64", func(t *testing.T) {
		t.Parallel()
		testScanRoundTrip[Int8Brand, int8](t, int64(42), int8(42))
	})

	t.Run("int16 from int64", func(t *testing.T) {
		t.Parallel()
		testScanRoundTrip[Int16Brand, int16](t, int64(42), int16(42))
	})

	t.Run("uint from int64", func(t *testing.T) {
		t.Parallel()
		testScanRoundTrip[UintBrand, uint](t, int64(42), uint(42))
	})

	t.Run("uint8 from int64", func(t *testing.T) {
		t.Parallel()
		testScanRoundTrip[Uint8Brand, uint8](t, int64(42), uint8(42))
	})

	t.Run("uint16 from int64", func(t *testing.T) {
		t.Parallel()
		testScanRoundTrip[Uint16Brand, uint16](t, int64(42), uint16(42))
	})

	t.Run("uint32 from int64", func(t *testing.T) {
		t.Parallel()
		testScanRoundTrip[Uint32Brand, uint32](t, int64(42), uint32(42))
	})
}

func TestValueAllTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		id          any
		expectedVal any
	}{
		{"string non-zero", NewID[StringBrand](testIDValue), testIDValue},
		{"int non-zero", NewID[IntBrand, int](42), int64(42)},
		{"int8 non-zero", NewID[Int8Brand, int8](42), int64(42)},
		{"int16 non-zero", NewID[Int16Brand, int16](42), int64(42)},
		{"uint non-zero", NewID[UintBrand, uint](42), int64(42)},
		{"uint8 non-zero", NewID[Uint8Brand, uint8](42), int64(42)},
		{"uint16 non-zero", NewID[Uint16Brand, uint16](42), int64(42)},
		{"uint32 non-zero", NewID[Uint32Brand, uint32](42), int64(42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			val, err := valuer(t, tt.id)
			if err != nil {
				t.Fatalf("Value failed: %v", err)
			}

			if val != tt.expectedVal {
				t.Errorf("expected %v (%T), got %v (%T)", tt.expectedVal, tt.expectedVal, val, val)
			}
		})
	}
}

//nolint:cyclop // exhaustive type switch
func valuer(t *testing.T, v any) (any, error) {
	t.Helper()

	switch id := v.(type) {
	case ID[StringBrand, string]:
		return id.Value()
	case ID[IntBrand, int]:
		return id.Value()
	case ID[Int8Brand, int8]:
		return id.Value()
	case ID[Int16Brand, int16]:
		return id.Value()
	case ID[Int32Brand, int32]:
		return id.Value()
	case ID[Int64Brand, int64]:
		return id.Value()
	case ID[UintBrand, uint]:
		return id.Value()
	case ID[Uint8Brand, uint8]:
		return id.Value()
	case ID[Uint16Brand, uint16]:
		return id.Value()
	case ID[Uint32Brand, uint32]:
		return id.Value()
	case ID[Uint64Brand, uint64]:
		return id.Value()
	default:
		t.Fatalf("unsupported type: %T", v)

		return nil, nil //nolint:nilnil // unreachable after t.Fatalf
	}
}

func TestScanNilAllTypes(t *testing.T) {
	t.Parallel()

	t.Run("int nil", func(t *testing.T) {
		t.Parallel()
		testScanNil[IntBrand, int](t, "int ID")
	})
	t.Run("int8 nil", func(t *testing.T) {
		t.Parallel()
		testScanNil[Int8Brand, int8](t, "int8 ID")
	})
	t.Run("int16 nil", func(t *testing.T) {
		t.Parallel()
		testScanNil[Int16Brand, int16](t, "int16 ID")
	})
	t.Run("uint nil", func(t *testing.T) {
		t.Parallel()
		testScanNil[UintBrand, uint](t, "uint ID")
	})
	t.Run("uint8 nil", func(t *testing.T) {
		t.Parallel()
		testScanNil[Uint8Brand, uint8](t, "uint8 ID")
	})
	t.Run("uint16 nil", func(t *testing.T) {
		t.Parallel()
		testScanNil[Uint16Brand, uint16](t, "uint16 ID")
	})
}

func TestUnmarshalTextAllNumericTypes(t *testing.T) {
	t.Parallel()

	t.Run("int from text", func(t *testing.T) {
		t.Parallel()
		testUnmarshalTextRoundTrip[IntBrand, int](t, "42", 42)
	})

	t.Run("int64 from text", func(t *testing.T) {
		t.Parallel()
		testUnmarshalTextRoundTrip[Int64Brand, int64](t, "42", 42)
	})

	t.Run("uint64 from text", func(t *testing.T) {
		t.Parallel()
		testUnmarshalTextRoundTrip[Uint64Brand, uint64](t, "42", 42)
	})

	t.Run("negative int64 from text", func(t *testing.T) {
		t.Parallel()
		testUnmarshalTextRoundTrip[Int64Brand, int64](t, "-42", -42)
	})

	t.Run("invalid int64 text", func(t *testing.T) {
		t.Parallel()

		var id ID[Int64Brand, int64]

		err := id.UnmarshalText([]byte("not-a-number"))
		if err == nil {
			t.Error("expected error for invalid int64 text")
		}
	})

	t.Run("invalid uint64 text", func(t *testing.T) {
		t.Parallel()

		var id ID[Uint64Brand, uint64]

		err := id.UnmarshalText([]byte("not-a-number"))
		if err == nil {
			t.Error("expected error for invalid uint64 text")
		}
	})
}
