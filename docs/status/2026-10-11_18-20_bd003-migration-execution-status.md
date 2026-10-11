# BD003 Migration Execution Status — Full Session Report

**Date:** 2026-10-11 18:20
**Author:** Crush (execution session)
**Plan:** `docs/planning/2026-10-11_04-16_SUPERB-strongid-migration-bd003.md`
**Scope executed:** T01–T15 (partial) of the 18-task plan; Stage 1 complete and shipped, Stage 2 shipped, Stage 3 opened.

---

## a) FULLY DONE ✅

### Stage 1 — BD003 port into `linter/` (T01–T12)

| #   | Task                                                                                                                                                                                                                                                                                                          | Evidence                                                                                                                      |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| T01 | Name heuristics ported (`splitNameTokens`, `containsIDPattern`)                                                                                                                                                                                                                                               | `linter/idpattern.go` + table tests incl. English-word sweep (`solid`, `acid`, `rapid`, `candid`… all negative)               |
| T02 | Primitive table + AST scan (params, named returns, struct fields incl. function-local types, multi-name fields; `String`/`Error` skip; ported exclusion table `cmd/`, `ui/`, `_test.go`, …)                                                                                                                   | `linter/primitives.go`, fixtures `linter/testdata/bd003/{params,fields,returns,suppressed}.go` + `cmd/main.go` exclusion case |
| T03 | BD003 finding construction: advisory severity mapping (high→error, medium→warning, low→info), Range via DeclEndLine, Snippet `name type`                                                                                                                                                                      | `linter/finding.go:findingForPrimitiveID`, shape test pins all fields                                                         |
| T04 | Suggestion engine ported: `extractIDRoot`, `suggestIDTypeName`, `suggestIDBrand`, verb-prefix stripping, before/after code, complete `typeDefinition` metadata                                                                                                                                                | `linter/suggest_id.go` + `suggest_id_test.go`                                                                                 |
| T05 | `-rules` CLI flag (comma list, unknown rule = hard error, BD002 not selectable, `-fix` requires BD001) + `RuleSet` API (`DefaultRuleSet`, `AllRules`, `ParseRuleSet`, `DetectPathWithRules`)                                                                                                                  | `linter/rules.go`, `cmd/brandid-lint/main.go`, CLI + rule tests                                                               |
| T06 | Suppression: BD003 joined the known-rule set; `//brandid-lint:ignore(BD003) <reason>` (trailing / above line / whole-declaration doc directive); suppresses-nothing and redundant directives are BD002; inactive-rule directives silent, malformed always loud                                                | `suppress.go` restructure (suppressionOutcome split), `primitives.go:suppressPrimitiveIDs`, placement tests                   |
| T07 | **Parity gate GREEN**: A/B diff vs branching-flow strong-id on 49-file corpus + fixture dir. 26/26 findings identical; 2 documented accepts (legacy `//lint:ignore STRONG_ID` form; new suppression contract). Gate: `scripts/parity-bd003.sh`, report: `docs/status/2026-10-11_11-50_bd003-parity-report.md` | exit 0                                                                                                                        |
| T08 | Severity ranking + edge guards (embedded structs, multiple names, unnamed params)                                                                                                                                                                                                                             | folded into T03/T04 tests                                                                                                     |
| T09 | Lint + fmt + tests green: `nix run .#test` (both modules, both JSON modes), `nix run .#lint` (0 issues), `nix flake check` (all checks passed), golangci via `buildflow -s golangci-lint` (3/3)                                                                                                               | session logs                                                                                                                  |
| T10 | Docs: `doc.go` semantic split (hygiene vs adoption, opt-in contract, suppression law), website guide BD003 section + `-rules` usage, README table row                                                                                                                                                         | committed + deployed                                                                                                          |
| T11 | **Released `linter/v0.4.0`**: CHANGELOG section cut, Version bumped, annotated+signed tag pushed, GitHub Release verified (full CHANGELOG body), website changelog updated + Firebase deployed (`brandedid.web.app`)                                                                                          | release live 2026-10-11 10:32                                                                                                 |
| T12 | Provider Spec description updated (BD001/BD002/BD003 semantics) + test                                                                                                                                                                                                                                        | `provider/provider.go`, `provider_test.go`                                                                                    |

### Stage 2 + fleet (T13–T15 partial)

