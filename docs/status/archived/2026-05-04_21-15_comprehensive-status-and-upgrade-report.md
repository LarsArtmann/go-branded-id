# Status Report: go-branded-id

**Date:** 2026-05-04 21:15\
**Generated from:** `master` branch at `e64969e`\
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

1. ~~**Test coverage — 80.0%** — Good but not great. Gaps:~~ done (88.5% after successive coverage passes (latest: delegate-path sentinel tests 2026-09-22))
   ~~- `scanIntegerID` — 33.3% (SQL deserialization integer helper)~~
   ~~- `UnmarshalText` — 65.6% (Text deserialization)~~
   ~~- `String` — 66.7% (string representation)~~
   ~~- `Value` — 70.0% (SQL value driver)~~
   ~~- `Scan` — 70.2% (SQL scan)~~
   ~~- `UnmarshalBinary` — 78.7% (binary deserialization)~~
   ~~- `Format` — 80.0% (fmt.Formatter)~~
   ~~- `MarshalBinary` — 88.9%~~
   ~~- `Compare` — 92.3%~~
   ~~- `MarshalJSON` — 83.3%~~

2. ~~**v0.2.0 release** — CHANGELOG is ready but **not tagged**. Consumer (`go-composable-business-types`) still pins `v0.1.0` with a `replace` directive.~~ done (shipped inside v0.3.0 (044bd67); the CHANGELOG 0.2.0 note records the fold)

3. ~~**README** — Solid but could use: `UnmarshalText` example, `Gob` example, `Format` verb examples~~ done (README rewritten 2026-05-20 and expanded through v0.6.0 (serialization, error handling))

---

## c) NOT STARTED ⬜

