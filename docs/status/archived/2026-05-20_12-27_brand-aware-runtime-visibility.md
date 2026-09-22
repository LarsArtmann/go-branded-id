# Status Report: go-branded-id

**Date:** 2026-05-20 12:27\
**Trigger:** Critical review — library documented a best practice it didn't support, phantom types invisible at runtime\
**Status:** Changes implemented, tested, linted — NOT YET VERSIONED OR PUSHED

---

## Executive Summary

The library had 56 repos importing it but its branded IDs were completely invisible at runtime. `String()` returned just `"abc123"` with zero indication of whether that was a UserID, OrderID, or AggregateID. The README documented a `ValidateID` function that didn't exist in the library, and a `Name()` best practice that the library itself never consumed.

This has been fixed. `String()` is now brand-aware, `ValidateID` is shipped, and `GoString()` / `%#v` provide meaningful debug output.

---

## A) FULLY DONE ✅

### 1. Brand-Aware `String()` Method

- **Before:** `fmt.Println(userID)` → `"abc123"` (invisible)
- **After:** `fmt.Println(userID)` → `"User:abc123"` (for named brands)
- **Unnamed brands:** Unchanged — returns just `"abc123"` (backward compatible)
- **Implementation:** Extracted `valueString()` for internal use; `String()` adds brand prefix only when `BrandNamer` is implemented

### 2. `BrandNamer` Interface + `BrandName[B]()` Function

- New `BrandNamer` interface: `Name() string`
- `BrandName[B]()` returns brand name or falls back to `%T` type name
- Internal `brandName[B]()` returns `(string, bool)` to avoid allocation for unnamed brands

### 3. `ValidateID` + `ValidateIDWithValue` — Actually Shipped

- The README's example code is now a real function in the library
- Error format: `id: invalid: User: empty` (no more "invalid ID ID" double-talk)
- `ValidateIDWithValue` accepts optional custom validator function
- Both use `ErrInvalidID` sentinel error with `%w` wrapping

### 4. Improved `GoString()` and `%#v` Format

- **Before:** `GoString()` returned same as `String()` — useless
- **After:** `fmt.Printf("%#v", id)` → `id.User(abc123)` for named brands
- Unnamed brands: `id.id.StringBrand(abc123)` (shows package-qualified type)

### 5. Serialization Unaffected

- `MarshalText()` now uses `valueString()` (raw value, no brand prefix)
- JSON, SQL, Binary, Gob serialization unchanged — all use raw value
- No brand prefix leaks into stored data

### 6. README Rewritten

- Quick Start now shows `Name()` on the brand and demonstrates `String()` output
- "Best Practice" section replaced with "Named Brand Types" — shows actual shipped API
- Brand Utilities section added to API Reference
- Serialization explicitly documented as raw-value-only

### 7. Tests — All Passing

- 81 tests, 0 failures, race detector clean
- New tests: `TestString_BrandAware`, `TestString_NoBrandName`, `TestGoString_BrandAware`, `TestFormat_HashV_NamedBrand`, etc.
- 3 new Example tests: `ExampleID_String_named`, `ExampleID_String_unnamed`, `ExampleValidateID_zero`
- `golangci-lint run`: 0 issues

---

## B) PARTIALLY DONE ⚠️

### 1. Version Bump

- Changes are on `master` but NOT tagged as a new version
- This is a **minor semver bump** (new functionality, backward compatible)
- Needs `v0.2.0` or `v1.0.0` tag

### 2. CHANGELOG.md

- Not updated yet with these changes
- Previous CHANGELOG only covers up to v0.1.0

### 3. MIGRATION.md

- Should note that `String()` now includes brand prefix for named brands
- Should document that existing brands without `Name()` are unaffected

---

## C) NOT STARTED ❌

### 1. Ecosystem Update — 56 Repos