- **T13 — branching-flow v0.6.6 deprecation RELEASED**: CLI help leads with DEPRECATED + migration commands, `docs/CLI/*` regenerated, `docs/analyzer-reference.md` gained the "Rule-ID / Baseline Migration" table (old flag/rule/suppression → new equivalent, baseline continuity), CHANGELOG section cut, AGENTS.md timeline note, full test suite green, tag `v0.6.6` pushed, GitHub release created manually (workflow blocked by billing — see d).
- **T14 — journal budget test migrated**: `internal/policy/strong_id_budget_test.go` now runs `brandid-lint -rules BD003 -format sarif`, counts BD003 results, treats exit 1 (findings) as a successful scan, and gained a liveness probe (one known primitive ID in a temp module must yield exactly one finding) replacing the old filesScanned check. Budget 69 carried over — test GREEN. Committed + pushed (`7b02cc8`).
- **T15.3 — go-business-rules BDD gate migrated**: the "no strong-id violations" spec now shells `brandid-lint -rules BD003` (exit 0 expected, skip when binary absent). Suite green. Committed + pushed (`a6c2324`).
- **T15.4 (partial) — BuildFlow re-pin**: BuildFlow `go.mod` bumped `go-branded-id/linter` v0.2.0 → v0.4.0 (ends the embedded-provider fixture-corruption class once shipped). ⚠️ landed via daemon commit on `feat/toolsdk-provider-switchover`, unpushed, and the repo has since switched branches — see d) and questions.

### Incidents fixed during the session

- **Fixture-corruption incident (3rd historical instance, new mechanism)**: `buildflow format`/`--fix` runs corrupted `linter/testdata/*.go` (duplicate `Name()` stubs inserted). Root cause: BuildFlow's _embedded_ brandid-lint provider (pinned v0.2.0) repaired testdata files — single-file/replayed invocations bypass walk-level dir skipping. Fixes: (1) `underTestDataPath` guard in `RepairPath` — repair now refuses ANY file under a testdata path segment, with regression test `TestRepairPath_NeverTouchesTestData`; (2) BuildFlow result cache purged (927 MB → 82 MB); (3) fixtures restored from the last clean commit and committed (`a419545`); (4) BuildFlow re-pin (above) as the fleet-level fix. Guard shipped in linter/v0.4.0.
- **CI: validate-docs workflow fixed**: `md-go-validator@latest` now needs go ≥ 1.27 but the job pinned 1.26 (`GOTOOLCHAIN=local`) — every master push failed doc validation. Bumped that job to 1.27 (matches linter pin; go-pins tripwire green). Committed + pushed.
- **Website deploy unblocked**: dependabot bumped `package.json` without the lockfile; `pnpm install` resync + deploy succeeded; lockfile committed + pushed.

---

## b) PARTIALLY DONE ⚠️

### `linter/v0.5.0` — provider `rules` tool option (discovered mid-T15)

Found a **lockstep blocker** during the fleet survey: BuildFlow dogfoods `branching-flow:STRONG_ID` as an `error_rules` REGRESSION guard, and no repo could enable BD003 through BuildFlow at all (provider always ran the default rule set). Implemented:

- `linter.DetectWithRules(ctx, rules)` (analog of Detect with explicit rules)
- Provider `rules` tool option (toolsdk `Option`, parsed via `ParseRuleSet`, invalid value = hard Detect error, absent = default set byte-identical) — requires toolsdk v1.15.0 (bumped from v1.13.1)
- Provider tests: default-without-options stays BD003-free, `rules: BD003` enables it, invalid rule errors
- CHANGELOG `[0.5.0]` cut, Version const bumped to 0.5.0

**INCOMPLETE**: `linter/vendor/` re-vendored but **flake's `vendorHash` NOT updated** → `nix flake check` RED on the linter derivation; not formatted/lint-finalized; **not tagged, not pushed, not released**. 5 files currently uncommitted (daemon may sweep them).

### T15 fleet re-pointing

- webphone: **DEFER** decided (baselined STRONG_ID noise deliberately; error_rules elevation pending upstream anyway) — note only, no change made.
- DiscordSync: **DEFER** decided (STRONG_ID inside 2200 warnings budget = advisory; adoption optional later) — note only.
- Verified: no fleet repo error-gates on STRONG_ID except BuildFlow's own dogfood entry (the blocker above).

---

## c) NOT STARTED ❌

- **T16** — branching-flow removal release (delete `pkg/strongid` + CLI + LinterKey + stats + goldens + toolsdk description, `all` 14→13). Blocked on: v0.5.0 release + BuildFlow re-point of its STRONG_ID regression guard (lockstep guardrail).
- **T17** — fleet BD003 verification scan + triage report (counts vs old STRONG_ID baselines, budgets recorded).
- **T18** — knowledge capture: go-branded-id AGENTS.md (BD003 contract, opt-in default, fleet state, BuildFlow-embedded-provider gotcha), branching-flow AGENTS.md final state, TODO_LIST harvest, archive/cross-link planning docs.
- BuildFlow's `.buildflow.yml` error_rules re-point (`branching-flow:STRONG_ID` → `brandid-lint:BD003` + `tool_options: brandid-lint.rules`) — needs v0.5.0.
- Website redeploy with the 0.5.0 provider-option docs (pages not yet updated for it either).

---

