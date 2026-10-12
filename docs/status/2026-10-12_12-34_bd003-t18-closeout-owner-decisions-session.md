# BD003 T18 Close-Out + Owner Decisions — Session Status Report

**Date:** 2026-10-12 12:34
**Scope:** This session only. Predecessor and task context:
`docs/status/2026-10-11_20-10_bd003-migration-t16-t18-execution-status.md`
(T16+T17 done; T18 pending; 3 owner questions open). User directive on
resume: "READ, UNDERSTAND, RESEARCH, REFLECT… Execute and Verify them one
step at a time. Repeat until done." — interpreted as: execute T18 in full,
then the 20:10 report's f)-list down the executable spine, then put the 3
owner questions to the owner. **The SUPERB plan T01–T18 is now COMPLETE.**

## a) FULLY DONE (this session)

### T18 knowledge capture (the session's core mandate)

1. **go-branded-id AGENTS.md — three knowledge sections added** (the exact
   f.1–f.3 content): "BD003 — Opt-In Primitive-ID Rule" (opt-in mechanics
   via `-rules`/`DetectWithRules`/BuildFlow `tool_options`; purely-syntactic
   detection + the `containsIDPattern` name heuristic incl. what does NOT
   match; detect-only + metadata payload; severity tables error/warning/info
   with the domain prefix lists; BD003-only path exclusions; generated-file
   skip; `//brandid-lint:ignore(BD003) <reason>` placement contract;
   legacy-rule-gone note); "Linter vendorHash — Nix FOD Content-Addressing
   Trap" (stale hash silently injects OLD vendor tree; `go mod vendor &&
   nix hash path linter/vendor` procedure; dummy-SRI escape hatch;
   BuildFlow-embeds-via-flake.lock-revs corollary); "Parity Gate +
   Testdata-Corruption Postmortem" (self-healing v0.6.6 worktree;
   empty-side-diff-is-not-parity rule; three-layer corruption fix incl.
   "check which buildflow binary the daemon runs"). Linter file map in Code
   Organization refreshed (primitives/idpattern/suggest_id/generated/rules).
2. **TODO_LIST harvest, both repos** (f.4): go-branded-id — BD003 migration
   row closed as done; new rows for parity generated fixtures (High),
   drift-check script mode (High), fleet baseline publish; ROADMAP Theme 3
   gained 8 brandid-lint follow-on ideas (json format, min-severity,
   suppression file, SARIF category, BD001/BD003 dedup, more rule ports,
   corpus tag/fuzz, Go 1.27 floor decision). branching-flow — profile
   rebuild item refreshed v0.6.5→v0.6.7 with the SystemNix decision
   pointer; severity-cap owner item corrected for the strong-id exit (now
   4 BOOL_BLIND in scope, the 3 STRONG_ID errors live as BD003 under
   brandid-lint, outside branching-flow's gate); strong-id drafts archival
   and sweep-linters cleanup added.
3. **Report cross-links + refresh** (f.5/f.6): 18:20 and 19:18 reports
   carry superseded-by pointers in their headers; 19:18 has an inline
   UPDATE block (T16 tail described in its §c is DONE); 20:10 footer has a
   full T18-executed annotation with per-item routing.
4. **New session/baseline doc committed**:
   `docs/status/2026-10-11_20-25_bd003-fleet-baseline-t18-closeout.md` —
   the fleet BD003 baseline table (f.11) with per-repo totals AND severity
   splits (measured fresh this session; all 17 totals match T17b exactly —
   independent re-verification), the key invariant (every error-severity
   finding lives in a non-gated repo; both gated repos hold zero), a
   drift-check procedure, and the journal 69-finding adjudication inventory
   (per-file grouping + the 4 warnings itemized with file:line).

### f)-list execution beyond T18

5. **branching-flow hygiene**: stale worktrees cleaned (prunable
   `/tmp/baseline-bf` pruned; `.corpus-pins/branching-flow@dbdeb0f1`
   removed after verifying the parity script and all of go-branded-id
   reference it nowhere; other repos' corpus pins untouched) — f.42.
   Auto-tag double-fire checked: exactly one `v0.6.7` tag + one GitHub
   release — f.43. Coverage ratchet re-run: **80.3% vs 79.5 baseline,
   fail-below 79.0 — green, gain locked** into `.coverage-baseline` —
   f.29. Stale `docs/SPLIT-BRAIN.html` (still describing strong-id,
   referenced by nothing living) archived via `git mv` to
   `docs/archive/` — f.22.
6. **CV leg closed** (f.7/f.40/f.27): the daemon-swept gate changes
   verified intact in HEAD (brandid leg + legs-manifest `brandid` entry);
   gate `--self-test` green; **full `go-change-gate.sh --quick` run green
   across all legs including `brandid` (0 findings) and `leg-inventory`**;
   `glossary-check.sh` header got the one-line strong-id→BD003 pointer and
   still runs green (43 identity types rowed).
7. **Owner questions asked and partially answered** (see d) for the rough
   edges): Q2 fleet debt policy → **conversion waves decided**; recorded in
   TODO_LIST (wave 1 = journal, inventory ready). Q1 tool installs →
   owner's "why can't we just be on master?" answered with research and
   EXECUTED as master-tracking (below). Q3 → explained (see g), decision
   still open.
8. **f.24 DONE — brandid-lint is now a nix artifact**: go-branded-id flake
   gained `packages.brandid-lint`/`packages.default` and
   `apps.brandid-lint` (buildGoModule binding extracted from
   `checks.linter`; derivation `version` also fixed `0.1.0`→`0.5.1` — it
   had been stale since the first release, the same drift class as
   branching-flow's flake version). Verified: `nix build .#brandid-lint`,
   `nix run .#brandid-lint -- -rules BD003 <fixture>` (findings fire),
   `nix flake check --no-build` green. Committed and **pushed to
   origin/master (pre-push dual-test hook passed)** — required so the
   SystemNix input could resolve the new output.
