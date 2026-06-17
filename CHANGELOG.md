# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.3.1] - 2026-06-17

### Changed

- **`Format()` and `GoString()` allocate significantly less**: `Format()` now writes brand name and value directly to `fmt.State` via `io.WriteString` instead of building intermediate strings; `GoString()` uses concatenation instead of `fmt.Sprintf`. The new internal `writeTo(io.Writer)` helper powers both paths and bypasses string concatenation entirely. Benchmarks for `fmt.Sprintf("%s", id)` (the common hot path):
  - String, no brand: 3 allocs / 48B → 2 allocs / 32B (−33%)
  - String, named brand: 4 allocs / 80B → 2 allocs / 40B (−50%)
  - Int64, named brand: 4 allocs / 80B → 3 allocs / 40B (−25%)
  - Direct `String()` calls are unaffected: 0 allocs for unbranded string values, 1 alloc for named brands (inevitable concatenation)

### Fixed

- `nix flake check` build check failed in Nix sandbox: Go build cache could not initialize at the default `$HOME/.cache/go-build` path (`mkdir /homeless-shelter: permission denied`). Now sets `GOCACHE` to a writable `$TMPDIR` subdirectory.

## [0.3.0] - 2026-05-20

### Added

- `BrandNamer` interface for brand types to provide human-readable names
- `BrandName[B]()` function — returns brand name (or type name fallback) for logging and introspection
- `ValidateID[B, V]()` — validates ID is not zero, returns brand-aware error messages
- `ValidateIDWithValue[B, V]()` — validates ID and optionally validates the value with a custom function
- `MustValidateID[B, V]()` — panic version of ValidateID for init-time validation
- `ErrInvalidID` sentinel error for ID validation failures
- `GoString()` now returns `id.BrandName(value)` instead of mirroring `String()`
- `%#v` format now shows `id.BrandName(value)` for meaningful debug output
- 15 new tests and 3 new Example tests for brand utilities

### Changed

- **`String()` is now brand-aware**: returns `"Brand:value"` for named brands (e.g., `"User:abc123"`), value-only for unnamed brands (backward compatible)
- `MarshalText()` uses internal `valueString()` — serialization never includes brand prefix
- README rewritten: Quick Start shows named brands, "Named Brand Types" section replaces "Best Practice", Brand Utilities in API Reference

### Fixed

- `ValidateID` function now actually exists in the library (was only documented as example code before)
- `Name()` method on brand types is now consumed by the library (was documented but ignored before)

## [0.2.0] - 2026-05-04

### Added

- `Ptr()` method returning `*ID[B, V]` for optional ID fields
- `FromPtr()` function dereferencing pointers with nil→zero fallback
- `Format` method implementing `fmt.Formatter` for custom formatting (`%s`, `%d`, `%v`, `%#v`, `%q`)
- Comprehensive test coverage for all integer types across serialization methods
- Fuzz tests for JSON and binary round-trips
- CI workflow for build, test (with race detector), and lint
- MIT license (changed from Proprietary)

### Changed

- Removed `float64` from `Compare` — floats are not valid ID types; no serialization format supports them

### Fixed

- `Scan` for `int8`, `int16`, `uint8`, `uint16` — missing type cases caused silent failures
- Inconsistent `Scan` implementation for `int` — now uses shared `scanIntegerID` helper
- `readByte` redundant double type assertion
- `readUnsigned` panic for `uint16`/`uint32` during binary deserialization

### Removed

- `float64` support from `Compare` — eliminated split-brain (no serialization format supports float64)

## [0.1.0] - 2026-01-01

### Added

- Initial release
- Core `ID[B, V]` type with phantom typing for compile-time type safety
- Serialization support: JSON, SQL, Binary, Text, Gob
- Comparison, equality, and zero-value semantics
- All integer types: string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64
