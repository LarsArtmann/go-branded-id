# Migration Report: branching-flow strong-id → go-branded-id/linter

_Status: analysis, 2026-10-11. No code changed._

## Question

Should the strong-id linter currently living in `~/projects/branching-flow` (module `github.com/larsartmann/go-design-smells`, package `pkg/strongid`) be migrated into `go-branded-id/linter` (the `brandid-lint` sub-module)? If so, how?

## What exists today

### Source: `branching-flow/pkg/strongid`

- Detects function params, named return values, and struct fields whose names contain an ID pattern (`userId`, `order_id`, `PID`) typed as bare primitives (`string`, `int`) and suggests branded ID types from `github.com/larsartmann/go-branded-id` (import path hardcoded in `types.go:16`).
- Size: ~1,100 LOC non-test (analyzer 163, types 317, suggestion 266, to_findings 42, doc 7) + ~1,650 LOC tests (analysis, nolint, integration, bench, types, to_findings) + ~420 LOC CLI wiring (`strong_id.go`, `strong_id_fix.go`, `strong_id_output.go`).
- Rule ID in findings: `STRONG_ID` (go-finding, advisory severity, carries before/after code + suggested type definition metadata).
- Fix mode is **display-only**: it prints detailed suggestions (before/after code), it does not apply edits.
- Suppression: `//nolint:branching-flow[:strong-id]` parsed by `pkg/core/ignore_comments.go`; suppressed violations are kept and marked, hidden by default, visible via `--include-suppressed`.
- Deeply wired into branching-flow: `pkg/analysis.RunAll`, `pkg/stats` collectors (suppression-marking analyzer contract), `LinterKey` enum and the `all` command (14 CLI linters), golden tests, CLI docs, the toolsdk provider (`branching-flow` Spec wraps `RunAll`), and the baseline system.
- Internal dependencies that do NOT exist in go-branded-id/linter: `pkg/astutil` (AnalyzerCore, declaration inspector, type-string helpers), `pkg/core` (ViolationData, IgnoreComments), `pkg/primitivetype` (enum + exclusion tables), `pkg/text`, `pkg/finding` (builders). All are branching-flow-internal; a port must re-implement the needed pieces stdlib-only.

### Target: `go-branded-id/linter` (brandid-lint)

- Rules: **BD001** (brand type used with `id.ID[...]` but missing `Name() string`), **BD002** (broken `//brandid-lint:ignore(...)` directive: missing reason, unknown rule, stale, duplicate, unplaced). Rule IDs documented as stable forever.
- ~1,000 LOC non-test. Own parse-only scanning (`walkGoFiles`, `scanFile` shared by detect and repair), own suppression machinery that is rule-ID-generic.
- Repair is auto-applying (inserts `Name()` stubs, idempotent).
- Deps: go-finding + go-finding/toolsdk only. Registered as BuildFlow provider `brandid-lint` with Detect + Repair, `ModuleFanOut: true`.
- Layering rules (from AGENTS.md): root module stays zero-dependency and must never depend on go-finding or `linter/`; `linter/` does not import the root module (detection is purely syntactic) and must not import branching-flow either.

### Ecosystem facts (verified 2026-10-11)

- `//nolint:branching-flow:strong-id` source suppressions outside branching-flow: **0**. Repos gate strong-id via `.buildflow.yml` instead.
- ~12 repos' `.buildflow.yml` reference branching-flow; several skip or cap it:
  - `webphone` skips branching-flow entirely; its triage note cites `STRONG_ID x2` (pbx UUIDs are upstream wire format).
  - `DiscordSync` caps branching-flow at `max_findings: 2200` against 1,872 advisory warnings including STRONG_ID.
- brandid-lint already runs in every BuildFlow Go repo via the orchestrator's blank import — adding a default-on rule changes findings across the whole ecosystem in one step.

## Options

### Option A — Full migration: port as BD003, later remove from branching-flow

**PRO**

- Right home: strong-id is the only branching-flow analyzer whose fix is literally "import go-branded-id". Its purpose is this library's adoption; it is misplaced in a generic design-smell suite.
- Completes the lifecycle in one tool: BD003 "adopt branded IDs" → BD001 "name your brands" → Repair scaffolds the `Name()` stub. One story, one binary, one BuildFlow provider.
- One suppression syntax (`//brandid-lint:ignore(BD003) <reason>`), reusing the existing rule-generic directive machinery; BD002 automatically polices broken BD003 directives once BD003 joins the known-rule set.
- Zero ecosystem source suppressions to migrate (verified: 0 `nolint:branching-flow:strong-id` outside branching-flow).
- Kills future split brain: suggestions and brand expectations can no longer drift between two tools.
- Sharpens branching-flow: it keeps generic primitive-obsession detection (`primitivetype`, `phantom`), sheds the library-specific rule.

