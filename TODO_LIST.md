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

| Task                                             | Status       | Impact | Effort | Evidence                                                                                                                                                        |
| ------------------------------------------------ | ------------ | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Re-tag and re-push v0.3.1 after GOEXPERIMENT fix | 🔴 `TODO`    | High   | 10min  | Root cause found: CI failed because `GOEXPERIMENT=jsonv2` was not set. Fixed in `.github/workflows/release.yml`. Re-tag at HEAD and push to trigger release CI. |
| Bump 14 downstream ecosystem repos to v0.3.1     | 🔵 `BLOCKED` | High   | Days   | Source fixes applied & pushed; `go.mod` bump not done. Blocked until v0.3.1 GitHub Release exists (consumers need `go get @v0.3.1` to resolve).                 |

## Medium Impact

| Task                                            | Status    | Impact | Effort | Evidence                                                                                                                            |
| ----------------------------------------------- | --------- | ------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------- |
| Add CI integration test against ecosystem repos | 🔴 `TODO` | Med    | 2h     | `cmd/namer/main.go` exists; no representative cross-repo compile test in CI yet                                                     |
| Add tests for `cmd/namer` codemod               | 🔴 `TODO` | Med    | 1h     | `cmd/namer/main.go` has 0% coverage (`GOEXPERIMENT=jsonv2 go test ./cmd/namer/ -cover`)                                             |
| Evaluate json/v2 as long-term choice            | 🔴 `TODO` | Med    | 2h     | `encoding/json/v2` requires `GOEXPERIMENT=jsonv2` in Go 1.26. Consider whether to stay on v2 or fall back to v1 before v1.0 freeze. |

## Low Impact

| Task                                                 | Status    | Impact | Effort | Evidence                                                                        |
| ---------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------- |
| Document why go-cqrs-lite marker types skip `Name()` | 🔴 `TODO` | Low    | 15min  | Marker types are internal storage/stream keys; `Name()` would break key format. |

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