9. **SystemNix master-tracking prepared and verified** (Q1 execution):
   new `brandid-lint` input (`github:LarsArtmann/go-branded-id?ref=master`,
   follows wired; named `brandid-lint` because `go-branded-id` was already
   taken by an existing non-flake source pin — see d) for how I found
   that), `lars-packages.nix` entry (flows into environment.systemPackages),
   lock updated: `brandid-lint` → f5b52cd, `branching-flow` → v0.6.7 tip
   (go-finding transitive bump included). Verified: `nix build
   .#brandid-lint` → 0.5.1 binary runs; `nix build .#branching-flow` →
   0.6.7 builds clean (no FOD mismatch from the go-finding bump). `nix
   fmt` clean. Remaining: owner `sudo nixos-rebuild switch` — then
   `/run/current-system/sw/bin/{branching-flow,brandid-lint}` serve
   current releases and CV's brandid leg becomes SKIP-proof system-wide.

## b) PARTIALLY DONE

- **SystemNix deliberate commit**: my annotated `git commit` of the lock
  update was ABORTED by the repo's pre-commit hook — its `nix flake check
  --no-build --all-systems` leg was **OOM-killed mid-eval** ("Killed",
  line 474 of the hook) and reported as failure. The eval itself was
  progressing normally (warnings only, dozens of derivations checked).
  The daemon subsequently swept the lock (commit `412dff75`), so the tree
  is complete and consistent — but no deliberate commit message documents
  the wiring on SystemNix master; the explanation lives only in the
  lars-packages/flake comments and here.
- **Q3 (BuildFlow re-embed)**: explained to the owner, not yet decided —
  TODO row stays 🔵 blocked.

## c) NOT STARTED (deliberately routed, not forgotten)

- Journal wave 1 conversion (decided "waves", wave 1 defined, inventory
  ready — a multi-hour repo-local session of its own).
- Parity corpus generated fixtures (f.19), drift-check script mode (f.11b),
  go-cqrs-lite +1 adjudication (f.14), CV mixins drift (f.15),
  DiscordSync/webphone + big-repo waves, BuildFlow dogfooding finding
  (f.18), branching-flow strong-id drafts archival (f.37), sweep-linters
  cleanup (f.23/f.28), CI-billing acceptance decision (f.8/f.10), ROADMAP
  items. All routed into TODO_LIST/ROADMAP with evidence this session.

## d) TOTALLY FUCKED UP (honest ledger)

1. **Baseline totals row was wrong on first write**: I hand-estimated the
   fleet table totals (wrote 4745/264/279/4202; correct: 4965/262/293/4410
   — per-repo rows were all correct, the sums were not). Caught it in
   self-review and fixed, but the doc hit the repo briefly wrong, and my
   TODO row first referenced a wrong filename (`21-*`). Lesson
   re-confirmed: COMPUTE sums mechanically, never hand-add a totals row.
2. **Added a duplicate SystemNix input**: I wrote a `go-branded-id` input
   block after reading only the neighboring input section — the name was
   already taken at flake.nix:368 by a non-flake source pin. `nix flake
   metadata` caught it ("attribute already defined"); renamed to
   `brandid-lint`. One `grep go-branded-id flake.nix` before writing
   would have avoided the whole cycle. Read-before-write applies to
   namespace surveys, not just the edit site.
3. **Repeated a documented rg trap**: I ran `rg -rn` — the previous
   session's lesson #7 says `-r` is REPLACE. Combined with `-l` it was
   harmless this time, but I typed the exact flag combo the handoff
   warned about. Muscled memory beats written memory; alias it or stop
   using `-r`.
4. **Two exact-match edit failures on daemon-reformatted tables**: the
   daemon's table re-padding broke my `old_string`s twice (go-branded-id
   TODO_LIST); I also initially misdiagnosed WHICH of three multiedit
   blocks failed. Cost: three retries. The python line-indexed fallback
   worked; the pattern for daemon-churned markdown tables is now: index
   by line number, assert the prefix, replace.
5. **First severity-split jq was garbage** (`group_by` + `//` precedence
   produced "1 null" rows); recognized immediately because anchoring
   counts still matched, switched to the brief's proven simple form.
   Cheap lesson: use the documented working command first.
