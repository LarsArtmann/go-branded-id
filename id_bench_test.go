package id

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"unicode/utf8"
)

const benchIDValue = "test-id-12345"

// Fuzz tests

func FuzzIDJSONString(f *testing.F) {
	testcases := []string{
		"test",
		"hello-world",
		"123",
		"",
		"unicode-日本語", //nolint:gosmopolitan // intentional unicode test data
	}
	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, orig string) {
		if !utf8.ValidString(orig) {
			t.Skip("skipping invalid UTF-8 string")
		}

		id := NewID[StringBrand](orig)

		data, err := json.Marshal(id)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var restored ID[StringBrand, string]

		err = json.Unmarshal(data, &restored)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		assertCmpEqual(t, restored.Get(), orig)
	})
}

func FuzzIDJSONInt64(f *testing.F) {
	for _, tc := range []int64{0, 1, -1, 42, math.MaxInt64, math.MinInt64} {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, orig int64) {
		id := NewID[Int64Brand, int64](orig)

		data, err := json.Marshal(id)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var restored ID[Int64Brand, int64]

		err = json.Unmarshal(data, &restored)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		assertCmpEqual(t, restored.Get(), orig)
	})
}

func FuzzIDBinaryString(f *testing.F) {
	testcases := []string{
		"test",
		"hello-world",
		"123",
		"unicode-日本語", //nolint:gosmopolitan // intentional unicode test data
	}
	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, orig string) {
		if orig == "" {
			return // empty string is zero value
		}

		id := NewID[StringBrand](orig)

		data, err := id.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary failed: %v", err)
		}

		var restored ID[StringBrand, string]

		err = restored.UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary failed: %v", err)
		}

		assertCmpEqual(t, restored.Get(), orig)
	})
}

func FuzzIDBinaryInt64(f *testing.F) {
	for _, tc := range []int64{0, 1, -1, 42, math.MaxInt64, math.MinInt64} {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, orig int64) {
		id := NewID[Int64Brand, int64](orig)

		data, err := id.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary failed: %v", err)
		}

		var restored ID[Int64Brand, int64]

		err = restored.UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary failed: %v", err)
		}

		assertCmpEqual(t, restored.Get(), orig)
	})
}

func FuzzIDBinaryUint64(f *testing.F) {
	testcases := []uint64{0, 1, 42, math.MaxUint64}
	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, orig uint64) {
		id := NewID[Uint64Brand, uint64](orig)

		data, err := id.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary failed: %v", err)
		}

		var restored ID[Uint64Brand, uint64]

		err = restored.UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary failed: %v", err)
		}

		assertCmpEqual(t, restored.Get(), orig)
	})
}

// Benchmarks

func BenchmarkNewID(b *testing.B) {
	for b.Loop() {
		_ = NewID[StringBrand](benchIDValue)
	}
}

func BenchmarkIDGet(b *testing.B) {
	benchmarkIDMethod(b, NewID[StringBrand](benchIDValue), func(id ID[StringBrand, string]) {
		_ = id.Get()
	})
}

func BenchmarkIDString(b *testing.B) {
	benchmarkIDMethod(b, NewID[StringBrand](benchIDValue), func(id ID[StringBrand, string]) {
		_ = id.String()
	})
}

func BenchmarkIDStringInt64(b *testing.B) {
	id := NewID[Int64Brand, int64](123456789)
	for b.Loop() {
		_ = id.String()
	}
}

func BenchmarkIDIsZero(b *testing.B) {
	benchmarkIDMethod(b, NewID[StringBrand](benchIDValue), func(id ID[StringBrand, string]) {
		_ = id.IsZero()
	})
}

// Helper for benchmarking ID methods.
func benchmarkIDMethod[B, V comparable](b *testing.B, id ID[B, V], fn func(ID[B, V])) {
	b.Helper()

	for b.Loop() {
		fn(id)
	}
}

func BenchmarkIDEqual(b *testing.B) {
	id1 := NewID[StringBrand](benchIDValue)

	id2 := NewID[StringBrand](benchIDValue)
	for b.Loop() {
		_ = id1.Equal(id2)
	}
}

func BenchmarkIDCompare(b *testing.B) {
	id1 := NewID[Int64Brand, int64](100)

	id2 := NewID[Int64Brand, int64](200)
	for b.Loop() {
		_, _ = id1.Compare(id2)
	}
}

func BenchmarkIDMarshalJSON(b *testing.B) {
	benchmarkIDMethodWithError(
		b,
		NewID[StringBrand](benchIDValue),
		func(id ID[StringBrand, string]) ([]byte, error) {
			return id.MarshalJSON()
		},
	)
}

func BenchmarkIDMarshalJSONInt64(b *testing.B) {
	benchmarkIDMethodWithError(
		b,
		NewID[Int64Brand, int64](123456789),
		func(id ID[Int64Brand, int64]) ([]byte, error) {
			return id.MarshalJSON()
		},
	)
}

func benchmarkIDMethodWithError[B, V comparable, R any](
	b *testing.B,
	id ID[B, V],
	fn func(ID[B, V]) (R, error),
) {
	b.Helper()

	for b.Loop() {
		_, _ = fn(id)
	}
}

func BenchmarkIDUnmarshalJSON(b *testing.B) {
	benchmarkIDUnmarshalJSON[StringBrand, string](b, []byte(`"test-id-12345"`))
}

