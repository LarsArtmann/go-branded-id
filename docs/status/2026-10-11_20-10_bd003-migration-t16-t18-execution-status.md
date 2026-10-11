# BD003 Migration — Execution Status (T16–T18 Session)

**Date:** 2026-10-11 20:10
**Scope:** This session only — resuming from the paused state documented in `2026-10-11_19-18_bd003-migration-phase-execution-status.md`. User directive: finish T16 (branching-flow strong-id removal → v0.6.7), then T17 (fleet verification + CV gate migration per the session's fleet-grep finding), then T18 (knowledge capture). "GET SHIT DONE" mode.

---

## a) FULLY DONE (this session)

### T16 — branching-flow strong-id removal → v0.6.7 RELEASED

1. **All 4 remaining test failures fixed** (exact fixes from handoff):
   - `linters_test.go`: pin 14→13 linters; removed duplicate boolblind line in `TestLinterSet_CommandSuggestions` expected-usage list.
   - `rule_registry_test.go:19`: `"panic-condition"` → `"panic_condition"` (registry token uses underscore).
   - `TestBDDSuite` (2 specs): both fixtures relied on removed strong-id detection of `Process(userID string)`. Re-triangulated on the phantom swap-hazard (`Process(userID string, sessionID string) error`) — the finding names `userID`, satisfying both the exit-code spec and the markdown-row spec; also un-vacuums the "explicitly disabled" sibling spec which was passing on zero findings.
2. **Full verification green**: `go build` + `go vet` + `go test ./...` (35 packages, 0 FAIL), `go-arch-lint check` OK (after removing the stale `pkg/strongid` component + 3 `mayDependOn` entries from `.go-arch-lint.yml`), golangci-lint 0 issues (after `--fix` cleaned 3 gci import-order leftovers), `nix fmt` 0 changed.
   - **Discovery:** tests require `GOEXPERIMENT=jsonv2` (dev-shell default via flake `env`); without it the BDD suite fails confusingly ("13 of 13 linter(s) FAILED") on a different spec set than the real failures.
3. **Docs regen + full living-docs sweep** (generate-docs diff sane; orphaned `branching-flow_strong-id.md` trashed):
   - README.md: intro, Quick Start, command table, strong-id section → migration note pointing at `brandid-lint -rules BD003`, SARIF examples, GH Actions examples, `//lint:ignore` examples re-triangulated on live rules (BOOL_BLIND).
   - FEATURES.md: 16→15 analyzers, linter table renumbered (13 CLI + 2 SDK-only), all count claims updated, dogfooding row removed.
   - QUICK_START.md, format-support-matrix.md, integrations.md (4 spots), analyzer-reference.md (file table → removal header; `strongIDConfig` row deleted; helper user-lists pruned), DOMAIN_LANGUAGE.md (16→15), stats-json-schema.md, PUBLIC_OR_PRIVATE.md (10→9), DOGFOODING.md + dedup-acceptance.md left as historical records.
   - `docs/agents-gotchas.md`: 17 surgical edits — suppression contract (all+7 rules, 3 suppression-marking analyzers), astutil usage notes, explain contract typed-nolint list, go-finding model note, Footer row (infra survives, zero users now), `all` count 14→13, docs/CLI file count 19, workflow references removed.
   - `pkg/stats/collect.go` comment, `tools/sweep-linters.sh` (strong-id column retired), `tools/README.md` (reference table annotated).
   - **Code deletions**: `.github/workflows/strong-id.yml` (whole SARIF workflow), ci.yml linter loop entry.
   - AGENTS.md: package map, Violation-interface counts (3/15), deprecation section → removal-complete record.
4. **v0.6.7 released end-to-end**: CHANGELOG `[0.6.7]` (carries the accumulated [Unreleased] features honestly + Removed section), **flake.nix version 0.6.5→0.6.7** (it was stale from v0.6.6 — the exact AGENTS.md gotcha), release commit `8edb44df`, annotated tag pushed, manual `gh release create` (billing blocks GoReleaser) → https://github.com/LarsArtmann/branching-flow/releases/tag/v0.6.7
   - **Note:** repo CI workflow has been dead since 2026-07-16 (billing — 4s runner spawn failures, pre-existing); local gates are the verification of record. Docs Health workflow: green.

### Linter v0.5.1 — CV parity gap found, fixed, released

5. **Discovered a real BD003 parity gap via the CV gate**: old strong-id skipped generated files by default (`excludeGenerated`); BD003 scanned them → ~90 findings in CV's sqlc (`.sql.go`) and templ (`*_templ.go`) output vs the old zero. The frozen parity corpus had no generated files, so the 26/26 gate couldn't catch it.
6. **Root-cause fix in brandid-lint** (per the "fix detection in brandid-lint, never back here" contract): `linter/generated.go` — `isGenerated()` matches the ecosystem-standard `Code generated ... DO NOT EDIT.` header (sqlc/templ/protoc/mockgen/stringer) before the package clause; skip lives in `scanFile` so detection AND repair see it (repair never inserts Name() stubs into machine output). Pinned by `TestGeneratedFilesAreSkipped` (6 subtests incl. after-package-clause comment not skipping, repair no-op).
7. **Released linter/v0.5.1**: Version bump, CHANGELOG, full `nix flake check` green, commit `507b01d`, tag pushed, GitHub Release live, **parity gate re-run GREEN** (28 findings, 1 accept — identical to pre-removal record), website changelog.mdx + brandid-lint guide updated, Firebase deployed.
8. **Parity gate made self-healing** (`scripts/parity-bd003.sh`): post-removal it silently built an old-side binary WITHOUT the strong-id subcommand → compared against empty → garbage verdict. Now materializes a detached v0.6.6 worktree when the checkout lacks `pkg/strongid` (loud failure if it can't). Verified: worktree created, used, cleaned up.

### T17a — CV gate migrated to BD003

9. **CV scans clean**: v0.5.1 binary → `brandid-lint -rules BD003 /home/lars/projects/CV` exit 0, zero findings — the zero-conversion ratchet carries over exactly.
10. **CV scripts rewired** (uncommitted; daemon will sweep):
    - `scripts/go-change-gate.sh` 4f-ter: strong-id leg → **brandid leg** (`brandid-lint -rules BD003 .`, exit code IS the zero-ratchet, `CV_BRANDID_LINT_BIN` override, SKIP-with-reason shape preserved, boundary-table adjudication now via `//brandid-lint:ignore(BD003) <reason>`).
    - `scripts/detector-baseline.sh`: strong-id sub retired (header comments, ALL_SUBS, sub_header entry), empty `baselines/strong-id.txt` trashed, `--self-test` PASS, full check runs (dupe/do/anti-patterns clean).
    - `scripts/go-change-gate-legs.txt`: strong-id → brandid (leg-inventory ratchet is bidirectional).
    - CV AGENTS.md: standing-gate description updated.
    - Gate `--self-test` green; `bash -n` clean.
11. **Local binary installed**: `~/go/bin/brandid-lint` rebuilt from v0.5.1 (was v0.4.1-dev pseudo-version).

### T17b — fleet verification scan (measurement complete)

12. **160+ repos scanned** with `brandid-lint -rules BD003`. Key results:
    - **Parity anchors**: branching-flow 0 (was 0), go-finding 2 (was 2), CV 0 (zero-ratchet preserved).
    - go-cqrs-lite 224 (was 223 on 2026-10-05 — +1, consistent with 6 days of repo churn, not a detection drift; frozen-corpus parity gate is authoritative and green).
    - go-branded-id 0 in non-testdata code (the 28 parity findings live in skipped testdata fixtures + the legacy sample.go).
    - Error-gated repos clean: **journal** = 65 note + 4 warning, **zero error-severity BD003** → BuildFlow guard (error_rules: BD003) not breached; go-business-rules = 0 findings.
    - DEFER repos confirmed comment-only (no live STRONG_ID config): DiscordSync 218 findings (advisory policy, was never error-gated), webphone 6.
    - BuildFlow repo itself: 1 finding (info-tier, not gated there).
    - Large advisory counts (informational, not gates): Kernovia 1013, CreditReformBilanzampel 945, ActaFlow 223, DiscordSync 218, go-cqrs-lite 224, ci-siblings 335, games 727, standard-bug-tracking-schema 692, picoclaw 510.

---

## b) PARTIALLY DONE

- **T17b report**: the scan ran and anchor comparisons are done (above), but the committed status document (planned `docs/status/<date>_bd003-fleet-verification.md`) is THIS document instead of a dedicated one — acceptable consolidation, noting it here deliberately.
- **T18 knowledge capture**: NOT started as a task, but a slice landed incidentally: branching-flow AGENTS.md removal section, CV AGENTS.md gate description, go-branded-id website docs, both CHANGELOGs. The planned go-branded-id AGENTS.md additions (BD003 contract details, FOD content-addressing trap, `nix hash path` procedure, corruption postmortem) are still MISSING.
- **CV changes uncommitted**: the daemon will sweep them into an auto-commit; no deliberate CV commit message exists yet.
- **BuildFlow embedded linter**: installed binary (`5aaf336`, branch `chore/linter-v050-guard`) embeds linter v0.5.0, NOT v0.5.1 — it lacks the generated-file skip. Harmless today (the only error-gated BD003 repo pair, journal/go-business-rules, has zero error findings and no generated-file exposure), but the embedding is one patch behind.

## c) NOT STARTED

- **T18 proper**: go-branded-id AGENTS.md knowledge capture (BD003 contract, generated-file skip semantics, parity-gate self-healing note, FOD trap write-up), branching-flow TODO_LIST/FEATURES harvest in ITS repo, cross-linking the three status reports, refreshing the 19:18 report (now stale on T16/T17 state).
- **branching-flow TODO_LIST.md**: stale item about "rebuild system profile onto v0.6.5" still present; v0.6.7 is out now and the SystemNix pin serves `/run/current-system/sw/bin/branching-flow` — the installed binary there is whatever NixOS last activated, NOT v0.6.7.
- **Fleet `.buildflow.yml` BD003 adoption beyond the guard repos** — never planned for this session beyond verification.

## d) TOTALLY FUCKED UP (honest ledger)

1. **The parity gate would have silently lied forever** after v0.6.7 removal (empty old-side comparison) — I only noticed because I re-ran it as part of verification luck, not because the plan scheduled a post-removal gate check. Class: unverified-gate failure mode. Fixed (self-healing v0.6.6 worktree), but the plan should have called it out at T16 planning time.
2. **BD003 shipped for a full day (v0.4.0–v0.5.0) with a generated-file parity hole** — ~90 false findings on the first real-world consumer (CV). The parity corpus selection (this repo's files only) was too narrow; no generated fixture existed on either side. The gap was catchable pre-release with one sqlc-shaped fixture.
3. **`nix fmt` churn race**: my first README section replacement failed because the daemon rewrote a URL's letter case between my read and my edit — cost one retry. Known hazard, still bit me.
4. **Earlier handoff's BDD diagnosis was wrong-ish**: the brief said "strong-id spec remnants in the CLI BDD suite"; reality was two fixture-dependent specs (exit-code + markdown rows) whose fixtures relied on strong-id detection. Diagnosed correctly on arrival, but the handoff's precision was misleading — a lesson for handoff quality.
5. **Noticed, not caused**: CV's detector-baseline mixins leg shows pre-existing drift (2 names: `models.ATeamApplicationResponse`, `worker.TaskResult`) — outside my scope, unreported to CV's owner channel until now (it IS reported here).

## e) WHAT WE SHOULD IMPROVE

1. **Parity corpora must include generated files** (one sqlc-shaped + one templ-shaped fixture, both sides zero) — the ecosystem runs on generated code.
2. **Gates that compare two sides must fail loud when a side cannot run** (empty-old-side diff ≠ parity). Generalize the parity script's loud-failure pattern to any A/B instrument.
3. **flake.nix version pins drift on every branching-flow release** — a release-checklist step (or tripwire script like go-branded-id's check-go-pins.sh) should assert tag == flake version at tag time. It was stale for v0.6.6 AND would have been for v0.6.7 without the AGENTS gotcha catching it.
4. **`GOEXPERIMENT=jsonv2` is load-bearing for branching-flow tests** but invisible when invoking `nix develop -c bash -c "go test"` without the flake env — document the exact invocation (I hit a misleading failure set without it).
5. **BuildFlow embedding lags the linter by design friction** (branch + flake.lock + profile reinstall). A make-target or script "re-embed brandid-lint@latest" would close the gap mechanically.
6. **Fleet BD003 counts are now measured but not tracked** — a committed baseline table (like CV's detector-baselines) with the 160-repo scan would make the next sweep a drift check instead of a re-measurement.

## f) NEXT — up to 50 things

**Close out the migration (T18 + hygiene) — do first:**
1. go-branded-id AGENTS.md: BD003 section (contract, `-rules` opt-in, suppression directive, generated-file skip, path exclusions, severity model).
2. go-branded-id AGENTS.md: FOD content-addressing trap + `go mod vendor && nix hash path linter/vendor` procedure (from v0.5.0 session).
3. go-branded-id AGENTS.md: testdata-fixture corruption postmortem + resolution (BuildFlow binary re-embed) + parity-gate self-healing note.
4. Harvest both repos' TODO_LIST.md (branching-flow stale v0.6.5 profile item; CV items).
5. Cross-link the three status reports (19:18, 19:50-ish plan, this one) with one-line pointers.
6. Refresh/annotate the 19:18 phase-execution report (T16/T17 now done).
7. Commit CV's script changes deliberately (daemon sweep is shapeless; a real commit message documents the gate swap).
8. CI on branching-flow: replace the dead billing-blocked runners or accept local-gates-only officially in AGENTS.md (the Release workflow failure is now expected noise — silence it).
9. branching-flow FEATURES.md: nothing further needed — verified.
10. SystemNix: rebuild the branching-flow profile binary onto v0.6.7 (sudo step, owner action) so `/run/current-system/sw/bin/branching-flow` stops serving the old build.

**Fleet hardening:**
11. Commit a fleet BD003 baseline table (repo → count → severity split) in go-branded-id docs/status; wire a drift-check mode into a script.
12. Re-embed brandid-lint v0.5.1 into BuildFlow (go.mod bump + flake.lock update + profile reinstall) — kills the last v0.5.0 embedding.
13. Sweep the 65+ note / 4 warning findings in journal: convert or suppress with directives (owner adjudication).
14. go-cqrs-lite +1 finding: adjudicate (new primitive ID since 2026-10-05).
15. CV mixins baseline drift (2 names): adjudicate or re-snapshot.
16. DiscordSync (218) / webphone (6): schedule the BD003 adoption waves or write standing suppressions — the DEFER debt.
17. ActaFlow (223), Kernovia (1013), CreditReformBilanzampel (945): owner decisions — convert waves or advisory-forever policy notes.
18. BuildFlow's own 1 BD003 finding: brand it (dogfooding).
19. Add a generated-fixture pair (sqlc + templ shapes) to the frozen parity corpus so the class stays covered.
20. Consider `brandid-lint` on PATH via SystemNix pin (currently ~/go/bin manual build) — makes CV's gate leg SKIP-proof on other machines.

**Quality-of-life / small repairs noticed en route:**
21. branching-flow `explain_test.go` nolint warning: "unknown linters in //nolint directives: erraudit" (pre-existing golangci warning).
22. `docs/SPLIT-BRAIN.html` still references strong-id (generated artifact; regenerate or archive).
23. `tools/sweep-linters.sh` REPOS list is 3 repos; consider adding CV/journal for BD003-adjacent sweeps or retire the script's README cross-refs.
24. go-branded-id: add a nix flake app/package attr for the brandid-lint CLI (currently only `checks.x86_64-linux.linter` builds it) so consumers don't hand-build.
25. Linter CHANGELOG: note in 0.5.1 that BuildFlow consumers need re-embedding for the fix.
26. Website: add a "generated files are skipped" line to the guide's suppression section FAQ (one-liner).
27. CV `glossary-check.sh` header still narrates the strong-id era — fine as history, but a one-line "now BD003" pointer would help.
28. branching-flow TODO_LIST: add "port remaining sweep-linters reference numbers post-v0.6.7" or drop the stale table.
29. The `.coverage-baseline` / coverage ratchet in branching-flow may have moved after test ports — run `nix run .#coverage` once.
30. Consider tagging the parity corpus snapshot (git tag on go-branded-id) so "frozen" is mechanically enforced, not convention.

**Bigger follow-ons (post-migration):**
31. BD001/BD003 finding dedup advisory: same decl flagged by both rules → one finding with two rule IDs? (design question).
32. `brandid-lint -format json` (finding JSON) for machine consumers like CV's name-set ratchets.
33. Severity CLI filter (`-min-severity`) mirroring branching-flow's knob.
34. Suppression-file support (`.brandid-lint.yml`) for vendored/generated-adjacent exceptions.
35. SARIF: include `category` per rule for GitHub Security tab splitting.
36. Evaluate porting more branching-flow rules (BOOL_BLIND? FLAG_PARAM?) into go-finding-native linters — the extraction pipeline is now proven twice.
37. branching-flow: retire or archive `docs/drafts/*strong-id*` per the archived/ convention.
38. go-branded-id root module: Go 1.27 floor decision when jsonv2 becomes default (the dual-mode gotcha notes say 1.27 flips the tag).
39. Fleet-wide BD001 sweep re-run with v0.5.1 (generated skip may have REMOVED some BD001 findings repos were suppressing).
40. CV: run the FULL go-change-gate once (not just self-test + leg) to prove the brandid leg in situ.
41. Investigate the `BuildFlow.vendor.bak` directory (stale vendor copy still referencing pkg/strongid in its collect.go — delete or archive).
42. `/home/lars/projects/.corpus-pins/branching-flow@dbdeb0f1` worktree + prunable `/tmp/baseline-bf` worktree: clean up branching-flow's worktree list.
43. Auto-tag workflow on branching-flow fired for v0.6.7 (flake version change) — verify it didn't double-tag.
44. go-finding deep-dive docs in branching-flow still describe STRONG_ID correlation pairs — historical, annotate.
45. Consider a `parity-bd003.sh --fuzz` mode (random small corpora from ~/projects) as a stronger regression harness.

## g) Questions for the owner

1. **SystemNix pin**: `~/go/bin/brandid-lint` is a manual build and `/run/current-system/sw/bin/branching-flow` serves an old version — should I prepare a NixOS/home-manager change pinning BOTH tools to current releases (v0.6.7 / linter v0.5.1), or is the profile rebuild deliberately owner-managed (sudo)?
2. **Fleet BD003 debt policy**: journal has 69 advisory findings (65 note + 4 warning, zero error) and the big advisory repos have 200–1000 each — do you want a conversion-wave roadmap (like CV's zero-conversion) or a standing "advisory forever, error-gate only" policy documented per repo?
3. **BuildFlow master**: it remains red on origin (another session's unreleased toolsdk `Fact` API WIP, daemon-pushed) while my buildable state sits on `chore/linter-v050-guard` — should the v0.5.1 re-embed branch off origin/master again and become the new install source now, or wait until the Fact-API session releases go-finding?

---

*Execution paused here per instruction. Next action on resume: T18 items 1–6, then the f) list in order.*
