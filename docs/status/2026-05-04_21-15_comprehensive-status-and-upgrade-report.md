# Status Report: go-branded-id

**Date:** 2026-05-04 21:15  
**Generated from:** `master` branch at `e64969e`  
**Session work:** Migration guide upgrade, CI workflow, v0.2.0 prep, cleanup

---

## Project Overview

| Metric                 | Value                                                                                          |
| ---------------------- | ---------------------------------------------------------------------------------------------- |
| **Module**             | `github.com/larsartmann/go-branded-id`                                                         |
| **Go version**         | 1.26.2                                                                                         |
| **Dependencies**       | 0 (pure stdlib)                                                                                |
| **Source files**       | 8 (`id.go`, `id_binary.go`, `id_gob.go`, `id_json.go`, `id_ptr.go`, `id_sql.go`, `id_text.go`) |
| **Test files**         | 6                                                                                              |
| **Total LoC**          | 3,393 (1,260 production, 2,133 test)                                                           |
| **Test-to-code ratio** | 1.69:1                                                                                         |
| **Coverage**           | 80.0%                                                                                          |
| **License**            | MIT                                                                                            |
| **Current tag**        | `v0.1.0`                                                                                       |
| **Pending release**    | `v0.2.0` (ready)                                                                               |
| **Size**               | 172KB                                                                                          |
| **Lint**               | 0 issues                                                                                       |
| **Race detector**      | Clean                                                                                          |
| **Consumers**          | 6 files in `go-composable-business-types`                                                      |

---

## a) FULLY DONE ✅

1. **Core library** — `ID[B, V comparable]` phantom type with full branded type safety
2. **All serialization formats** — JSON, SQL, Binary, Text, Gob — all 11 numeric types + string
3. **Full API surface** — `NewID`, `Get`, `IsZero`, `Reset`, `Equal`, `Compare`, `Or`, `String`, `GoString`, `Format`, `Ptr`, `FromPtr`
4. **SQL support** — `Scan` and `Value` for string, all int/uint types, nil handling
5. **Performance** — Zero-allocation core ops; `NewID` ~0.22ns, `Get` ~1ns, `Equal` ~0.22ns
6. **Comprehensive tests** — Unit, integration, fuzz tests for JSON/Binary round-trips
7. **Benchmarks** — 19 benchmarks covering all major operations
8. **Lint** — 0 issues with aggressive golangci-lint config (50+ linters)
9. **MIT license** — Changed from Proprietary
10. **Migration guide** — Fully rewritten with prerequisites, verification, troubleshooting, bonus features
11. **Docs validation CI** — `validate-docs.yml` validates all Go code blocks in Markdown
12. **Go CI workflow** — NEW: `go.yml` with build + test (race + cover) + golangci-lint
13. **CHANGELOG v0.2.0** — Cut from `[Unreleased]` to `[0.2.0] - 2026-05-04` with all changes documented
14. **Cleanup** — Removed stale `docs/status/` (5 historical files), empty `report/`, empty `docs/`
15. **git-town config** — Configured with `main = "master"`

---

## b) PARTIALLY DONE 🔧

1. **Test coverage — 80.0%** — Good but not great. Gaps:
   - `scanIntegerID` — 33.3% (SQL deserialization integer helper)
   - `UnmarshalText` — 65.6% (Text deserialization)
   - `String` — 66.7% (string representation)
   - `Value` — 70.0% (SQL value driver)
   - `Scan` — 70.2% (SQL scan)
   - `UnmarshalBinary` — 78.7% (binary deserialization)
   - `Format` — 80.0% (fmt.Formatter)
   - `MarshalBinary` — 88.9%
   - `Compare` — 92.3%
   - `MarshalJSON` — 83.3%

2. **v0.2.0 release** — CHANGELOG is ready but **not tagged**. Consumer (`go-composable-business-types`) still pins `v0.1.0` with a `replace` directive.

3. **README** — Solid but could use: `UnmarshalText` example, `Gob` example, `Format` verb examples

---

## c) NOT STARTED ⬜

1. **No `go.sum` file** — `go.mod` has zero dependencies, so it's empty/absent. Not a problem but unusual.
2. **No `CONTRIBUTING.md`** — README mentions contributing guidelines but no dedicated file
3. **No Godoc site** — No pkg.go.dev badge or godoc integration
4. **No release automation** — No goreleaser, no tag-triggered CI release
5. **No dependabot/renovate** — No automated dependency scanning (moot since zero deps)
6. **No code owners** — No `CODEOWNERS` file
7. **No PR/issue templates** — No `.github/ISSUE_TEMPLATE/` or `.github/PULL_REQUEST_TEMPLATE.md`
8. **No `go vet` standalone** — Only runs through golangci-lint
9. **No mutation testing** — No `go-mutesting` or similar
10. **No `//go:generate`** — No code generation setup
11. **No `doc.go`** — No package-level doc example file
12. **No security policy** — No `SECURITY.md`
13. **No reproducible builds** — No `GOFLAGS=-trimpath` in CI
14. **No coverage enforcement** — No minimum coverage threshold in CI

---

## d) TOTALLY FUCKED UP 💥

1. **Consumer still on v0.1.0 with `replace` directive** — `go-composable-business-types/go.mod` line 15: `replace github.com/larsartmann/go-branded-id => ../go-branded-id`. This is a local dev hack that will break for anyone else cloning that repo. **Must be removed after v0.2.0 tag.**

