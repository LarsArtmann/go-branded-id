package id

import (
	"encoding"
	"errors"
	"testing"
)

// Test types for sentinel error tests. Defined at package level so methods can
// be attached.

type (
	sentinelUnsupportedBrand struct{}
	sentinelUnsupportedType  struct{ X int } // comparable, no serialization interface
)

type sentinelMarshalBrand struct{}

// sentinelFailingBinary implements BinaryMarshaler but always returns an error.
type sentinelFailingBinary struct{ X int }

func (sentinelFailingBinary) MarshalBinary() ([]byte, error) {
	return nil, errors.New("intentional marshal failure")
}

// Compile-time interface assertion.
var _ encoding.BinaryMarshaler = sentinelFailingBinary{}

func assertErrorIs(tb testing.TB, err error, sentinel error) {
	tb.Helper()

	if err == nil {
		tb.Fatalf("expected error matching %v, got nil", sentinel)
	}

	if !errors.Is(err, sentinel) {
		tb.Errorf("expected errors.Is(err, %v), got: %v", sentinel, err)
	}
}

func TestSentinelErrInvalidID(t *testing.T) {
	t.Parallel()

	t.Run("ValidateID zero", func(t *testing.T) {
		t.Parallel()

		var zero ID[StringBrand, string]
		assertErrorIs(t, ValidateID(zero), ErrInvalidID)
	})

	t.Run("ValidateIDWithValue zero", func(t *testing.T) {
		t.Parallel()

		var zero ID[StringBrand, string]
		assertErrorIs(t, ValidateIDWithValue[StringBrand, string](zero, nil), ErrInvalidID)
	})
}

func TestSentinelErrNotOrdered(t *testing.T) {
	t.Parallel()

	type FloatBrand struct{}

	t.Run("Compare float64", func(t *testing.T) {
		t.Parallel()

		id := NewID[FloatBrand, float64](1.5)
		_, err := id.Compare(NewID[FloatBrand, float64](2.5))
		assertErrorIs(t, err, ErrNotOrdered)
	})
}

func TestSentinelErrUnsupportedType(t *testing.T) {
	t.Parallel()

	t.Run("binary marshal", func(t *testing.T) {
		t.Parallel()

		id := NewID[sentinelUnsupportedBrand, sentinelUnsupportedType](
			sentinelUnsupportedType{X: 1},
		)
		_, err := id.MarshalBinary()
		assertErrorIs(t, err, ErrUnsupportedType)
	})

	t.Run("binary unmarshal", func(t *testing.T) {
		t.Parallel()

		var id ID[sentinelUnsupportedBrand, sentinelUnsupportedType]
		err := id.UnmarshalBinary([]byte{1, 2, 3})
		assertErrorIs(t, err, ErrUnsupportedType)
	})

	t.Run("text unmarshal", func(t *testing.T) {
		t.Parallel()

		var id ID[sentinelUnsupportedBrand, sentinelUnsupportedType]
		err := id.UnmarshalText([]byte("data"))
		assertErrorIs(t, err, ErrUnsupportedType)
	})

	t.Run("SQL value", func(t *testing.T) {
		t.Parallel()

		id := NewID[sentinelUnsupportedBrand, sentinelUnsupportedType](
			sentinelUnsupportedType{X: 1},
		)
		_, err := id.Value()
		assertErrorIs(t, err, ErrUnsupportedType)
	})
}

func TestSentinelErrCannotScan(t *testing.T) {
	t.Parallel()

	t.Run("string ID scan int", func(t *testing.T) {
		t.Parallel()

		var id ID[StringBrand, string]
		err := id.Scan(42)
		assertErrorIs(t, err, ErrCannotScan)
	})

	t.Run("int64 ID scan string", func(t *testing.T) {
		t.Parallel()

		var id ID[Int64Brand, int64]
		err := id.Scan("not-a-number")
		assertErrorIs(t, err, ErrCannotScan)
	})
}

func TestSentinelErrInsufficientData(t *testing.T) {
	t.Parallel()

	t.Run("int64 needs 8 bytes", func(t *testing.T) {
		t.Parallel()

		var id ID[Int64Brand, int64]
		err := id.UnmarshalBinary([]byte{1, 2, 3})
		assertErrorIs(t, err, ErrInsufficientData)
	})

	t.Run("int32 needs 4 bytes", func(t *testing.T) {
		t.Parallel()

		var id ID[Int32Brand, int32]
		err := id.UnmarshalBinary([]byte{1})
		assertErrorIs(t, err, ErrInsufficientData)
	})
}

func TestSentinelErrNilReceiver(t *testing.T) {
	t.Parallel()

	t.Run("Scan on nil pointer", func(t *testing.T) {
		t.Parallel()

		var id *ID[StringBrand, string]
		err := id.Scan("test")
		assertErrorIs(t, err, ErrNilReceiver)
	})
}

func TestSentinelErrMarshal(t *testing.T) {
	t.Parallel()

	t.Run("binary marshal delegate failure", func(t *testing.T) {
		t.Parallel()

		id := NewID[sentinelMarshalBrand, sentinelFailingBinary](sentinelFailingBinary{X: 1})
		_, err := id.MarshalBinary()
		assertErrorIs(t, err, ErrMarshal)
	})
}

func TestSentinelErrUnmarshal(t *testing.T) {
	t.Parallel()

	t.Run("text unmarshal invalid int64", func(t *testing.T) {
		t.Parallel()

		var id ID[Int64Brand, int64]
		err := id.UnmarshalText([]byte("not-a-number"))
		assertErrorIs(t, err, ErrUnmarshal)
	})

	t.Run("text unmarshal invalid uint64", func(t *testing.T) {
		t.Parallel()

		var id ID[Uint64Brand, uint64]
		err := id.UnmarshalText([]byte("not-a-number"))
		assertErrorIs(t, err, ErrUnmarshal)
	})

	t.Run("json unmarshal invalid", func(t *testing.T) {
		t.Parallel()

		var id ID[Int64Brand, int64]
		err := id.UnmarshalJSON([]byte("not-json"))
		assertErrorIs(t, err, ErrUnmarshal)
	})
}

func TestSentinelErrInternal_Defensive(t *testing.T) {
	t.Parallel()

	// ErrInternal guards type assertions that the outer type switch guarantees
	// will succeed (e.g. any(string(data)).(V) after case string). These code
	// paths are unreachable without runtime reflection hacking. We verify the
	// sentinel is defined and has a stable message so consumers can rely on it.
	if ErrInternal == nil {
		t.Fatal("ErrInternal sentinel must not be nil")
	}

	if ErrInternal.Error() != "id: internal error" {
		t.Errorf("unexpected ErrInternal message: %q", ErrInternal.Error())
	}
}
