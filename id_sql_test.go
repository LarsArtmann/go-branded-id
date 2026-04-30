package id

import (
	"testing"
)

func testScanSubTest[B any, V comparable](t *testing.T, name string, input any, expected V) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		t.Parallel()
		testScanRoundTrip[B, V](t, input, expected)
	})
}

func testScanInvalidSubTest[B any, V comparable](
	t *testing.T,
	name, typeName string,
	invalidValue any,
) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		t.Parallel()
		testScanInvalidType[B, V](t, typeName, invalidValue)
	})
}

func TestIDScan(t *testing.T) {
	t.Parallel()
	testScanSubTest[StringBrand, string](t, "string ID from string", testIDValue, testIDValue)

	t.Run("string ID from []byte", func(t *testing.T) {
		t.Parallel()

		var id ID[StringBrand, string]

		err := id.Scan([]byte(testIDValue))
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		assertCmpEqual(t, id.Get(), testIDValue)
	})

	t.Run("string ID from nil", func(t *testing.T) {
		t.Parallel()
		testScanNil[StringBrand, string](t, "string ID")
	})

	testScanInvalidSubTest[StringBrand, string](t, "string ID from invalid type", "string ID", 123)

	testIDScanTests[Int64Brand, int64](t, "int64 ID", 42, int64(42))
	testIDScanTests[Int32Brand, int32](t, "int32 ID", 42, int32(42))

	testScanRoundTripVariants[Int64Brand, any, int64](t, []scanTestCase[any, int64]{
		{"int64 ID from int64", int64(42), int64(42)},
		{"int64 ID from int", 42, 42},
		{"int64 ID from float64", float64(42), 42},
	})

	testScanRoundTripVariants[Uint64Brand, any, uint64](t, []scanTestCase[any, uint64]{
		{"uint64 ID from int64", int64(42), uint64(42)},
	})
}

func testScanRoundTrip[B any, V comparable](t *testing.T, input any, expected V) {
	t.Helper()

	var id ID[B, V]

	err := id.Scan(input)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	assertCmpEqual(t, id.Get(), expected)
}

type scanTestCase[S any, V comparable] struct {
	name     string
	source   S
	expected V
}

func testScanRoundTripVariants[B, S any, V comparable](t *testing.T, cases []scanTestCase[S, V]) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testScanRoundTrip[B, V](t, tc.source, tc.expected)
		})
	}
}

func testScanNil[B any, V comparable](t *testing.T, name string) {
	t.Helper()

	var id ID[B, V]

	err := id.Scan(nil)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !id.IsZero() {
		t.Errorf("%s should be zero", name)
	}
}

func testScanInvalidType[B any, V comparable](t *testing.T, name string, invalidValue any) {
	t.Helper()

	var id ID[B, V]

	err := id.Scan(invalidValue)
	if err == nil {
		t.Errorf("%s should reject invalid type %T", name, invalidValue)
	}
}

func testIDScanTests[B any, V comparable](t *testing.T, typeName string, source int64, expected V) {
	t.Helper()

	t.Run("from nil", func(t *testing.T) {
		t.Parallel()
		testScanNil[B, V](t, typeName)
	})

	t.Run("from invalid type", func(t *testing.T) {
		t.Parallel()

		var zero V
		switch any(zero).(type) {
		case int64, int32, uint64:
			testScanInvalidType[B, V](t, typeName, "not-a-number")
		case string:
			testScanInvalidType[B, V](t, typeName, 123)
		default:
			testScanInvalidType[B, V](t, typeName, nil)
		}
	})

	t.Run("from int64", func(t *testing.T) {
		t.Parallel()
		testScanRoundTrip[B, V](t, source, expected)
	})
}

func testValueNonZero[B any, V comparable](t *testing.T, id ID[B, V], expected any) {
	t.Helper()

	val, err := id.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	if val != expected {
		t.Errorf("expected %v, got %v", expected, val)
	}
}

func testValueZero[B any, V comparable](t *testing.T, id ID[B, V]) {
	t.Helper()

	val, err := id.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}
}

func testIDValueTests[B any, V comparable](
	t *testing.T,
	typeName string,
	value V,
	expectedValue any,
) {
	t.Helper()

	t.Run(typeName+"/non-zero", func(t *testing.T) {
		t.Parallel()

		id := NewID[B, V](value)
		testValueNonZero(t, id, expectedValue)
	})

	t.Run(typeName+"/zero", func(t *testing.T) {
		t.Parallel()

		var id ID[B, V]
		testValueZero(t, id)
	})
}

func TestIDValue(t *testing.T) {
	t.Parallel()
	testIDValueTests[StringBrand, string](t, "string ID", testIDValue, testIDValue)
	testIDValueTests[Int64Brand, int64](t, "int64 ID", int64(42), int64(42))
	testIDValueTests[Int32Brand, int32](t, "int32 ID", int32(42), int64(42))
	testIDValueTests[Uint64Brand, uint64](t, "uint64 ID", uint64(42), int64(42))
}
