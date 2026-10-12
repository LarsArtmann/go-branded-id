# BD003 — Re-Embed b)-List Closure + BuildFlow Lint-Gate Recovery — Session Report

**Date:** 2026-10-12 19:10 (Monday)
**Trigger:** Owner directive (twice repeated): "READ, UNDERSTAND, RESEARCH, REFLECT. Break this down... Execute and Verify them one step at a time. Keep going until everything works."
**Scope:** Close the 15:32 report's b)-list (verification gaps from the BuildFlow re-embed) plus every autonomous f)-list item that needs no owner decision. The three g) questions were answered by reality where reality had already answered them.
**Repos touched:** BuildFlow (code + docs), go-branded-id (docs), SystemNix (docs). CV, go-cqrs-lite, journal: read-only verification/scans only.

---

## a) FULLY DONE

1. **b.0 — Repo/origin reality check:** Both repos were already PUSHED since the last session (BuildFlow master @ `8a349283a` = origin; go-branded-id @ `70e3226` = origin) — g.1 and f.8 resolved without action. **BuildFlow CI is red on GitHub Actions BILLING** ("recent account payments have failed / spending limit"), not code: both Build Gate jobs (go build, link audit) never started. This extends the known f.8/10 billing issue to BuildFlow's org — owner action. go-branded-id CI green (Go + Validate Docs on the pushed docs commits).
2. **b.1 — erraudit gate #132 (missed after the dep sweep): FAILED then FIXED.** First run: 1 violation — `compileGate` (tools/gomod/update.go:297) returned bare `error` (`generic_return`, pre-existing from the BuildFlow#45 WIP). Fixed at the signature level (no suppression; the gate runs `--no-suppress` anyway): `compileGate` and `NewToolchainFloorError` narrowed to `*errorfamily.Error`, matching the repo's newer constructor pattern (`NewVendoringInconsistencyError`). Typed-nil boxing audited per AGENTS gotcha #213 — caller only uses `buildErr` behind a concrete `!= nil`, all returns wrap it, success path is literal `nil`. Gate now **34/34 modules, 0 violations**.
3. **b.2 — Release-gating pair:** `buildflow doctor --strict` exit 0 (13 pass / 1 advisory / 0 fail); `buildflow docs --check` exit 0. The advisory (`vendor/fod-hash-freshness`) is the DESIGNED committed-split-order branch (TODO MS4): adjudicated green via `buildflow -s nix-hash-fix` (no drift) + `nix build .` exit 0. Not a bug.
4. **b.3 — Live BD003 probe through the installed binary (the gold-standard gap): CLOSED.** Isolated scratch repo (git init + one violating file + `.buildflow.yml` with `tool_options: brandid-lint: {rules: BD003}`) run through the profile-installed buildflow 8a34928: full structured finding rendered end-to-end (`orderID` fn param → suggested `OrderID = id.ID[OrderBrand, string]`, severity error, category type-safety), findings gate exit 69. Negative control (`rules: BD001`) exit 0, zero findings — opt-in semantics verified through the pipeline. Scratch trashed. Side observation: summary shows `filesScanned: 0` with `filesAffected: 1` — diagnosed as the provider still using the v1 findings-only `DetectorFunc` while go-finding v1.15.0 ships `DetectResult.Summary.FilesScanned`; recorded as a TODO row (next linter release).
5. **b.4 — LSP:** project-wide diagnostics on go-branded-id: empty. The ~45 stale "unused" warnings are gone.
6. **b.5 — Stale generation comment:** BuildFlow flake.nix's go-structure-linter input comment rewritten truthfully: pin `d77b3f30` (still == upstream master) requires go-finding v1.14.0/toolsdk v1.15.0, one generation behind the shared input's v1.15.0; combination gate-green; realign on the next go-structure-linter release. (Previous claim "same generation as the shared input" had become false.)
7. **Post-sweep tidy gap (new catch):** the preflight flagged **18 modules needing `go mod tidy`** — the dep sweep's `go get` left old-version go.sum residue. `nix run .#deps` run: replace graph converged, v1.14.0 residue gone, 20 files updated, **f.9 resolved by MVS**: all modules now uniformly require root v1.15.0 / pipeline v1.14.0 / analysis v1.13.0 (the cross-module split is gone; forcing sub-module max tags would fight MVS). vendorHash re-verified current.
8. **Lint-gate recovery (new catch):** `buildflow format` exposed **15 golangci-lint error findings on master** (invisible while CI was billing-blocked; the last session's green claim was tests-only). All fixed:
   - `tools/providers/tmp_region_test.go` DELETED (documented decision, pre-edit this time): swept debug scaffolding — zero assertions, reads a `/tmp/dprint_region.txt` scratch fixture, pure fmt.Println walk, t.Skips on every clean machine. Carried 11 of the 15 findings (wsl_v5/nlreturn).
   - 7× ineffassign dead stores (`x := make(...)` immediately overwritten by `lo.Map/Filter/SliceToMap`) across workflow_summary_output, list_cmd_json (2), language_cmd_json (2), doctor_only, live_dashboard_smoke_test, duration_json_audit_test — collapsed to direct `lo.*` calls.
   - 1× nonamedreturns in go_auto_upgrade_generated.go (hand-written file despite the name) — unnamed returns + locals.
   - **1 REAL BUG fixed:** `renderConfigSkipped` built the `"N skipped via config:"` header then overwrote it with the `lo.Map` body — the header silently never rendered (the ineffassign finding was the symptom). Now appends the body to the header, matching its sibling `renderFilteredTools`.
   - Verification: `buildflow -s golangci-lint` exit 0; full workspace test gate **34/34 GREEN exit 0** (after ALL of today's changes).
9. **f.31 (scoped) — em-dash ratchet:** the 4 baselined legacy dashes in tools/gomod/update.go converted to ASCII punctuation; baseline regenerated (226 entries, update.go line removed); `TestNoNewEmDashes` PASS. The full drive-to-zero (~226 files) stays ZCC11's owner decision.
10. **f.14/18 — CV gate re-run against the new binary:** `scripts/go-change-gate.sh --quick` at CV @ `2c3b77671`: **all checks green, exit 0** through buildflow 8a34928 (brandid-lint v0.5.1 embedded).
11. **f.17 — go-cqrs-lite +1 finding IDENTIFIED:** Oct-5 tree (`git archive 9dccf72`, read-only) rescanned = exactly 223 vs 224 on master; finding-signature diff proves **net churn, not detection drift**: +9 new INFO findings (7× `systemscenario/chaos.go` streamID/eventID test-DSL params, `core/v5/id/actor_id.go` NewServiceActor serviceID, `metaengine/enginetest/keyset_pagination.go` keysetRow.ID) offset by 8 genuine fixes (dedup/ring ×2, mesh-demo ×4+, projectionhost ×2, taskmanager ×3, scheduler ×1); `record/` → `core/v5/record/` are 7 relocations. All new findings are INFO in test/example infra — convert-or-suppress stays scoped to the next go-cqrs-lite pass. TODO row updated with the full identification.
12. **f.30 — SystemNix FEATURES.md:** `art-dupl + 10 Go tools` row → `+ 11` with `brandid-lint` added (package already wired at `lib/lars-packages.nix:36`, master-tracking; only the inventory row was stale).
13. **Knowledge patches:** BuildFlow AGENTS.md post-dep-sweep checklist gained the missing tidy step (`nix run .#deps` — exactly the gap that bit this and the last session); BuildFlow TODO_LIST DD20 annotated (re-baseline confirmed again); go-branded-id AGENTS.md gained the validated end-to-end probe recipe + filesScanned gap note; go-branded-id TODO_LIST: re-embed row updated (pushed; CI billing), new FilesScanned/DetectResult row, f.17 row enriched.

**Verification spine for everything above:** builds + vet green per module; erraudit 34/34; golangci-lint exit 0; full test gate 34/34 exit 0; nix build exit 0; doctor/docs/nix-hash-fix all exit 0; CV gate green.

---

## b) PARTIALLY DONE

1. **Today's BuildFlow commits are local-only** (daemon-swept, not pushed — push stays owner-gated per standing rule; ~5 new commits on top of `8a349283a`). Same for go-branded-id/SystemNix doc commits.

---

## c) NOT STARTED (owner-blocked or wave-scoped)

1. **GitHub Actions billing fix** (owner account action) — blocks ALL BuildFlow CI signal; local gates are the only evidence until fixed.
2. **SystemNix sudo activation** (`sudo nixos-rebuild switch`) + g.3 commit-discipline decision — unchanged.
3. **g.2 Wave 1 (journal, 69 findings)** — still awaiting owner "start now".
4. **Fleet bump** of remaining downstream repos to library v0.7.0 / linter v0.5.1 (go-ecosystem-upgrade wave).
5. **f.18 BuildFlow dogfood BD003** (enabling via `.buildflow.yml` would need an inventory first — BuildFlow's own code has ID-ish params; policy decision).
6. Linter hardening f.20–28 (parity fixtures, drift-check script mode, json format, min-severity, SARIF category, dedup, FilesScanned/DetectResult upgrade, Go 1.27 floor decision).
7. branching-flow housekeeping f.29 (drafts archival, sweep-linters, CI-billing items), ROADMAP Theme 3.

---

## d) TOTALLY FUCKED UP

**Nothing unrecovered.** Two near-misses, both caught by process:

1. **Daemon mtime-trap twice** on flake.nix and update.go edits ("modified since last read") — both recovered by re-reading and re-applying; content was unchanged (mtime-only churn). The trap is now expected behavior, not a surprise.
2. **Broken forensic intersection in the f.17 analysis** (comm over `path:line` vs `path` lists silently produced an empty intersection that briefly suggested "no findings in changed files" — contradicted by the count delta). Caught because the empty result contradicted the known +1; fixed the comparison granularity before drawing conclusions. Same failure class as d)-2 from the 15:32 report (silent empties); the reflex that caught it is the improvement working.

---

## e) WHAT WE SHOULD IMPROVE

1. **"Green master" must include the LINT gate, not just tests.** The 15:32 report's green claim was the 34/34 TEST gate; the lint gate was red with 15 findings (billing-blocked CI hid it). Baseline definition now: tests + lint + erraudit + doctor/docs + nix build. The 15:32 e.2 lesson ("baseline means the full gate") under-scoped "full".
2. **The post-dep-sweep checklist in BuildFlow AGENTS.md was incomplete** (no tidy step) — patched this session so the next sweep doesn't rediscover it.
3. **Scratch-probe pattern worked well** — the isolated git-init fixture gave the gold-standard end-to-end evidence without touching any fleet repo; keep it as the canonical probe method (now documented in go-branded-id AGENTS.md).

---

## f) NEXT (autonomous vs owner)

1. (owner) Fix GitHub Actions billing; then re-check BuildFlow CI on master.
2. (owner) g.2 Wave 1 journal — start?
3. (owner) g.3 SystemNix commit discipline + sudo activation.
4. (owner) Push today's local commits (BuildFlow ~5, go-branded-id docs, SystemNix docs) or leave to the daemon/owner flow.
5. (autonomous, next session) Fleet bump wave (14 repos → library v0.7.0 / linter v0.5.1) via go-ecosystem-upgrade.
6. (autonomous) Linter release cycle: FilesScanned/DetectResult provider upgrade (+ the BD001 ignore directives for go-cqrs-lite marker brands while at it).
7. (owner decision) ZCC11 em-dash prune sweep; f.18 BuildFlow dogfood.

---

## g) QUESTIONS FOR THE OWNER

1. **CI billing** — BuildFlow's Build Gate can't run (payments/spending limit). Fix in GitHub settings? (go-branded-id CI still runs; the block appears per-account/org.)
2. **Wave 1 (journal, 69 findings): start now?** (Carried over unchanged.)
3. **SystemNix commit discipline (g.3)** — retry the deliberate lock commit or accept the daemon sweep? (Carried over unchanged.)
4. **Push today's BuildFlow commits?** The erraudit fix, tidy sweep, lint recovery, and comment updates are local-only; a push would also finally exercise CI — once billing is fixed.

---

**Session verdict:** The re-embed is closed to gold standard (live end-to-end probe included), the BuildFlow gates are ALL green again (tests, lint, erraudit, doctor, docs, build) after two hidden debt classes were caught and fixed (post-sweep residue + billing-blocked lint drift), and every f)-list item that needed no owner decision is either done or precisely identified. Remaining opens are owner decisions (billing, waves, commits, pushes) and the long-tail linter-hardening backlog.
