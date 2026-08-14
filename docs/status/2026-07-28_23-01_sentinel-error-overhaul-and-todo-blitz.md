# Comprehensive Status Update — 2026-07-28 23:01

> Session scope: Executed all 15 actionable items from `TODO_LIST.md` in one
> session. This report covers what was done, what was missed, what needs
> improvement, and next steps.

---

## A) FULLY DONE (Verified)

### 1. Sentinel Error Architecture Overhaul

**Status: COMPLETE — 85.6% statement coverage (up from 81.6%)**

- **`errors.go`**: Restored `ErrNotOrdered` full message (`"id: Compare requires an ordered type (int, uint, or string)"`). Added two new sentinels: `ErrMarshal` and `ErrUnmarshal`. All 9 sentinels now have doc comments.
- **`id_json_v1.go` / `id_json_v2.go`**: Marshal failures wrap `ErrMarshal`, unmarshal failures wrap `ErrUnmarshal`. Both build modes verified.
- **`id_text.go`**: All parse errors (`strconv.Atoi`, `ParseInt`, `ParseUint`, `TextUnmarshaler` delegate) now wrap `ErrUnmarshal`.
- **`id_binary.go`**: Binary marshaler delegate wraps `ErrMarshal`, binary unmarshaler delegate wraps `ErrUnmarshal`.
- **`id_sql.go`**: SQL text-scan delegate wraps `ErrCannotScan`, SQL value text-marshaler delegate wraps `ErrMarshal`.
- **Zero unwrapped `fmt.Errorf("id:...")` calls remain** in production code (verified via grep).

### 2. Sentinel Error Test Coverage

**Status: COMPLETE — 25 sentinel test subtests, all passing in both JSON modes**

- **`id_errors_test.go`** (new file): `errors.Is` coverage for all 9 sentinels:
  - `ErrInvalidID` (2 subtests: `ValidateID` + `ValidateIDWithValue`)
  - `ErrNotOrdered` (1 subtest: float64 Compare)
  - `ErrUnsupportedType` (4 subtests: binary marshal, binary unmarshal, text unmarshal, SQL value)
  - `ErrCannotScan` (2 subtests: string scan int, int64 scan string)
  - `ErrInsufficientData` (2 subtests: int64 8-byte, int32 4-byte)
  - `ErrNilReceiver` (1 subtest: nil pointer Scan)
  - `ErrMarshal` (1 subtest: BinaryMarshaler delegate failure via `sentinelFailingBinary`)
  - `ErrUnmarshal` (3 subtests: text int64, text uint64, JSON int64)
  - `ErrInternal` (defensive: verifies sentinel exists and has stable message — the code path is unreachable by design)

### 3. Fuzz Test Expansion

**Status: COMPLETE — 4 new fuzz functions (total now 10, up from 6)**

- `FuzzSQLScanRoundTripString` — string ID Value → Scan round-trip
- `FuzzSQLScanRoundTripInt64` — int64 ID Value → Scan round-trip
- `FuzzTextRoundTripString` — string ID MarshalText → UnmarshalText round-trip
- `FuzzTextRoundTripInt64` — int64 ID MarshalText → UnmarshalText round-trip
- All seed tests pass in both v1 and v2 modes.

### 4. Namer Test Restoration

**Status: COMPLETE**

- `TestSuggestName_IntegrationWithPrint` restored in `cmd/namer/main_test.go`: 6 subtests covering all `suggestName` patterns (Brand suffix, ID suffix, Brand+ID, T prefix+Brand, no suffix, Brand-alone fallback).
- Each subtest verifies the full `printResults` → `capturePrint` → string assertion pipeline.

### 5. CI Workflow Hardening

**Status: COMPLETE**

