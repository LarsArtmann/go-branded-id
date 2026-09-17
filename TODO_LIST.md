# TODO List — go-branded-id

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.
> Completed work is removed from this list and recorded in `CHANGELOG.md`.

## Status legend

| Status           | Meaning                                                 |
| ---------------- | ------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                               |
| 🟡 `IN_PROGRESS` | Actively being worked on.                               |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed. |

## High Impact

| Task                                                                              | Status | Impact | Evidence                                                                                                                                                                                                                   |
| --------------------------------------------------------------------------------- | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Add a CI/release guard that rejects a tracked compiled binary at repo root        | 🔴     | High   | Prevents recurrence of the v0.5.0 incident where a tracked `namer` binary inflated release source archives ~10x. Currently relies on `.gitignore` only — no workflow checks for build artifacts.                           |
| Bump `js-yaml` to ≥4.3.2 in `website/` (`pnpm update js-yaml`)                    | 🔴     | Low    | Lockfile resolves 4.3.1; open Dependabot alert (high, CPU-use DoS in merge keys). The other 6 open alerts (fast-uri ×4, svgo ×2) reference the pre-regeneration lockfile and auto-close on the next Dependabot scan — verified 2026-09-17: count already dropped 10→7 after the lockfile push. |

## Medium Impact

| Task                                                                            | Status | Impact | Evidence                                                                                                                                                       |
| ------------------------------------------------------------------------------- | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Add `ErrMarshal`/`ErrUnmarshal` test coverage for the remaining delegate paths   | 🔴     | Med    | `id_errors_test.go` proves `ErrMarshal` only via `MarshalBinary` (one subtest). Untested delegates: JSON marshaler, SQL `Value()` TextMarshaler, `UnmarshalBinary` custom-type delegate, `unmarshalTextDefault` TextUnmarshaler path. |
| Guard the JSON v1 imports at build time (the contract test cannot catch this)    | 🔴     | Med    | `TestDualJSONContract_Imports` runs inside the package: when v1 imports are corrupted, the package fails to compile and the test never executes (`docs/status/2026-08-02_16-11`, e1). Add a grep-based import check to the pre-push hook (before `go test`) and/or CI, and make the hook report both modes instead of dying on v1 (`set -e`). |

## Ecosystem tracking

| Task                                         | Status       | Impact | Evidence                                                                                                                                                                                                                   |
| -------------------------------------------- | ------------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Bump 14 downstream ecosystem repos to v0.6.0 | 🔴     | Med    | v0.6.0 released 2026-09-17 (tag pushed, proxy indexed, `go get` verified in a clean module). Source fixes from the v0.3.x cycle (added `Name()` methods, `.String()` → `.Get()`) are applied and pushed to all repos; the `go.mod` dependency bump per repo remains — use the go-ecosystem-upgrade skill. |

---

## Ecosystem detail

Downstream repos: InboxClean, CreditReformBilanzampel, ActaFlow, SEC, storbi,
ChastityAPI, smart-configs, StopTube, universal-workflow, Zlota44, timesheets,
complaints-mcp (archived), cqrs-htmx, emeet-pixyd.

Brands deliberately **not** changed (correct as-is — go-cqrs-lite marker types,
BerryBig, Cyberdom): see AGENTS.md "Brands That Deliberately Skip `Name()`".

Pre-existing test failures **not** caused by this library: CreditReformBilanzampel
(BDD undefined step), timesheets (fuzz hours overflow), emeet-pixyd (PipeWire
state file), Zlota44 (internal/discovery).