6. **Deliberate SystemNix commit aborted by an OOM-killed hook** (see b):
   I attempted a hook-gated commit without anticipating that a full
   all-systems eval of a ~100-derivation NixOS config under load can be
   memory-killed. The fallback (daemon sweep, hookless) is acceptable
   here ONLY because the content was already build-verified directly; a
   less-verified change should not silently ride that path.

## e) WHAT WE SHOULD IMPROVE

1. **Mechanical totals**: every table with a sums row gets computed by a
   script (python `sum`), like BUILD-mode doc counts. This is the
   second-fasteest recurring doc bug class this week.
2. **Pre-edit namespace greps in nix flake inputs**: input names are a
   global namespace — grep the whole `inputs = {` block before adding.
3. **SystemNix pre-commit flake-check leg is OOM-prone** under load
   (evidence: this session's kill). Consider splitting the check or
   adding an eval-cache path — owner's call; documented here.
4. **AGENTS.md command table not updated for the new flake app** (found
   while writing this report): `nix run .#brandid-lint` /
   `packages.brandid-lint` deserve a row in go-branded-id AGENTS.md
   "Essential Commands" — small follow-up, listed below.
5. **SystemNix's own docs were not swept** for the new tool (FEATURES.md /
   lars-packages README inventory) — the repo's doc conventions are
   outside what I read this session; left deliberately rather than
   half-guessed. Listed below.
