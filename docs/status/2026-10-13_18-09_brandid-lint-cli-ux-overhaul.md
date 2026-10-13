# Status Report — brandid-lint CLI UX Overhaul (session snapshot)

**Date:** 2026-10-13 18:09 CEST
**Session trigger:** "What is the cli name of our linter?" → config verification → three rounds of pasted CLI failures from go-credentials-store → full CLI UX overhaul, implemented and verified in-session.
**Scope:** This session only. Q&A + verification work + the CLI fix. No other research performed.

---

## Self-Review (asked directly: what did you forget / do better / still improve?)

**What did you forget:**

- **The fix is not in the user's hands yet.** The PATH binary (SystemNix, master-pinned lock) is still the old 0.5.1 until: push → SystemNix lock bump → rebuild. I mentioned this once in passing instead of leading with it. The very next `brandid-lint ./...` from a normal shell will STILL fail.
- The FIRST paste already showed the triple-wrapped `stat` error (`stat ./...: stat ./...:`). I diagnosed it internally but only fixed it two turns later, bundled with the bigger fix.
- No TDD: tests were written after implementation; the user's exact failing commands were never reproduced as failing tests first (they were verified manually after the fact).
- Did not run `nix flake check` (full gate) — only `nix run .#test` and `nix run .#lint`.
- Did not check whether root README or the website document the CLI flags I changed (staleness risk left open).
- `doc-files-age-check` exposure: code changed, but README.md and TODO_LIST.md were not touched (BuildFlow pre-commit requires contact within 3 weeks).
- Stale LSP diagnostics ("args.go: valueFlags is unused") kept repeating in tool output long after the symbol was deleted — noise I never explained to the user.

**What could I have done better:**

- **Ship after paste 1.** The full scope was visible in the first failure: usage text showed every flag, the error showed the wrap bug. Instead the fix landed only after the third frustration message. Two turns announced fixes without shipping (interrupted, but I also kept no todo list to resume from).
- State the deployment chain (push → lock bump → rebuild) proactively, before the user hits the stale binary again.

**What can I still improve (process level):**

- Golden-output integration tests: run the BUILT binary against fixture trees and assert complete output, instead of only in-process `run()` tests — that layer catches flag-package behavior end-to-end.
- Real-repo regression fixtures: the go-credentials-store BD003 paste is perfect testdata material for suggestion quality.
- Close the report loop: section (f) below must be harvested into TODO_LIST.md, not entombed here.

**Did I lie?** No. Every claim above is backed by a command run this session (test/lint/build outputs quoted in the session log).

**Ghost systems / split brains:** none created — args.go is fully wired into main.go; no concept duplicated. One stale-doc risk flagged: `provider.go`'s tool-option description still says "BD001, BD003" and does not mention the new `all` alias or BD002 no-op (task below).

