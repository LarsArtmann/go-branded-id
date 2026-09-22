# Status Report — 2026-07-27 10:39 CEST

## Flake `outputs` Bug Fix + Brutal Self-Review

**Session scope:** Fix three failing BuildFlow steps (`nix-fmt`, `nix-build-verify`, `nix-hash-fix`) and honestly assess what was done well, what was forgotten, and what should improve.

**Trigger:** User pasted BuildFlow output showing `error: function 'outputs' called with unexpected argument 'nixpkgs'` cascading into all three Nix-based steps.

---

## The Bug, The Fix, The Verification

### What was broken

Commit `530702f` ("chore(main): comprehensive project maintenance and quality improvements") removed `nixpkgs` from the `outputs` destructure pattern in `flake.nix:17-23`:

```diff
   outputs =
     inputs@{
       self,
-      nixpkgs,
       flake-parts,
       treefmt-nix,
       systems,
     }:
```

The pattern had **no ellipsis (`...`)**, making it strict. Nix always calls `outputs` with **every** input declared in the `inputs` block — including `nixpkgs`. An unnamed-but-present argument is rejected → `unexpected argument 'nixpkgs'` → every flake evaluation explodes → `nix-fmt`, `nix-build-verify`, `nix-hash-fix` all fail.

### The fix

Added `...` to the pattern (`flake.nix:23`):

```nix
outputs =
  inputs@{
    self,
    flake-parts,
    treefmt-nix,
    systems,
    ...
  }:
```

**Why `...` and not re-adding `nixpkgs`:** `nixpkgs` is never referenced directly in the `outputs` body. It flows through `inputs` → `mkFlake { inherit inputs; }` → arrives as `pkgs` in `perSystem`. Naming it explicitly in the destructure would imply it's used there, which is a lie. The ellipsis is the honest, idiomatic flake-parts pattern. (Confirmed: `website/flake.nix:21` uses exactly this pattern — `inputs@{ self, flake-parts, ... }`.)

### Verification (all green)

| Check                                   | Result                                                                             |
| --------------------------------------- | ---------------------------------------------------------------------------------- |
| `nix flake check --no-build`            | all checks passed                                                                  |
| `nix build .#checks.x86_64-linux.build` | succeeds                                                                           |
| `nix fmt`                               | 0 changed (already formatted)                                                      |
| `buildflow -s nix-fmt`                  | exit 0                                                                             |
| `buildflow -s nix-hash-fix`             | passed (1/1)                                                                       |
| `buildflow -s nix-build-verify`         | cascade-only step — no longer triggers (was only failing because `nix-fmt` failed) |
| **`buildflow` (full suite)**            | **37/38 passed, 0 failed, 1 skipped by config (`gitleaks`)**                       |

---

## a) FULLY DONE

1. **Root-caused the error** — traced the `unexpected argument 'nixpkgs'` to the strict `outputs` pattern introduced in `530702f`.
2. **Applied the minimal correct fix** — one line (`...`), no scope creep, no unrelated changes.
3. **Verified end-to-end** — ran `nix flake check`, `nix build`, `nix fmt`, and the full `buildflow` suite. All previously-failing steps now pass.
4. **Confirmed blast radius is contained** — checked `website/flake.nix`; it already uses the correct `...` pattern. Only the root `flake.nix` was affected.
5. **Identified the original commit** — `530702f` removed both `nixpkgs` from `outputs` AND `system` from `perSystem`. The `perSystem` removal was safe (it has `...`), but the `outputs` removal was the landmine.

---

## b) PARTIALLY DONE

Nothing. The fix itself is complete and verified.

---

## c) NOT STARTED

