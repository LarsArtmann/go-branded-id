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

| Task                                                  | Status    | Impact | Effort | Evidence                                                                                                                                                                           |
| ----------------------------------------------------- | --------- | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Add `errors.Is` tests for 5 new sentinel errors       | 🔴 `TODO` | High   | 1h     | Only `ErrNotOrdered` has an `errors.Is` test (`id_test.go:332`). `ErrUnsupportedType`, `ErrCannotScan`, `ErrInsufficientData`, `ErrInternal`, `ErrNilReceiver` have zero coverage. |
| Fix Validate Docs CI failure                          | 🔴 `TODO` | High   | 30min  | `validate-docs.yml` fails: `md-go-validator@latest` module v1.2.0 exists but root package doesn't. Needs install path corrected (likely `cmd/md-go-validator`).                    |
| Remove tracked `namer` binary and add to `.gitignore` | 🔴 `TODO` | High   | 5min   | `git ls-files namer` confirms binary is tracked at repo root. BuildFlow flags it. Should be gitignored, not committed.                                                             |
| Investigate Dependabot vulnerabilities                | 🔴 `TODO` | High   | 30min  | GitHub reported 2 vulnerabilities (1 high, 1 moderate) on v0.5.0 push. Not investigated. Run `gh api repos/LarsArtmann/go-branded-id/dependabot/alerts`.                           |

## Medium Impact

| Task                                                          | Status       | Impact | Effort | Evidence                                                                                                                                                               |
| ------------------------------------------------------------- | ------------ | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Bump 14 downstream ecosystem repos to v0.5.0                  | 🔵 `BLOCKED` | Med    | Days   | v0.5.0 tagged and pushed. Source fixes applied to all repos. `go.mod` bump not done in any repo. Requires per-repo access. See ecosystem section.                      |
| Document sentinel errors in README                            | 🔴 `TODO`    | Med    | 30min  | README has no mention of `errors.Is`, sentinel errors, or the error handling pattern. Consumers don't know they can branch on error category.                          |
| Add `cmd/namer` and sentinel errors to website docs           | 🔴 `TODO`    | Med    | 2h     | Website (`website/src/content/docs/`) documents neither the namer codemod tool nor the sentinel error taxonomy. Both shipped in v0.5.0.                                |
| Add website `changelog.mdx` entries for v0.4.0 and v0.5.0     | 🔴 `TODO`    | Med    | 30min  | Website changelog only goes up to 0.3.2. Dual-mode support, namer tool, and sentinel errors are invisible on the public site.                                          |
| Pin GitHub Actions to SHA hashes                              | 🔴 `TODO`    | Med    | 1h     | BuildFlow flags 17 `github-actions-pinned` errors. All actions use `@v4`/`@v5`/`@v7` tags instead of commit SHAs.                                                      |
| Add `nix flake check` to CI workflows                         | 🔴 `TODO`    | Med    | 30min  | `nix flake check` (sandbox build in both JSON modes) only runs locally. Not in `go.yml` or `release.yml`.                                                              |
| Restore `ErrNotOrdered` full message                          | 🔴 `TODO`    | Med    | 5min   | Sentinel shortened from `"id: Compare requires an ordered type (int, uint, or string)"` to `"id: Compare requires an ordered type"`. `errors.go:14`.                   |
| Review remaining `fmt.Errorf` calls without sentinel wrapping | 🔴 `TODO`    | Med    | 1h     | 7 `fmt.Errorf` calls wrap external errors (strconv, TextUnmarshaler, etc.) without a library sentinel. E.g., `id_sql.go:261` wraps `err` but not `ErrUnsupportedType`. |

## Low Impact

| Task                                                    | Status    | Impact | Effort | Evidence                                                                                                                                    |
| ------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Restore `TestSuggestName_IntegrationWithPrint`          | 🔴 `TODO` | Low    | 30min  | Deleted in commit `0e73d12` (regression). Guards against future expectation-weakening. Not in `cmd/namer/main_test.go`.                     |
| Add `printResults` test asserting suggested name string | 🔴 `TODO` | Low    | 30min  | Existing `TestPrintResults_*` only check counts/headers, never the suggested `Name()` string value.                                         |
| Add fuzz tests for SQL `Scan` and Text `UnmarshalText`  | 🔴 `TODO` | Low    | 1h     | Only JSON (`FuzzMarshalJSON`) and Binary (`FuzzMarshalBinary`) have fuzz tests. SQL and Text round-trips are untested with arbitrary input. |
| Add `outputs` pattern gotcha to AGENTS.md               | 🔴 `TODO` | Low    | 10min  | Flake `outputs` must include `...` if not all inputs are named. Not documented in AGENTS.md "Critical Gotchas".                             |
| Add pre-commit hook for dual-mode `go test`             | 🔴 `TODO` | Low    | 30min  | Developers running plain `go test` only test v1 mode. A pre-commit hook running both modes prevents single-mode blind spots.                |

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
