package id

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"testing"
)

const testIDValue = "test-id"

type (
	StringBrand struct{}
	IntBrand    struct{}
	Int8Brand   struct{}
	Int16Brand  struct{}
	Int32Brand  struct{}
	Int64Brand  struct{}
	UintBrand   struct{}
	Uint8Brand  struct{}
	Uint16Brand struct{}
	Uint32Brand struct{}
	Uint64Brand struct{}
)

func assertIDValue[B any, V comparable](t *testing.T, v, expected V) {
	assertCmpEqual(t, NewID[B](v).Get(), expected)
}

//nolint:cyclop // exhaustive type switch over all brand types
func assertIDValueMatches(t *testing.T, v, expected any) {
	t.Helper()

	switch val := v.(type) {
	case ID[IntBrand, int]:
		if val.Get() != expected.(int) { //nolint:forcetypeassert // guaranteed by test construction
			t.Errorf("expected %v, got %v", expected, val.Get())
		}
	case ID[Int8Brand, int8]:
		if val.Get() != expected.(int8) { //nolint:forcetypeassert // guaranteed by test construction
			t.Errorf("expected %v, got %v", expected, val.Get())
		}
	case ID[Int16Brand, int16]:
		if val.Get() != expected.(int16) { //nolint:forcetypeassert // guaranteed by test construction
			t.Errorf("expected %v, got %v", expected, val.Get())
		}
	case ID[Int32Brand, int32]:
		if val.Get() != expected.(int32) { //nolint:forcetypeassert // guaranteed by test construction
			t.Errorf("expected %v, got %v", expected, val.Get())
		}
	case ID[Int64Brand, int64]:
		if val.Get() != expected.(int64) { //nolint:forcetypeassert // guaranteed by test construction
			t.Errorf("expected %v, got %v", expected, val.Get())
		}
	case ID[UintBrand, uint]:
		if val.Get() != expected.(uint) { //nolint:forcetypeassert // guaranteed by test construction
			t.Errorf("expected %v, got %v", expected, val.Get())
		}
	case ID[Uint8Brand, uint8]:
		if val.Get() != expected.(uint8) { //nolint:forcetypeassert // guaranteed by test construction
			t.Errorf("expected %v, got %v", expected, val.Get())
		}
	case ID[Uint16Brand, uint16]:
		if val.Get() != expected.(uint16) { //nolint:forcetypeassert // guaranteed by test construction
			t.Errorf("expected %v, got %v", expected, val.Get())
		}
	case ID[Uint32Brand, uint32]:
		if val.Get() != expected.(uint32) { //nolint:forcetypeassert // guaranteed by test construction
			t.Errorf("expected %v, got %v", expected, val.Get())
		}
	case ID[Uint64Brand, uint64]:
		if val.Get() != expected.(uint64) { //nolint:forcetypeassert // guaranteed by test construction
			t.Errorf("expected %v, got %v", expected, val.Get())
		}
	case ID[StringBrand, string]:
		if val.Get() != expected.(string) { //nolint:forcetypeassert // guaranteed by test construction
			t.Errorf("expected %v, got %v", expected, val.Get())
		}

		if !val.IsZero() {
			t.Error("empty string should be zero")
		}
	}
}

func TestNewID(t *testing.T) {
	t.Parallel()

	id := NewID[StringBrand]("user-123")
	if id.Get() != "user-123" {
		t.Errorf("expected user-123, got %s", id.Get())
	}

	if id.IsZero() {
		t.Error("expected non-zero id")
	}
}

func TestNewIDNumeric(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		brand    any
		value    any
		expected any
	}{
		{"int64", Int64Brand{}, int64(42), int64(42)},
		{"int32", Int32Brand{}, int32(42), int32(42)},
		{"uint64", Uint64Brand{}, uint64(42), uint64(42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch v := tt.value.(type) {
			case int64:
				assertIDValue[Int64Brand](
					t,
					v,
					tt.expected.(int64), //nolint:forcetypeassert // guaranteed by type switch
				)
			case int32:
				assertIDValue[Int32Brand](
					t,
					v,
					tt.expected.(int32), //nolint:forcetypeassert // guaranteed by type switch
				)
			case uint64:
				assertIDValue[Uint64Brand](
					t,
					v,
					tt.expected.(uint64), //nolint:forcetypeassert // guaranteed by type switch
				)
			}
		})
	}
}

func TestIDIsZero(t *testing.T) {
	t.Parallel()

	var zeroID ID[StringBrand, string]
	if !zeroID.IsZero() {
		t.Error("expected zero ID to be zero")
	}

	nonZeroID := NewID[StringBrand]("test")
	if nonZeroID.IsZero() {
		t.Error("expected non-zero ID to not be zero")
	}
}

