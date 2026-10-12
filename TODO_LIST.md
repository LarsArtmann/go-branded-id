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

| Task                                                                                       | Status    | Impact | Evidence                                                                                                                                                                                                                                           |
| ------------------------------------------------------------------------------------------ | --------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Add generated-file fixtures (sqlc + templ shapes) to the frozen BD003 parity corpus        | 🔴 `TODO` | High   | The v0.4.0–v0.5.0 generated-file parity hole shipped ~90 false findings on CV's first real scan; the corpus has no generated fixture on either side, so the class is still uncovered. (20:10 report f.19; `scripts/parity-bd003.sh` corpus build.) |
| Fleet BD003 baseline drift-check mode: wire `scripts/` re-scan against the committed table | 🔴 `TODO` | Med    | Baseline table committed 2026-10-11 (see `docs/status/` fleet verification); a script mode makes the next sweep a drift check instead of a re-measurement. (20:10 report f.11.)                                                                    |

## Medium Impact

_All medium-impact items completed 2026-09-29 (website page audit,
BD001 in-source suppression via BD002, brandid-lint Repair + `-fix`) — see
`CHANGELOG.md` [Unreleased] and `linter/CHANGELOG.md` [Unreleased]._

| Task                                                                         | Status           | Impact | Evidence                                                                                                                                                                                   |
| ---------------------------------------------------------------------------- | ---------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| DONE 2026-10-11: nix flake `packages.brandid-lint` + `apps.brandid-lint`        | ✅      | Med    | `nix build .#brandid-lint` + `nix run .#brandid-lint -- -rules BD003 <path>` verified; derivation `version` also fixed `0.1.0` → `0.5.1` (was stale since the first release). Consumers can stop hand-building. (Was 20:10 report f.24.) |
| Publish the BD003 fleet baseline table (done 2026-10-11, drift mode pending) | 🟡 `IN_PROGRESS` | Med    | Table committed in `docs/status/2026-10-11_20-25_bd003-fleet-baseline-t18-closeout.md`; the script drift-check mode is tracked in High Impact above. (20:10 report f.11.)                  |

## Ecosystem tracking

| Task                                                                                  | Status | Impact | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------------- | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Bump 14 downstream ecosystem repos to current releases                                | 🔵     | Med    | Open since the v0.6.0 cycle (released 2026-09-17): source fixes applied and pushed everywhere; the per-repo `go.mod` bump remains (current targets: library v0.7.0, linter v0.5.1) — use the go-ecosystem-upgrade skill. While there, go-cqrs-lite's marker brands should gain `//brandid-lint:ignore(BD001) <reason>` directives now that brandid-lint supports them.                                                                       |
| go-cqrs-lite: adjudicate the +1 BD003 finding                                         | 🔵     | Low    | Fleet scan 2026-10-11 shows 224 vs 223 on the 2026-10-05 baseline (+1, consistent with repo churn, not detection drift — frozen-corpus parity gate is green). Identify and convert-or-suppress on the next go-cqrs-lite pass. (20:10 report f.14.)                                                                                                                                                                                           |
| DONE 2026-10-12: BuildFlow re-embed brandid-lint v0.5.1                              | ✅     | Med    | Unblocked by the go-finding v1.15.0 release (Fact API) + green master. Flake pins moved to tag commits (go-finding v1.15.0, go-branded-id linter/v0.5.1); requires bumped go-finding v1.15.0 in all 18 workspace modules + linter v0.5.1 in root/tools/execution; 34/34 workspace test gate green (incl. sibling API contract + toolsdk parity tests); profile reinstalled (buildflow 8a34928, BD003 visible in `buildflow list providers`). BuildFlow master committed locally, not yet pushed. (Was 20:10 report g.3.) |
| SystemNix: master-track brandid-lint (branching-flow already master-tracked)  | 🟡     | Med    | OWNER ANSWER 2026-10-11: "why can't we just be on master" — yes. branching-flow input already follows master (`SystemNix/flake.nix` `ref=refs/heads/master`); it serves an old binary only because the lock/activation is stale. brandid-lint needs the new `packages.brandid-lint` flake output (done, see Medium Impact) + a SystemNix input/package entry; then `nix flake update` + owner sudo activation. Makes CV's gate leg SKIP-proof everywhere (f.20). |
| BD003 conversion-wave roadmap (journal 69 → DiscordSync 218 → webphone 6 → big advisory repos) | 🔴 `TODO` | Med    | OWNER DECIDED 2026-10-11: conversion waves (not advisory-forever). Wave 1 = journal (inventory + itemized 4 warnings in `docs/status/2026-10-11_20-25_bd003-fleet-baseline-t18-closeout.md`); CV's zero-conversion wave is the template; each wave keeps the error-gate green. (20:10 report g.2/f.13/f.16/f.17.) |

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