1. ~~**AGENTS.md update** — the "Critical Gotchas" section should document: _"The `outputs` pattern must include `...` if it doesn't name every input. Nix passes all declared inputs; a strict pattern without `...` breaks the entire flake."_ This is textbook "hard to discover from code alone" context.~~ done (done — 'Flake outputs Must Include ...' gotcha added to AGENTS.md (2026-07-28))
2. ~~**Commit message quality** — see section (d) below. The auto-git daemon committed the fix with a generic message.~~ **Won't implement — history stands as-is; the lesson is recorded here and in later reports.**
3. ~~**Root-cause of the root-cause** — WHY did `530702f` remove those lines? Was it an agent refactor? A linter suggestion? Understanding this prevents the same class of bug recurring.~~ **Won't implement — origin unprovable after the fact; fix-on-sight + explicit-commit policy adopted instead.**

---

## d) TOTALLY FUCKED UP

### The auto-git daemon wrote a lying commit message

The daemon committed my fix as `4b6b4b5`:

```
chore(nix): update flake.nix configuration

- Update Nix flake inputs and dependencies to latest stable versions
- Refresh Go toolchain and development shell packages
- Ensure reproducible build environment across development machines
- Maintain compatibility with latest nixpkgs release
- Update flake.lock hashes for transitive dependency updates
```

**Every single bullet point is false.** I changed ONE line — added `...` to a pattern. I did not touch inputs, dependencies, the toolchain, the dev shell, or any lock hashes. This commit message actively misleads anyone reading `git log`. If someone bisects a flake issue in the future, this commit will send them down completely wrong paths ("must be a dependency update!") when the actual change was a one-character bug fix that unblocked the build.

**I should have committed immediately** with a proper message like:

```
fix(nix): add ellipsis to outputs pattern to accept all flake inputs

Commit 530702f removed nixpkgs from the outputs destructure but left the
pattern strict, causing "unexpected argument 'nixpkgs'" and breaking all
flake evaluations (nix-fmt, nix-build-verify, nix-hash-fix).
```

By not committing, I ceded the commit message to the daemon. **Lesson: for a fix this critical (unblocks entire CI), commit immediately with a precise message.**

### What I forgot during the session

1. **Did not check `website/flake.nix` during the fix** — only checked it just now for this report. If the same bug had been there, the website build would still be broken. I got lucky; the website already used `...`.
2. **Did not update AGENTS.md** — direct violation of my own "Aggressive Update Protocol" which states: _"Update at the moment of discovery, not end of session."_ I discovered a non-obvious gotcha and walked away without recording it.
3. **Did not investigate the human/agent intent behind `530702f`** — I treated the symptom (bad pattern) without asking why the lines were removed. If this was an agent making "improvements," the same agent will break the next flake it touches.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Commit critical fixes immediately** — don't let the auto-git daemon write the message for one-line build-unblocking fixes. The daemon's templated messages are fine for routine churn but actively harmful for surgical bug fixes.
2. **Add a `flake check` gate to the daemon** — or to a pre-commit hook — that fails if the flake doesn't evaluate. `nix flake check --no-build` takes <1 second on evaluation and would have caught `530702f` before it landed.
3. **Document the `outputs` pattern rule in AGENTS.md** — this is a Nix gotcha that's invisible until it breaks everything.
4. **Review agent-generated "maintenance" commits before merging** — `530702f` bundled a flake refactor, lint config changes, new tests, and AGENTS.md updates into one commit. The flake change was buried and unreviewed. Large "comprehensive maintenance" commits are where bugs hide.

### Codebase improvements (noted but not addressed — out of session scope)

1. ~~**`go-structure-linter` warnings (21 errors)** — all "root-package-files" complaints.~~ Won't implement — intentional flat layout documented in AGENTS.md; tool noise accepted. AGENTS.md says the flat single-package layout is intentional. These warnings are noise; the linter config should be told to accept the intentional layout rather than emitting 21 errors every BuildFlow run.
2. ~~**`gitleaks` skipped by config** — is this intentional? If so, document why. If not, re-enable.~~ done — the current `.buildflow.yml` declares its skip list explicitly (`go-mod-update` only); no gitleaks skip remains.
3. ~~**`assets/` and `internal/` directory warnings** — same class as above; intentional layout flagged by a generic linter.~~ Won't implement — intentional layout.

