// Package linter implements brandid-lint: a static analyzer that finds brand
// types used with id.ID[Brand, Value] that are missing their Name() string
// method (rule BD001).
//
// A brand type is identified by two signals:
//   - It is declared as an empty struct.
//   - It is used as the first type argument of id.ID[...] (qualified or
//     same-package).
//
// Both signals must hold, so empty structs that never appear in an id.ID
// instantiation are never flagged. Findings are emitted as go-finding
// Finding values positioned at the type declaration, where the fix lands.
//
// # Suppression (rule BD002)
//
// A brand that deliberately skips Name() (for example when String() output
// is a storage key) is suppressed in source, right where the decision lives:
//
//	//brandid-lint:ignore(BD001) event stream marker: String() is the stream name
//	type StreamMarker struct{}
//
// The directive names the rule (only BD001 is suppressible) and requires a
// reason. It may sit on the declaration line or as the last line of the
// comment group directly above it. Suppressed brands produce no finding and
// are never repaired. A directive that cannot do its job — malformed syntax,
// missing reason, unknown rule, stale (the brand has Name() by now),
// duplicate, or annotating no declaration — is itself reported as a BD002
// finding, because a silent no-op directive is worse than none.
//
// # Repair
//
// Repair (and the CLI's -fix flag) inserts the suggested Name() stub after
// the type declaration of every unsuppressed BD001 finding, as byte-level
// insertion edits applied back to front so nothing shifts. It is idempotent
// (a second run finds nothing to insert), preserves file permissions, and
// never touches suppressed brands.
//
// Detection is purely syntactic (go/parser): files that fail to parse are
// skipped, and the scanner never resolves imports across packages. The
// directories .git, vendor, and node_modules are never scanned.
package linter