## d) TOTALLY FUCKED UP 💥

1. **Fixture corruption loop — my own doing, twice over.** I ran `buildflow format` / `buildflow --fix` in this repo while the orchestrator embedded brandid-lint v0.2.0. First corruption: I restored the fixtures and moved on WITHOUT understanding the writer. It corrupted again on my next full run. Only then did I root-cause (embedded old provider + single-file repair path) and fix at four levels. **Lesson: restore-and-continue without root cause = guaranteed repeat.**
2. **Wasted debugging cycles on my own test bug.** `TestParseRuleSet/BD001,` flip-flopped confusingly across runs; I suspected daemon rewrites and SplitSeq semantics before noticing my test compared rule-set VALUES on error paths (where partial rules legitimately remain). Diagnosed only via an in-process probe.
3. **Release sequencing miss:** the provider `rules` option should have been part of v0.4.0 — I discovered the BuildFlow-regression-guard dependency only during the T15 fleet survey, forcing a second release (v0.5.0). A 10-minute earlier read of BuildFlow's `.buildflow.yml` would have caught it.
4. **Edit-tool friction vs the daemon:** several edits rejected by mtime checks after daemon commits; re-reads cost time. Should have batched file writes and expected churn (my own context notes even warned me).
5. **Branching-flow release workflow blocked by GitHub billing** ("recent account payments have failed") — not my failure, but I worked around it with a manual `gh release create`; automation remains broken for future releases.

---

## e) WHAT WE SHOULD IMPROVE 🛠️

