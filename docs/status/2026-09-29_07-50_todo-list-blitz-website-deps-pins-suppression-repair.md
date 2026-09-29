# Status Report — 2026-09-29 07:50 — TODO-List Blitz: Website Deps, Go-Pins Tripwire, BD001 Suppression + Repair

> Session scope: work through every actionable item of `TODO_LIST.md` in
> go-branded-id (plus one small go-error-family interlude at the end).
> Repo state at close: working tree clean (auto-commit daemon swept all
> changes), root `go 1.26` + linter `go 1.27.1` floors correct, all test
> suites green, `nix flake check` passed, golangci-lint 0 issues in all three
> phases. All claims below were verified live during the session unless
> marked otherwise.

## a) FULLY DONE

1. **Daemon `linter/go.mod` bump reverted — twice.** The auto-commit daemon
   swept `go 1.27.1` → `go 1.27` in commit `33bc879` before the session
   start, and did it AGAIN mid-session (~07:30, fifth and sixth instances of
   the documented class). Both restored to `1.27.1`, verified by building +
   testing the linter module in the dev shell after each restore.
2. **Website dependency advisories cleared (High-impact TODO #1).**
   `website/pnpm-workspace.yaml` overrides: `fast-uri ^3.1.4 → ^3.1.7`, added
   `svgo ^4.1.0` and `devalue ^5.9.2`. Lockfile resolved to fast-uri 3.1.8 /
   svgo 4.1.0 / devalue 5.9.4. `pnpm audit`: **No known vulnerabilities**
   (was 7 Dependabot alerts: 4× fast-uri high, 2× svgo, 1× devalue).
   Website build green (14 pages). Direct deps refreshed as a side effect
   (astro 7.3.5, starlight 0.42.4, html-validate 11.16.1).
3. **Go-pins consistency tripwire (High-impact TODO #2).** New
   `scripts/check-go-pins.sh` asserts the three-way pin agreement for BOTH
   modules (go.mod directive ≤ flake `go_1_2x` pin on major.minor; every CI
   `go-version` in go.yml/release.yml/validate-docs.yml matches a pin). Wired
   into: new CI `go-pins` job in `go.yml`, and as the first guard in the
   pre-push hook (hook reinstalled at `.git/hooks/pre-push`). **Mutation
   tested**: fails on a bumped root directive and on a foreign CI go-version;
   files restored byte-identical afterwards.
4. **BD001 in-source suppression + BD002 (Medium TODO).**
   `//brandid-lint:ignore(BD001) <reason>` on the declaration line or as the
   last line of the comment group directly above. Reason mandatory, only
   BD001 suppressible. Broken directives (malformed grammar, missing rule,
   unknown rule, missing reason, stale on a named brand, duplicate,
   unplaced/annotating nothing) are each reported as **BD002** — a silent
   no-op directive never stays silent. Suppressed brands produce no finding
   and are excluded from repair. Files: `linter/suppress.go` (parser +
   association + BD002 findings), `detect.go` reworked around a shared
   `fileScan`, 4 new testdata fixtures, `suppress_test.go` (~13 table cases +
   finding-shape test). **Dogfooded against the real false-positive class**:
   a copy of go-cqrs-lite's `stream_id.go` flags `StreamMarker`; adding the
   directive takes the run to exit 0.
5. **brandid-lint Repair capability (Medium TODO).** `linter.Repair(ctx)` /
   `RepairPath(path)` insert the suggested `Name()` stub after the enclosing
   declaration of every unsuppressed, unnamed brand — byte-offset insertions
   spliced ascending (grouped `type (...)` blocks keep stub order = decl
   order), idempotent (second run: zero changes, byte-identical file),
   preserves file permissions, gofmt-stable output for EOF-no-newline /
   blank-line / adjacent-decl / type-block cases (all edge-tested via
   `format.Source`). CLI gained `-fix` (repair → re-detect → exit by what
   remains; dry-run remains default). Provider registers a real toolsdk
   `Repairer` (BuildFlow measures the delta by re-running Detect, by design).
   BD001 findings now carry `FixStrategy: direct` + Before/AfterCode.
   **Design pivot mid-task**: built against the _pinned_ go-finding v1.13.0 —
   the local checkout's `TextEdit`/`WithEdits` API is 25 commits past the
   last tag and unreleased, so findings use Before/After codes and repair
   drives the same shared `fileScan` as detection (alignment proven by test:
   repair → DetectPath = 0 findings).
6. **Website audit beyond /changelog/ (Medium TODO).** `api-reference.mdx`:
   all symbols + 9 sentinel errors verified against `go doc -all`; added the
   missing `Format` verb set (%v/%s/%d/%q/%#v). `guides/namer-tool.mdx` was
   documenting the **deleted** `cmd/namer` tool — fully rewritten as the
   brandid-lint guide (usage, exit codes, `-fix`, SARIF, suppression
   directive, BuildFlow provider); sidebar label updated. Site rebuilds green.
7. **Pre-existing lint-gate breakage fixed on sight.** golangci-lint renamed
   `exhaustruct` → `exhaustruct_v5`; the stale `//nolint:exhaustruct` in
   `id_ptr.go` had silently stopped suppressing (root lint failing on a file
   I never touched). Fixed the nolint + the `.golangci.yml` test-exclusion
   entry + my own three fresh nolints in `linter/suppress.go`.
8. **Documentation brought current.** Root `CHANGELOG.md` Unreleased (+2
   Added, +3 Fixed entries), `linter/CHANGELOG.md` Unreleased (4 entries),
   `AGENTS.md` (7 areas: file tree, pins tripwire, linter Repair+suppression,
   "Brands That Deliberately Skip Name()" now prescribes the directive,
   pre-push hook, exhaustruct_v5 rename gotcha, dev-shell jsonv2 gotcha),
   `TODO_LIST.md` rewritten (all actionable items closed; ecosystem item
   annotated with the suppression follow-up).
9. **Final gates, all green at close:** root tests JSON v1 + v2, linter
   module tests (3 packages), golangci-lint v1+v2+linter = 0 issues each,
   `nix flake check` (root build/test both modes + linter build/test +
   treefmt) = all checks passed, go-pins check consistent.

## b) PARTIALLY DONE

1. **go-error-family docs validation (end-of-session interlude).** The pasted
   `md-go-validator` failure (`http-and-cli.mdx` block #6: top-level
   `mux.Handle` statement → "expected declaration, found mux") was **already
   fixed by a parallel session** — daemon commit `9090f0f` (07:30) wrapped
   both `mux.Handle` calls in real functions and removed two `// skip-validate`
   markers. I verified the fix in the diff and on disk, but did **not** re-run
   the validator to green: `md-go-validator` requires go ≥ 1.27 while the
   ambient shell has go 1.26.7 with `GOTOOLCHAIN=local` — I was starting the
   `nix develop` retry when this report was requested. Remaining: one
   validator run in the right environment.
2. **BuildFlow as a LOCAL gate for this repo.** Outside the dev shell: 8 step
   failures (ambient go 1.26.7 vs linter floor 1.27.1 for golangci-lint/test
   fan-out; missing govalid/go-licenses binaries; tsconfig/pnpm-audit env
   mismatches). Inside the dev shell: 2 failures (root `test-compile` hits
   the `json.Marshal requires go1.27 (module is go1.26)` language gate;
   go-licenses still absent). Loop-detector confirmed these failures are
   identical across pre-session runs — **pre-existing, not caused by this
   session**. I analyzed and documented the dev-shell gotcha in AGENTS.md but
   deliberately did not "fix" it (the canonical gates — flake apps + CI — are
   green and orchestrate Go versions explicitly). Open decision: what the ONE
   canonical local BuildFlow invocation should be.
3. **Linter release v0.2.0.** CHANGELOG Unreleased section is written, but
   the `Version` const is still `0.1.0` and no `linter/v0.2.0` tag exists —
   by policy both happen only at release time, and no release was requested.

## c) NOT STARTED

1. **Ecosystem bump: 14 downstream repos to v0.6.0** — was 🔵 BLOCKED before
   the session; unchanged (separate sessions, go-ecosystem-upgrade skill).
2. **go-cqrs-lite suppression directives** — `StreamMarker`/`TimerMarker`
   should now carry `//brandid-lint:ignore(BD001) <reason>`; noted in
   TODO_LIST, not executed (different repo).
3. **Website deploy** — `nix run .#deploy` (from `website/`) is a
   release-time step; the live site still serves the old content (stale
   namer page included) until the next release deploy.
4. **`website/src/content/docs/changelog.mdx`** — no Unreleased-preview
   section added (deliberately release-time, but factually not done).

## d) TOTALLY FUCKED UP (my mistakes this session — all caught and fixed)

1. **suppress_test.go multiedit misfire**: one edit renamed
   `TestSuppression_BrokenDirectivesStayLoud` into a duplicate
   `TestSuppression_BD002FindingShape` → redeclaration compile failure.
2. **repair_test.go syntax typos shipped to disk** (missing comma
   `writeTestFile(tb dir, …)`; `writeTestFile(tb := t, …)`) — two rounds of
   compile-error fixing for sloppiness that a re-read would have caught.
3. **spliceInsertions assembled the file TAIL-FIRST**: my descending-order
   splicer emitted stub + reversed file content. Real logic bug; my "apply
   back to front" reasoning was wrong for a builder-style assembly. Tests
   caught it; rewritten ascending with a prev cursor.
4. **Four rounds of bad test inputs** (not repair bugs): wrong line-number
   expectation, a fixture missing its `id.ID` usage, non-gofmt-clean test
   sources (double blank line, unaligned type block, unspaced doc-comment
   directive) that made honest repairs look gofmt-unstable. Cost: ~30 min.
5. **Wrote three NEW nolints with the dead `exhaustruct` name** in my own
   suppress.go — the exact rename failure class I later fixed repo-wide.
6. **Designed against an unreleased API first**: built finding/repair on the
   local go-finding checkout's `TextEdit`/`WithEdits` (25 commits past
   v1.13.0), forcing a mid-task redesign. Should have resolved the pinned
   module dir (`go list -m -f '{{.Dir}}'`) before designing.
7. **Edit/daemon mtime races**: several "file modified since read" failures
   (daemon commits + `nix fmt` reformatting) — recoverable but noisy; one
   malformed splice left a dangling `return suppressed, findings` + double
   `}}` that needed two extra fix rounds.
8. **Near-miss (caught in design, never shipped)**: naive `sort -V` pin
   comparison would false-fail `linter 1.27.1 ≤ go_1_27`; replaced with
   major.minor comparison before writing the script.

## e) WHAT WE SHOULD IMPROVE (forgotten / better / noticed)

1. **Forgot: race detector never ran on the linter code.** I ran plain
   `go test -count=1` locally; CI's linter-test job has `-race` but nothing
   was pushed this session, so nothing raced. `nix run .#test-race` and
   `.#coverage` were also never invoked.
2. **Forgot: no automated test for `scripts/pre-push-dual-test.sh` or
   `check-go-pins.sh` living in CI beyond the one job** (script itself is
   only proven by the mutation test I ran manually).
3. **Forgot: Dependabot alert closure not verified** — lockfile is patched
   and `pnpm audit` is clean, but I never re-checked the GitHub alerts page
   (they may take a push to re-evaluate).
4. **Better: design against the PINNED dependency API first** (module cache),
   local checkouts second — would have avoided the whole Edits redesign.
5. **Better: write gofmt-canonical test inputs from the start** (run
   `format.Source` over fixtures before running the suite).
6. **Better: check linter-name currency before writing nolint comments**
   (the config's `exhaustruct_v5` was visible in the enable list).
7. **Better: re-verify `go.mod` files right before declaring final gates** —
   the daemon's second bump sat ~an hour because I only noticed it via a
   BuildFlow error. The pins check exists now; use it between phases.
8. **Noticed: the LSP/golangci_lint_ls diagnostics were stale all session**
   (kept reporting typecheck errors on files that compile and test green) —
   unreliable mid-edit signal; wasted attention second-guessing green code.
9. **Noticed: BuildFlow's local-run value for this two-module repo is low**
   — env-sensitive (inside vs outside dev shell differ; 9 tools missing),
   while the flake apps + CI are the authoritative gates. Pick ONE canonical
   local invocation and document it, or accept advisory status.
10. **Improve: pin `golangci-lint` version in CI** instead of `version:
    latest` — the exhaustruct rename class will recur on every upstream
    release.
11. **Improve: `md-go-validator` needs a sane distribution** (flake/devShell
    presence + documented go floor) so docs validation doesn't depend on
    ambient toolchain luck (go-error-family hit exactly this).

## f) Up to 50 things to do next

**Releases & deployment**

1. Cut linter v0.2.0: bump `Version` const, date the `linter/CHANGELOG.md` section, tag `linter/v0.2.0` (signed, annotated), push tag.
2. Cut next root release: date the CHANGELOG Unreleased section, tag `vX.Y.Z`, push tag (CI builds the GitHub Release from the section).
3. Add the new version section to `website/src/content/docs/changelog.mdx` and run `nix run .#deploy` from `website/` (release-time step per AGENTS.md).
4. Verify the 7 Dependabot alerts actually closed after the patched lockfile lands on GitHub.

**Ecosystem (blocked item + follow-ups)**
5. Bump the 14 downstream repos to v0.6.0 (go-ecosystem-upgrade skill; the tracked 🔵 item).
6. go-cqrs-lite: add `//brandid-lint:ignore(BD001) <reason>` above `StreamMarker`/`TimerMarker`; verify with the brandid-lint CLI.
7. After bumps: run brandid-lint over all downstream repos to hunt remaining unnamed brands.
8. Bump the tooling consumers (go-output, cmdguard, go-finding CLI) from v0.5.1.

**Linter hardening**
9. Migrate findings to `WithEdits`/`TextEdit` once go-finding ships the API in a release (currently 25 commits past v1.13.0).
10. Fuzz the suppression-directive parser (grammar fuzz, mirroring the root module's JSON fuzz style).
11. Property test: repair output is gofmt-idempotent across generated file shapes.
12. Type-block suppression fixture (directive inside/above a `type (…)` group) — noticed gap.
13. Baseline/ratchet mode (`--save-baseline`) for incremental downstream adoption.
14. golangci-lint module-plugin distribution (`.custom-gcl.yml`) for IDE integration.
15. SARIF e2e test through the provider path.
16. Benchmarks for Detect/Repair on large trees (`b.Loop`).
17. CLI `-version` flag (version currently only visible in SARIF ToolInfo).
18. Multi-rule directive grammar (`ignore(BD001,BD002)`) — only when a second suppressible rule exists.
19. Consider `//brandid-lint:ignore` block form (start/stop) if downstream wants range suppression.

**CI / gates / scripts**
20. Pin `golangci-lint` version in CI (drop `version: latest`) to stop rename-class breakage.
21. Add a website build job (or extend validate-docs.yml) so content regressions fail CI.
22. Extend pre-push hook to also test the linter module (currently root dual-mode only).
23. Teach `check-go-pins.sh` about `toolchain` directives if they ever appear.
24. Add bats/shunit2-style tests for `scripts/*.sh`.
25. Decide + document ONE canonical local BuildFlow invocation (e.g. `nix develop -c buildflow`) in AGENTS.md.
26. Wire govalid/go-licenses into the dev shell (or drop those steps from local expectations).

**Docs & website**
27. Website `named-brands.mdx`: add the suppression directive section + cross-link to the brandid-lint guide.
28. Website: Unreleased-preview block on `changelog.mdx` (or automate from CHANGELOG.md at build time).
29. README: feature list — mention brandid-lint, suppression, and `-fix` (freshness check).
30. Check `CONTRIBUTING.md`/`SECURITY.md` freshness (validate-docs.yml scope).
31. Consider a short `linter/README.md` pointing at the website guide.
32. Add `ExampleDetect`/`ExampleRepair` doc examples in the linter package.
33. Post-deploy: confirm pagefind search indexes the rewritten brandid-lint page.

**Upstream (other repos, noticed this session)**
34. go-finding: release v1.13.x/v1.14.0 containing the Edits API (unblocks item 9).
35. go-finding: shared in-source suppression-directive filter (so cqrs-linter and brandid-lint stop each shipping a parser).
36. md-go-validator: distribute via flake/devShell + document its go floor (unblocked the go-error-family verification).
37. go-error-family: re-run md-go-validator over `website/` in the right env to confirm commit `9090f0f` is green (my unfinished verification).
38. BuildFlow upstream: make `go-mod-update` respect flake `go_1_2x` pins fleet-wide (kills the recurring daemon bump class at the source).

**Housekeeping**
39. Harvest this report into TODO_LIST/ROADMAP via docs-health HARVEST (top items only; the rest is roadmap fuel).
40. Run `nix run .#test-race` and `.#coverage` once on the final state (the missed verification).
41. Review linter coverage % and set a floor if it matters.
42. Daemon mitigation decision: pre-commit guard vs periodic pins-check vs upstream fix (question g.2).
43. gopls/AGENTS note for stable LSP across the jsonv2 build tags (diagnostics were stale all session).
44. Check `.github/dependabot.yml` covers the `linter/` module and `website/` (npm) ecosystems.
45. Add `Format` examples to the formatting guide if one exists (verb set now documented in API ref only).
46. Consider exporting `ErrOffsetOutOfBounds`-style internals? (Currently private — verify nothing leaks via API surface.)
47. Website: stale `minimumReleaseAgeExclude: astro@7.3.3` entry (astro now 7.3.5) — prune or refresh.
48. Root: when the `go 1.27` floor eventually moves, use the new AGENTS checklist (go.mod + flake + CI + website in one commit).
49. Skim `docs/status/` for older reports whose "next steps" this session completed (ANNOTATE via docs-health).
50. Triage this list — most of 40–50 is ROADMAP fuel, not TODO_LIST material.

## g) Questions I can NOT figure out myself

1. **Release timing**: cut `linter/v0.2.0` and a root release now (website
   deploy included), or hold until the 14-repo v0.6.0 ecosystem session so
   downstreams jump straight to the newest versions in one hop?
2. **Daemon mitigation**: `go-mod-update` is skipped in `.buildflow.yml` yet
   the daemon still re-bumped `linter/go.mod` twice _during_ this session. Do
   you want a local guard (pre-commit hook running `check-go-pins.sh`, or a
   periodic auto-revert loop), or is CI + pre-push enough and the root cause
   should be chased in the daemon/BuildFlow itself?
3. **BuildFlow locally**: fix the environment so `buildflow` is green for
   this two-module repo (dev-shell wiring for govalid/go-licenses, one
   canonical invocation), or accept flake apps + CI as authoritative and
   treat local buildflow runs as advisory-only?
