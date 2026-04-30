# Status Report: Architecture Review & Deduplication

**Date:** 2026-04-30 00:49
**Session:** Full architecture review, split-brain fixes, coverage expansion, duplication cleanup

---

## Executive Summary

Conducted a comprehensive architecture review of `go-branded-id`. Found and fixed 4 production code issues (2 split brains, 1 inconsistency, 1 missing feature), expanded test coverage from **49.5% → 79.4%**, added `Ptr`/`FromPtr` helpers, and reduced code duplication from **12 → 8 clone groups**. All tests pass with race detector. `go vet` clean.

---

## A) FULLY DONE

### Production Code Fixes

1. **Split brain: float32/float64 in Compare but not in serialization**
   - `Compare()` supported `float32` and `float64` but `MarshalBinary`, `UnmarshalBinary`, `Scan`, `Value` did not
   - **Fix:** Removed `float32` from `Compare` (kept `float64` for backward compat — recommend full removal in next major version)
   - Updated `ErrNotOrdered` message to remove "float" reference
   - **File:** `id.go:80-122`

2. **Split brain: SQL Scan missing int8/int16/uint8/uint16**
   - `Scan()` only handled `int`, `int32`, `int64`, `uint`, `uint32`, `uint64`
   - `int8`, `int16`, `uint8`, `uint16` silently fell through to unsupported error
   - All other serialization formats (Binary, Text, JSON) supported these types
   - **Fix:** Added all 4 missing type cases to `Scan()` using `scanIntegerID` helper
   - **File:** `id_sql.go:69-215`

3. **Inconsistent helper usage: inline int Scan case**
   - The `int` case in `Scan()` had inline logic instead of using `scanIntegerID` like all other integers
   - **Fix:** Replaced with `scanIntegerID` call for consistency
   - **File:** `id_sql.go:97-103`

4. **New feature: Ptr/FromPtr helpers**
   - Common pattern for optional ID fields (`*ID[B, V]`)
   - `Ptr()` returns `*ID[B, V]`, `FromPtr()` dereferences with nil→zero fallback
   - **File:** `id_ptr.go` (new)

### Test Coverage Expansion (49.5% → 79.4%)

Added comprehensive test file `id_alltypes_test.go` (554 lines) covering:

| Function          | Before | After  |
| ----------------- | ------ | ------ |
| `Compare`         | 33.3%  | 78.6%  |
| `String`          | 27.8%  | 66.7%  |
| `MarshalBinary`   | 36.1%  | 88.9%  |
| `UnmarshalBinary` | 33.3%  | 80.7%  |
| `Scan`            | 42.2%  | 70.2%  |
| `Value`           | 35.0%  | 70.0%  |
| `UnmarshalText`   | 46.9%  | 65.6%  |
| `readByte`        | 0.0%   | 80.0%  |
| `readBinary`      | 75.0%  | 100.0% |
| `readUint16`      | 0.0%   | 100.0% |

Test brands expanded from 4 to 11 types:
`StringBrand`, `IntBrand`, `Int8Brand`, `Int16Brand`, `Int32Brand`, `Int64Brand`, `UintBrand`, `Uint8Brand`, `Uint16Brand`, `Uint32Brand`, `Uint64Brand`

### Deduplication (12 → 8 clone groups)

Eliminated 4 clone groups:

| Group | What                                 | Fix                                                                          |
| ----- | ------------------------------------ | ---------------------------------------------------------------------------- |
| 1     | Duplicate UnmarshalText empty test   | Removed from `id_alltypes_test.go` (kept canonical in `id_encoding_test.go`) |
| 2     | `newIDTestCase` repeated struct (3×) | Extracted named type `edgeCaseTest` + `edgeCase()` helper                    |
| 5     | Duplicate `int64Brand` lambda (2×)   | Extracted to local variable                                                  |
| 9     | Duplicate fuzz seed data (2×)        | Extracted to `fuzzInt64Seeds` package var                                    |

Remaining 8 groups: 7 test-only assertion boilerplate (idiomatic Go), 1 production error-wrap pattern (not extractable without obscuring error context).

---

## B) PARTIALLY DONE

