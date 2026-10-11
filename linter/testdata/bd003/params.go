// Package testdata holds parse-only BD003 fixtures. This file exercises
// function parameters: primitive-typed ID params are flagged, non-ID names
// and non-primitive types are not, and String/Error method declarations are
// skipped (interface contracts must stay primitive).
package testdata

// flagged: param userId string -> BD003.
func getUser(userId string) {}

// flagged: standalone id -> BD003 with the domain derived from the func name.
func findInvoice(id string) {}

// flagged: all-caps acronym param.
func lookupPID(PID string) {}

// not flagged: isValid is not an ID name.
func check(isValid bool) {}

// not flagged: non-primitive type (qualified branded ID).
func byBranded(userId UserID) {}

// not flagged: interface methods must stay primitive.
func (r repo) String(id string) string { return id }

// not flagged: interface error message.
func (r repo) Error(id string) error { return nil }

// repo is the receiver for the interface-method fixtures above.
type repo struct{}

// UserID is a named type, not primitive.
type UserID struct{ inner string }
