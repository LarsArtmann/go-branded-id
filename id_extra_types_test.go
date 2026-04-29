package id

import "testing"

func TestBinaryRoundTripAllTypes(t *testing.T) {
	t.Parallel()

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

	t.Run("int", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[IntBrand, int](t, 42)
		testBinaryRoundTrip[IntBrand, int](t, -1000000)
	})

	t.Run("uint", func(t *testing.T) {
		t.Parallel()
		testBinaryRoundTrip[UintBrand, uint](t, 42)
		testBinaryRoundTrip[UintBrand, uint](t, 0)
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
}

func TestBinaryZeroValueAllTypes(t *testing.T) {
	t.Parallel()

	testBinaryZeroRoundTrip[Int8Brand, int8](t)
	testBinaryZeroRoundTrip[Int16Brand, int16](t)
	testBinaryZeroRoundTrip[IntBrand, int](t)
	testBinaryZeroRoundTrip[UintBrand, uint](t)
	testBinaryZeroRoundTrip[Uint8Brand, uint8](t)
	testBinaryZeroRoundTrip[Uint16Brand, uint16](t)
	testBinaryZeroRoundTrip[Uint32Brand, uint32](t)
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