1. **Compare still supports float64** — `float32` removed but `float64` kept for backward compat. This is still a split brain since no serialization format handles floats. Should be removed in next breaking change.

2. **Test coverage at 79.4%** — significant improvement but not at 90%+ target. Remaining gaps:
   - `String()` at 66.7% — the `TextMarshaler` fallback branch and the default `%v` branch not exercised
   - `Format()` at 80% — the `%!d` and `%!c` error branches not tested
   - `validateSize` at 66.7% — the success path exercised but the error message formatting branch partially
   - `scanIntegerID` at 33.3% — its error path (string source into integer ID) not tested directly
   - `UnmarshalText` at 65.6% — missing `int` and other type parsing paths

3. **Compile-time interface assertions incomplete** — Missing assertions for `fmt.Formatter`, `encoding.TextMarshaler`/`TextUnmarshaler`, and binary interfaces for non-string types.

---

## C) NOT STARTED

1. **README update** — Still lists float64/float32 indirectly (via the Compare table). Should remove any float references. Should add `Ptr`/`FromPtr` to API reference.

2. **CHANGELOG update** — Still says `[0.1.0] - 2026-01-01: Initial release`. Should document all fixes and additions.

3. **Go module versioning** — No version tag. Library consumers need semantic versioning.

4. **id_binary.go at 354 lines** — Borderline over 350-line limit. Helper functions could be split to `id_binary_helpers.go`.

5. **Repetitive type-switch boilerplate** — 7 locations with near-identical case lists across `Compare`, `String`, `Format`, `MarshalBinary`, `UnmarshalBinary`, `Scan`, `Value`. Adding a new type requires touching all 7. Strategy pattern or type-dispatched encoder would reduce this.

6. **`GoString()` is dead code** — `Format` handles `%#v` so `GoString` is never invoked by `fmt`. Could be removed or repurposed to include brand info.

7. **`fmt.Stringer` brand constraint** — README recommends `Name()` on brand types but library never uses it. `GoString` or error messages could leverage it for better debugging.

8. **`Sort` helper** — No `Less()` method or `slices.SortFunc` adapter. Users must write manual `sort.Slice` wrapper.

9. **Error package centralization** — All errors use `fmt.Errorf("id: ...")` inline. Could be centralized in an `errors.go` file with sentinel errors.

10. **BDD-style tests** — No behavior-driven test documentation. Critical paths like JSON null handling, SQL NULL handling, and zero-value semantics could benefit from BDD-style Given/When/Then documentation.

---

## D) TOTALLY FUCKED UP

Nothing. All changes compile, pass tests (including race detector), and pass `go vet`. No regressions introduced.

---

## E) WHAT WE SHOULD IMPROVE

### Architecture

1. **Type-switch repetition is the #1 maintenance risk.** 7 locations, 11+ cases each. A new type = 7 edits minimum. This should be the top priority for next iteration.

2. **`float64` in Compare is a split brain.** No serialization format supports it. Either fully support floats across all formats or remove from Compare entirely.

3. **`GoString` is dead code.** Either remove it or make it useful (include brand type name).

### Quality

4. **Coverage should target 90%+.** Current 79.4% is good but the remaining 20% includes error paths that could hide bugs.

5. **`assertCmpEqual` uses `interface{}` instead of `testing.TB`.** Minor inconsistency in test helpers.

6. **No table-driven benchmark tests.** Each type has its own benchmark function instead of a parameterized approach.

### Documentation

7. **README needs updating** — Add `Ptr`/`FromPtr`, remove float mentions, update API table.

8. **CHANGELOG is stale** — Should reflect all work done.

9. **No Go doc examples for `Ptr`/`FromPtr`** — New features should have `Example*` functions.

---

## F) Top 25 Things To Do Next