func BenchmarkIDUnmarshalJSONInt64(b *testing.B) {
	benchmarkIDUnmarshalJSON[Int64Brand, int64](b, []byte(`123456789`))
}

func benchmarkIDUnmarshalJSON[B, V comparable](b *testing.B, data []byte) {
	b.Helper()

	for b.Loop() {
		var id ID[B, V]

		_ = id.UnmarshalJSON(data)
	}
}

func BenchmarkIDMarshalBinary(b *testing.B) {
	id := NewID[Int64Brand, int64](123456789)
	for b.Loop() {
		_, _ = id.MarshalBinary()
	}
}

func BenchmarkIDUnmarshalBinary(b *testing.B) {
	id := NewID[Int64Brand, int64](123456789)
	data, _ := id.MarshalBinary()

	for b.Loop() {
		var restored ID[Int64Brand, int64]

		_ = restored.UnmarshalBinary(data)
	}
}

func BenchmarkIDScan(b *testing.B) {
	benchmarkIDScan[StringBrand, string](b, benchIDValue)
}

func BenchmarkIDScanInt64(b *testing.B) {
	benchmarkIDScan[Int64Brand, int64](b, int64(123456789))
}

func benchmarkIDScan[B, V comparable](b *testing.B, value V) {
	b.Helper()

	for b.Loop() {
		var id ID[B, V]

		_ = id.Scan(value)
	}
}

func BenchmarkIDValue(b *testing.B) {
	benchmarkIDMethodWithError(
		b,
		NewID[StringBrand](benchIDValue),
		func(id ID[StringBrand, string]) (driver.Value, error) {
			return id.Value()
		},
	)
}

func BenchmarkIDValueInt64(b *testing.B) {
	benchmarkIDMethodWithError(
		b,
		NewID[Int64Brand, int64](123456789),
		func(id ID[Int64Brand, int64]) (driver.Value, error) {
			return id.Value()
		},
	)
}

func BenchmarkJSONRoundTrip(b *testing.B) {
	benchmarkJSONRoundTrip(b, NewID[StringBrand](benchIDValue))
}

func BenchmarkJSONRoundTripInt64(b *testing.B) {
	benchmarkJSONRoundTrip(b, NewID[Int64Brand, int64](123456789))
}

func benchmarkJSONRoundTrip[B, V comparable](b *testing.B, id ID[B, V]) {
	b.Helper()

	for b.Loop() {
		data, _ := json.Marshal(id) //nolint:errchkjson // Benchmark: error check not needed

		var restored ID[B, V]

		_ = json.Unmarshal(data, &restored)
	}
}

// Example functions for godoc

func ExampleNewID() {
	type UserBrand struct{}

	userID := NewID[UserBrand]("user-123")
	fmt.Println(userID.Get())
	// Output: user-123
}

func ExampleID_String() {
	type OrderBrand struct{}

	orderID := NewID[OrderBrand, int64](12345)
	fmt.Println(orderID.String())
	// Output: 12345
}

func ExampleID_Equal() {
	type UserBrand struct{}

	id1 := NewID[UserBrand]("user-123")
	id2 := NewID[UserBrand]("user-123")
	id3 := NewID[UserBrand]("user-456")

	fmt.Println(id1.Equal(id2))
	fmt.Println(id1.Equal(id3))
	// Output:
	// true
	// false
}

func ExampleID_Compare() {
	type OrderBrand struct{}

	id1 := NewID[OrderBrand, int64](100)
	id2 := NewID[OrderBrand, int64](200)
	id3 := NewID[OrderBrand, int64](100)

	cmp, _ := id1.Compare(id2)
	fmt.Println(cmp)
	cmp, _ = id2.Compare(id1)
	fmt.Println(cmp)
	cmp, _ = id1.Compare(id3)
	fmt.Println(cmp)
	// Output:
	// -1
	// 1
	// 0
}

func ExampleID_Or() {
	type UserBrand struct{}

	id := NewID[UserBrand]("user-123")
	defaultID := NewID[UserBrand]("unknown")

	fmt.Println(id.Or(defaultID).Get())

	var zeroID ID[UserBrand, string]
	fmt.Println(zeroID.Or(defaultID).Get())
	// Output:
	// user-123
	// unknown
}

func ExampleID_IsZero() {
	type UserBrand struct{}

	id := NewID[UserBrand]("user-123")

	var zeroID ID[UserBrand, string]

	fmt.Println(id.IsZero())
	fmt.Println(zeroID.IsZero())
	// Output:
	// false
	// true
}

func ExampleID_Reset() {
	type UserBrand struct{}

	id := NewID[UserBrand]("user-123")
	id.Reset()
	fmt.Println(id.IsZero())
	// Output: true
}

func ExampleID_Ptr() {
	type UserBrand struct{}

	id := NewID[UserBrand]("user-123")
	p := id.Ptr()
	fmt.Println(p.Get())
	// Output: user-123
}

func ExampleFromPtr() {
	type UserBrand struct{}

	// With non-nil pointer
	id := NewID[UserBrand]("user-123")
	fmt.Println(FromPtr(id.Ptr()).Get())

	// With nil pointer
	var nilPtr *ID[UserBrand, string]
	fmt.Println(FromPtr(nilPtr).IsZero())
	// Output:
	// user-123
	// true
}
