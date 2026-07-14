# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.3.2] - 2026-07-13

### Fixed

- **CI could not build the library**: `id_json.go` and `id_sql.go` import `encoding/json/v2`, which requires `GOEXPERIMENT=jsonv2`. All CI workflows (`go.yml`, `release.yml`, `validate-docs.yml`) and the Nix flake now set `GOEXPERIMENT=jsonv2`. This was the root cause of the missing v0.3.1 GitHub Release — the release workflow's `go test -race ./...` step failed because the build constraints excluded `encoding/json/v2`.
- Release workflow used `golangci-lint-action@v6` while the CI workflow used `@v7`; both now use `@v7`.

### Changed

- README performance table updated with verified benchmark numbers from Go 1.26.4 (`MarshalJSON` was listed as ~60ns/2allocs, actual is ~200ns/3allocs; `Compare` was ~1.5ns, actual is ~3ns; several other corrections).
- `CONTRIBUTING.md` rebuilt: removed stale references to `just`, `pkg/errors/`, `go-arch-lint`, and `CONTRIBUTING-setup.sh` (none exist); replaced with accurate Nix-based instructions.
- Installation instructions now document the `GOEXPERIMENT=jsonv2` prerequisite.
- `MIGRATION.md` updated: Go version requirement is `1.26+` (not `1.26.2`); added troubleshooting for the `encoding/json/v2` build constraint error.

### Added

- `FEATURES.md` — honest feature inventory with code-verified `file:line` citations (was missing).
- `ROADMAP.md` — long-term vision (ecosystem adoption, v1.0 stability, tooling).
- `docs/DOMAIN_LANGUAGE.md` — added `Phantom Type` and `Zero Value` terms; removed inapplicable empty DDD template sections; fixed `.` placeholder.
- CHANGELOG compare links.

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

> Note: this version was never tagged as a separate release; the changes below
> are included in the `v0.3.0` tag. Kept here for chronological accuracy.

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

---

[0.3.2]: https://github.com/larsartmann/go-branded-id/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/larsartmann/go-branded-id/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/larsartmann/go-branded-id/compare/v0.1.0...v0.3.0
[0.2.0]: https://github.com/larsartmann/go-branded-id/compare/v0.1.0...v0.3.0
[0.1.0]: https://github.com/larsartmann/go-branded-id/releases/tag/v0.1.0
