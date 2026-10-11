# BD003 Fleet Baseline + T18 Close-Out

**Date:** 2026-10-11 20:25
**Purpose:** (1) The committed fleet BD003 baseline table the 20:10 report
asked for (f.11) — the next fleet sweep becomes a drift check against this
table instead of a re-measurement. (2) The T18 knowledge-capture close-out
record for the BD003 migration (plan T01–T18 now COMPLETE).

Companion reports, in order: `2026-10-11_18-20_bd003-migration-execution-status.md`
→ `2026-10-11_19-18_bd003-migration-phase-execution-status.md` →
`2026-10-11_20-10_bd003-migration-t16-t18-execution-status.md` → this one.

## 1) Fleet BD003 baseline (measured 2026-10-11 ~20:20, linter v0.5.1)

Scan: `brandid-lint -rules BD003 -format sarif <repo>` (binary rebuilt from
`linter/v0.5.1`), severity via
`jq -r '.runs[0].results[]? | .properties.severity // .level'`.
All 17 totals independently re-verified this session and match the T17b
record exactly (anchors: branching-flow 0=0, go-finding 2=2, CV 0, journal
69, go-cqrs-lite 224).

| Repo                         | Total | Error | Warning | Note | Gate status                                                                        |
| ---------------------------- | ----- | ----- | ------- | ---- | ---------------------------------------------------------------------------------- |
| branching-flow               | 0     | 0     | 0       | 0    | clean (parity anchor; strong-id removed v0.6.7)                                    |
| go-finding                   | 2     | 0     | 0       | 2    | advisory (parity anchor)                                                           |
| CV                           | 0     | 0     | 0       | 0    | zero-ratchet via `go-change-gate.sh` brandid leg                                   |
| go-branded-id                | 0     | 0     | 0       | 0    | clean outside skipped testdata/fixtures                                            |
| journal                      | 69    | 0     | 4       | 65   | **error-gated** (`error_rules: brandid-lint: [BD003]`) — error count 0, gate green |
| go-business-rules            | 0     | 0     | 0       | 0    | **error-gated** — clean                                                            |
| DiscordSync                  | 218   | 2     | 12      | 204  | advisory (DEFER repo, comment-only config)                                         |
| webphone                     | 6     | 0     | 0       | 6    | advisory (DEFER repo, comment-only config)                                         |
| Kernovia                     | 1013  | 14    | 29      | 970  | advisory                                                                           |
| CreditReformBilanzampel      | 945   | 32    | 49      | 864  | advisory                                                                           |
| ActaFlow                     | 223   | 0     | 0       | 223  | advisory                                                                           |
| ci-siblings                  | 335   | 8     | 24      | 303  | advisory                                                                           |
| games                        | 727   | 81    | 93      | 553  | advisory                                                                           |
| standard-bug-tracking-schema | 692   | 107   | 41      | 544  | advisory                                                                           |
| picoclaw                     | 510   | 18    | 27      | 465  | advisory                                                                           |
| BuildFlow                    | 1     | 0     | 0       | 1    | advisory (dogfooding candidate, f.18)                                              |
| **Totals**                   | 4965  | 262   | 293     | 4410 |                                                                                    |

**Key invariant:** every error- severity finding lives in a repo whose
BuildFlow config does NOT error-gate BD003. The two gated repos (journal,
go-business-rules) hold zero error findings — the guard has never fired.

**Drift check (next sweep):** re-run the scan command above and compare
against this table; any anchor change (branching-flow/go-finding/CV/journal
totals) or any new error-severity finding in a gated repo is a detection
regression to triage against the frozen-corpus parity gate FIRST (it is
authoritative for detection drift; fleet deltas are usually repo churn).
A script mode for this is tracked in `TODO_LIST.md` (High Impact).

### Journal 69-finding inventory (adjudication prep, f.13)

Grouped by file (`brandid-lint -rules BD003 journal`, severity split 65 note

- 4 warning):