---

## f) Up to 50 things we should get done next

Ranked by impact (Pareto: top ~5 deliver 80% of value).

### High impact — do first

1. ~~**Amend commit `4b6b4b5`** with an honest message describing the actual one-line fix (if history rewrite is acceptable; otherwise add a clarifying follow-up commit).~~ **Won't implement — history stands as-is.**
2. ~~**Add `outputs` pattern gotcha to `AGENTS.md`** "Critical Gotchas" section.~~ done (done — AGENTS.md gotcha added 2026-07-28)
3. ~~**Add `nix flake check --no-build`** as a pre-commit or BuildFlow step to catch flake evaluation errors before they land.~~ done (flake-check CI job in go.yml)
4. ~~**Investigate `530702f`** — determine if an agent or tool suggested removing `nixpkgs`/`system`. If an agent, add a guardrail so it doesn't recur.~~ **Won't implement — origin unprovable after the fact.**
5. ~~**Audit other repos in the ecosystem** (14 downstream repos per AGENTS.md) — if they use the same flake-parts pattern, they may have the same latent bug.~~ done (go-linter-sdk self-regression found and fixed in the 2026-09-17 fleet pass)

### Medium impact — technical debt

6. ~~**Configure `go-structure-linter`** to suppress the intentional root-package layout (21 errors → 0). Either via config ignore or a `//nolint`-equivalent.~~ **Won't implement — intentional layout documented; noise accepted.**
7. ~~**Document or re-enable `gitleaks`** in BuildFlow config.~~ done (.buildflow.yml declares its skip list explicitly (go-mod-update only))
8. ~~**Add a regression note** — somewhere testable (CI or a flake eval smoke test) that the `outputs` function accepts all inputs.~~ done (the CI flake-check job is the regression gate)
9. ~~**Review the `530702f` commit holistically** — it changed `.golangci.yml`, added tests, updated AGENTS.md, and broke the flake. Were the other changes reviewed? Are the new tests correct?~~ **Won't implement — superseded by later full audits.**
10. ~~**Standardize the `outputs` pattern across all flakes** in the ecosystem — enforce `...` unless all inputs are explicitly named and used.~~ done (linter-stack flakes verified in the 2026-09-17 pass)
11. ~~**Add a `flake eval` smoke test** to CI (`go.yml`) that runs `nix flake show` or `nix flake check --no-build` before the full build.~~ done (flake-check job runs the --all-systems eval plus current-system build)
12. ~~**Update `AGENTS.md` "Essential Commands"** to mention `nix flake check --no-build` as a quick smoke test.~~ done (AGENTS.md lists nix flake check)

### Lower impact — polish

