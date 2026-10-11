// Package testdata exercises struct fields: primitive-typed ID fields are flagged
// (including every name of a multi-name field and fields of function-local
// struct types), embedded structs are not name-matched, and non-ID field
// names stay silent.
package testdata

// Order carries flagged and unflagged fields.
type Order struct {
	OrderID   string // flagged: snake-free camel ID
	userID    int64  // flagged: private fields count too
	Note      string // not flagged: no ID name
	Threshold int    // not flagged
}

// MultiName shows one declaration, two names: both are flagged.
type MultiName struct {
	ownerID, tenantID string
}

// LocalType carries a function-local struct with an ID field.
func LocalType() {
	type inner struct {
		sessionTokenID uint32
	}

	_ = inner{}
}

// Embedded embeds a type; the embedded name is not an ID field name.
type Embedded struct {
	Order
}
