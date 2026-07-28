# TODO List — go-branded-id

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## High Impact

| Task                                                  | Status    | Impact | Evidence                                                                                                  |
| ----------------------------------------------------- | --------- | ------ | --------------------------------------------------------------------------------------------------------- |
| Add `errors.Is` tests for sentinel errors             | 🟢 `DONE` | High   | All 9 sentinels covered in `id_errors_test.go`. `ErrMarshal` and `ErrUnmarshal` also added.               |
| Fix Validate Docs CI failure                          | 🟢 `DONE` | High   | Install path corrected to `cmd/md-go-validator` in `validate-docs.yml`.                                   |
| Remove tracked `namer` binary and add to `.gitignore` | 🟢 `DONE` | High   | Binary removed in `c29a034` and `/namer` is gitignored. Released in v0.5.1.                               |
| Investigate Dependabot vulnerabilities                | 🟢 `DONE` | High   | 2 alerts in website npm deps: astro XSS (→ `^7.1.0`), fast-uri host confusion (→ override `^3.1.4`). Fixed. |

## Medium Impact

| Task                                                          | Status       | Impact | Evidence                                                                                                                       |
| ------------------------------------------------------------- | ------------ | ------ | ------------------------------------------------------------------------------------------------------------------------------ |
| Bump 14 downstream ecosystem repos to v0.5.0                  | 🔵 `BLOCKED` | Med    | v0.5.0 tagged and pushed. Source fixes applied to all repos. `go.mod` bump not done. Requires per-repo access. See ecosystem. |
| Document sentinel errors in README                            | 🟢 `DONE`    | Med    | README "Error Handling" section with sentinel table and `errors.Is` examples added.                                            |
| Add `cmd/namer` and sentinel errors to website docs           | 🟢 `DONE`    | Med    | `guides/namer-tool.mdx`, `guides/error-handling.mdx`, and API reference sentinel table added.                                  |
| Add website `changelog.mdx` entries for v0.4.0 and v0.5.0     | 🟢 `DONE`    | Med    | Website changelog now includes v0.5.1, v0.5.0, and v0.4.0 entries.                                                             |
| Pin GitHub Actions to SHA hashes                              | 🟢 `DONE`    | Med    | All actions in `go.yml`, `release.yml`, `validate-docs.yml` pinned to commit SHAs with `# vX` comments.                        |
| Add `nix flake check` to CI workflows                         | 🟢 `DONE`    | Med    | `flake-check` job added to `go.yml` using `DeterminateSystems/nix-installer-action`.                                           |
| Restore `ErrNotOrdered` full message                          | 🟢 `DONE`    | Med    | Sentinel restored to `"id: Compare requires an ordered type (int, uint, or string)"` in `errors.go:14`.                        |
| Review remaining `fmt.Errorf` calls without sentinel wrapping | 🟢 `DONE`    | Med    | All marshal/unmarshal paths now wrap `ErrMarshal`/`ErrUnmarshal`/`ErrCannotScan`. Zero unwrapped error paths remain.           |

## Low Impact

| Task                                                    | Status    | Impact | Evidence                                                                                                        |
| ------------------------------------------------------- | --------- | ------ | --------------------------------------------------------------------------------------------------------------- |
| Restore `TestSuggestName_IntegrationWithPrint`          | 🟢 `DONE` | Low    | Comprehensive multi-case test in `cmd/namer/main_test.go` covering all suffix patterns.                         |
| Add `printResults` test asserting suggested name string | 🟢 `DONE` | Low    | Covered by `TestSuggestName_IntegrationWithPrint` (6 cases) and `TestPrintResults_SuggestNameIntegration`.      |
| Add fuzz tests for SQL `Scan` and Text `UnmarshalText`  | 🟢 `DONE` | Low    | `FuzzSQLScanRoundTripString/Int64`, `FuzzTextRoundTripString/Int64` in `id_bench_test.go`.                      |
| Add `outputs` pattern gotcha to AGENTS.md               | 🟢 `DONE` | Low    | "Flake `outputs` Must Include `...`" section added to AGENTS.md Critical Gotchas.                               |
| Add pre-commit hook for dual-mode `go test`             | 🟢 `DONE` | Low    | `scripts/pre-push-dual-test.sh` installed as `.git/hooks/pre-push`. Documented in AGENTS.md.                    |

---

## Ecosystem tracking

This library has 14 downstream repos. The v0.3.0/v0.3.1 source changes (`Name()`
methods added, `.String()` → `.Get()` fixes, test fixes) are **applied and pushed**
to all of them. The **`go.mod` dependency bump to v0.5.0 is not yet done**
in any repo.

Downstream repos: InboxClean, CreditReformBilanzampel, ActaFlow, SEC, storbi,
ChastityAPI, smart-configs, StopTube, universal-workflow, Zlota44, timesheets,
complaints-mcp (archived), cqrs-htmx, emeet-pixyd.

Deliberately **not** changed (correct as-is): go-cqrs-lite (marker types — see
AGENTS.md "Brands That Deliberately Skip `Name()`"), BerryBig (test brands
only), Cyberdom (no brand types).

Pre-existing test failures **not** caused by this library: CreditReformBilanzampel
(BDD undefined step), timesheets (fuzz hours overflow), emeet-pixyd (PipeWire
state file), Zlota44 (internal/discovery).
