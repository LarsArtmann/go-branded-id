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

| Task                                         | Status    | Impact | Effort | Evidence                                                                                                            |
| -------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------- |
| Bump 14 downstream ecosystem repos to v0.4.0 | 🔴 `TODO` | High   | Days   | Tags v0.3.1–v0.4.0 exist on remote. Source fixes applied & pushed to all repos. `go.mod` bump not done in any repo. |

## Medium Impact

| Task                                            | Status    | Impact | Effort | Evidence                                                                                               |
| ----------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------ |
| Add CI integration test against ecosystem repos | 🔴 `TODO` | Med    | 2h     | `cmd/namer/main.go` exists with 93% test coverage; no representative cross-repo compile test in CI yet |

## Low Impact

| Task                                                 | Status    | Impact | Effort | Evidence                                                                  |
| ---------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------- |
| Document why go-cqrs-lite marker types skip `Name()` | 🟢 `DONE` | Low    | 15min  | Documented in `AGENTS.md` under "Brands That Deliberately Skip `Name()`". |

---

## Completed (log in CHANGELOG.md before removing)

| Task                                             | Status    | Evidence                                                                                   |
| ------------------------------------------------ | --------- | ------------------------------------------------------------------------------------------ |
| Re-tag and re-push v0.3.1 after GOEXPERIMENT fix | 🟢 `DONE` | Tags v0.3.1–v0.4.0 exist on remote. Release CI fires successfully.                         |
| Add tests for `cmd/namer` codemod                | 🟢 `DONE` | 93% coverage (was 0%). Found and fixed nil-pointer bug in `isNameMethod`.                  |
| Evaluate json/v2 as long-term choice             | 🟢 `DONE` | Dual-supports both v1 and v2 via build tags. CI tests both modes. See `id_json_v{1,2}.go`. |

---

## Ecosystem tracking

This library has 14 downstream repos. The v0.3.0/v0.3.1 source changes (`Name()`
methods added, `.String()` → `.Get()` fixes, test fixes) are **applied and pushed**
to all of them. The **`go.mod` dependency bump itself is not yet done**
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