| File                          | Findings | Dominant shape                                                                        |
| ----------------------------- | -------- | ------------------------------------------------------------------------------------- |
| `internal/cqrs/events.go`     | 23       | event struct string ID fields (note)                                                  |
| `internal/cqrs/readmodels.go` | 21       | read-model row struct ID fields (note)                                                |
| `internal/config/config.go`   | 9        | config struct ID fields (note) — likely suppress-not-convert (config is presentation) |
| `internal/app/ingest.go`      | 6        | ingest row structs (note)                                                             |
| `internal/app/export.go`      | 6        | export row structs (note)                                                             |
| `internal/app/convert.go`     | 3        | `bankSyncExportRow` ID fields (note)                                                  |
| `internal/app/book.go`        | 1        | `TaxKeyYAML.RuleID` (note)                                                            |

The 4 warnings, itemized:

- `internal/app/export.go:439` — `PaymentAllocationV1Row.TransactionID` (string)
- `internal/config/config.go:219` — param `entityID` (string) in `checkDimension`
- `internal/config/config.go:251` — param `entityID` (string) in `Establishment`
- `internal/config/config.go:267` — param `entityID` (string) in `Jurisdiction`

Adjudication (convert waves vs `//brandid-lint:ignore(BD003) <reason>`
suppressions vs advisory-forever) is the owner decision tracked as the
"Fleet BD003 debt policy" row in `TODO_LIST.md` (20:10 report g.2).

## 2) T18 close-out record (this session)

1. **go-branded-id AGENTS.md** — three knowledge sections added:
   BD003 contract (opt-in `-rules`/RuleSet/tool_options, purely-syntactic
   detection + name heuristic, detect-only + metadata payload, severity
   tables, BD003-only path exclusions, generated-file skip, suppression
   directive placement, legacy-rule-gone note); the Nix FOD
   content-addressing vendorHash trap (+ `go mod vendor && nix hash path
   linter/vendor` procedure, dummy-hash escape hatch, BuildFlow
   flake.lock-rev embedding); parity-gate self-healing + the
   testdata-corruption postmortem (three-layer fix). Linter file map in
   Code Organization refreshed (primitives/idpattern/suggest_id/generated/
   rules).
2. **TODO_LIST harvest, both repos** — go-branded-id: BD003 migration row
   closed (done), fleet drift-mode + parity generated fixtures + flake CLI
   attr added, four BLOCKED owner-decision rows routed (BuildFlow re-embed,
   SystemNix pin, fleet debt policy, go-cqrs-lite +1); ROADMAP Theme 3
   gained the brandid-lint follow-on ideas (json format, min-severity,
   suppression file, SARIF category, BD001/BD003 dedup, more rule ports,
   corpus tag/fuzz, Go 1.27 floor). branching-flow: profile-rebuild item
   refreshed v0.6.5→v0.6.7, severity-cap item corrected for the strong-id
   exit (4 BOOL_BLIND remain in scope), strong-id drafts archival +
   sweep-linters cleanup added.
3. **Report cross-links** — 18:20 and 19:18 headers now carry superseded-by
   pointers; 20:10 footer points here.
4. **branching-flow hygiene (f)-list)** — stale worktrees pruned
   (`/tmp/baseline-bf` prunable entry, `.corpus-pins/branching-flow@dbdeb0f1`
   removed after verifying nothing references it); auto-tag double-fire
   checked (exactly one `v0.6.7` tag + one release); coverage ratchet
   re-run (80.3% vs 79.5 baseline — gain locked into `.coverage-baseline`);
   stale `docs/SPLIT-BRAIN.html` (still describing strong-id) archived to
   `docs/archive/`.
5. **CV** — gate-script changes verified intact in HEAD after the daemon
   sweep; full gate re-run + glossary-check pointer done in CV this session
   (recorded in CV's git history).

Owner questions still open: the three from the 20:10 report (g.1 SystemNix
pin, g.2 fleet debt policy, g.3 BuildFlow re-embed timing).
