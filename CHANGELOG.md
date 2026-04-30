# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `Ptr()` method returning `*ID[B, V]` for optional ID fields
- `FromPtr()` function dereferencing pointers with nil→zero fallback
- `Format` method implementing `fmt.Formatter` for custom formatting (`%s`, `%d`, `%v`, `%#v`, `%q`)
- Comprehensive test coverage for all integer types across serialization methods
- Fuzz tests for JSON and binary round-trips

### Changed

- Removed `float64` from `Compare` — floats are not valid ID types; no serialization format supports them

### Fixed

- `Scan` for `int8`, `int16`, `uint8`, `uint16` — missing type cases caused silent failures
- Inconsistent `Scan` implementation for `int` — now uses shared `scanIntegerID` helper
- `readByte` redundant double type assertion
- `readUnsigned` panic for `uint16`/`uint32` during binary deserialization

### Removed

- `float64` support from `Compare` — eliminated split-brain (no serialization format supports float64)

### Deprecated

### Security

## [0.1.0] - 2026-01-01

### Added

- Initial release
- Core `ID[B, V]` type with phantom typing for compile-time type safety
- Serialization support: JSON, SQL, Binary, Text, Gob
- Comparison, equality, and zero-value semantics
- All integer types: string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64
