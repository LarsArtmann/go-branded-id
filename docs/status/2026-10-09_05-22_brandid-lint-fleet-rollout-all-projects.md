# Status Report: brandid-lint Fleet Rollout — All of ~/projects

**Date:** 2026-10-09 05:22 CEST
**Session scope:** Run `brandid-lint` against all of `/home/lars/projects`, then "fix all" BD001 findings in live code (user excluded `archived/`, `BuildFlow.vendor.bak/`, `ci-siblings/go-cqrs-lite/`).
**Baseline:** 84 findings fleet-wide → 64 in 10 live projects after exclusions.
**End state:** **0 findings in all 10 live projects.** 24 remaining scan hits are exclusively in excluded dirs, go-branded-id's own documented test-brand case, and linter testdata fixtures.

---

## Resolution Summary

| Resolution                              | Count   | Projects                                                                                |
| --------------------------------------- | ------- | --------------------------------------------------------------------------------------- |
| `Name()` stubs kept                     | 41      | accountability-system (19), GmbH (12), Rolls-Royce (8), gomend (1), SKILLS snippet (1)  |
| Suppressed — `String()` is load-bearing | 22      | CV (4), SwettySwipperWeb (6), standard-bug-tracking-schema (10), template-arch-lint (2) |
| Malformed directives repaired           | 2       | go-output (`D2NodeIDBrand`, `GraphNodeIDBrand`)                                         |
| Real bugs fixed (surfaced by stubs)     | 6 sites | GmbH (1), accountability-system (2 production + 3 test)                                 |

---

## a) FULLY DONE

