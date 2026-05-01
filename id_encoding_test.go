package id

import (
	"bytes"
	"encoding/gob"
	"testing"
)

func assertCmpEqual[T comparable](tb testing.TB, got, want T) {
	tb.Helper()

	if got != want {
		tb.Errorf("expected %v, got %v", want, got)
	}
}

func TestIDText(t *testing.T) { //nolint:funlen // table-driven test with multiple sub-tests
	t.Parallel()
	t.Run("marshal non-zero string", func(t *testing.T) {
		t.Parallel()

		id := NewID[StringBrand](testIDValue)

		data, err := id.MarshalText()
		if err != nil {
			t.Fatalf("MarshalText failed: %v", err)
		}

		if string(data) != testIDValue {
			t.Errorf("expected test-id, got %s", string(data))
		}
	})

	t.Run("marshal zero string", func(t *testing.T) {
		t.Parallel()

		var id ID[StringBrand, string]

		data, err := id.MarshalText()
		if err != nil {
			t.Fatalf("MarshalText failed: %v", err)
		}

		if data != nil {
			t.Errorf("expected nil, got %s", string(data))
		}
	})

	t.Run("marshal int64", func(t *testing.T) {
		t.Parallel()

		id := NewID[Int64Brand, int64](42)

		data, err := id.MarshalText()
		if err != nil {
			t.Fatalf("MarshalText failed: %v", err)
		}

		assertCmpEqual(t, string(data), "42")
	})

	t.Run("unmarshal valid string", func(t *testing.T) {
		t.Parallel()
		testUnmarshalTextRoundTrip[StringBrand, string](t, testIDValue, testIDValue)
	})

	t.Run("unmarshal empty", func(t *testing.T) {
		t.Parallel()

		var id ID[StringBrand, string]

		err := id.UnmarshalText([]byte{})
		if err != nil {
			t.Fatalf("UnmarshalText failed: %v", err)
		}

		if !id.IsZero() {
			t.Error("expected zero ID")
		}
	})

	t.Run("numeric IDs", func(t *testing.T) {
		t.Parallel()

		testIDAllTypesUnmarshalText(t, textUnmarshalTestImpl{})
	})
}

type textUnmarshalTest interface {
	TestInt64(t *testing.T)
	TestUint64(t *testing.T)
}

type textUnmarshalTestImpl struct{}

func (t textUnmarshalTestImpl) TestInt64(tx *testing.T) {
	tx.Parallel()
	testUnmarshalTextRoundTrip[Int64Brand, int64](tx, "42", 42)
}

func (t textUnmarshalTestImpl) TestUint64(tx *testing.T) {
	tx.Parallel()
	testUnmarshalTextRoundTrip[Uint64Brand, uint64](tx, "42", 42)
}

func testIDAllTypesUnmarshalText(t *testing.T, ut textUnmarshalTest) {
	t.Helper()

	t.Run("int64 ID", ut.TestInt64)
	t.Run("uint64 ID", ut.TestUint64)
}

func testUnmarshalTextRoundTrip[B any, V comparable](t *testing.T, input string, expected V) {
	t.Helper()

	var id ID[B, V]

	err := id.UnmarshalText([]byte(input))
	if err != nil {
		t.Fatalf("UnmarshalText failed: %v", err)
	}

	assertCmpEqual(t, id.Get(), expected)
}

// testIDRoundTrip is a shared helper for round-trip serialization tests.
func testIDRoundTrip[B any, V comparable](
	t *testing.T,
	value V,
	marshal func(ID[B, V]) ([]byte, error),
	unmarshal func(*ID[B, V], []byte) error,
) {
	t.Helper()

	original := NewID[B, V](value)

	data, err := marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var restored ID[B, V]

	err = unmarshal(&restored, data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	assertCmpEqual(t, original.Get(), restored.Get())
}

func TestIDBinary(t *testing.T) {
	testIDAllTypesRoundTrip(t, binaryRoundTripTest{})

	t.Run("zero ID", func(t *testing.T) {
		t.Parallel()

		var original ID[StringBrand, string]

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary failed: %v", err)
		}

		if data != nil {
			t.Errorf("expected nil, got %v", data)
		}

		var restored ID[StringBrand, string]

		err = restored.UnmarshalBinary(nil)
		if err != nil {
			t.Fatalf("UnmarshalBinary failed: %v", err)
		}

		if !restored.IsZero() {
			t.Error("expected zero ID")
		}
	})
}

type binaryRoundTripTest struct{}

func (b binaryRoundTripTest) TestString(t *testing.T) {
	t.Parallel()
	testIDRoundTrip(
		t,
		testIDValue,
		func(id ID[StringBrand, string]) ([]byte, error) { return id.MarshalBinary() },
		func(id *ID[StringBrand, string], data []byte) error { return id.UnmarshalBinary(data) },
	)
}

func (b binaryRoundTripTest) TestInt64(t *testing.T) {
	t.Parallel()
	testIDRoundTrip(
		t,
		int64(42),
		func(id ID[Int64Brand, int64]) ([]byte, error) { return id.MarshalBinary() },
		func(id *ID[Int64Brand, int64], data []byte) error { return id.UnmarshalBinary(data) },
	)
}

func (b binaryRoundTripTest) TestInt32(t *testing.T) {
	t.Parallel()
	testIDRoundTrip(
		t,
		int32(42),
		func(id ID[Int32Brand, int32]) ([]byte, error) { return id.MarshalBinary() },
		func(id *ID[Int32Brand, int32], data []byte) error { return id.UnmarshalBinary(data) },
	)
}

func (b binaryRoundTripTest) TestUint64(t *testing.T) {
	t.Parallel()
	testIDRoundTrip(
		t,
		uint64(42),
		func(id ID[Uint64Brand, uint64]) ([]byte, error) { return id.MarshalBinary() },
		func(id *ID[Uint64Brand, uint64], data []byte) error { return id.UnmarshalBinary(data) },
	)
}

func TestIDGob(t *testing.T) {
	t.Parallel()
	t.Run("string ID", func(t *testing.T) {
		t.Parallel()

		original := NewID[StringBrand](testIDValue)

		var buf bytes.Buffer

		enc := gob.NewEncoder(&buf)

		err := enc.Encode(original)
		if err != nil {
			t.Fatalf("GobEncode failed: %v", err)
		}

		var restored ID[StringBrand, string]

		dec := gob.NewDecoder(&buf)

		err = dec.Decode(&restored)
		if err != nil {
			t.Fatalf("GobDecode failed: %v", err)
		}

		assertCmpEqual(t, original.Get(), restored.Get())
	})

	t.Run("int64 ID", func(t *testing.T) {
		t.Parallel()

		original := NewID[Int64Brand, int64](42)

		var buf bytes.Buffer

		enc := gob.NewEncoder(&buf)

		err := enc.Encode(original)
		if err != nil {
			t.Fatalf("GobEncode failed: %v", err)
		}

		var restored ID[Int64Brand, int64]

		dec := gob.NewDecoder(&buf)

		err = dec.Decode(&restored)
		if err != nil {
			t.Fatalf("GobDecode failed: %v", err)
		}

		assertCmpEqual(t, original.Get(), restored.Get())
	})
}