13. ~~**Review whether the `checks.build` derivation** needs the `GOCACHE` workaround documented in AGENTS.md — is it still necessary with current nixpkgs?~~ done (still required; documented in AGENTS.md (Nix Sandbox GOCACHE))
14. ~~**Add `meta.description` to all `apps`** — BuildFlow emits warnings for every app lacking description (7 warnings).~~ **Won't implement — not added.**
15. ~~**Consider a `flake-parts` module** for the shared `mkApp` helper — currently duplicated if other repos copy this flake.~~ **Won't implement — not extracted.**
16. ~~**Document the `GOEXPERIMENT=jsonv2` requirement** more prominently in the flake itself (a comment near the devShell).~~ **Won't implement — obsolete — the GOEXPERIMENT requirement was removed in v0.5.0.**
17. ~~**Review the `cmd/namer` tool** — was it affected by `530702f`? Does it still build?~~ done (builds and tests at 93.2% coverage)
18. ~~**Check if the website's `flake.lock`** is also stale (the root one was updated by the daemon).~~ done (website flake maintained independently; refreshed with later dep regenerations)
19. ~~**Run `nix flake update`** deliberately (not via daemon) to see if any input bumps cause issues.~~ done (flake.lock updated by the daemon/BuildFlow since)
20. ~~**Add a `just`/`make` compatibility shim** if any downstream tooling expects it (CONTRIBUTING.md still references `just` — stale per AGENTS.md).~~ **Won't implement — never needed; CONTRIBUTING.md fixed instead (v0.3.2).**
21. ~~**Fix or remove `CONTRIBUTING.md`** — it references nonexistent `pkg/errors/`, `go-arch-lint`, and `just`. Either rewrite or delete.~~ done (rebuilt in v0.3.2 (ed5ee4b))
22. ~~**Review the 4 new test files** added in `530702f` (`id_bench_test.go`, `id_brand_test.go`, `id_json_test.go`, `id_alltypes_test.go`) for correctness and coverage gaps.~~ done (superseded — suite now at 431 subtests with contract tests)
23. ~~**Benchmark the `...` pattern fix** — confirm no evaluation performance regression (negligible, but verify).~~ **Won't implement — negligible; not measured.**
24. ~~**Add a `CHANGELOG.md` entry** for the flake fix if it warrants a patch release.~~ **Won't implement — not released; documented in reports and AGENTS.md.**
25. ~~**Review the v0.3.1 release** — AGENTS.md says it "never fired" due to missing `GOEXPERIMENT`. Is the fix actually deployed? Did the tag get re-pushed?~~ done (v0.3.2 fixed the release pipeline; the proxy serves v0.3.1)

### Ecosystem / strategic

26. ~~**Notify downstream repos** (InboxClean, CreditReformBilanzampel, ActaFlow, etc.) of the flake pattern gotcha if they copy this flake structure.~~ done (2026-09-17 fleet pass covered the linter-stack repos)
27. ~~**Create a shared flake template** (or flake-parts module) so all LarsArtmann Go repos use a consistent, tested flake pattern.~~ **Won't implement — not created.**
28. ~~**Consider a `nix-health` check** for the `outputs` pattern across all repos.~~ **Won't implement — not built.**
29. ~~**Review whether `encoding/json/v2` is still experimental** — if it's stabilized in Go 1.26+, the `GOEXPERIMENT` flag may be removable.~~ done (dual-mode support shipped v0.5.0; v1 remains the default)
30. ~~**Document the release process end-to-end** — the v0.3.1 saga suggests gaps.~~ done (AGENTS.md release process section is the checklist)
31. ~~**Set up dependabot / flake-update automation** for flake inputs (if not already).~~ **Won't implement — not configured.**
32. ~~**Review the `git-town.toml`** config — still accurate?~~ done (main = master still correct)
33. ~~**Audit the BuildFlow pre-commit hook** (34 checks) — are any stale or redundant?~~ done (superseded — .buildflow.yml is explicit; AGENTS.md no longer hardcodes a count)
34. ~~**Add a `docs/status/` index** or README listing all status reports chronologically.~~ **Won't implement — not created.**
35. ~~**Review the `domains/` repo** DNS status — AGENTS.md says CNAME is "pending terraform apply."~~ done (domain live (AGENTS.md Website section))
36. ~~**Verify `branded-id.lars.software`** is live (DNS may have propagated since last report).~~ done (live; v0.6.0 deployed and fetched)
37. ~~**Run `nix flake check --all-systems`** to verify cross-platform compatibility (the default skips darwin/aarch64).~~ done (CI runs it (flake-check job))
38. ~~**Consider splitting `flake.nix`** into a `flake-parts` module if it grows further.~~ **Won't implement — not needed at current size.**
39. ~~**Review `treefmt-nix` config** — are all 4 formatters (gofumpt, goimports, golines, nixfmt) still needed and non-conflicting?~~ done (all four formatters current; treefmt-check green)
40. ~~**Add `GOEXPERIMENT=jsonv2` to the `checks.build`** — wait, it's already there. Verify all derivations that run `go` have it.~~ done (checks.test runs both modes)
41. ~~**Document the `cmd/namer` codemod** usage in README or a dedicated doc.~~ done (website guides/namer-tool.mdx documents the tool)
42. ~~**Review whether `flake-parts` `mkFlake`** could catch this pattern error at definition time (feature request upstream?).~~ **Won't implement — not filed.**
43. ~~**Add a `direnv` `.envrc`** if not present, using `nix develop` — improves DX.~~ **Won't implement — not added.**
44. ~~**Review the `devShells.ci`** — is it used by CI? If not, remove (YAGNI).~~ done (devShells.ci exists in flake.nix:77)
45. ~~**Consolidate `GOEXPERIMENT=jsonv2`** — it's repeated in 8+ places. Could it be set once via `mkShellNoCC` default or an env wrapper?~~ **Won't implement — obsolete — requirement removed in v0.5.0.**
46. ~~**Check if `go_1_26`** is the right attribute or if it should be `go` (defaulting to latest) for less churn.~~ done (go_1_26 pinned deliberately (AGENTS.md version-pins gotcha))
47. ~~**Review the `lib.fileset.gitTracked`** usage in `checks.build` — does it handle `vendor/` correctly?~~ done (no vendor/ directory; fileset works as-is)
48. ~~**Add a `flake.nix` smoke test to the website** — `website/flake.nix` should also be checked in CI.~~ **Won't implement — not added.**
49. ~~**Document the relationship between root `flake.nix` and `website/flake.nix`** — two independent flakes in one repo is unusual; explain why.~~ done (AGENTS.md Website section notes the independent website flake)
50. ~~**Celebrate** — the build is unblocked and all 37 checks pass. Then tackle items 1-5.~~ done (done — the build stayed unblocked through v0.6.0)

