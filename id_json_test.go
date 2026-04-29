package id

import (
	"encoding/json"
	"fmt"
	"testing"
)

// assertUnmarshalError verifies that JSON unmarshaling fails with an error.
func assertUnmarshalError[B any, V comparable](t *testing.T, data string) {
	t.Helper()

	var id ID[B, V]

	err := json.Unmarshal([]byte(data), &id)
	if err == nil {
		t.Error("expected error for unmarshaling")
	}
}

func testMarshalZeroID[B any, V comparable](t *testing.T) {
	t.Helper()

	var id ID[B, V]

	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	if string(data) != "null" {
		t.Errorf("expected null, got %s", string(data))
	}
}

func testMarshalNonZeroID[B any, V comparable](t *testing.T, value V, format string) {
	t.Helper()

	id := NewID[B, V](value)

	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := fmt.Sprintf(format, value)
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func testUnmarshalNonZeroID[B any, V comparable](t *testing.T, jsonData string, expected V) {
	t.Helper()

	var id ID[B, V]

	err := json.Unmarshal([]byte(jsonData), &id)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if id.Get() != expected {
		t.Errorf("expected %v, got %v", expected, id.Get())
	}
}

func TestIDJSON(t *testing.T) {
	t.Parallel()
	t.Run("string ID non-zero", func(t *testing.T) {
		t.Parallel()
		testMarshalNonZeroID[StringBrand, string](t, "abc123", "%q")
	})

	t.Run("string ID zero", func(t *testing.T) {
		t.Parallel()
		testMarshalZeroID[StringBrand, string](t)
	})

	t.Run("int64 ID non-zero", func(t *testing.T) {
		t.Parallel()
		testMarshalNonZeroID[Int64Brand, int64](t, 42, "%v")
	})

	t.Run("int64 ID zero", func(t *testing.T) {
		t.Parallel()
		testMarshalZeroID[Int64Brand, int64](t)
	})

	t.Run("int32 ID non-zero", func(t *testing.T) {
		t.Parallel()
		testMarshalNonZeroID[Int32Brand, int32](t, 42, "%v")
	})

	t.Run("uint64 ID non-zero", func(t *testing.T) {
		t.Parallel()
		testMarshalNonZeroID[Uint64Brand, uint64](t, 42, "%v")
	})
}

func TestIDUnmarshalJSON(t *testing.T) {
	t.Parallel()
	t.Run("string ID from string", func(t *testing.T) {
		t.Parallel()

		var id ID[StringBrand, string]

		err := json.Unmarshal([]byte(`"test-id"`), &id)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		if id.Get() != "test-id" {
			t.Errorf("expected test-id, got %s", id.Get())
		}
	})

	t.Run("string ID from null", func(t *testing.T) {
		t.Parallel()

		var id ID[StringBrand, string]

		err := json.Unmarshal([]byte("null"), &id)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		if !id.IsZero() {
			t.Error("expected zero ID")
		}
	})

	t.Run("numeric IDs", func(t *testing.T) {
		testIDAllTypesUnmarshalJSON(t, jsonUnmarshalTestImpl{})
	})

	t.Run("invalid inputs", func(t *testing.T) {
		t.Parallel()

		stringInvalidTests := []struct {
			name  string
			input string
		}{
			{"invalid JSON", `invalid`},
			{"number into string ID", "123"},
		}
		for _, tc := range stringInvalidTests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				assertUnmarshalError[StringBrand, string](t, tc.input)
			})
		}

		int64InvalidTests := []struct {
			name  string
			input string
		}{
			{"string into int64 ID", `"not-a-number"`},
		}
		for _, tc := range int64InvalidTests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				assertUnmarshalError[Int64Brand, int64](t, tc.input)
			})
		}
	})
}

type jsonUnmarshalTestImpl struct{}

func runTest(t *testing.T, testFunc func(tx *testing.T)) {
	t.Run("", func(tx *testing.T) {
		tx.Parallel()
		testFunc(tx)
	})
}

func (j jsonUnmarshalTestImpl) TestInt64(t *testing.T) {
	runTest(t, func(tx *testing.T) {
		testUnmarshalNonZeroID[Int64Brand, int64](tx, "42", 42)
	})
}

func (j jsonUnmarshalTestImpl) TestInt32(t *testing.T) {
	runTest(t, func(tx *testing.T) {
		testUnmarshalNonZeroID[Int32Brand, int32](tx, "42", 42)
	})
}

func (j jsonUnmarshalTestImpl) TestUint64(t *testing.T) {
	runTest(t, func(tx *testing.T) {
		testUnmarshalNonZeroID[Uint64Brand, uint64](tx, "42", 42)
	})
}

type jsonUnmarshalTest interface {
	TestInt64(t *testing.T)
	TestInt32(t *testing.T)
	TestUint64(t *testing.T)
}

func testIDAllTypesUnmarshalJSON(t *testing.T, ut jsonUnmarshalTest) {
	t.Parallel()
	t.Run("int64 ID", ut.TestInt64)
	t.Run("int32 ID", ut.TestInt32)
	t.Run("uint64 ID", ut.TestUint64)
}

func testJSONRoundTrip[B any, V comparable](t *testing.T, value V) {
	t.Helper()

	original := NewID[B, V](value)

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var restored ID[B, V]

	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	assertCmpEqual(t, original.Get(), restored.Get())
}

func TestIDJSONRoundTrip(t *testing.T) {
	testIDAllTypesRoundTrip(t, jsonRoundTripTest{})
}

type jsonRoundTripTest struct{}

func (j jsonRoundTripTest) TestString(t *testing.T) {
	t.Parallel()
	testJSONRoundTrip[StringBrand, string](t, "test-id")
}

func (j jsonRoundTripTest) TestInt64(t *testing.T) {
	t.Parallel()
	testJSONRoundTrip[Int64Brand, int64](t, 42)
}

func (j jsonRoundTripTest) TestInt32(t *testing.T) {
	t.Parallel()
	testJSONRoundTrip[Int32Brand, int32](t, 42)
}

func (j jsonRoundTripTest) TestUint64(t *testing.T) {
	t.Parallel()
	testJSONRoundTrip[Uint64Brand, uint64](t, 42)
}