6. **Stale LSP noise** (45 "unused" warnings on linter/*.go) contradicted
   by grep and a green release gate — I verified via grep instead of
   restarting the LSP; do both next time (cheaper for the next session).

## f) NEXT — up to 50 things

**Activation + wiring tail:**
1. Owner: `sudo nixos-rebuild switch` (SystemNix) — serves
   branching-flow 0.6.7 + brandid-lint 0.5.1 system-wide; verify
   `branching-flow version` = 0.6.7 and `brandid-lint --help` on PATH.
2. Retry a deliberate SystemNix lock commit when memory pressure allows
   (or accept the daemon sweep — content is verified; decision is the
   owner's).
3. go-branded-id AGENTS.md Essential Commands: add
   `nix run .#brandid-lint -- -rules BD003 <path>` /
   `nix build .#brandid-lint` row.
4. SystemNix docs sweep for the new brandid-lint package
   (FEATURES.md/tools inventory per that repo's conventions).
5. CV AGENTS/gate docs: note that brandid-lint now arrives via SystemNix
   PATH (the SKIP fallback message's "owner: the SystemNix pin" clause is
   now resolved once activated).

**Decided-but-not-started (owner already said "waves"):**
6. Wave 1 — journal: convert/suppress the 69 findings (inventory +
   itemized 4 warnings in the 20:25 closeout doc); keep the error-gate
   green; glossary rows per CV's template.
7. Wave 2 — DiscordSync (218); Wave 3 — webphone (6); then
   ActaFlow/Kernovia/CreditReformBilanzampel/ci-siblings/games/
   standard-bug-tracking-schema/picoclaw by owner-prioritized order.

**Still-blocked / owner-decision items (carried):**
8. BuildFlow v0.5.1 re-embed (g.3 — see g) below).
9. go-cqrs-lite +1 BD003 finding adjudication (f.14).
10. CV detector-baseline mixins drift, 2 names (f.15, pre-existing).
11. BuildFlow repo's own 1 BD003 finding — dogfood-brand it (f.18).
12. CI billing: fix or formally accept local-gates-only (f.8/f.10).

**Hardening + quality (routed in TODO/ROADMAP this session):**
13. Parity corpus: add sqlc-shaped + templ-shaped generated fixtures
    (f.19 — High row).
14. Fleet drift-check script mode against the committed baseline (f.11b).
15. Fleet-wide BD001 re-run under v0.5.1 (f.39 — generated skip may have
    freed suppressions).
16. Tag the frozen parity corpus snapshot (f.30 / ROADMAP).
17. branching-flow: archive `docs/drafts/*strong-id*` (f.37);
    sweep-linters REPOS/README cleanup (f.23/f.28).
18. ROADMAP Theme 3 follow-ons: `-format json`, `-min-severity`,
    `.brandid-lint.yml`, SARIF category, BD001/BD003 dedup, more rule
    ports, Go 1.27 floor decision.

## g) Questions for the owner (cannot figure this out myself)

1. **BuildFlow v0.5.1 re-embed timing** (carried from 20:10 g.3, now with
   the explanation you asked for): "Fact-API" = the unreleased go-finding
   `Fact` type from the `feat/detector-v2-coverage` branch. Another
   session's toolsdk provider-switchover WIP in BuildFlow
   (`sdk_registry.go` et al.) references it; the daemon pushed that WIP to
   origin/master, and BuildFlow CI jobs have been failing since (their
   failures currently show the 4s billing-spawn signature, so code-red vs
   billing-red is not cleanly separable from here). Re-embedding v0.5.1
   now would branch off that master; waiting means waiting for the
   go-finding release + a green master. My recommendation: WAIT (the
   v0.5.0 embedding is verified-harmless today: only error-gated repos
   matter and both are clean with no generated-file exposure). Which way?
2. **SystemNix commit discipline**: accept the daemon-swept lock commit
   (`412dff75`, content build-verified by me directly), or should I retry
   a deliberate annotated commit once the pre-commit hook's all-systems
   eval survives memory pressure?
3. **Wave 1 scheduling**: journal conversion is a multi-hour repo-local
   session (69 findings, mostly event/read-model struct fields). Start it
   on your next "go", or is another repo's wave more urgent?

---

*Session closed here: plan T01–T18 complete, migration finished, decisions
recorded, activation pending (sudo). Companion docs: 18:20 → 19:18 → 20:10
→ 20:25 → this report.*