1. **Full fleet scan** — 53,595 Go files, 324 projects; findings grouped and reported.
2. **Classification** — every one of the 64 live findings triaged into stub vs suppress, with documented evidence per repo (CV ADR `2026-08-16_branded-ids-without-name.md`; sbts package doc + `TestStringReturnsRawValue`; SwettySwipperWeb file-header comment requiring raw snowflakes for URLs; template-arch-lint's repo-wide raw-carrier idiom; go-output map-key regression notes).
3. **41 `Name()` stubs inserted** via `-fix` (byte-level, grouped-decl aware, gofmt-stable — verified per file).
4. **22 suppression directives written** with repo-specific reasons referencing the governing ADR/doc.
5. **2 broken go-output directives repaired** — reasons had wrapped onto a second comment line, so suppression silently failed since 2026-10-05; rewrapped to single-line-last.
6. **GmbH zero-UUID corruption root-caused and fixed** — `auth_handler.go:452` did `uuid.Parse(id.String())`; with named brands this yielded `uuid.Nil` → every second registration hit a PK collision (409). Fixed to `id.Get()`; suite now 3/3 green, full suite green.
7. **accountability-system contract fixes** — `middleware.go:163` context value `.String()` → `.Get()`; DB seed in `auth_service_di_test.go` → `.Get()`; 5 assertions in `ids_test.go` migrated to the raw-value contract. Full suite: 0 failures.
8. **Stub over-reach reverted cleanly** — 12 stubs in standard-bug-tracking-schema (10) and template-arch-lint (2) removed and replaced with suppressions after their own docs/tests proved raw-`String()` is deliberate.
9. **Pre-existing failure attribution via throwaway worktrees** — baselines proved unrelated: sbts (7 failures: cache/github/sync), Rolls-Royce (`visualtest` chromedp API), CV (adoption-policy wording), SwettySwipperWeb api (0 failures at my commit; breakage introduced by later daemon-swept dependency changes), gomend (go.mod `replace ../BuildFlow` desync).
10. **SKILLS snippet re-synced** — `domain-types-01/main.go` was a stale manual extract (old module path, `nanoid.NanoID`, no `Name()`, uncompilable); rewritten faithfully from its source block in `domain-types.md`, which already had `Name()`.
11. **gofmt verified on all touched files**; go-output also got a pre-existing trailing-newline fix.
12. **Final verification scan** — 0 live findings; two regressions found by that scan were both fixed and re-verified (see d).

## b) PARTIALLY DONE

1. **gomend stub** — inserted and gofmt-clean, but the repo cannot build (pre-existing go.mod desync), so no test verification is possible. Accepted risk, documented.
2. **SwettySwipperWeb api module** — directives proven inert (0 failures at my commit), but the module's suite is currently red from a later daemon dependency sweep, so a green-baseline full-suite run is impossible right now.
3. **Fleet `.String()`-as-data audit** — done only where tests forced it (GmbH, accountability, template-arch-lint). No systematic audit of the other stubbed repos; weak tests could hide more `uuid.Parse(id.String())`-class bugs.
4. **Per-repo lint suites** — gofmt checked everywhere; no repo's own golangci-lint/BuildFlow pipeline was run against the new stubs/directives.
5. **Knowledge capture** — new precedent case studies (CV, sbts, SwettySwipperWeb, go-output, template-arch-lint) are in this report but not yet folded into go-branded-id's `AGENTS.md` "Brands That Deliberately Skip Name()" section or the linter docs.
6. **CHANGELOGs** — GmbH and accountability-system received real behavior fixes (`.Get()` migration) with no CHANGELOG entries.
7. **Self-review coverage** — the linter's own test suite was never run this session (no linter source was modified, so risk is nil, but it was never checked either).

## c) NOT STARTED

1. Fixing any of the 5 pre-existing broken things (see f items 3–7).
2. TODO_LIST.md / CHANGELOG.md updates in touched repos (docs-health HARVEST of section f).
3. In-source suppressions for the excluded dirs (go-cqrs-lite markers are the canonical directive use case and are still unfixed in their own repo).
4. Fleet scan automation (CI scheduled job with a checked-in exclusion list).
5. Linter feature work identified this session (see f items 1–2, 9).
6. SKILLS `domain-types-02..08` staleness audit; identifying the process that regenerated snippet 01 over my edit.
7. Attribution investigation of the daemon dependency sweeps that broke gomend and SwettySwipperWeb api (suspect per AGENTS.md: BuildFlow `go-auto-upgrade` directive/dependency bumps; check audit-log).

## d) TOTALLY FUCKED UP

1. **I replicated the exact bug I had repaired 30 minutes earlier.** go-output's failure mode is "directive reason wrapped onto a continuation line → suppression silently fails". My template-arch-lint `UserBrand` directive was — two lines. BD001 kept firing and BD002 flagged the unplaced directive. Caught only by the final fleet re-scan; without it, that repo ships a silent no-op. Root cause: I hand-wrote directives in batches and scanned only at the end instead of after each repo.
2. **I ran `-fix` before the classification work was actually complete.** 12 of 53 stubs went into repos whose own package docs/tests explicitly forbid named brands. My pre-flight grep (`BrandNamer|intentionally do NOT|skip.*Name()`) missed standard-bug-tracking-schema because its doc says "deliberately unnamed" — a vocabulary mismatch, and I never grepped the target repos' tests (`TestStringReturns*`) or package docs before inserting. Test failures were the detector — it worked, but the cost was a revert plus an afternoon of failure archaeology that a 30-second doc read per repo would have avoided.
3. **Suppression is fragile by syntax, and I had no guardrail while hand-writing 22 of them.** The single-line-last-of-group rule is documented but easy to violate (I did). The tool tells you afterward (BD002) but does not help you get it right, and one wrapped directive = silent no-op.
4. **Pre-existing, fleet-blocking, unowned:** gomend and SwettySwipperWeb api are unbuildable/red from daemon-swept dependency changes; CV's wording gate is red. Left alone deliberately (not my changes), but the fleet is not actually "all green" — it is "all green except 5 things nobody owns".

## e) WHAT WE SHOULD IMPROVE

1. **Rollout playbook**: read the target repo's ids package doc + `rg 'TestStringReturns|unnamed|BrandNamer'` BEFORE any `-fix`. 30 seconds per repo would have saved the entire revert cycle.
2. **Scan after every repo**, not once at the end — manual directive edits need immediate tool feedback.
3. **Linter should emit the exact directive text** in the BD001 finding (copy-pasteable, single-line, with the suggested reason). Hand-writing is the failure mode; the tool authoring the text is the fix.
4. **`-fix` should repair BD002 placement** (rewrap wrapped directives automatically) — the same idempotent-repair philosophy as BD001 stubs.
5. **Worktree baselines immediately** when a test fails post-change — the GmbH 3-run flakiness dance was resolved in minutes once baselined; do it first, not third.
6. **Package-level suppression** for repos like sbts where 10 brands share one package-doc reason — 10 near-identical directives is noise that invites malformation.
7. **Capture the stub-vs-suppress decision table durably** (this report + AGENTS.md) so the next fleet scan starts from policy, not archaeology.

## Self-Review (brutal)

- **Did I lie?** No. Every claim above has a command/checkpoint behind it. Two claims started wrong and were corrected in-session: "CV failures might be mine" (baselined: no) and "Swetty failures might be mine" (proven: no — 0 at my commit).
- **What did I forget?** Per-repo lint runs; linter's own test suite; CHANGELOG/AGENTS.md capture; a systematic `.String()` data-use audit per stubbed repo; checking whether CV's source-text scanners could trip on my new comments (they didn't — CV suite ran green apart from the pre-existing wording failure — but by luck, not by design).
- **Ghost systems?** The `SKILLS/references/domain-types-0N/` generated snippets: 8 hand-extracted files with no generator, at least one provably stale, reverted by an unidentified sync process. Value is questionable; either wire them to a real generator or delete the directory and point at `domain-types.md`.
- **Split brains?** Mild one created deliberately: CV's suppression directives duplicate the ADR's decision in a second home (mitigated: each directive cites the ADR). Real one found: template-arch-lint's `values.UserID = ids.UserID` alias chain means the "raw carrier" decision now lives in the ids package doc, the values package, and two directive comments.
- **Scope creep trap?** Held. I did not fix the 5 pre-existing breakages, did not refactor template-arch-lint's wire seams to `.Value()`, did not touch excluded dirs.

## f) Up to 50 things to get done next (impact-sorted; most are ROADMAP fuel, not commitments)

| #  | Task                                                                                                           | Repo                                      | Impact |
| -- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------------- | ------ |
| 1  | Linter: emit copy-pasteable single-line directive text in BD001 findings                                       | go-branded-id/linter                      | High   |
| 2  | Linter: `-fix` repairs BD002 placement (rewrap wrapped directives)                                             | go-branded-id/linter                      | High   |
| 3  | SwettySwipperWeb api: revert or finish the daemon dependency sweep; get suite green                            | SwettySwipperWeb                          | High   |
| 4  | gomend: repair go.mod `replace ../BuildFlow` desync; verify ViolationGroupBrand stub builds                    | gomend                                    | High   |
| 5  | Systematic `\.String\(\)`-as-data audit in every stubbed repo (GmbH bug class)                                 | accountability, GmbH, Rolls-Royce, gomend | High   |
| 6  | Fleet-wide grep for `uuid.Parse(<ID>.String())` and similar parse-display patterns                             | all ~/projects                            | High   |
| 7  | CV: fix adoption-policy wording drift (AGENTS.md vs .goreleaser.yml)                                           | CV                                        | Med    |
| 8  | sbts: triage 7 pre-existing failures (cache/github/tracker/sync)                                               | standard-bug-tracking-schema              | Med    |
| 9  | Linter: package/file-level suppression syntax for uniform-reason packages                                      | go-branded-id/linter                      | Med    |
| 10 | Linter: config/exclusion-file support for fleet scans (no more grep -v pipes)                                  | go-branded-id/linter                      | Med    |
| 11 | Rollout playbook doc (pre-flight classification checklist) into linter docs                                    | go-branded-id/linter                      | Med    |
| 12 | AGENTS.md: add 5 new precedent links to "Brands That Deliberately Skip Name()"                                 | go-branded-id                             | Med    |
| 13 | Add testdata fixture: directive line + continuation comment (my exact BD002 case) if not covered               | go-branded-id/linter                      | Med    |
| 14 | go-cqrs-lite: add the canonical BD001 directives to StreamMarker/TimerMarker                                   | ci-siblings/go-cqrs-lite                  | Med    |
| 15 | RR: fix chromedp API usage in visualtest                                                                       | Rolls-Royce                               | Med    |
| 16 | Run each touched repo's own golangci-lint/BuildFlow against new stubs/directives                               | 10 repos                                  | Med    |
| 17 | GmbH + accountability: CHANGELOG entries for the `.Get()` fixes                                                | GmbH, accountability                      | Med    |
| 18 | Verify BuildFlow provider picks up new directives (run buildflow in go-output)                                 | go-output                                 | Med    |
| 19 | template-arch-lint: decide raw-carrier forever vs migrate seams to `.Value()`; document                        | template-arch-lint                        | Med    |
| 20 | SKILLS: find the sync that reverted snippet 01, or delete generated dir                                        | SKILLS                                    | Med    |
| 21 | SKILLS: audit domain-types-02..08 for the same staleness                                                       | SKILLS                                    | Low    |
| 22 | CI: scheduled fleet-wide brandid-lint scan with checked-in exclusion list                                      | go-branded-id                             | Med    |
| 23 | Release linter/vX.Y.Z if items 1–2 land (CHANGELOG first)                                                      | go-branded-id/linter                      | Med    |
| 24 | Run linter's own test suite + `nix run .#test-race` (not run this session)                                     | go-branded-id                             | Low    |
| 25 | Document "String() vs Get() at wire seams" contract in each consumer repo's AGENTS.md                          | 5 repos                                   | Med    |
| 26 | Investigate daemon sweep attribution (BuildFlow audit-log) for gomend/Swetty go.mod                            | cross                                     | Low    |
| 27 | Decide policy for excluded dirs: delete `BuildFlow.vendor.bak`, suppress or leave archived/                    | cross                                     | Low    |
| 28 | doc.go: add URL/query-param use case (SwettySwipperWeb) to the suppression guidance                            | go-branded-id/linter                      | Low    |
| 29 | Check ops/log-parsing impact of new `"Brand:"` log prefixes in GmbH/accountability                             | GmbH, accountability                      | Low    |
| 30 | Linter: SARIF output covered by CI test? verify                                                                | go-branded-id/linter                      | Low    |
| 31 | Linter: suggestion text for lowercase marker types reads awkward — polish or skip                              | go-branded-id/linter                      | Low    |
| 32 | Confirm `brandid-lint` actually runs in each repo's next `buildflow` run (provider wiring proof)               | 10 repos                                  | Low    |
| 33 | Baseline scan performance (~53k files) noted; decide scan cadence (monthly?)                                   | cross                                     | Low    |
| 34 | Harvest this report's section f into TODO_LIST.md / ROADMAP.md (docs-health HARVEST)                           | go-branded-id                             | Med    |
| 35 | Add `.String()` data-use audit as a documented step in the rollout playbook                                    | go-branded-id/linter                      | Low    |
| 36 | CV: confirm new directive comments never trip CV source-text scanners (add a test?)                            | CV                                        | Low    |
| 37 | go-output: CHANGELOG note for directive repair (comment-only)                                                  | go-output                                 | Low    |
| 38 | sbts: fold "brands must stay unnamed" package-doc reason into directive texts verbatim (currently paraphrased) | standard-bug-tracking-schema              | Low    |
| 39 | template-arch-lint: resolve the three-home split brain for the raw-carrier decision                            | template-arch-lint                        | Low    |
| 40 | Publish fleet-rollout story to branded-id.lars.software changelog page                                         | go-branded-id/website                     | Low    |
| 41 | Add brandid-lint pre-commit hook template for consumer repos                                                   | go-branded-id/linter                      | Low    |
| 42 | Consider `//brandid-lint:ignore` for BD002 — currently only BD001; document why                                | go-branded-id/linter                      | Low    |
| 43 | Benchmark linter on 53k files; record runtime in README                                                        | go-branded-id/linter                      | Low    |
| 44 | Consider detecting "intentionally do NOT implement BrandNamer" comments → suggest suppression automatically    | go-branded-id/linter                      | Low    |
| 45 | Verify go-output downstream consumers (nom etc.) unaffected — comment-only, but confirm CI                     | go-output                                 | Low    |
| 46 | Add fleet scan baseline (0 live findings) somewhere machine-checkable (CI job #22)                             | go-branded-id                             | Low    |
| 47 | Roll the 22 new directives' wording through each repo's own linter (line-length)                               | 5 repos                                   | Low    |
| 48 | Delete stale `/tmp` scan artifacts (`/tmp/brandid-*.txt`) or archive into docs                                 | local                                     | Low    |
| 49 | Update go-branded-id AGENTS.md: linter saw first fleet rollout; record counts + date                           | go-branded-id                             | Low    |
| 50 | Re-run `nix flake check` on go-branded-id (untouched this session; confirm clean)                              | go-branded-id                             | Low    |

## g) Questions I cannot figure out myself

1. **SwettySwipperWeb api and gomend were broken by auto-daemon dependency sweeps, not by this session.** Do you want those sweep commits reverted (restore green) or completed (finish the upgrade properly) — and is that my task or another session's?
2. **Is raw-`String()` a permanent convention** for standard-bug-tracking-schema and template-arch-lint (suppress forever), or a migration state you eventually want fixed (wire seams → `.Get()`, then enable `Name()`)? This decides whether those 12 suppressions are final or scaffolding.
3. **What regenerates `SKILLS/how-to-golang/references/domain-types-0N/`?** I found no generator script, yet my edit to snippet 01 was reverted within the hour. If the sync runs again it will clobber my re-synced version. Should I audit snippets 02–08 and/or delete the generated directory in favor of the source doc?

---

_Point-in-time snapshot. Section f is HARVEST input for TODO_LIST/ROADMAP (docs-health), not yet harvested. Format note: user explicitly requested `.md`; status-report skill's canonical format is HTML — flagged, not propagated as default._