2. **No `go.sum` = no integrity verification** — While there are zero dependencies, the lack of `go.sum` means downstream consumers who expect it may have issues. (Low severity — `go mod tidy` generates it.)

3. **Stale `coverage.out` in project root** — Leftover from this session's coverage analysis. Should be gitignored or removed.

---

## e) WHAT WE SHOULD IMPROVE 🚀

### High Impact

1. **Close coverage gaps** — Get to 90%+. `scanIntegerID` at 33.3% is embarrassing. Add tests for all integer type branches in `Scan`, `Value`, `UnmarshalText`, `UnmarshalBinary`.
2. **Tag v0.2.0** — Everything is ready. Just needs `git tag v0.2.0 && git push --tags`.
3. **Remove `replace` directive in consumer** — After v0.2.0 is tagged, remove `replace` from `go-composable-business-types/go.mod` and update to real `v0.2.0`.
4. **Add `.gitignore` entry for `coverage.out`** — Prevent artifact commits.

### Medium Impact

5. **Add coverage threshold to CI** — Fail CI if coverage drops below 85%.
6. **Add `doc.go` with runnable examples** — Better pkg.go.dev experience.
7. **Add `CONTRIBUTING.md`** — Since README mentions it.
8. **Add `SECURITY.md`** — Standard for open source libraries.
9. **Add release automation** — Tag-triggered GitHub Actions that creates a GitHub release.

### Low Impact

10. **Add PR/issue templates** — Professional OSS polish.
11. **Add `CODEOWNERS`** — If multiple contributors ever join.
12. **Explore `go:generate` for type-switch boilerplate** — The binary/SQL code has repetitive type switches that could be generated.

---

## f) Top #25 Things We Should Get Done Next

| #   | Priority | Task                                                                  | Est. Effort |
| --- | -------- | --------------------------------------------------------------------- | ----------- |
| 1   | P0       | Tag `v0.2.0` and push to remote                                       | 1 min       |
| 2   | P0       | Remove `replace` directive from `go-composable-business-types/go.mod` | 2 min       |
| 3   | P0       | Update consumer to `v0.2.0` (remove local replace)                    | 2 min       |
| 4   | P0       | Add `coverage.out` to `.gitignore`                                    | 1 min       |
| 5   | P0       | Delete stale `coverage.out` from project root                         | 1 min       |
| 6   | P1       | Add tests for `scanIntegerID` (33.3% → 90%+)                          | 30 min      |
| 7   | P1       | Add tests for `UnmarshalText` error paths (65.6% → 90%+)              | 20 min      |
| 8   | P1       | Add tests for `String()` `TextMarshaler` fallback path (66.7% → 90%+) | 15 min      |
| 9   | P1       | Add tests for `Value()` all int/uint types (70% → 90%+)               | 20 min      |
| 10  | P1       | Add tests for `Scan()` all int/uint types (70.2% → 90%+)              | 20 min      |
| 11  | P1       | Add tests for `UnmarshalBinary` error paths (78.7% → 90%+)            | 15 min      |
| 12  | P1       | Add tests for `Format` all verbs (80% → 95%+)                         | 15 min      |
| 13  | P1       | Add coverage threshold to CI (`go.yml`) — fail below 85%              | 5 min       |
| 14  | P1       | Add `SECURITY.md`                                                     | 10 min      |
| 15  | P2       | Add `CONTRIBUTING.md`                                                 | 15 min      |
| 16  | P2       | Add `doc.go` with package examples                                    | 10 min      |
| 17  | P2       | Add pkg.go.dev badge to README                                        | 5 min       |
| 18  | P2       | Add tag-triggered release GitHub Action                               | 30 min      |
| 19  | P2       | Add `UnmarshalText` example to README                                 | 5 min       |
| 20  | P2       | Add `Gob` example to README                                           | 5 min       |
| 21  | P2       | Add `Format` verb examples to README                                  | 5 min       |
| 22  | P3       | Add `.github/ISSUE_TEMPLATE/` (bug + feature)                         | 15 min      |
| 23  | P3       | Add `.github/PULL_REQUEST_TEMPLATE.md`                                | 10 min      |
| 24  | P3       | Add reproducible build flags to CI (`GOFLAGS=-trimpath`)              | 5 min       |
| 25  | P3       | Explore code generation for repetitive type-switch patterns           | 2 hr        |

---

## g) Top #1 Question I Cannot Answer Myself

**Should this project support UUID (`[16]byte`) as a value type?**

- UUID is `comparable` in Go and would technically work with `ID[B, [16]byte]`
- But none of the serialization formats (JSON, SQL, Binary, Text, Gob) have specialized support for `[16]byte` — it would fall through to the `default` case in every type switch
- The `String()` method would return `"id:%!v(MISSING)"` for UUID values
- This is a **design decision** that affects the API contract — do we add full UUID serialization support, explicitly document it as unsupported, or add a `uuid` build-tag module?
- **I cannot decide this without your product direction.**

---

## Session Changes Summary

| File                       | Change                                                                        |
| -------------------------- | ----------------------------------------------------------------------------- |
| `MIGRATION.md`             | Rewritten: added prerequisites, verification, bonus features, troubleshooting |
| `CHANGELOG.md`             | Cut v0.2.0 release (2026-05-04)                                               |
| `.github/workflows/go.yml` | NEW: CI build + test (race) + lint                                            |
| `docs/status/*`            | DELETED: 5 stale status reports                                               |
| `report/`                  | DELETED: empty directory                                                      |
| `docs/`                    | DELETED: was empty after status removal                                       |