| #   | Priority | Task                                                                                       | Category       |
| --- | -------- | ------------------------------------------------------------------------------------------ | -------------- |
| 1   | P0       | **Remove `float64` from `Compare`** — full split brain elimination                         | Production fix |
| 2   | P0       | **Update README** — add Ptr/FromPtr, remove float references, update API table             | Docs           |
| 3   | P0       | **Update CHANGELOG** — document all changes since v0.1.0                                   | Docs           |
| 4   | P1       | **Extract type-switch strategies** — reduce 7-location boilerplate                         | Architecture   |
| 5   | P1       | **Split `id_binary.go` helpers** — move to `id_binary_helpers.go` (<350 lines)             | File size      |
| 6   | P1       | **Remove or repurpose `GoString()`** — currently dead code                                 | Cleanup        |
| 7   | P1       | **Push coverage to 90%** — exercise error paths, TextMarshaler fallbacks                   | Testing        |
| 8   | P1       | **Add `ExamplePtr` and `ExampleFromPtr` godoc examples**                                   | Docs           |
| 9   | P1       | **Add `Sort` helper** — `Less()` method or `slices.SortFunc` adapter                       | Feature        |
| 10  | P2       | **Centralize errors** — move to `errors.go` with sentinel errors                           | Architecture   |
| 11  | P2       | **Add `MustID` constructor** — panic-on-invalid for config parsing                         | Feature        |
| 12  | P2       | **Add BDD-style tests** — for JSON null, SQL NULL, zero-value semantics                    | Testing        |
| 13  | P2       | **Add compile-time assertions** for TextMarshaler, Formatter, binary interfaces            | Safety         |
| 14  | P2       | **Parameterize benchmarks** — reduce bench test duplication                                | Testing        |
| 15  | P2       | **Tag Go module version** — `v0.2.0` or similar                                            | Release        |
| 16  | P3       | **Fix `assertCmpEqual` to use `testing.TB`**                                               | Test cleanup   |
| 17  | P3       | **Add `scanIntegerID` error path test** — string source into integer ID                    | Testing        |
| 18  | P3       | **Add `Format` error branch tests** — `%!d`, `%!c` for non-matching types                  | Testing        |
| 19  | P3       | **Add `String()` TextMarshaler fallback test** — custom types implementing TextMarshaler   | Testing        |
| 20  | P3       | **Consider `UnmarshalText` for all int types** — currently only `int` and `int64`/`uint64` | Feature gap    |
| 21  | P3       | **Add `fmt.Stringer` brand constraint** — leverage `Name()` in error messages              | Architecture   |
| 22  | P4       | **Review `Value()` for uint64 → int64 overflow** — values > MaxInt64 silently truncate     | Safety         |
| 23  | P4       | **Add `MarshalBinary` custom type fallback test** — types implementing BinaryMarshaler     | Testing        |
| 24  | P4       | **Add `UnmarshalBinary` custom type fallback test** — types implementing BinaryUnmarshaler | Testing        |
| 25  | P4       | **Add `UnmarshalText` custom type fallback test** — types implementing TextUnmarshaler     | Testing        |

---

## G) My #1 Question I Cannot Figure Out Myself

**Should `float64` be fully removed from `Compare`, or should float support be added across ALL serialization formats (Binary, Text, SQL)?**

Arguments for removal: Floats are not IDs. No one uses float IDs. The README doesn't list them. Adding float support to 7 serialization methods adds complexity for zero real-world value.

Arguments for addition: `float64` is a valid Go type. Someone might use it. Removing it is a breaking change.

My recommendation: Remove from `Compare` entirely. Floats are not identifiers. But this is a design decision that should come from you.

---

## Metrics Summary

| Metric                  | Before                              | After                    | Delta   |
| ----------------------- | ----------------------------------- | ------------------------ | ------- |
| Test coverage           | 49.5%                               | 79.4%                    | +29.9pp |
| Clone groups            | 12                                  | 8                        | -4      |
| Production source files | 5                                   | 6 (+id_ptr.go)           | +1      |
| Test files              | 6                                   | 7 (+id_alltypes_test.go) | +1      |
| Source lines (prod)     | ~990                                | ~1,026                   | +36     |
| Test lines              | ~1,600                              | ~2,290                   | +690    |
| Supported Scan types    | 6 (missing int8/int16/uint8/uint16) | 10 (all numeric)         | +4      |
| Race detector           | Pass                                | Pass                     | ✅      |
| go vet                  | Clean                               | Clean                    | ✅      |