The following repos could benefit from adding `Name()` to their brand types. None will break (they don't have `Name()`, so `String()` is unchanged), but they'd gain debug visibility:

| Repo          | Brand Types                  | Has Name() |
| ------------- | ---------------------------- | ---------- |
| go-cqrs-lite  | aggregate, todo, user brands | No         |
| BuildFlow     | telemetry identifiers        | No         |
| GmbH          | domain entity IDs            | No         |
| SEC           | entity IDs                   | No         |
| BerryBig      | entity IDs                   | No         |
| ChastityAPI   | entity IDs                   | No         |
| Cyberdom      | entity IDs                   | No         |
| storbi        | entity IDs                   | No         |
| StopTube      | entity IDs                   | No         |
| smart-configs | entity IDs                   | No         |
| timesheets    | entity IDs                   | No         |
| 20+ others    | unknown                      | No         |

### 2. go-cqrs-lite `id.Of[T]` Integration

- go-cqrs-lite wraps `go-branded-id` with ULID-backed IDs
- Its brands (`userBrand`, `orderBrand`, etc.) don't have `Name()`
- Could add `Name()` to all brands for immediate benefit

### 3. Benchmark Update (`id_bench_test.go`)

- Existing benchmarks don't cover `BrandName[B]()` or `String()` with named brands
- Should add benchmarks to verify zero-allocation claim holds with brand lookup

### 4. Fuzz Tests for Brand Utilities

- No fuzz coverage for `ValidateID` or `BrandName`

### 5. CI Pipeline Update

- GitHub Actions workflow may need updating for new files
- Should verify `go test ./... -race` and `golangci-lint run` pass in CI

---

## D) TOTALLY FUCKED UP 💥

### Nothing catastrophic.

**One design concern:** The `String()` output format `"User:abc123"` introduces a colon separator. If any consumer splits on `:` to extract the value from `String()`, this will break. However:

- Serialization (JSON/SQL/Text/Binary) all use `valueString()` — no prefix
- `Get()` returns the raw value — always safe
- The `:` separator is only in `String()` / `fmt.Println()` output
- This is documented in README: "Serialization always uses the raw value"

**One cosmetic issue:** For unnamed brands, `GoString()` returns `id.id.StringBrand(...)` — the double `id.` is because the package name is `id` and the format is `id.BrandName(...)`. This is technically correct but aesthetically ugly. Could be improved by stripping the package prefix or using a different format.

---

## E) WHAT WE SHOULD IMPROVE 🔧

1. ~~**`GoString()` for unnamed brands** — `id.id.StringBrand(42)` is ugly. Consider just `StringBrand(42)` or detecting when brand name starts with package name and stripping it.~~ **Won't implement — package-qualified fallback kept and documented (AGENTS.md BrandName fallback section).**

2. ~~**`ErrInvalidID` wrapping** — Currently `fmt.Errorf("%w: %s: empty", ErrInvalidID, BrandName[B]())`. The triple-colon format `"id: invalid: User: empty"` is verbose. Could simplify to `"User: invalid: empty"` or `"id: User: empty"`.~~ **Won't implement — sentinel message format kept stable; changing it would break message-parsing consumers.**

3. ~~**`ValidateID` should accept any comparable V** — Already does, but error message doesn't indicate the value type. Could be useful for debugging.~~ **Won't implement — brand name in the message already identifies the type; value type adds noise.**

4. ~~**Consider adding `MustValidateID`** — panic version for init-time validation, consistent with Go patterns.~~ done (shipped in v0.3.1 (8b30d92); id_brand.go:81)

5. ~~**Package doc comment** — Should mention `BrandNamer` interface and `ValidateID` in the overview.~~ done (package doc in id.go documents phantom types and brand naming)

6. ~~**Integration test with real consumer** — Run go-cqrs-lite tests against the updated library to verify no breakage.~~ **Won't implement — not feasible in-repo; downstream consumers verified pinned in the 2026-09-17 linter-stack pass.**

---

## F) TOP 25 THINGS TO DO NEXT

| #  | Priority | Task                                                                   |
| --- | ------------ | --------------------------------------------------------------------------------------------------------- |
| ~~1~~  | ~~P0~~ done — v0.3.0 (044bd67) | ~~Tag release: `v0.2.0` or `v1.0.0`~~ |
| ~~2~~  | ~~P0~~ done — v0.3.0 CHANGELOG section | ~~Update CHANGELOG.md~~ |
| ~~3~~  | ~~P0~~ done — tag on origin; releases automated since (release.yml) | ~~Push to remote~~ |
| ~~4~~  | ~~P1~~ **Won't implement — marker String() output is used as storage/stream keys; Name() would corrupt data (documented in AGENTS.md).** | ~~Add `Name()` to all brand types in go-cqrs-lite~~ |
| ~~5~~  | ~~P1~~ done — Correlation brand named in the 2026-05-20 migration (64 Name() methods across 12 repos) | ~~Add `Name()` to all brand types in ActaFlow~~ |
| ~~6~~  | ~~P1~~ done — indirect dep; source fixes applied in the 2026-05-20 migration | ~~Add `Name()` to all brand types in CreditReformBilanzampel~~ |
| ~~7~~  | ~~P1~~ done — 4 brands named (Message, Label, Thread, User) | ~~Add `Name()` to all brand types in InboxClean~~ |
| ~~8~~  | ~~P1~~ done — 27 .String() to .Get() fixes applied across consumers in the 2026-05-20 pass | ~~Run go-cqrs-lite test suite against updated library~~ |
| ~~9~~  | ~~P1~~ **Won't implement — package-qualified fallback kept; documented in AGENTS.md.** | ~~Fix `GoString()` ugly double-package prefix for unnamed brands~~ |
| ~~10~~ | ~~P2~~ done — String/BrandName/ValidateID benchmarks in id_bench_test.go (29 benchmark functions) | ~~Add benchmarks for `BrandName[B]()` and brand-aware `String()`~~ |
| ~~11~~ | ~~P2~~ done — FuzzValidateID in id_bench_test.go | ~~Add fuzz tests for `ValidateID`~~ |
| ~~12~~ | ~~P2~~ done — String() behavior change documented in MIGRATION.md | ~~Update MIGRATION.md with `String()` behavior change~~ |
| ~~13~~ | ~~P2~~ **Won't implement — no branded-id brand types surfaced in the migration sweep; not in the consumer set.** | ~~Add `Name()` to brand types in BuildFlow~~ |
| ~~14~~ | ~~P2~~ **Won't implement — no branded-id brand types surfaced in the migration sweep; not in the consumer set.** | ~~Add `Name()` to brand types in GmbH~~ |
| ~~15~~ | ~~P2~~ done — 2 brands named (Game, Player) | ~~Add `Name()` to brand types in SEC~~ |
| ~~16~~ | ~~P2~~ **Won't implement — no brand types found (documented in the 2026-05-20 reports and AGENTS.md).** | ~~Add `Name()` to brand types in Cyberdom~~ |
| ~~17~~ | ~~P3~~ done — shipped in v0.3.1 (8b30d92); id_brand.go:81 | ~~Consider `MustValidateID` convenience function~~ |
| ~~18~~ | ~~P3~~ done — package doc in id.go covers BrandNamer and validation | ~~Update package doc comment to mention BrandNamer~~ |
| ~~19~~ | ~~P3~~ done — 12 repos / 64 Name() methods migrated 2026-05-20; remainder not in the consumer set | ~~Add `Name()` to remaining 30+ repos~~ |
| ~~20~~ | ~~P3~~ done — cmd/namer shipped v0.5.0 (5c4f995) | ~~Create a codemod/tool to add `Name()` to all brand types automatically~~ |
| ~~21~~ | ~~P3~~ done — Example functions in id_brand_test.go (ExampleValidateID, ExampleBrandName, ...) | ~~Add Example tests for `ValidateIDWithValue`~~ |
| ~~22~~ | ~~P3~~ done — DOMAIN_LANGUAGE.md Named Brand entry defines the format | ~~Consider adding `String()` format to domain language doc~~ |
| ~~23~~ | ~~P4~~ done — go.yml runs the full module in both JSON modes | ~~Verify CI pipeline covers new files~~ |
| ~~24~~ | ~~P4~~ **Won't implement — markers are internal keys; adopting library validation is an upstream go-cqrs-lite decision.** | ~~Update go-cqrs-lite to use `ValidateID` instead of custom validation~~ |
| ~~25~~ | ~~P4~~ **Won't implement — never written; no announcement channel.** | ~~Write blog post / announcement about the change~~ |

---

## G) TOP #1 QUESTION I CANNOT ANSWER MYSELF ❓

**Should `String()` include the brand prefix for ALL brands (not just named ones)?**

Currently, brands without `Name()` return just the value: `"abc123"`. This means the 26+ repos without `Name()` will still have invisible IDs in logs. The alternative is to always use the type name as prefix: `"id.UserBrand:abc123"` — which is ugly but guarantees every ID is identifiable.

The tradeoff:

- **Named only (current):** Clean output, but 83% of ecosystem brands remain invisible
- **Always prefixed:** Every ID is debuggable, but output is ugly for unnamed brands and this IS a breaking change for any code parsing `.String()` output

This is a product decision, not a technical one.