---

## g) Questions I CANNOT figure out myself

1. **Should I amend the auto-git daemon's commit `4b6b4b5`** to fix its misleading message, or leave history as-is and add a clarifying follow-up commit? (Amending rewrites history; the daemon may re-commit if the tree changes.)

2. **What originated commit `530702f`?** Was it an automated agent, a linter suggestion, or a manual edit? Knowing the source determines whether I need to add guardrails against recurrence. (I can see the diff but not the intent or tool that produced it.)

3. **Is the `gitleaks` skip in BuildFlow intentional?** If secrets scanning was deliberately disabled (e.g., false positives on test fixtures), I should document why. If it's stale config, I should re-enable it. (The skip reason isn't recorded anywhere I can find.)

---

## Honest self-assessment

**What went well:** Correct root-cause analysis on the first try. Minimal, honest fix (didn't re-add `nixpkgs` just to match the old shape — used the idiomatic `...`). Thorough verification across `nix flake check`, `nix build`, `nix fmt`, and full `buildflow`. Didn't touch unrelated code.

**What went poorly:** I violated my own "update memory immediately" rule by not recording the gotcha in AGENTS.md. I let the auto-git daemon write a lying commit message by not committing first. I didn't check `website/flake.nix` during the session (only in this report). I treated the symptom without investigating the intent behind `530702f`.

**Grade:** B-. The fix is an A; the surrounding process hygiene is a C.

---

## Resolution (2026-07-28)

The flake fix is **permanent** — `flake.nix:23` retains the `...` pattern.
BuildFlow passes (37/38, `gitleaks` skipped by config).

**Items from section (f) addressed by later sessions:**

- CONTRIBUTING.md stale references (`just`, `pkg/errors/`) — fixed in v0.3.2
- goimports corruption hazard — documented in AGENTS.md + contract test added
  (see `2026-07-27_11-35` and `2026-07-27_16-44` reports)

~~**Still open:** `nix flake check --no-build` not in CI or pre-commit; `gitleaks`
skip reason undocumented; AGENTS.md lacks the `outputs` pattern gotcha (the
goimports gotcha was documented instead).~~ All three closed since: the
`flake-check` CI job exists (go.yml), `.buildflow.yml` declares its skip list
explicitly, and the `outputs` gotcha is in AGENTS.md. All work culminated in
**v0.5.0**; the flake-guard story completed around v0.6.0.