1. ~~**No `go.sum` file** — `go.mod` has zero dependencies, so it's empty/absent. Not a problem but unusual.~~ **Won't implement — zero dependencies by design — go.sum only exists when deps exist.**
2. ~~**No `CONTRIBUTING.md`** — README mentions contributing guidelines but no dedicated file~~ done (rebuilt in v0.3.2 (ed5ee4b) with accurate Nix-based instructions)
3. ~~**No Godoc site** — No pkg.go.dev badge or godoc integration~~ done (module indexed on pkg.go.dev; badge in README)
4. ~~**No release automation** — No goreleaser, no tag-triggered CI release~~ done (release.yml cuts GitHub Releases from tags; notes extracted from CHANGELOG (a03780b))
5. ~~**No dependabot/renovate** — No automated dependency scanning (moot since zero deps)~~ done (.github/dependabot.yml covers gomod + github-actions)
6. ~~**No code owners** — No `CODEOWNERS` file~~ **Won't implement — single maintainer, direct-push repo.**
7. ~~**No PR/issue templates** — No `.github/ISSUE_TEMPLATE/` or `.github/PULL_REQUEST_TEMPLATE.md`~~ **Won't implement — single maintainer, direct-push repo.**
8. ~~**No `go vet` standalone** — Only runs through golangci-lint~~ done (nix run .#vet app plus CI)
9. ~~**No mutation testing** — No `go-mutesting` or similar~~ **Won't implement — no demand signal for a stdlib-only type library.**
10. ~~**No `//go:generate`** — No code generation setup~~ **Won't implement — type switches maintained manually, locked by exhaustive all-types tests.**
11. ~~**No `doc.go`** — No package-level doc example file~~ done (pkg.go.dev renders package docs; Example_ test functions provide runnable docs)
12. ~~**No security policy** — No `SECURITY.md`~~ done (SECURITY.md created 2026-09-22)
13. ~~**No reproducible builds** — No `GOFLAGS=-trimpath` in CI~~ **Won't implement — source-only library; no binary releases to reproduce.**
14. ~~**No coverage enforcement** — No minimum coverage threshold in CI~~ **Won't implement — FEATURES.md verification snapshot re-derives coverage each docs pass; CI runs -cover.**

---

## d) TOTALLY FUCKED UP 💥

1. ~~**Consumer still on v0.1.0 with `replace` directive** — `go-composable-business-types/go.mod` line 15: `replace github.com/larsartmann/go-branded-id => ../go-branded-id`. This is a local dev hack that will break for anyone else cloning that repo. **Must be removed after v0.2.0 tag.**~~ done (removed during the 2026-05-20 ecosystem migration (v0.3.0 cycle))

2. ~~**No `go.sum` = no integrity verification** — While there are zero dependencies, the lack of `go.sum` means downstream consumers who expect it may have issues. (Low severity — `go mod tidy` generates it.)~~ **Won't implement — zero dependencies — absence is correct.**

3. ~~**Stale `coverage.out` in project root** — Leftover from this session's coverage analysis. Should be gitignored or removed.~~ done (file absent from the repo; *.out gitignored)

---

## e) WHAT WE SHOULD IMPROVE 🚀

### High Impact

1. ~~**Close coverage gaps** — Get to 90%+. `scanIntegerID` at 33.3% is embarrassing. Add tests for all integer type branches in `Scan`, `Value`, `UnmarshalText`, `UnmarshalBinary`.~~ done (coverage gaps closed across sessions; statement coverage is 88.5% (2026-09-22))
2. ~~**Tag v0.2.0** — Everything is ready. Just needs `git tag v0.2.0 && git push --tags`.~~ done (shipped inside v0.3.0 (044bd67))
3. ~~**Remove `replace` directive in consumer** — After v0.2.0 is tagged, remove `replace` from `go-composable-business-types/go.mod` and update to real `v0.2.0`.~~ done (replace removed and consumers migrated in the 2026-05-20 ecosystem pass)
4. ~~**Add `.gitignore` entry for `coverage.out`** — Prevent artifact commits.~~ done (*.out gitignored)

### Medium Impact

5. ~~**Add coverage threshold to CI** — Fail CI if coverage drops below 85%.~~ **Won't implement — FEATURES.md re-derives coverage every docs pass; a hardcoded threshold hides drift.**
6. ~~**Add `doc.go` with runnable examples** — Better pkg.go.dev experience.~~ **Won't implement — pkg.go.dev renders package docs; Example_ functions cover runnable docs.**
7. ~~**Add `CONTRIBUTING.md`** — Since README mentions it.~~ done (CONTRIBUTING.md rebuilt in v0.3.2 (ed5ee4b))
8. ~~**Add `SECURITY.md`** — Standard for open source libraries.~~ done (SECURITY.md created 2026-09-22)
9. ~~**Add release automation** — Tag-triggered GitHub Actions that creates a GitHub release.~~ done (release.yml creates GitHub Releases from tags with CHANGELOG-derived notes (a03780b))

### Low Impact

10. ~~**Add PR/issue templates** — Professional OSS polish.~~ **Won't implement — single-maintainer direct-push repo.**
11. ~~**Add `CODEOWNERS`** — If multiple contributors ever join.~~ **Won't implement — single maintainer.**
12. ~~**Explore `go:generate` for type-switch boilerplate** — The binary/SQL code has repetitive type switches that could be generated.~~ **Won't implement — type switches locked by exhaustive all-types tests; codegen adds more complexity than it removes.**

---

## f) Top #25 Things We Should Get Done Next

| #      | Priority                                                                                                      | Task                                                                      | Est. Effort |
| ------ | ------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- | ----------- |
| ~~1~~  | ~~P0~~ done — shipped inside v0.3.0 (044bd67)                                                                 | ~~Tag `v0.2.0` and push to remote~~                                       | ~~1 min~~   |
| ~~2~~  | ~~P0~~ done — in the 2026-05-20 ecosystem migration                                                           | ~~Remove `replace` directive from `go-composable-business-types/go.mod`~~ | ~~2 min~~   |
| ~~3~~  | ~~P0~~ done — in the 2026-05-20 ecosystem migration                                                           | ~~Update consumer to `v0.2.0` (remove local replace)~~                    | ~~2 min~~   |
| ~~4~~  | ~~P0~~ done — *.out in .gitignore                                                                             | ~~Add `coverage.out` to `.gitignore`~~                                    | ~~1 min~~   |
| ~~5~~  | ~~P0~~ done — file absent from the repo                                                                       | ~~Delete stale `coverage.out` from project root~~                         | ~~1 min~~   |
| ~~6~~  | ~~P1~~ done — coverage passes closed the gap (id_sql_test.go, id_alltypes_test.go)                            | ~~Add tests for `scanIntegerID` (33.3% → 90%+)~~                          | ~~30 min~~  |
| ~~7~~  | ~~P1~~ done — coverage passes closed the gap                                                                  | ~~Add tests for `UnmarshalText` error paths (65.6% → 90%+)~~              | ~~20 min~~  |
| ~~8~~  | ~~P1~~ done — coverage passes closed the gap                                                                  | ~~Add tests for `String()` `TextMarshaler` fallback path (66.7% → 90%+)~~ | ~~15 min~~  |
| ~~9~~  | ~~P1~~ done — coverage passes closed the gap                                                                  | ~~Add tests for `Value()` all int/uint types (70% → 90%+)~~               | ~~20 min~~  |
| ~~10~~ | ~~P1~~ done — coverage passes closed the gap                                                                  | ~~Add tests for `Scan()` all int/uint types (70.2% → 90%+)~~              | ~~20 min~~  |
| ~~11~~ | ~~P1~~ done — coverage passes closed the gap                                                                  | ~~Add tests for `UnmarshalBinary` error paths (78.7% → 90%+)~~            | ~~15 min~~  |
| ~~12~~ | ~~P1~~ done — id_test.go Format verb table covers %q, %#v and friends                                         | ~~Add tests for `Format` all verbs (80% → 95%+)~~                         | ~~15 min~~  |
| ~~13~~ | ~~P1~~ **Won't implement — FEATURES.md snapshot re-derives coverage every docs pass.**                        | ~~Add coverage threshold to CI (`go.yml`) — fail below 85%~~              | ~~5 min~~   |
| ~~14~~ | ~~P1~~ done — SECURITY.md created 2026-09-22                                                                  | ~~Add `SECURITY.md`~~                                                     | ~~10 min~~  |
| ~~15~~ | ~~P2~~ done — rebuilt in v0.3.2 (ed5ee4b)                                                                     | ~~Add `CONTRIBUTING.md`~~                                                 | ~~15 min~~  |
| ~~16~~ | ~~P2~~ **Won't implement — pkg.go.dev plus Example_ functions suffice.**                                      | ~~Add `doc.go` with package examples~~                                    | ~~10 min~~  |
| ~~17~~ | ~~P2~~ done — pkg.go.dev badge in README                                                                      | ~~Add pkg.go.dev badge to README~~                                        | ~~5 min~~   |
| ~~18~~ | ~~P2~~ done — release.yml ships CHANGELOG-derived notes (a03780b)                                             | ~~Add tag-triggered release GitHub Action~~                               | ~~30 min~~  |
| ~~19~~ | ~~P2~~ **Won't implement — serialization docs live on the website guides and pkg.go.dev; README stays lean.** | ~~Add `UnmarshalText` example to README~~                                 | ~~5 min~~   |
| ~~20~~ | ~~P2~~ **Won't implement — serialization docs live on the website guides and pkg.go.dev; README stays lean.** | ~~Add `Gob` example to README~~                                           | ~~5 min~~   |
| ~~21~~ | ~~P2~~ done — README Named Brand Types section demonstrates %#v                                               | ~~Add `Format` verb examples to README~~                                  | ~~5 min~~   |
| ~~22~~ | ~~P3~~ **Won't implement — single-maintainer direct-push repo.**                                              | ~~Add `.github/ISSUE_TEMPLATE/` (bug + feature)~~                         | ~~15 min~~  |
| ~~23~~ | ~~P3~~ **Won't implement — single-maintainer direct-push repo.**                                              | ~~Add `.github/PULL_REQUEST_TEMPLATE.md`~~                                | ~~10 min~~  |
| ~~24~~ | ~~P3~~ **Won't implement — source-only library; no binaries to reproduce.**                                   | ~~Add reproducible build flags to CI (`GOFLAGS=-trimpath`)~~              | ~~5 min~~   |
| ~~25~~ | ~~P3~~ **Won't implement — type switches locked by exhaustive tests; codegen not worth the complexity.**      | ~~Explore code generation for repetitive type-switch patterns~~           | ~~2 hr~~    |

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
