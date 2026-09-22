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

| Task                                                                                     | Status | Impact | Evidence                                                                                                                                                                                                                                                                                   |
| ---------------------------------------------------------------------------------------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Refresh website dependency overrides for `fast-uri`, `svgo`, `devalue`                    | 🔴     | High   | 7 Dependabot alerts open as of 2026-09-22 (4× fast-uri high, 2× svgo, 1× devalue medium) — all website-side. The lockfile (`website/pnpm-lock.yaml`) pins fast-uri 3.1.5 / svgo 4.0.2 / devalue 5.9.0, so these advisories are newer than the locked versions; overrides + `pnpm install` needed. |
| Add a Go-pins consistency CI job (assert `go.mod` `go` directive ≤ flake `go_1_2x` pin)  | 🔴     | High   | The daemon re-bumped `go 1.26` → `1.27.1` twice (caught by CI both times); `go-mod-update` is skipped via `.buildflow.yml` but a grep guard is belt-and-suspenders (`docs/status/2026-09-17_19-11`, f.4).                                                                                     |

## Medium Impact

| Task                                             | Status | Impact | Evidence                                                                                                                                     |
| ------------------------------------------------ | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Audit website pages beyond `/changelog/` for staleness | 🔴     | Med    | The v0.6.0 session verified only the changelog page (`docs/status/2026-09-17_19-11`, c.5); API reference vs the v0.6.0 surface is unverified. |

## Ecosystem tracking

| Task                                         | Status | Impact | Evidence                                                                                                                                                                                                                                                                                                  |
| -------------------------------------------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Bump 14 downstream ecosystem repos to v0.6.0 | 🔵     | Med    | v0.6.0 released 2026-09-17 (tag pushed, proxy indexed, `go get` verified in a clean module). Source fixes from the v0.3.x cycle (added `Name()` methods, `.String()` → `.Get()`) are applied and pushed to all repos; the `go.mod` dependency bump per repo remains — use the go-ecosystem-upgrade skill. |

---

## Ecosystem detail

Downstream repos: InboxClean, CreditReformBilanzampel, ActaFlow, SEC, storbi,
ChastityAPI, smart-configs, StopTube, universal-workflow, Zlota44, timesheets,
complaints-mcp (archived), cqrs-htmx, emeet-pixyd. Consumers of the library
tooling ecosystem additionally include go-output, cmdguard, and the go-finding
CLI (all pin v0.5.1).

Brands deliberately **not** changed (correct as-is — go-cqrs-lite marker types,
BerryBig, Cyberdom): see AGENTS.md "Brands That Deliberately Skip `Name()`".

Pre-existing test failures **not** caused by this library: CreditReformBilanzampel
(BDD undefined step), timesheets (fuzz hours overflow), emeet-pixyd (PipeWire
state file), Zlota44 (internal/discovery).
