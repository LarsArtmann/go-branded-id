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
// Detection is purely syntactic (go/parser): files that fail to parse are
// skipped, and the scanner never resolves imports across packages. The
// directories .git, vendor, and node_modules are never scanned.
package linter
