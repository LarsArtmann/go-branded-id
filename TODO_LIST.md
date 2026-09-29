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

_All high-impact items completed 2026-09-29 (website dependency advisories,
Go-pins CI tripwire) — see `CHANGELOG.md` [Unreleased]._

## Medium Impact

_All medium-impact items completed 2026-09-29 (website page audit,
BD001 in-source suppression via BD002, brandid-lint Repair + `-fix`) — see
`CHANGELOG.md` [Unreleased] and `linter/CHANGELOG.md` [Unreleased]._

## Ecosystem tracking

| Task                                         | Status | Impact | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| -------------------------------------------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Bump 14 downstream ecosystem repos to v0.6.0 | 🔵     | Med    | v0.6.0 released 2026-09-17 (tag pushed, proxy indexed, `go get` verified in a clean module). Source fixes from the v0.3.x cycle (added `Name()` methods, `.String()` → `.Get()`) are applied and pushed to all repos; the `go.mod` dependency bump per repo remains — use the go-ecosystem-upgrade skill. While there, go-cqrs-lite's marker brands should gain `//brandid-lint:ignore(BD001) <reason>` directives now that brandid-lint supports them. |

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
