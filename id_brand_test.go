package id

import (
	"errors"
	"fmt"
	"testing"
)

type testUserBrand struct{}

func (testUserBrand) Name() string { return "User" }

type testIntBrand struct{}

func (testIntBrand) Name() string { return "Order" }

const testErrInvalidEmpty = "id: invalid: User: empty"

func TestBrandName_WithNameMethod(t *testing.T) {
	t.Parallel()

	name := BrandName[testUserBrand]()
	if name != "User" {
		t.Errorf("expected 'User', got %q", name)
	}
}

func TestBrandName_WithoutNameMethod(t *testing.T) {
	t.Parallel()

	type orderBrand struct{}

	name := BrandName[orderBrand]()
	if name == "" {
		t.Error("expected non-empty type name")
	}
}

func TestString_BrandAware(t *testing.T) {
	t.Parallel()

	userID := NewID[testUserBrand]("abc123")
	if got := userID.String(); got != "User:abc123" {
		t.Errorf("expected 'User:abc123', got %q", got)
	}
}

func TestString_BrandAwareInt(t *testing.T) {
	t.Parallel()

	orderID := NewID[testIntBrand, int64](42)
	if got := orderID.String(); got != "Order:42" {
		t.Errorf("expected 'Order:42', got %q", got)
	}
}

func TestString_NoBrandName(t *testing.T) {
	t.Parallel()

	type anonBrand struct{}

	id := NewID[anonBrand]("value")
	if got := id.String(); got != "value" {
		t.Errorf("expected 'value', got %q", got)
	}
}

func TestGoString_BrandAware(t *testing.T) {
	t.Parallel()

	userID := NewID[testUserBrand]("abc123")
	if got := userID.GoString(); got != `id.User(abc123)` {
		t.Errorf("expected 'id.User(abc123)', got %q", got)
	}
}

func TestGoString_NoBrandName(t *testing.T) {
	t.Parallel()

	type anonBrand struct{}

	id := NewID[anonBrand]("value")
	if got := id.GoString(); got == "value" {
		t.Errorf("expected debug format, got plain value %q", got)
	}
}

func TestFormat_HashV_NamedBrand(t *testing.T) {
	t.Parallel()

	userID := NewID[testUserBrand, int64](42)

	got := fmt.Sprintf("%#v", userID)
	if got != "id.User(42)" {
		t.Errorf("expected 'id.User(42)', got %q", got)
	}
}

func TestFormat_HashV_UnnamedBrand(t *testing.T) {
	t.Parallel()

	id := NewID[Int64Brand, int64](42)

	got := fmt.Sprintf("%#v", id)
	if got != "id.id.Int64Brand(42)" {
		t.Errorf("expected 'id.id.Int64Brand(42)', got %q", got)
	}
}

func TestValidateID_Valid(t *testing.T) {
	t.Parallel()

	userID := NewID[testUserBrand]("user-123")

	err := ValidateID(userID)
	if err != nil {
		t.Errorf("expected no error for valid ID, got %v", err)
	}
}

func TestValidateID_Zero(t *testing.T) {
	t.Parallel()

	var zeroUserID ID[testUserBrand, string]

	err := ValidateID(zeroUserID)
	if err == nil {
		t.Error("expected error for zero ID")
	}

	expected := testErrInvalidEmpty
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestValidateID_NoNameMethod(t *testing.T) {
	t.Parallel()

	type orderBrand struct{}

	var zeroOrderID ID[orderBrand, string]

	err := ValidateID(zeroOrderID)
	if err == nil {
		t.Error("expected error for zero ID")
	}

	if err.Error() == "id: invalid:  ID: empty" {
		t.Error("error message should contain type name")
	}
}

func TestValidateIDWithValue_ValidValue(t *testing.T) {
	t.Parallel()

	userID := NewID[testUserBrand]("user-123")

	err := ValidateIDWithValue(userID, func(v string) error {
		if len(v) < 3 {
			return errors.New("too short")
		}

		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateIDWithValue_InvalidValue(t *testing.T) {
	t.Parallel()

	userID := NewID[testUserBrand]("ab")

	err := ValidateIDWithValue(userID, func(v string) error {
		if len(v) < 3 {
			return errors.New("too short")
		}

		return nil
	})
	if err == nil {
		t.Error("expected error for invalid value")
	}

	expected := "id: invalid: User: too short"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestValidateIDWithValue_NilValidator(t *testing.T) {
	t.Parallel()

	userID := NewID[testUserBrand]("user-123")

	err := ValidateIDWithValue(userID, nil)
	if err != nil {
		t.Errorf("expected no error with nil validator, got %v", err)
	}
}

func TestValidateIDWithValue_ZeroID(t *testing.T) {
	t.Parallel()

	var zeroUserID ID[testUserBrand, string]

	err := ValidateIDWithValue(zeroUserID, func(_ string) error {
		return errors.New("should not be called")
	})
	if err == nil {
		t.Error("expected error for zero ID")
	}

	expected := testErrInvalidEmpty
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func ExampleBrandName() {
	fmt.Println(BrandName[testUserBrand]())
	// Output: User
}

func ExampleID_String_named() {
	userID := NewID[testUserBrand]("abc123")
	fmt.Println(userID)
	// Output: User:abc123
}

func ExampleID_String_unnamed() {
	type AnonBrand struct{}

	id := NewID[AnonBrand]("abc123")
	fmt.Println(id)
	// Output: abc123
}

func ExampleValidateID() {
	userID := NewID[testUserBrand]("user-123")
	if err := ValidateID(userID); err != nil {
		fmt.Println(err)
	}
}

func ExampleValidateID_zero() {
	var emptyUserID ID[testUserBrand, string]
	if err := ValidateID(emptyUserID); err != nil {
		fmt.Println(err)
	}
	// Output: id: invalid: User: empty
}

func TestMustValidateID_Valid(t *testing.T) {
	t.Parallel()
	MustValidateID(NewID[testUserBrand]("user-123")) // should not panic
}

func TestMustValidateID_Zero(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for zero ID")
		}

		err, ok := r.(error)
		if !ok {
			t.Fatalf("expected error, got %T", r)
		}

		expected := testErrInvalidEmpty
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	}()

	var zero ID[testUserBrand, string]
	MustValidateID(zero)
}
