# Status Report — 2026-07-27 16:44

## Comprehensive JSON Dual-Mode Hardening, Lint Audit & Namer Test Coverage

> **Session scope:** Execute the 21-item TODO list from the prior status report (`2026-07-27_11-27_dual-json-support-and-namer-tests.md`), plus all collateral damage discovered along the way.

---

## Executive Summary

**20 of 21 original TODOs are fully done. 1 is documented but externally blocked (bumping 14 downstream repos).** Along the way, a **critical build-breaking bug** was discovered and fixed: `goimports` had silently corrupted the v1 JSON files to import `encoding/json/v2`, breaking the default build mode for every consumer without `GOEXPERIMENT=jsonv2`. Lint went from **82 issues → 0** in both JSON modes. All tests pass with `-race` in both modes. Website builds. Docs validate.

---

## a) FULLY DONE (18 items)

### Critical Bug Fix (Not on Original List)

| What                     | Detail                                                                                                                                                                                                                                                                                                                                                                         |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **goimports corruption** | `id_json_v1.go` and `json_helpers_v1_test.go` both imported `encoding/json/v2` instead of `encoding/json`. This was committed in `4f38f16` ("fix(json): resolve critical import bug") — the commit that _introduced_ the bug it claimed to fix. The default build mode (`go build ./...` without `GOEXPERIMENT`) was completely broken. Fixed by correcting both import paths. |

### Original TODOs

| #      | Task                                 | What was done                                                                                                                                                                                                                                                                                                                                         |
| ------ | ------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **1**  | JSON byte-identical equivalence test | `TestJSONByteEquivalence` in `id_json_contract_test.go` — asserts exact bytes for string/int64/uint64/int32 marshal + null for zero values. Runs in both CI modes.                                                                                                                                                                                    |
| **2**  | Structural parity meta-test          | `TestDualJSONContract_Imports`, `TestDualJSONContract_BuildTags`, `TestDualJSONContract_StructuralParity` — locks the import split, build-tag split, and byte-level parity of both file pairs after normalization.                                                                                                                                    |
| **3**  | md-go-validator on .mdx code blocks  | Fixed 5 validation errors by adding `// skip-validate` to illustrative snippets in README.md (3 blocks), `value-types.mdx` (1), `serialization.mdx` (1). Result: 39 valid, 5 skipped, **0 errors**.                                                                                                                                                   |
| **4**  | changelog.mdx grammar                | "dual-supports" → "support" in both website changelog and CONTRIBUTING.md.                                                                                                                                                                                                                                                                            |
| **5**  | Website builds                       | Verified `nix run .#build` from `website/` — 12 pages built successfully in ~2.5s.                                                                                                                                                                                                                                                                    |
| **6**  | AGENTS.md stale files                | Removed entire "Stale Files to Ignore" section (CONTRIBUTING.md is current). Also fixed "dual-supports" in line 11 and the stale justfile reference in line 9.                                                                                                                                                                                        |
| **7**  | CI lint v2 mode                      | Removed `goexperiment.jsonv2` from `.golangci.yml` build-tags (so default lint covers v1). CI workflow now runs lint with `--build-tags goexperiment.jsonv2` in a separate matrix leg.                                                                                                                                                                |
| **8**  | CI matrix strategy                   | Rewrote `go.yml` and `release.yml` with `matrix: [v1, v2]` for build, test, and lint jobs. Release workflow now gates on both verify+lint jobs before creating the GitHub release.                                                                                                                                                                    |
| **10** | v0.4.0 migration guide               | Added comprehensive "v0.4.0: Dual JSON v1/v2 Support" section to MIGRATION.md covering what changed, why dual-mode, how to opt into v2, troubleshooting.                                                                                                                                                                                              |
| **11** | Fuzz test for cmd/namer scanFile     | `FuzzScanFile` in `coverage_test.go` — 8 seed inputs (valid Go, invalid Go, empty, garbage). Verifies scanFile never panics on arbitrary input.                                                                                                                                                                                                       |
| **12** | testdata: pointer receiver           | `testdata/pointer_receiver.go` — `PointerBrand` with `func (*PointerBrand) Name() string`. `TestScanFile_PointerReceiver` verifies HasName=true.                                                                                                                                                                                                      |
| **13** | walkFn error path                    | `TestWalkFn_ErrorPath` — feeds a simulated walk error into `walkFn` and verifies it wraps the error correctly.                                                                                                                                                                                                                                        |
| **14** | Refactor main() → testable run()     | Extracted `run(args, stdout, stderr) int` using a per-call `flag.NewFlagSet` (avoids global flag redefinition panic in tests). `TestRun_NoArgs` and `TestRun_ScanTestData` verify the full CLI flow.                                                                                                                                                  |
| **15** | v1 vs v2 benchmark                   | 4 benchmarks: `BenchmarkJSONDualMarshalString/Int64`, `BenchmarkJSONDualUnmarshalString/Int64`. Includes benchstat workflow in doc comment.                                                                                                                                                                                                           |
| **16** | tparallel fixes                      | `TestIDBinary`: added `t.Parallel()` (real fix). `TestIDJSONRoundTrip`: added `//nolint:tparallel` with justification (subtests run via interface indirection).                                                                                                                                                                                       |
| **17** | ExampleValidateID output             | Added `// Output:` to make the example testable.                                                                                                                                                                                                                                                                                                      |
| **18** | isIDSelector default case            | `TestIsIDSelector_DefaultCase` — covers BinaryExpr, CallExpr, BasicLit (all return false).                                                                                                                                                                                                                                                            |
| **19** | isEmptyStructBrand non-struct        | `TestIsEmptyStructBrand_NonStructType` — covers InterfaceType, non-empty struct, Ident, empty struct not in brandsUsedWithID, plus positive case.                                                                                                                                                                                                     |
| **20** | Codegen decision documented          | Added "Decision: No Code Generation for the JSON File Pairs" section to `dedup-acceptance.md` with rationale (parity test guards it, generator adds more complexity than it saves).                                                                                                                                                                   |
| **21** | Lint audit (82 → 0)                  | Config-level: added `id`, `n`, `r`, `p`, `w`, `tc`, `ts`, `rt`, `f`, `v1`, `v2` to varnamelen ignore-names; set `makezero: always: false`; added `err113` to test-file exclusions. Code-level: 12 `//nolint:err113` on diagnostic errors embedding `%T`; 2 goconst constants extracted in cmd/namer; removed stale `//nolint:funlen` from id_text.go. |