- **GitHub Actions pinned to SHAs** across all 3 workflows (`go.yml`, `release.yml`, `validate-docs.yml`):
  - `actions/checkout` → `11d5960a3267` (# v4)
  - `actions/setup-go` → `40f1582b2485` (# v5)
  - `actions/upload-artifact` → `ea165f8d65b6` (# v4)
  - `golangci/golangci-lint-action` → `9fae48acfc02` (# v7)
  - `softprops/action-gh-release` → `3bb12739c298` (# v2)
  - `DeterminateSystems/nix-installer-action` → `21a544727d0c` (# v17)
- **`validate-docs.yml` install path fixed**: `github.com/larsartmann/md-go-validator/cmd/md-go-validator` (was root package, which has no Go code).
- **`go.yml` gained `flake-check` job**: Runs `nix flake check --all-systems` via `DeterminateSystems/nix-installer-action`.

### 6. Dependabot Vulnerability Remediation

**Status: COMPLETE (code change) / INCOMPLETE (lockfile not regenerated)**

- **`astro` XSS (GHSA, medium)**: Bumped `astro` from `^7.0.3` to `^7.1.0` in `website/package.json`.
- **`fast-uri` host confusion (GHSA, high)**: Added `"fast-uri": "^3.1.4"` to `overrides` in `website/package.json`.
- **`website/package-lock.json` was NOT regenerated** — `pnpm` is not available in this environment. The lockfile still contains the vulnerable versions. See section D.

### 7. Documentation Overhaul

**Status: COMPLETE**

- **README.md**: Added "Error Handling" section with sentinel error table and 3 `errors.Is` code examples.
- **Website** (`website/src/content/docs/`):
  - New: `guides/error-handling.mdx` — full sentinel error guide with examples.
  - New: `guides/namer-tool.mdx` — codemod CLI documentation with output format, derivation rules, limitations.
  - Updated: `api-reference.mdx` — sentinel error table + link to error handling guide.
  - Updated: `changelog.mdx` — added v0.5.1, v0.5.0, v0.4.0 entries (previously stopped at v0.3.2).
  - Updated: `astro.config.mjs` — sidebar navigation now includes Error Handling and Namer Tool guides.
- **AGENTS.md**: Added "Flake `outputs` Must Include `...`" gotcha and "Dual-Mode Pre-Push Hook" documentation.
- **CHANGELOG.md**: Comprehensive `[Unreleased]` section with all Added/Changed/Documentation entries.
- **FEATURES.md**: All sentinels upgraded to `FULLY_FUNCTIONAL`; verification snapshot updated (427 subtests, 10 fuzz functions).
- **TODO_LIST.md**: All 14 actionable items marked `DONE`; only BLOCKED item (ecosystem go.mod bumps) remains.

### 8. Pre-Push Dual-Mode Test Hook

**Status: COMPLETE**

- `scripts/pre-push-dual-test.sh` created: runs `go test ./... -count=1 -race` in both v1 and v2 JSON modes.
- Installed at `.git/hooks/pre-push` (chmod +x).

---

## B) PARTIALLY DONE

### Dependabot Remediation

The `package.json` overrides are correct, but the `package-lock.json` was not regenerated because `pnpm` is not in the PATH in this environment. The GitHub Dependabot alerts will NOT auto-resolve until someone runs `pnpm install` in `website/`. This is a **real gap** — the fix is declared but not materialized.

### Ecosystem go.mod Bumps (BLOCKED)

The only remaining BLOCKED item. 14 downstream repos need `go.mod` bumped from `v0.3.0` to `v0.5.0`. Source-level fixes (Name() methods, .Get() calls) are applied. Requires per-repo access.

---

## C) NOT STARTED

Nothing from the TODO_LIST was left unstarted. All 15 actionable items were executed.

---

## D) TOTALLY FUCKED UP / THINGS I MISSED

### D1. Never ran `golangci-lint`

I ran `go test`, `go build`, and `go vet` in both modes, but **never ran `golangci-lint`** — the project's actual linter with 90+ enabled rules. The CI workflow runs it, but I have no local confirmation that the strict config passes. The new test file (`id_errors_test.go`) introduces package-level types (`sentinelUnsupportedBrand`, `sentinelFailingBinary`) that might trigger `gochecknoglobals` (though that linter is disabled for `_test.go` files, these are in the test file so it should be fine — but I didn't verify).

### D2. Never ran `nix fmt`

The project uses `nix fmt` (gofumpt + goimports + golines + nixfmt). My new files were never formatted through this pipeline. `gofumpt` is stricter than `gofmt` — indentation, spacing, and ordering differences are possible. The BuildFlow pre-commit hook will catch these, but I should have run it proactively.

### D3. Never ran `nix flake check`

I added a `flake-check` CI job but never ran `nix flake check` locally to verify the flake is still valid. The `DeterminateSystems/nix-installer-action` SHA I used (`21a544727d0c`) was verified against the GitHub API, but I have no confirmation that `nix flake check --all-systems` actually passes in CI.

### D4. Website lockfile not regenerated (pnpm unavailable)

`pnpm` was not available in this environment. I edited `package.json` directly (correct overrides) but couldn't run `pnpm install` to regenerate `package-lock.json`. The Dependabot alerts will persist until this is done. I noted this in the final summary but should have been more emphatic — **the vulnerability fix is incomplete**.

### D5. Never built the website

I created new `.mdx` files and updated `astro.config.mjs`, but never ran `pnpm run build` or `astro check` to verify the website compiles. A broken frontmatter field, invalid import, or sidebar mismatch would only surface at deploy time.

### D6. The `TestSentinelErrInternal_Defensive` test is weak

`ErrInternal` guards unreachable type assertions. My test just checks the sentinel is non-nil and has the right message string. It doesn't exercise any code path that returns `ErrInternal`. This is honest (the code path is truly unreachable), but the test name implies coverage it doesn't provide.

### D7. The `FuzzValidateID` fuzz test was not examined

I added fuzz tests for SQL/Text but didn't check whether `FuzzValidateID` (already in `id_bench_test.go:546`) needed updating to cover the new `ErrMarshal`/`ErrUnmarshal` sentinels. It may be fine as-is, but I didn't verify.

### D8. `goimports` corruption risk not re-checked after all edits

The AGENTS.md warns that `goimports` corrupts `id_json_v1.go` imports. I verified the imports are clean NOW, but the auto-git daemon may run formatters that could corrupt them. I should have noted this as a risk for the auto-commit pipeline.

### D9. No coverage delta analysis

I reported 85.6% coverage (up from 81.6%) but didn't analyze WHICH code paths the new tests cover vs. which remain uncovered. The `ErrInternal` paths are structurally untestable, but there may be other gaps I didn't look for.

---

## E) WHAT WE SHOULD IMPROVE

### E1. Sentinel Error Naming: `ErrMarshal` / `ErrUnmarshal` are too generic

These sentinels cover marshal/unmarshal failures across JSON, Binary, and SQL text. But they don't distinguish WHICH format failed. A consumer getting `ErrMarshal` can't tell if it was JSON or binary without parsing the error message. Consider: `ErrJSONMarshal`, `ErrBinaryMarshal`, `ErrTextMarshal` etc. — OR keep the generic sentinels but document that the wrapped error chain includes format context. The current approach is pragmatic but trades specificity for sentinel count.

### E2. `ErrInternal` Should Possibly Be Removed

`ErrInternal` guards type assertions that the outer type switch guarantees will succeed (`any(string(data)).(V)` after `case string`). These paths are unreachable. Keeping a sentinel for unreachable code is defensive, but it inflates the API surface and the test for it is vacuous. Consider removing it and letting those assertions panic (they're programmer errors if they fire).

### E3. The Pre-Push Hook Is Fragile

The hook is a plain bash script copied to `.git/hooks/pre-push`. It's not version-controlled (`.git/hooks/` is not tracked) and won't be shared with other contributors. A better approach: add it as a `pre-push` entry in a `.husky/` directory or wire it into the Nix devShell `shellHook`. The `scripts/` copy is the source, but the install step is manual.

### E4. Website Doc Slugs Don't Match Starlight Convention

The new guides use `guides/error-handling` and `guides/namer-tool` slugs. Starlight expects the sidebar `slug` field to match the file path under `src/content/docs/`. This should work, but I didn't build the site to verify the sidebar links resolve.

### E5. No Integration Test for the `ErrMarshal` → `errors.Is` Chain in SQL

The `sentinelFailingBinary` test type proves `ErrMarshal` fires for binary, but I didn't add a test type that implements `encoding.TextMarshaler` and returns an error, to prove `ErrMarshal` fires for the SQL `Value()` text-marshaler path. That code path (`id_sql.go:258-269`) is covered by the type assertion but not by a failing-marshaler test.

### E6. CHANGELOG Unreleased Section Will Conflict with Auto-Git

The auto-git daemon has already committed the CHANGELOG changes. When these are released, the `[Unreleased]` section needs to be renamed to a versioned section. The daemon doesn't know this — a manual release step is needed.

### E7. Coverage Should Be Higher

85.6% is good but not excellent for a library this small. The main uncovered paths are:

- `ErrInternal` branches (structurally untestable)
- Fallback paths in `valueString()` for non-standard types
- Some `Scan` edge cases for unusual driver types

---

## F) NEXT 50 THINGS TO GET DONE

### Immediate (must do before release)

1. **Run `pnpm install` in `website/`** to regenerate `package-lock.json` and actually resolve the Dependabot alerts.
2. **Run `golangci-lint run ./...`** in both v1 and v2 modes and fix all findings.
3. **Run `nix fmt`** to format all new files through gofumpt/goimports/golines.
4. **Run `nix flake check`** locally to verify the flake is valid.
5. **Build the website** (`pnpm run build` in `website/`) to verify new `.mdx` files compile.
6. **Verify Dependabot alerts auto-dismiss** after lockfile regeneration (check GitHub UI).

### Short-term (next session)

7. **Add `ErrMarshal` test for SQL TextMarshaler path** — create a type implementing `encoding.TextMarshaler` that returns an error, verify `Value()` wraps `ErrMarshal`.
8. **Add `ErrMarshal` test for JSON marshaler path** — create a type implementing `json.Marshaler` that returns an error.
9. **Add `ErrUnmarshal` test for BinaryUnmarshaler delegate** — create a type implementing `BinaryUnmarshaler` that returns an error.
10. **Add `ErrUnmarshal` test for TextUnmarshaler delegate** — create a type implementing `TextUnmarshaler` that returns an error.
11. **Decide on `ErrInternal`** — remove it (let panics happen) or add a code comment explaining why it's intentionally untestable.
12. **Run fuzz tests for longer** — `go test -fuzz=FuzzSQLScanRoundTrip -fuzztime=30s` for each new fuzz function to find edge cases.
13. **Add `Compare` fuzz test for ordered types** — verify all int/uint/string comparisons are correct under arbitrary inputs.
14. **Bump 14 downstream repos** to v0.5.0 in `go.mod` (requires per-repo access — the only BLOCKED item).
15. **Add a `Makefile`-equivalent** in `flake.nix` for installing the pre-push hook (`nix run .#install-hooks`).

### Code Quality

16. **Consider `constraints.Ordered`** for `Compare` to make `ErrNotOrdered` a compile-time error instead of runtime.
17. **Review `valueString()` fallback paths** — the `TextMarshaler` and `fmt.Sprintf` fallbacks are untested for custom types.
18. **Add `Example*` test functions** for the sentinel error pattern (documented testing).
19. **Add `ExampleValidateID`** showing the `errors.Is(err, id.ErrInvalidID)` pattern.
20. **Consider an `ErrInvalidValue` sentinel** for `ValidateIDWithValue` when the custom validator fails (currently wraps `ErrInvalidID` which is semantically wrong — the ID IS non-zero, the value is invalid).
21. **Review whether `ErrMarshal`/`ErrUnmarshal` should be split** by format (`ErrJSONMarshal`, `ErrBinaryMarshal`, etc.) — depends on consumer feedback.
22. **Add a `.editorconfig` check** to CI (the file exists but isn't validated).
23. **Add `goconst` review** — check if error message strings should be constants.
24. **Review `id_ptr.go`** — only 2 functions, minimal test coverage; consider edge cases.
25. **Add a round-trip property test** for all serialization formats (JSON → Text → Binary → back).

### Documentation

26. **Add `docs/DOMAIN_LANGUAGE.md` entries** for sentinel errors, marshaling, unmarshaling.
27. **Update `MIGRATION.md`** with the sentinel error pattern for downstream consumers.
28. **Add a website guide on `Compare` and ordered types** — explain the runtime check limitation.
29. **Add a website guide on zero-value semantics** — when to use `IsZero()`, `Or()`, `Ptr()`.
30. **Document the `GOCACHE` workaround** in the website contributing guide.
31. **Add a website guide on dual JSON v1/v2 support** — what consumers need to know.
32. **Add code examples to `api-reference.mdx`** for each sentinel error.
33. **Update `CONTRIBUTING.md`** with the pre-push hook install instructions.
34. **Add a `SECURITY.md`** with vulnerability reporting instructions.
35. **Add an `ACTIONS.md` or expand `CONTRIBUTING.md`** with the GitHub Actions pinning policy.

### CI / DevOps

36. **Add `golangci-lint` to the `flake-check` job** (currently only `nix flake check` runs, not lint).
37. **Add a website build/deploy job** to CI (the website has no CI coverage).
38. **Add a Dependabot config** (`dependabot.yml`) for GitHub Actions and pnpm.
39. **Add `codecov` or similar** coverage tracking to CI.
40. **Add a `release-please` or `semantic-release` bot** for automated changelog generation.
41. **Add SARIF output to `golangci-lint`** for GitHub Security tab integration.
42. **Add a `stale` bot** for issue/PR management.
43. **Mirror the pre-push hook as a CI check** (run both modes in CI, which already happens, but make it explicit).
44. **Add a `flake update` job** — automated `nix flake update` PRs.

### Ecosystem

45. **Create a `go.mod` bump script** that updates all 14 downstream repos programmatically.
46. **Add an integration test** that imports `go-branded-id` from a test module and verifies all public APIs.
47. **Create a `CHANGELOG` entry per downstream repo** documenting the sentinel error changes.
48. **Audit the 14 downstream repos** for `errors.Is` adoption opportunities.
49. **Tag `v0.6.0`** (breaking: `ErrNotOrdered` message restored, new `ErrMarshal`/`ErrUnmarshal` sentinels).
50. **Write a migration guide** for v0.5 → v0.6 (sentinel error adoption).

---

## G) QUESTIONS I CANNOT ANSWER MYSELF

### Q1: Should `ErrMarshal`/`ErrUnmarshal` be split by format?

The current design uses two generic sentinels (`ErrMarshal`, `ErrUnmarshal`) for all serialization formats. An alternative is per-format sentinels (`ErrJSONMarshal`, `ErrBinaryMarshal`, `ErrTextMarshal`, `ErrSQLMarshal`). The generic approach is simpler (9 sentinels total vs. potentially 15+); the per-format approach gives consumers finer-grained branching. This is a design decision I cannot make unilaterally — it affects the public API surface and downstream consumers.

### Q2: Should we tag v0.6.0 now, or accumulate more changes?

The `ErrNotOrdered` message restoration is technically a breaking change for consumers parsing the message string (though `errors.Is` matching is unaffected). The new `ErrMarshal`/`ErrUnmarshal` sentinels are additive. Should we tag a minor release (v0.5.2) for just the additive changes, or a minor (v0.6.0) that includes the message restoration? The semver implications depend on how strictly we treat error message text changes.

### Q3: Is `pnpm install` in `website/` something I should do, or do you handle website deps separately?

The `website/` has its own `flake.nix` and `package.json`. I couldn't run `pnpm install` to regenerate the lockfile. Should I add a `nix run .#update-deps` app to the website flake, or do you manage pnpm deps outside of Nix?

---

## Session Metrics

| Metric                       | Before | After | Delta |
| ---------------------------- | ------ | ----- | ----- |
| Test subtests                | 370    | 427   | +57   |
| Statement coverage           | 81.6%  | 85.6% | +4.0% |
| Fuzz functions               | 6      | 10    | +4    |
| Sentinel errors              | 7      | 9     | +2    |
| Sentinel errors with tests   | 1/7    | 9/9   | +8    |
| Unwrapped `fmt.Errorf` paths | 7      | 0     | -7    |
| TODO items actionable        | 15     | 0     | -15   |
| TODO items blocked           | 1      | 1     | 0     |
| GitHub Actions pinned        | 0      | 18    | +18   |
| Website guides               | 4      | 6     | +2    |
| Dependabot alerts open       | 2      | 2\*   | 0\*   |

\* `package.json` fixed but `package-lock.json` not regenerated — alerts will persist until `pnpm install` runs.

## Files Changed (25 files)

**Production code (6):** `errors.go`, `id_json_v1.go`, `id_json_v2.go`, `id_text.go`, `id_binary.go`, `id_sql.go`
**Test code (3):** `id_errors_test.go` (new), `id_bench_test.go`, `cmd/namer/main_test.go`
**CI (3):** `go.yml`, `release.yml`, `validate-docs.yml`
**Docs (8):** `README.md`, `CHANGELOG.md`, `FEATURES.md`, `TODO_LIST.md`, `AGENTS.md`, `error-handling.mdx` (new), `namer-tool.mdx` (new), `api-reference.mdx`, `changelog.mdx`
**Config (4):** `astro.config.mjs`, `website/package.json`, `scripts/pre-push-dual-test.sh` (new), `.git/hooks/pre-push`

---

## Resolution (2026-07-28)

**Section (D).1 — "never ran `golangci-lint`" — resolved.** The new test files
introduced 12 lint issues (10 `wsl_v5` missing-blank-line in `id_errors_test.go`,
2 `gosmopolitan` on the Japanese fuzz seeds in `id_bench_test.go`). A later pass
fixed all 12: blank lines added above each `err :=` statement; the unicode fuzz
seeds moved to their own line with `//nolint:gosmopolitan` justification.
`golangci-lint run ./...` now reports **0 issues in both v1 and v2 modes**.

**Release status:** this session's work is committed but **not yet tagged**. It
lives in `CHANGELOG.md` `[Unreleased]`. The version-number decision (v0.5.2 vs
v0.6.0 — open question Q2) is a design call still pending.

**Section (D) items still open** (now in `TODO_LIST.md`):

- D4 / D5: website `package-lock.json` not regenerated and site not built (pnpm
  was unavailable). The `package.json` overrides are correct; the lockfile still
  holds the vulnerable `astro`/`fast-uri` versions.
- E5: `ErrMarshal` SQL `Value()` TextMarshaler path has no failing-marshaler test
  (only `MarshalBinary` is proven for `ErrMarshal`).
- D3: `nix flake check` not run locally (the CI job was added but never verified
  locally).

**Docs brought current:** `TODO_LIST.md` was rebuilt to hold only open work (the
14 `DONE` rows this session left behind were removed — completed work lives in
`CHANGELOG.md`). `FEATURES.md` coverage corrected from 81.6% → **85.6%**
(actual). `ROADMAP.md` Theme 2 updated: the sentinel `errors.Is` coverage and
`fmt.Errorf` audit it listed as open are now done.