1. **Never run full `buildflow --fix` in go-branded-id** until the installed BuildFlow embeds linter ≥ v0.4.0. Targeted steps (`-s golangci-lint`) + `nix fmt` are safe. Record in AGENTS.md (T18).
2. **BuildFlow needs a re-pin discipline**: its embedded provider pins lag releases; a stale embedded linter can act on repos with old semantics. Consider a CI check in BuildFlow for provider-pin staleness vs latest tags.
3. **Pre-release checklist should include "who consumes this?"** — I shipped v0.4.0 without scanning BuildFlow's own config for STRONG_ID usage; the consumer survey belonged in release prep, not post-release.
4. **Test-design rule learned**: on error paths, never assert full struct equality unless the contract guarantees zero-value returns (partial-rule application bit me).
5. **`nix flake check` before every tag** caught formatting drift twice — keep it as the hard pre-tag gate (it already is; I just hit it twice).
6. **Dependabot lockfile drift**: website `package.json` bumps arrive without lockfile updates, breaking deploys; the dependabot job should sync both or the deploy should run `pnpm install` defensively (it now must, until fixed).
7. **linter/v0.3.0 was never tagged** (this morning's other session cut the CHANGELOG but the tag list shows no `linter/v0.3.0`) — release hygiene gap worth closing or consciously skipping.

---

## f) NEXT — up to 50 things, in priority order

**Unblock and finish v0.5.0 (critical path):**

1. Get new `vendorHash` from the FOD failure (`nix build .#checks.x86_64_linux.linter` → copy `got: sha256-…`), update `flake.nix:150`
2. `nix flake check` green
3. `buildflow -s golangci-lint` green on new provider code
4. Website guide: document the BuildFlow `tool_options: { brandid-lint: { rules: "BD001,BD003" } }` usage
5. Website changelog page: add `[linter/v0.5.0]` section
6. Commit remaining 5 WIP files deliberately, tag `linter/v0.5.0` (annotated + signed), push tag
7. Verify GitHub Release picked up the CHANGELOG section
8. `nix run .#deploy` (website) from `website/`

**BuildFlow (lockstep blocker for T16):**
9. Decide branch strategy for the re-pin (see question 1) — re-apply `go get github.com/larsartmann/go-branded-id/linter@linter/v0.5.0` on the right branch
10. Re-point BuildFlow's own `.buildflow.yml`: `error_rules: branching-flow:[…, STRONG_ID]` → `brandid-lint:[BD003]` + `tool_options` rules=BD001,BD003
11. Verify BuildFlow's 25-site stringly-ID regression guard fires under BD003 (seed one violation in a scratch branch)
12. Wait for/trigger a BuildFlow release so the installed binary embeds linter ≥ v0.4.0 (ends fixture-corruption exposure fleet-wide)
13. Add provider-pin-staleness CI check in BuildFlow (improvement #2)

**T16 — branching-flow removal (after lockstep satisfied):**
14. Delete `pkg/strongid/` + tests; fix `analysis.RunAll`; `LinterKey` 14→13
15. Delete `cmd/branching-flow/strong_id*.go`; update `allLinters` list; regen golden tests
16. Remove `IgnoreCommentTypeStrongID` + stats suppression split entries
17. toolsdk Spec description 14→13 linters; regen docs (`generate-docs`)
18. Update `docs/analyzer-reference.md` (drop strong-id rows, keep migration section as historical)
19. Full branching-flow suite green; cut CHANGELOG `[0.6.7]`; tag + push (+ manual release if billing still broken)
20. Run parity gate one final time BEFORE deletion for the record

**T17 — fleet verification:**
21. Fleet scan `brandid-lint -rules BD003` across `~/projects/*` (excluding branching-flow); capture counts
22. Diff counts vs old STRONG_ID baselines (journal 69, DiscordSync ~1872-bucket, BuildFlow 0-after-cleanup)
23. Write `docs/status/<date>_bd003-fleet-verification.md` with per-repo budgets
24. Record webphone/DiscordSync DEFER decisions with rationale in that report

**T18 — knowledge capture:**
25. go-branded-id AGENTS.md: BD003 section (contract, opt-in, suppression law, exclusions, tool_options knob)
26. go-branded-id AGENTS.md: "never full-`buildflow --fix` until BuildFlow ships linter ≥ v0.4.0" gotcha
27. go-branded-id AGENTS.md: fixture-corruption 4th-instance postmortem (embedded-provider mechanism)
28. branching-flow AGENTS.md: final state after removal
29. TODO_LIST.md: mark migration row progress; harvest follow-ups from this report
30. Cross-link planning doc ↔ parity report ↔ this status doc
31. TODO_LIST: BuildFlow re-pin + release follow-up item

**Deferred/optional (from plan tail + improvements):**
32. Retro-tag `linter/v0.3.0` at commit `a2d351d` or document the skip
33. journal: consider enabling BD003 via BuildFlow tool_options once BuildFlow ships v0.5.0-embed (replaces the CLI-shelling budget test eventually)
34. go-business-rules: same consideration
35. DiscordSync: if adopting BD003, set `tool_options` + re-measure the 2200 budget without STRONG_ID
36. webphone: re-visit DEFER after their error_rules elevation lands
37. Website: add a "migrating from branching-flow" landing section (bigger than the guide note)
38. Consider `-format finding` JSON output for brandid-lint (journal's old envelope had summary counts; SARIF counting is clunkier)
39. Parity gate: wire into `linter/` CI as a job (currently manual script; branching-flow removal will break it → convert to frozen-corpus-only mode)
40. `linter/flake pname` still says `brandid-lint-0.1.0` in derivation names — cosmetic version-string plumbing
41. SECURITY.md missing 'Security Practices' section (buildflow securitymd finding, pre-existing)
42. go-structure-linter: AGENTS.md exceeds 220-line budget (pre-existing; grew again this session — T18 should trim, not grow it)
43. Root-module `.golangci.yml` duplicates `linter/.golangci.yml` 246 lines (art-dupl finding, pre-existing)
44. Dependabot lockfile-sync fix for website (improvement #6)
45. GitHub billing: fix account payments so Release/GoReleaser workflows run again
46. `nix flake check --all-systems` cross-arch eval in CI (already split passes; verify both green post-v0.5.0)
47. Bump BuildFlow go.sum tidy state after branch dance settles
48. Consider BD003 `Suggestion` metadata JSON schema doc on the website (tooling contract)
49. Fleet-wide suppression audit: any repo adopting BD003 writes `ignore(BD003) <reason>` — spot-check reasons exist (BD002 polices, but audit anyway)
50. Celebrate: single source of truth for ID linting once T16 lands

---

## g) QUESTIONS (cannot figure out myself) ❓

1. **BuildFlow branch strategy**: my linter-pin bump (v0.4.0, soon v0.5.0) was daemon-committed onto `feat/toolsdk-provider-switchover`, which is unpushed, and the repo has since moved to `fix/tailwind-build-excludes-minify` (another session's work, ahead 1 of master). Which branch should carry the re-pin, and do you want me to cherry-pick it there — or are you handling BuildFlow's branch convergence yourself?
2. **webphone + DiscordSync DEFER confirmation**: I decided both repos defer BD003 adoption (webphone baselined STRONG_ID as accepted noise; DiscordSync counts it inside a 2200 advisory budget). Confirm, or should I enable BD003 via `tool_options` for either once v0.5.0 + BuildFlow ship?
3. **GitHub billing**: branching-flow's Release workflow (GoReleaser job) fails with "recent account payments have failed or your spending limit needs to be increased" — I created the v0.6.6 release manually. Is that known/being fixed? It affects the T16 removal release automation too.

---

**Bottom line:** Stage 1 + 2 shipped (BD003 live as `linter/v0.4.0`, branching-flow v0.6.6 deprecation out, journal + go-business-rules migrated and green, parity proven). Stage 3 is blocked on finishing `linter/v0.5.0` (vendorHash!) and the BuildFlow re-pin/re-point. The one genuinely bad moment — repeated fixture corruption — is fixed at four levels and shipped inside the release.