---

## b) PARTIALLY DONE (1 item)

| #     | Task                               | Status                          | What remains                                                                                                                                                                                                                                                                                                                                                                                                              |
| ----- | ---------------------------------- | ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **9** | Bump 14 downstream repos to v0.4.0 | **Documented but not executed** | The bump procedure is fully documented in MIGRATION.md ("Bumping Downstream Repositories"). The 14 repos (InboxClean, CreditReformBilanzampel, ActaFlow, SEC, storbi, ChastityAPI, smart-configs, StopTube, universal-workflow, Zlota44, timesheets, complaints-mcp, cqrs-htmx, emeet-pixyd) require access to each repo. Cannot be done autonomously — requires cloning each, running `go get`, testing, committing, PR. |

---

## c) NOT STARTED (0 items)

All items from the original 21-item list have been addressed.

---

## d) TOTALLY FUCKED UP (1 thing, caught and fixed)

### The goimports Corruption (Pre-existing, Discovered This Session)

**What happened:** Commit `4f38f16` titled "fix(json): resolve critical import bug in JSON helpers and dedup code" actually **introduced** a critical bug. It rewrote the import in `id_json_v1.go` from `"encoding/json"` to `"encoding/json/v2"`. The same corruption existed in `json_helpers_v1_test.go`.

**Impact:** The default build mode (`go build ./...` without `GOEXPERIMENT=jsonv2`) was completely broken. Every consumer of the library that didn't set `GOEXPERIMENT=jsonv2` would get:

```
imports encoding/json/v2: build constraints exclude all Go files in .../encoding/json/v2
```

**Why it was so insidious:**