func TestIDReset(t *testing.T) {
	t.Parallel()

	id := NewID[StringBrand]("test")
	id.Reset()

	if !id.IsZero() {
		t.Error("expected zero ID after Reset")
	}
}

func TestIDEqual(t *testing.T) {
	t.Parallel()

	id1 := NewID[StringBrand]("test")
	id2 := NewID[StringBrand]("test")
	id3 := NewID[StringBrand]("other")

	if !id1.Equal(id2) {
		t.Error("expected equal IDs")
	}

	if id1.Equal(id3) {
		t.Error("expected unequal IDs")
	}
}

func testIDCompareGeneric[B any, V comparable](
	t *testing.T,
	createID func(V) ID[B, V],
	tests []struct {
		name     string
		a, b     V
		expected int
	},
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			idA := createID(tt.a)
			idB := createID(tt.b)

			result, err := idA.Compare(idB)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestIDCompare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"less", 1, 2, -1},
		{"equal", 5, 5, 0},
		{"greater", 3, 1, 1},
	}

	testIDCompareGeneric(
		t,
		func(v int) ID[Int64Brand, int] { return NewID[Int64Brand, int](v) },
		tests,
	)
}

func TestIDCompareString(t *testing.T) {
	t.Parallel()

	idA := NewID[StringBrand]("a")
	idB := NewID[StringBrand]("b")
	idC := NewID[StringBrand]("a")

	cmp, err := idA.Compare(idB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmp != -1 {
		t.Error("expected 'a' < 'b'")
	}

	cmp, err = idA.Compare(idC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmp != 0 {
		t.Error("expected 'a' == 'a'")
	}

	cmp, err = idB.Compare(idA)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmp != 1 {
		t.Error("expected 'b' > 'a'")
	}
}

func TestIDCompareInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		a, b     int64
		expected int
	}{
		{"less", 100, 200, -1},
		{"equal", 100, 100, 0},
		{"greater", 200, 100, 1},
	}

	testIDCompareGeneric(
		t,
		func(v int64) ID[Int64Brand, int64] { return NewID[Int64Brand, int64](v) },
		tests,
	)
}

func TestIDCompareUint64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		a, b     uint64
		expected int
	}{
		{"less", 100, 200, -1},
		{"equal", 100, 100, 0},
		{"greater", 200, 100, 1},
	}

	testIDCompareGeneric(
		t,
		func(v uint64) ID[Uint64Brand, uint64] { return NewID[Uint64Brand, uint64](v) },
		tests,
	)
}

func TestIDCompareFloat64ReturnsErrNotOrdered(t *testing.T) {
	t.Parallel()

	type FloatBrand struct{}

	id := NewID[FloatBrand, float64](1.5)
	_, err := id.Compare(NewID[FloatBrand, float64](2.5))

	if err == nil {
		t.Fatal("expected error for float64 Compare")
	}

	if !errors.Is(err, ErrNotOrdered) {
		t.Errorf("expected ErrNotOrdered, got %v", err)
	}
}

func TestIDOr(t *testing.T) {
	t.Parallel()
	t.Run("non-zero returns self", func(t *testing.T) {
		t.Parallel()

		id := NewID[StringBrand]("test")
		defaultID := NewID[StringBrand]("default")

		assertCmpEqual(t, id.Or(defaultID).Get(), "test")
	})

	t.Run("zero returns default", func(t *testing.T) {
		t.Parallel()

		var id ID[StringBrand, string]

		defaultID := NewID[StringBrand]("default")

		assertCmpEqual(t, id.Or(defaultID).Get(), "default")
	})
}

func TestIDString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       any
		expected string
	}{
		{"string", NewID[StringBrand](testIDValue), testIDValue},
		{"int64", NewID[Int64Brand, int64](42), "42"},
		{"int32", NewID[Int32Brand, int32](42), "42"},
		{"uint64", NewID[Uint64Brand, uint64](42), "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got string

			switch v := tt.id.(type) {
			case ID[StringBrand, string]:
				got = v.String()
			case ID[Int64Brand, int64]:
				got = v.String()
			case ID[Int32Brand, int32]:
				got = v.String()
			case ID[Uint64Brand, uint64]:
				got = v.String()
			}

			assertCmpEqual(t, got, tt.expected)
		})
	}
}

