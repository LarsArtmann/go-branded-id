// Package testdata exercises the ignore(BD003) suppression contract: a directive on
// the declaration line, directly above it, or attached to the enclosing
// declaration suppresses; a directive that suppresses nothing is a BD002
// finding.
package testdata

// suppressed: trailing directive on the param line.
func keepTrailing(userId string) {} //brandid-lint:ignore(BD003) display layer passthrough

// suppressed: directive as the last line of the comment group above the line.
func keepAbove(
	//brandid-lint:ignore(BD003) external API contract, name is fixed
	userId string,
) {
}

// suppressed: doc directive covers every ID param and named return of the func.
// brandid-lint:ignore(BD003) legacy wire format, whole function pinned
func keepWhole(orderId string) (customerId string) { return }

// suppressed: trailing directive on the struct field's declaration line.
type KeptFields struct {
	//brandid-lint:ignore(BD003) protobuf-generated, field names are frozen
	productId string
}

// brandid-lint:ignore(BD003) suppresses nothing here
func noIDNames(count int) {}
