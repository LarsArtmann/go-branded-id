# TODO List — go-branded-id

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md` (not yet created).
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## High Impact

| Task                                         | Status       | Impact | Effort | Evidence                                                                                                                                                                 |
| -------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Investigate missing v0.3.1 GitHub Release    | 🔴 `TODO`    | High   | 30min  | Tag `v0.3.1` pushed to remote (`8b30d92`) but `gh release view v0.3.1` → "release not found". Release workflow (`.github/workflows/release.yml`) did not fire or failed. |
| Bump 14 downstream ecosystem repos to v0.3.1 | 🔵 `BLOCKED` | High   | Days   | `Name()` + `.Get()` fixes already applied & pushed to all 14 repos (see "Ecosystem tracking" below); the `go.mod` dependency bump itself is not yet done in any of them. |

## Medium Impact

| Task                                            | Status    | Impact | Effort | Evidence                                                                        |
| ----------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------- |
| Add CI integration test against ecosystem repos | 🔴 `TODO` | Med    | 2h     | `cmd/namer/main.go` exists; no representative cross-repo compile test in CI yet |
| Add tests for `cmd/namer` codemod               | 🔴 `TODO` | Med    | 1h     | `cmd/namer/main.go` has 0% coverage (`go test ./cmd/namer/ -cover`)             |

## Low Impact

| Task                                                 | Status    | Impact | Effort | Evidence                                                                                                              |
| ---------------------------------------------------- | --------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------------- |
| Document why go-cqrs-lite marker types skip `Name()` | 🔴 `TODO` | Low    | 15min  | Marker types are internal storage/stream keys; `Name()` would break key format. Add a note in `cmd/namer/` or README. |

---

## Ecosystem tracking

This library has 14 downstream repos. The v0.3.0/v0.3.1 source changes (`Name()`
methods added, `.String()` → `.Get()` fixes, test fixes) are **applied and pushed**
to all of them. The **`go.mod` dependency bump to v0.3.1 itself is not yet done**
in any repo.

Downstream repos: InboxClean, CreditReformBilanzampel, ActaFlow, SEC, storbi,
ChastityAPI, smart-configs, StopTube, universal-workflow, Zlota44, timesheets,
complaints-mcp (archived), cqrs-htmx, emeet-pixyd.

Deliberately **not** changed (correct as-is): go-cqrs-lite (marker types),
BerryBig (test brands only), Cyberdom (no brand types).

Pre-existing test failures **not** caused by this library: CreditReformBilanzampel
(BDD undefined step), timesheets (fuzz hours overflow), emeet-pixyd (PipeWire
state file), Zlota44 (internal/discovery).

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding, many
    documented TODOs are already done.
  - One task per row. If it takes more than ~2 hours, split it into smaller tasks.
  - Cite evidence (file:line) so the next person can verify without re-deriving.
  - DONE items should be REMOVED, not kept. Use CHANGELOG.md for history.
  - If a task is vague ("improve X"), refine it or move it to ROADMAP.md.
-->