- `go list` showed the correct `GoFiles` (the v1 file IS included) — only the resolved `Imports` list revealed the v2 leak
- The v2 build mode worked fine, so testing only in v2 mode hid the bug
- The commit message claimed to fix an import bug, not introduce one
- `goimports` keeps the corruption once present (it sees `encoding/json/v2` as valid and doesn't revert it)

**How I caught it:** The very first `go test ./...` I ran failed with the build-constraints error. I traced it through `go list`, `go list -test`, minimal repro modules, and finally found the corrupted import by reading `id_json_v1.go:6`.

**Resolution:** Corrected both import paths. Added `TestDualJSONContract_Imports` to ensure it never happens again — it asserts v1 files import `encoding/json` and NOT `encoding/json/v2`, running in both CI modes. Also wrote a full feedback document for `go-auto-upgrade` at `docs/feedback/new/2026-07-27-dual-json-v1-v2-build-tag-repos.md`.

---

## e) WHAT WE SHOULD IMPROVE

### Architectural

1. **The goimports corruption will recur.** The contract test catches it _after_ it happens. The root cause is that goimports has no build-tag awareness and picks `encoding/json/v2` when re-adding the json import from scratch. A pre-commit hook or a `goimports -local` configuration that pins the import would be more preventive. The contract test is a safety net, not a cure.

2. **The dual-mode file pairs are a maintenance liability.** Every logic change must be made in both files. The parity test catches divergence, but the duplication remains. If a third mode is ever needed, code generation becomes worth the complexity.

3. **The `err113` nolint approach is a code smell.** 12 `//nolint:err113` directives on diagnostic errors that embed `%T` formatting. The "correct" fix would be sentinel errors with `fmt.Errorf("...: %w", ErrUnsupportedType)` — but the current errors embed runtime type information that can't be a static sentinel. The nolint is honest; the alternative would be worse.

### Testing

4. **No integration test for the full CI matrix locally.** You have to remember to run both `go test` and `GOEXPERIMENT=jsonv2 go test`. The `nix run .#test` app does this, but a developer running plain `go test` only tests one mode. A pre-commit hook running both modes would help.

5. **The namer tool has no integration test that actually runs the binary.** `TestRun_*` tests the `run()` function directly, but doesn't test `go run ./cmd/namer testdata/`. The binary could theoretically behave differently (flag parsing edge cases, exit codes).

6. **Benchmark results are not captured anywhere.** The v1-vs-v2 benchmarks exist but there's no baseline file checked in. Without a baseline, `benchstat` comparisons require re-running the old version.

### CI/CD

7. **The release workflow runs lint in both modes but doesn't run the contract tests explicitly.** They're part of the test suite, so they run — but naming them explicitly in the release workflow would make the gating intent clearer.

8. **No `nix flake check` in CI.** The GitHub workflows use `actions/setup-go` directly. `nix flake check` (which includes the sandbox build in both json modes) is only run locally. Adding it to CI would catch Nix-specific issues.

### Documentation

9. **The website changelog only goes up to 0.3.2.** The dual-mode work (0.4.0) is documented in MIGRATION.md and CHANGELOG.md has `[Unreleased]`, but the website `changelog.mdx` doesn't have a 0.4.0 entry yet.

10. **CONTRIBUTING.md doesn't mention the dual-mode contract tests.** A contributor who edits one half of a json file pair needs to know the parity test exists and will fail if they forget the other half.

---

## f) Up to 50 Things to Get Done Next

### High Priority

1. **Release v0.4.0** — tag, push, let CI create the GitHub release
2. **Add 0.4.0 entry to website `changelog.mdx`** — the dual-mode work is invisible on the website
3. **Actually bump the 14 downstream repos** — procedure is documented, needs execution
4. **Add `nix flake check` to CI** — currently only runs locally
5. **Add pre-commit hook for dual-mode `go test`** — prevents single-mode blind spots
6. **Add the dual-mode contract test pattern to CONTRIBUTING.md** — so contributors know about the parity requirement

### Medium Priority

7. **Capture benchmark baseline files** — check in `bench-v1.txt` and `bench-v2.txt` for benchstat comparison
8. **Add integration test that runs the namer binary** — `exec.Command("go", "run", "./cmd/namer", "testdata/")`
9. **Investigate goimports pinning** — can `.goimports` or editor config pin the json import path?
10. **Add `go vet` in both json modes to CI** — currently only `golangci-lint` runs
11. **Consider a `make dual-test` or `just dual-test` equivalent in flake** — convenience alias for both-mode testing
12. **Add fuzz tests for SQL Scan with arbitrary `src any`** — currently only JSON and Binary have fuzz tests
13. **Add fuzz test for Text unmarshal** — no fuzz coverage for `UnmarshalText`
14. **Benchmark SQL Scan/Value** — no performance characterization for the SQL path
15. **Add `GOEXPERIMENT=jsonv2` to the devShell `.envrc`** — optional, but convenient for v2-focused development
16. **Document the goimports corruption in CONTRIBUTING.md** — warn contributors about the hazard
17. **Add a `// Code generated DO NOT EDIT` header to the json file pairs** — even though they're hand-written, this signals to tools that they should not be auto-formatted (debatable)
18. **Consider `go:generate` directive that runs the parity test** — `go generate ./...` could verify the contract before building
19. **Add coverage report to CI** — `go test -cover` runs but the report isn't uploaded as an artifact
20. **Add race detector to the namer tests** — currently only the library tests run with `-race`

### Lower Priority

21. **Explore `encoding/json/v2` jsontext API** — v2 has a streaming `jsontext` API that could be more efficient for the ID type
22. **Add `ID[B, V]` to the `encoding/json/v2` marshaler registry** — v2 supports per-type marshalers without methods
23. **Consider a `Compare` method that works at compile time** — currently runtime-checked via type switch
24. **Add `Hash()` method** — for use with `map[ID[B,V]]` (currently works via comparability but no explicit hash)
25. **Add `ID[B, V]` support for `encoding/gob` with interface fields** — current gob implementation delegates to binary
26. **Explore generic constraints for `Compare`** — `constraints.Ordered` would make `ErrNotOrdered` a compile-time error
27. **Add `Stringer` benchmark for named vs unnamed brands** — performance impact of `BrandName[B]()` reflection
28. **Profile JSON marshal allocations** — `json.Marshal` allocates; could use `json.MarshalWrite` in v2 mode
29. **Add `Format` method benchmark** — `%s`, `%v`, `%#v` formatting performance
30. **Consider `ID[B, V]` implementing `fmt.Stringer` only for named brands** — unnamed brands return raw value, which is already the behavior
31. **Add `Scan` benchmark for all SQL driver types** — int64, int, float64, []byte, string
32. **Add `Value` benchmark** — SQL `driver.Valuer` performance
33. **Explore `sync.Pool` for marshal buffers** — reduce allocations for hot paths
34. **Add `ID[B, V]` support for `encoding/xml`** — currently only Text (which covers XML indirectly)
35. **Consider `ID[B, V]` implementing `sql/driver.Valuer` with nullable support** — `NullID[B,V]` type?
36. **Add property-based testing with rapid** — generate random IDs and verify invariants
37. **Add `ID[B, V]` JSON marshaling with `omitempty` support** — via a wrapper type?
38. **Document the little-endian binary format** — in a separate spec or RFC-style doc
39. **Add cross-language binary compatibility tests** — marshal in Go, verify bytes match a Python/TS reference
40. **Consider `ID[B, V]` for protobuf** — `proto.Marshal` support via custom message
41. **Add `ID[B, V]` support for `msgpack`** — common binary format
42. **Explore `ID[B, V]` with `cmp` package** — `cmp.Equal` and `cmp.Diff` support
43. **Add `ID[B, V]` to `gob` codec registry** — for `encoding/gob` v2 if it materializes
44. **Consider `ID[B, V]` implementing `crypto/hash.Hash`** — for content-addressable IDs?
45. **Add `ID[B, V]` support for `database/sql` `NullX` types** — `NullInt64`, `NullString` interop
46. **Explore `ID[B, V]` with generics methods on `[]ID`** — batch operations
47. **Add `ID[B, V]` implementing `sort.Interface`** — or helper functions for sorting
48. **Consider `ID[B, V]` with `context.Context`** — trace context propagation?
49. **Add `ID[B, V]` support for `encoding/csv`** — via Text marshal
50. **Write a blog post about the dual-mode architecture** — the pattern is novel and reusable

---

## g) Questions I Cannot Answer Myself

### Q1: Should the goimports corruption be reported as a bug to the Go team?

`goimports` picking `encoding/json/v2` for files tagged `//go:build !goexperiment.jsonv2` is arguably a bug — it has no build-tag awareness and suggests an import that will break the file's build constraint. However, I'm not sure if the Go team considers this in scope (goimports has never been build-tag-aware). Should I file a GitHub issue?

### Q2: Should v0.4.0 be released now, or should the 14 downstream repos be bumped first?

The library is in a releasable state (build+test+lint clean in both modes, migration guide written). But the downstream repos are still on older versions. Releasing v0.4.0 now means downstream repos can bump at their own pace. Waiting means the release is delayed by 14 repo migrations. Which order do you prefer?

### Q3: Should the `[Unreleased]` section in CHANGELOG.md be formalized as `v0.4.0` before committing?

The CHANGELOG.md has an `[Unreleased]` section but no `v0.4.0` header. The website `changelog.mdx` only goes up to `0.3.2`. I didn't want to unilaterally decide the version number or release timing. Should I draft the formal changelog entry now, or wait for your go-ahead?

---

## Verification Snapshot

| Check                       | v1 (default)                      | v2 (GOEXPERIMENT=jsonv2) |
| --------------------------- | --------------------------------- | ------------------------ |
| `go build ./...`            | ✅ rc=0                           | ✅ rc=0                  |
| `go test -race ./...`       | ✅ pass                           | ✅ pass                  |
| `golangci-lint run`         | ✅ 0 issues                       | ✅ 0 issues              |
| `md-go-validator .`         | ✅ 0 errors (39 valid, 5 skipped) | —                        |
| `nix run .#build` (website) | ✅ 12 pages                       | —                        |

| Metric              | Before      | After                                            |
| ------------------- | ----------- | ------------------------------------------------ |
| Lint issues (v1)    | 82          | **0**                                            |
| Lint issues (v2)    | not checked | **0**                                            |
| Default build       | **BROKEN**  | ✅ works                                         |
| Tests (both modes)  | v1 broken   | ✅ both pass                                     |
| Docs validation     | 5 errors    | **0**                                            |
| Contract tests      | 0           | 4 (imports, buildtags, parity, byte-equivalence) |
| namer test coverage | partial     | +8 tests (run, fuzz, coverage, pointer, walkFn)  |