**Tests:** linter module 3/3 packages green; root green in both JSON modes; golangci-lint 0 issues both modules; real-repo verification of all four previously failing command shapes passed.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                                 | Evidence                                                                                         |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------ |
| 1  | CLI name question answered (`brandid-lint`)                                                                                                                                                                                          | session Q&A                                                                                      |
| 2  | Config verification BOTH sides: local flake exposes package+app+default (evaluates to `brandid-lint-0.5.1`); SystemNix input master-tracked, lock entry present, follows wired, `nix build SystemNix#brandid-lint` succeeds          | flake.nix:57-168; SystemNix flake.nix:504, lib/lars-packages.nix:36; store path built in-session |
| 3  | cmdguard non-adoption explained with evidence (ecosystem layering: cmdguard is a downstream consumer pinning v0.5.1; vendor weight: charmbracelet/cobra stack vs. 3-module vendor tree; ROADMAP Theme 3 reserves cmdguard for namer) | README/go.mod/vendor read; ROADMAP.md:29                                                         |
| 4  | Rules-default question answered (BD001 default, BD002 automatic companion, BD003 opt-in) incl. fleet-flood rationale                                                                                                                 | linter/rules.go:32-35                                                                            |
| 5  | **Flags may appear anywhere** — arg permutation before flag parsing; `--` terminator and `-` as path preserved; duplicate paths collapsed                                                                                            | cmd/brandid-lint/args.go (`splitArgs`, `takesValue`); TestSplitArgs                              |
| 6  | **Go-style path patterns** — `./...`, `pkg/...`, `/abs/...`, bare `...` expand to walk roots                                                                                                                                         | args.go (`expandPattern`, `resolvePaths`); TestResolvePaths                                      |
| 7  | **No path = working directory** (like golangci-lint) instead of usage+exit 2                                                                                                                                                         | main.go; TestRun_NoPathsDefaultsToWorkingDir, TestRun_FixWithoutPathsDefaultsToWorkingDir        |
| 8  | **`-version` flag** — prints `brandid-lint 0.5.1`, exit 0                                                                                                                                                                            | main.go; TestRun_Version                                                                         |
| 9  | **BD002 accepted as no-op + `all` alias** — `-rules BD001,BD002,BD003` and `-rules all` now work; `errBD002NotSelectable` removed; messages/docs updated                                                                             | rules.go; TestRun_BD002AcceptedAsNoop, TestRun_AllRulesAlias                                     |
| 10 | **Errors name the path exactly once** — stat/walk wraps de-duplicated; phase labels kept (`error: scan: stat ./nope: ...`)                                                                                                           | detect.go, main.go; TestRun_MissingPathExitTwo asserts single `stat`                             |
| 11 | Usage text + package doc.go updated (patterns, default path, flag placement)                                                                                                                                                         | main.go printUsage, doc.go                                                                       |
| 12 | CHANGELOG: full Unreleased section (Added/Changed/Fixed) in house style                                                                                                                                                              | linter/CHANGELOG.md                                                                              |
| 13 | AGENTS.md updated: code tree (args.go line, CLI comment) + brandid-lint Behavior bullets                                                                                                                                             | AGENTS.md                                                                                        |
| 14 | All gates green: `go test ./...` (linter 3 pkgs), `nix run .#test` (root v1+v2 + linter), `nix run .#lint` (0 issues both modules)                                                                                                   | session log                                                                                      |
| 15 | Real-repo verification in go-credentials-store: `-version`, `-fix ./...` (exit 0), `. -rules BD001,BD002,BD003 -fix .` (works, BD003 findings correct), `-rules all` (exit 1 as expected)                                            | /tmp/brandid-lint build, session log                                                             |
| 16 | Root module untouched (zero-dependency library not affected); user's working tree in go-credentials-store not modified by verification (dirty `*_templ.go` files predate session)                                                    | git status check                                                                                 |

## b) PARTIALLY DONE

| # | Item                                | What remains                                                                                                                                                                                                                 |
| - | ----------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | BD003 output-quality problem        | Diagnosed (suggestion zoo: one-off types per callsite — `CredentialCreatedID`, `RotateFormURLID`, `RawID`, per-handler `URLID` variants where one `CredentialID`/`URLID` would do) and flagged to user; redesign NOT started |
| 2 | Release readiness                   | CHANGELOG Unreleased cut, but `linter.Version` still 0.5.1 (deliberate: bump at release), no `linter/v0.6.0` tag, nothing pushed                                                                                             |
| 3 | SystemNix propagation               | Not started — blocked on push + lock bump; PATH binary stale (old behavior)                                                                                                                                                  |
| 4 | provider.go tool-option description | Still reads "BD001, BD003" — does not mention `all` or the BD002 no-op                                                                                                                                                       |

## c) NOT STARTED

