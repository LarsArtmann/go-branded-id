package id

import (
	"errors"
	"fmt"
)

// BrandNamer is an optional interface that brand types can implement
// to provide a human-readable name for debugging and logging.
type BrandNamer interface {
	Name() string
}

// ErrInvalidID is returned when ID validation fails.
var ErrInvalidID = errors.New("id: invalid")

// brandName returns the brand name if it implements BrandNamer.
// Returns (name, true) for named brands, ("", false) for unnamed brands.
func brandName[B any]() (string, bool) {
	var brand B

	namer, ok := any(brand).(BrandNamer)
	if !ok {
		return "", false
	}

	return namer.Name(), true
}

// BrandName returns the name of the brand if it implements BrandNamer,
// otherwise returns the type name of the brand.
func BrandName[B any]() string {
	if name, ok := brandName[B](); ok {
		return name
	}

	var brand B

	return fmt.Sprintf("%T", brand) // e.g., "main.UserBrand"
}

// ValidateID validates that the given ID is not zero.
// Returns ErrInvalidID if the ID is zero.
//
// Example:
//
//	type UserBrand struct{}
//	func (UserBrand) Name() string { return "User" }
//
//	func main() {
//	    userID := NewID[UserBrand]("user-123")
//	    if err := ValidateID(userID); err != nil {
//	        log.Fatal(err)
//	    }
//	}
func ValidateID[B any, V comparable](id ID[B, V]) error {
	if id.IsZero() {
		return fmt.Errorf("%w: %s: empty", ErrInvalidID, BrandName[B]())
	}

	return nil
}

// ValidateIDWithValue validates that the ID is not zero and optionally
// validates the value using the provided validator function.
func ValidateIDWithValue[B any, V comparable](
	id ID[B, V],
	validate func(V) error,
) error {
	if id.IsZero() {
		return fmt.Errorf("%w: %s: empty", ErrInvalidID, BrandName[B]())
	}

	if validate != nil {
		if err := validate(id.value); err != nil {
			return fmt.Errorf("%w: %s: %w", ErrInvalidID, BrandName[B](), err)
		}
	}

	return nil
}

// MustValidateID is like ValidateID but panics if the ID is zero.
// Use for init-time validation or when a zero ID is a programming error.
func MustValidateID[B any, V comparable](id ID[B, V]) {
	if err := ValidateID[B, V](id); err != nil {
		panic(err)
	}
}