**CONTRA**

- Real port cost: the analyzer core is rewrite-not-copy (deps on astutil/core/primitivetype); ~1,100 LOC logic against new stdlib-only plumbing. The suggestion engine and name heuristics (`containsIDPattern`, `splitNameTokens`, `extractIDRoot`, `suggestIDTypeName`, `suggestIDBrand`) are the carry-over IP.
- Wide de-wiring in branching-flow: RunAll, stats collectors, LinterKey (`all` goes 14→13), golden tests, CLI subcommand + fix output, docs, toolsdk description, baseline migration (STRONG_ID is pinned in repos' triage notes and caps).
- Blast radius: brandid-lint runs everywhere BuildFlow runs. Default-on BD003 would flood repos that deliberately silenced strong-id (webphone skipped branching-flow over STRONG_ID; DiscordSync caps at 2200 including STRONG_ID). BD003 must ship opt-in/default-off or the migration is a hostile regression.
- Contract mismatch: brandid-lint is a strict contract linter (Detect + Repair, exit-code gates); strong-id is advisory adoption pressure with no auto-fix. BD003 would be detect-only while BD001 has Repair — asymmetric.
- Suggestion metadata (before/after code, type definitions) has no consumer in brandid-lint's current output paths beyond what go-finding already renders.

### Option B — Copy without deleting (dual home)

**PRO**: cheapest; no branching-flow churn; BD003 available immediately.

**CONTRA**: split brain. Two heuristic sets, two suppression syntaxes, two finding streams for the same problem, guaranteed drift. Worst option — dismissed.

### Option C — Status quo (keep the split)

**PRO**: zero work; separation is defensible (adoption funnel vs contract linter); branching-flow already delivers strong-id wherever it is not skipped.

**CONTRA**: the library's adoption engine lives in an unrelated multi-smell suite; two tools to explain and run; anyone capping or skipping branching-flow loses adoption pressure entirely; suggestions can drift from brandid-lint's brand expectations.

## Verdict

**Option A, staged.** Strong-id's whole purpose is this library; migrating it is the correct end-state. Stage it so the ecosystem does not flood:

1. **Port to `linter/` as BD003** — advisory severity, detect-only (Repair stays BD001-only), **opt-in via a `-rules` flag** so default BuildFlow runs are unchanged on day one. Add BD003 to the suppress.go known-rule set so BD002 automatically covers broken BD003 directives. Re-implement the needed AST helpers stdlib-only against the existing `walkGoFiles`/`scanFile` plumbing.
2. **Port the test corpus** — analysis/nolint/integration tests carry over nearly as-is; plumbing tests get rewritten; add testdata fixtures (parse-only, never compiled, per existing convention).
3. **After one green cycle, delete from branching-flow** — 14→13 linters, LinterKey, stats wiring, golden tests, CLI subcommand, docs; publish a baseline-migration note ("Rule-ID / Baseline Migration" runbook pattern). Repos then re-point `.buildflow.yml` from branching-flow caps to enabling BD003.

### Hard constraints (any implementation)

- `linter/` must NOT import the root module, branching-flow, or any new external dep beyond go-finding/toolsdk — detection stays purely syntactic.
- Root module stays zero-dependency; no `go.work`; linter releases tag as `linter/vX.Y.Z` with its CHANGELOG section cut before tagging.
- BD001/BD002 IDs remain untouched; BD003 is additive and, once shipped, stable forever like the others.
- Never auto-"fix" BD003 (rewriting signatures is beyond a linter's remit); suggestions are output metadata only.

### Open decisions before implementation

- Default rule set of the CLI/provider when no `-rules` is given (recommend: BD001+BD002, BD003 opt-in).
- Whether the toolsdk Spec description changes to advertise BD003, and whether BuildFlow needs a per-rule gate vs per-tool `max_findings` caps.
- Whether the ported suggestion engine should also emit the brand-type definition as Repair-adjacent scaffold output in a later minor (deferred; not part of stage 1).