1. BD003 suggestion aggregation/consolidation redesign
2. Release cut: version bump, CHANGELOG section finalize, signed annotated tag, push
3. SystemNix lock bump + toplevel rebuild + CV change-gate revalidation with new binary
4. `nix flake check` full pass (only #test/#lint ran)
5. Root README / website linter-page docs sync for new CLI UX
6. `doc-files-age-check` contact (README.md / TODO_LIST.md untouched during code change)
7. TODO_LIST.md HARVEST from this report's section (f)
8. Golden-output CLI integration tests (exec the built binary)
9. go-credentials-store fixture adoption for BD003 suggestion regressions

## d) TOTALLY FUCKED UP

**Nothing.** No broken builds, no data loss, no reverted user work, all gates green. The honest entry here is the process failure, not an artifact: the fix shipped two turns later than it should have (see Self-Review), which cost the user real frustration ("I am going crazy").

## e) WHAT WE SHOULD IMPROVE

1. **Latency to fix > correctness of fix.** Three user rage-messages before code landed. Ship the visible scope immediately; questions can be answered after.
2. **Todo-list discipline across interruptions** — keep an explicit list so an interrupted fix resumes itself.
3. **Lead with deployment reality**: a fix that lives only in the working tree is not a fix for the user's shell. Always state the push/lock/rebuild chain up front.
4. **TDD for CLI paper cuts** — turn each pasted failure into a failing test first.
5. **Golden integration tests** over in-process-only tests.
6. **Provider/docs parity** — behavior changes must reach provider descriptions, README, and website in the same session, not "later".
7. **Clear stale LSP noise** instead of letting it repeat across tool results.

## f) Next tasks (ranked by impact)

1. Owner decision + execution: push master, cut `linter/v0.6.0` (bump `Version`, finalize CHANGELOG section, signed annotated tag, push tag)
2. SystemNix: bump `brandid-lint` lock input, rebuild, verify PATH binary `brandid-lint -version` reports 0.6.0
3. Re-run go-credentials-store scans with the deployed binary (confirm the exact session failures are gone in the user's shell)
4. BD003 suggestion consolidation design: one ID type per domain entity (e.g. single `CredentialID` across all Credential* events)
5. BD003: `ID` fields should derive suggestions from the containing struct, not invent per-event types
6. BD003: cluster handler params named `id` per URL resource into one suggested type
7. BD003 output ergonomics: summary line first ("28 findings across 9 files"), then details
8. provider.go: option description updated for `all` + BD002 no-op
9. Root README: document CLI UX (patterns, `all`, `-version`, default path, flags anywhere)
10. Website linter page: same sync
11. `nix flake check` full pass
12. TODO_LIST.md HARVEST from this report
13. Touch README.md/TODO_LIST.md for doc-files-age-check window
14. Golden-output integration tests (exec built binary against testdata trees)
15. Port go-credentials-store snippets (routes.go, events.go) into linter testdata as BD003 regression fixtures
16. `-format json` output for editor/CI integration
17. Usage examples block (`brandid-lint ./...`, `-rules all`, `-fix` dry-run semantics)
18. Investigate go-finding `FormatTextRich` TTY/color behavior; decide NO_COLOR policy
19. Decide severity policy for BD003 flood (all INFO today)
20. Website changelog page entry at release time (manual, per AGENTS.md)
21. Confirm auto-commit daemon committed session changes; review the auto-commit content
22. Pre-push hook dual-JSON test run before any push
23. Keep cmdguard/fang reserved for namer (ROADMAP Theme 3); revisit only on owner signal (question 3)
24. Consider `-quiet` flag for large-repo scans
25. Consider test for `-rules=all` attached form edge (currently covered implicitly by table; make explicit)

## g) Questions I cannot answer myself

1. **Release now?** Shall I push master and cut `linter/v0.6.0` (CHANGELOG section + signed tag) so SystemNix and your PATH pick the fixes up — or do you want more changes (e.g. BD003 consolidation) in the same release? I never push without your go-ahead.
2. **BD003 suggestion policy:** consolidate to one ID type per domain entity (`CredentialID` even across different event structs), or keep today's per-declaration suggestions? This is a domain-taste call, not discoverable from code.
3. **Reopen the cmdguard/fang decision?** Your "proper fang colors" comment cut against my non-adoption rationale (layering + vendor weight). Keep brandid-lint minimal and reserve cmdguard for the future namer CLI — or port brandid-lint onto cmdguard v4 now?

---

_Format note: .md was explicitly demanded by the user (skill default is a styled HTML dashboard — override honored, flagged here). Commit of this report intentionally skipped: the harness forbids commits without explicit request; the auto-commit daemon picks the file up._
