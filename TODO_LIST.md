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

_BD003 migration COMPLETE 2026-10-11 (T01–T18 of
`docs/planning/2026-10-11_04-16_SUPERB-strongid-migration-bd003.md`): linter
v0.4.0→v0.5.1, branching-flow strong-id deprecated (v0.6.6) then removed
(v0.6.7), CV gate migrated, fleet verified, parity gate self-healing — see
`linter/CHANGELOG.md` and
`docs/status/2026-10-11_20-10_bd003-migration-t16-t18-execution-status.md`._

| Task                                                                                        | Status    | Impact | Evidence                                                                                                                                                                                                                                        |
| ------------------------------------------------------------------------------------------- | --------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Add generated-file fixtures (sqlc + templ shapes) to the frozen BD003 parity corpus          | 🔴 `TODO` | High   | The v0.4.0–v0.5.0 generated-file parity hole shipped ~90 false findings on CV's first real scan; the corpus has no generated fixture on either side, so the class is still uncovered. (20:10 report f.19; `scripts/parity-bd003.sh` corpus build.) |
| Fleet BD003 baseline drift-check mode: wire `scripts/` re-scan against the committed table   | 🔴 `TODO` | Med    | Baseline table committed 2026-10-11 (see `docs/status/` fleet verification); a script mode makes the next sweep a drift check instead of a re-measurement. (20:10 report f.11.)                                                                  |

## Medium Impact

_All medium-impact items completed 2026-09-29 (website page audit,
BD001 in-source suppression via BD002, brandid-lint Repair + `-fix`) — see
`CHANGELOG.md` [Unreleased] and `linter/CHANGELOG.md` [Unreleased]._

| Task                                                                        | Status    | Impact | Evidence                                                                                                                                                                     |
| --------------------------------------------------------------------------- | --------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Add a nix flake app/package attr for the brandid-lint CLI                   | 🔴 `TODO` | Med    | Only `checks.linter` builds it today; consumers (CV's gate) hand-build via `go build -o`. A `packages.brandid-lint`/`apps.brandid-lint` attr removes the manual step. (20:10 report f.24.) |
| Publish the BD003 fleet baseline table (done 2026-10-11, drift mode pending) | 🟡 `IN_PROGRESS` | Med | Table committed in `docs/status/2026-10-11_21-*_bd003-fleet-baseline.md`; the script drift-check mode is tracked in High Impact above. (20:10 report f.11.)                 |

## Ecosystem tracking

| Task                                                                                   | Status | Impact | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| -------------------------------------------------------------------------------------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Bump 14 downstream ecosystem repos to current releases                                | 🔵     | Med    | Open since the v0.6.0 cycle (released 2026-09-17): source fixes applied and pushed everywhere; the per-repo `go.mod` bump remains (current targets: library v0.7.0, linter v0.5.1) — use the go-ecosystem-upgrade skill. While there, go-cqrs-lite's marker brands should gain `//brandid-lint:ignore(BD001) <reason>` directives now that brandid-lint supports them. |
| go-cqrs-lite: adjudicate the +1 BD003 finding                                         | 🔵     | Low    | Fleet scan 2026-10-11 shows 224 vs 223 on the 2026-10-05 baseline (+1, consistent with repo churn, not detection drift — frozen-corpus parity gate is green). Identify and convert-or-suppress on the next go-cqrs-lite pass. (20:10 report f.14.) |
| BuildFlow: re-embed brandid-lint v0.5.1                                               | 🔵     | Med    | Installed BuildFlow binary embeds linter v0.5.0 (no generated-file skip). Harmless today (the only error-gated repos have zero error findings and no generated-file exposure), but one patch behind. BLOCKED on owner decision: branch off red origin/master now vs wait for the toolsdk Fact-API release (20:10 report g.3). Re-embed = go.mod bump + flake.lock input rev + profile reinstall (see AGENTS.md "Linter vendorHash" section). |
| SystemNix: pin branching-flow v0.6.7 + brandid-lint v0.5.1                            | 🔵     | Med    | `~/go/bin/brandid-lint` is a manual build; `/run/current-system/sw/bin/branching-flow` serves the pre-0.6.7 build. BLOCKED on owner decision: NixOS/home-manager pin for both tools vs owner-managed sudo rebuilds (20:10 report g.1); a pin also makes CV's gate leg SKIP-proof on other machines (f.20). |
| Fleet BD003 debt policy (journal 69, DiscordSync 218, webphone 6, big advisory repos) | 🔵     | Med    | All advisory (zero error-severity in every error-gated repo). BLOCKED on owner decision: conversion-wave roadmap vs standing "advisory forever, error-gate only" policy documented per repo (20:10 report g.2, f.13/f.16/f.17). Journal's 69 are inventoried in the 2026-10-11 fleet baseline table. |

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