func TestIDGoString(t *testing.T) {
	t.Parallel()

	id := NewID[StringBrand](testIDValue)
	if id.GoString() != testIDValue {
		t.Errorf("expected test-id, got %s", id.GoString())
	}
}

func TestIDFormat(t *testing.T) {
	t.Parallel()

	id := NewID[Int64Brand, int64](42)

	tests := []struct {
		format   string
		expected string
	}{
		{"%s", "42"},
		{"%d", "42"},
		{"%q", `"42"`},
		{"%v", "42"},
		{"%#v", "id(42)"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			t.Parallel()

			assertCmpEqual(t, fmt.Sprintf(tt.format, id), tt.expected)
		})
	}
}

func TestIDTypeSafety(t *testing.T) {
	t.Parallel()

	type UserBrand struct{}

	type OrderBrand struct{}

	userID := NewID[UserBrand]("user-123")
	orderID := NewID[OrderBrand]("order-456")

	_ = userID.Get()
	_ = orderID.Get()
}

func TestIDSorting(t *testing.T) {
	t.Parallel()

	ids := []ID[Int64Brand, int64]{
		NewID[Int64Brand, int64](3),
		NewID[Int64Brand, int64](1),
		NewID[Int64Brand, int64](2),
	}

	sort.Slice(ids, func(i, j int) bool {
		cmp, err := ids[i].Compare(ids[j])
		if err != nil {
			panic(err)
		}

		return cmp < 0
	})

	expected := []int64{1, 2, 3}
	for i, id := range ids {
		if id.Get() != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], id.Get())
		}
	}
}

type edgeCaseTest struct {
	name     string
	brand    func(v any) any
	value    any
	expected any
}

func edgeCase(name string, brandFunc func(v any) any, value any) edgeCaseTest {
	return edgeCaseTest{name: name, brand: brandFunc, value: value, expected: value}
}

func TestIDEdgeCases(t *testing.T) {
	t.Parallel()

	int64Brand := func(v any) any { return NewID[Int64Brand](v.(int64)) } //nolint:forcetypeassert // test construction

	tests := []edgeCaseTest{
		edgeCase("max int64", int64Brand, int64(math.MaxInt64)),
		edgeCase("min int64", int64Brand, int64(math.MinInt64)),
		edgeCase(
			"max uint64",
			func(v any) any { return NewID[Uint64Brand](v.(uint64)) }, //nolint:forcetypeassert // test construction
			uint64(math.MaxUint64),
		),
		{
			"empty string",
			func(v any) any { return NewID[StringBrand](v.(string)) }, //nolint:forcetypeassert // test construction
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id := tt.brand(tt.value)
			assertIDValueMatches(t, id, tt.expected)
		})
	}
}

type roundTripTest interface {
	TestString(t *testing.T)
	TestInt64(t *testing.T)
	TestInt32(t *testing.T)
	TestUint64(t *testing.T)
}

func testIDAllTypesRoundTrip(t *testing.T, rt roundTripTest) {
	t.Helper()

	t.Run("string ID", rt.TestString)
	t.Run("int64 ID", rt.TestInt64)
	t.Run("int32 ID", rt.TestInt32)
	t.Run("uint64 ID", rt.TestUint64)
}

func TestPtr(t *testing.T) {
	t.Parallel()

	t.Run("Ptr returns non-nil pointer", func(t *testing.T) {
		t.Parallel()

		id := NewID[StringBrand]("user-123")
		p := id.Ptr()

		if p == nil {
			t.Fatal("expected non-nil pointer")
		}

		if p.Get() != "user-123" {
			t.Errorf("expected user-123, got %s", p.Get())
		}
	})

	t.Run("Ptr of zero value is non-nil", func(t *testing.T) {
		t.Parallel()

		var id ID[StringBrand, string]
		p := id.Ptr()

		if p == nil {
			t.Fatal("expected non-nil pointer for zero ID")
		}

		if !p.IsZero() {
			t.Error("expected zero ID through pointer")
		}
	})
}

func TestFromPtr(t *testing.T) {
	t.Parallel()

	t.Run("FromPtr with nil returns zero", func(t *testing.T) {
		t.Parallel()

		var p *ID[StringBrand, string]
		id := FromPtr(p)

		if !id.IsZero() {
			t.Error("expected zero ID from nil pointer")
		}
	})

	t.Run("FromPtr with non-nil returns value", func(t *testing.T) {
		t.Parallel()

		id := NewID[StringBrand]("user-123")
		id2 := FromPtr(id.Ptr())

		if !id.Equal(id2) {
			t.Errorf("expected equal IDs, got %v", id2.Get())
		}
	})
}
