# Status Report: Deduplication Review & Critical Import Bug Fix

**Date:** 2026-07-27 11:35
**Session scope:** `art-dupl` deduplication review → bug discovery → fix → partial cleanup
**Working tree:** Clean (auto-commit daemon committed all changes)
**Branch:** `master` (16 commits ahead of `origin/master`)

---

## TL;DR

Ran `art-dupl --type-aware -t 1` to review code duplication. Found 3 clone groups. While analyzing them, discovered a **CRITICAL bug**: `id_json_v1.go` and `json_helpers_v1_test.go` were importing `encoding/json/v2` instead of `encoding/json` — meaning **the library could not build or test in v1 mode at all**. This was a goimports-corruption bug from the previous session that was supposedly fixed but wasn't. Fixed it. Removed one harmful duplicate function. Accepted two intentional clones with documented rationale.

---

## a) FULLY DONE ✓

### 1. Fixed critical import corruption bug (CRITICAL P0)
- **What:** `id_json_v1.go` (line 6) and `json_helpers_v1_test.go` (line 6) both had `import "encoding/json/v2"` instead of `import "encoding/json"`. The `//go:build !goexperiment.jsonv2` tag means these files compile in v1 mode, but the v2 import doesn't exist without `GOEXPERIMENT=jsonv2`.
- **Impact:** `go build ./...` and `go test ./...` (without `GOEXPERIMENT`) completely failed. The entire v1 code path — the default mode — was broken. Any downstream consumer not setting `GOEXPERIMENT=jsonv2` could not use the library.
- **Root cause:** The `goimports` formatter (run via `nix fmt`) sees `json.Marshal`/`json.Unmarshal` calls and "helpfully" rewrites the import to `encoding/json/v2`. This happened in the previous session and was documented as a known hazard, but the fix was either reverted by a subsequent `nix fmt` run or was never applied to all affected files.
- **Fix:** Manually corrected both imports back to `"encoding/json"`.
- **Verification:** Both `go test ./... -count=1` (v1) and `GOEXPERIMENT=jsonv2 go test ./... -count=1` (v2) pass.

### 2. Eliminated harmful duplicate function in cmd/namer
- **What:** `typeNameFromExpr` (main.go:299-308) and `receiverTypeName` (main.go:310-320) were 100% identical functions — same signature, same body, same logic. A pure semantic clone with zero differences.
- **Fix:** Deleted `receiverTypeName`, updated call site at main.go:384 to use `typeNameFromExpr`, removed redundant test subtest (`t.Run(tt.name+"/receiverTypeName", ...)`).
- **Verification:** `go test ./cmd/namer/ -cover` confirms 80.1% coverage maintained.

### 3. Accepted 2 intentional clone groups (documented)
Created `dedup-acceptance.md` with rationale:
- **`id_json_v1.go` ↔ `id_json_v2.go`** — Build-tag architecture requires identical files with different imports. Cannot merge.
- **`id_binary.go:135-141` ↔ `id_text.go:29-35`** — 4-line idiomatic empty-input guard. Extraction would be net-worse.

### 4. Updated AGENTS.md with goimports corruption warning
Added a prominent "CRITICAL: goimports corrupts v1 files" paragraph in the Dual JSON section, documenting the exact failure mode and the mandatory post-format verification step (`go build ./...` without GOEXPERIMENT).

### 5. Verified final state
- `art-dupl --type-aware -t 1` → down from 3 clone groups to 2 (both accepted)
- v1 tests pass, v2 tests pass, `go vet` clean, namer coverage 80.1%

---

## b) PARTIALLY DONE ⚠️

### Deduplication review itself
- **Done:** Analyzed all 3 clone groups, made extract/accept decisions, eliminated the harmful one.
- **Not done:** Did not run `art-dupl` at threshold 5 (the skill's default) to check for smaller duplications. The user explicitly asked for `-t 1`, so this is acceptable, but there may be smaller clones lurking.

### AGENTS.md "Stale Files to Ignore" section
- **Done:** Added the goimports warning to the Dual JSON section.
- **Not done:** The "Stale Files to Ignore" section at the bottom of AGENTS.md still says `CONTRIBUTING.md` is stale, but CONTRIBUTING.md was fixed in the previous session. This section is itself stale. Not touched this session.

---

## c) NOT STARTED ✗

1. **v1/v2 equivalence test** — No test verifies that marshaling/unmarshaling produces byte-identical output in both modes. The two build-tagged files could silently diverge.
2. **Meta-test for import correctness** — No test verifies that `id_json_v1.go` actually imports `encoding/json` (not v2). This would have caught the corruption automatically.
3. **Lint cleanup** — 79 lint issues exist (50 varnamelen, 16 err113, 8 makezero, etc.). All pre-existing, none introduced this session. Not addressed.
4. **Website grammar fix** — `changelog.mdx` still has "dual-supports" (should be "dual-support"). Not touched.
5. **Previous status report update** — `docs/status/2026-07-27_11-27_dual-json-support-and-namer-tests.md` claims everything works but doesn't mention the import corruption that was present at that time.
6. **CI hardening** — CI workflows run both modes but don't lint in v2 mode specifically.
7. **Downstream repo bumps** — 14 repos need version bumps. Waiting on version number decision.
8. **Release** — No tag created. Waiting on version number decision.

---

## d) TOTALLY FUCKED UP 💥

### The goimports corruption survived an entire session
This is the headline failure. The previous session's handoff document explicitly says:

> "After running `nix fmt` on build-tagged files, the `goimports` formatter corrupted `import "encoding/json"` to `import "encoding/json/v2"` in v1 files TWICE. This was caught only by running `go build ./...` after formatting."

The previous session KNEW about this, documented it, and claimed to have fixed it. **But the files were still broken when this session started.** Either:
1. The fix was never actually applied to all affected files, OR
2. The auto-commit daemon ran `nix fmt` again after the fix, re-corrupting the files, OR
3. The fix was applied but then reverted by another operation.

**The lesson:** Documenting a hazard is not enough. There needs to be an automated guard (test, lint rule, pre-commit check) that PREVENTS the corruption, not just a warning that says "remember to check after formatting."

### I didn't catch it until art-dupl surfaced it
I read both files at the start of this session and my eye scanned past the import line. The `art-dupl` clone report showing `id_json_v1.go:44-56 | return nil` is what made me look closer. I should have verified `go build ./...` (v1 mode) as the very first step of this session, before anything else.

### Auto-commit daemon committed with garbage messages
Recent commits include:
- `355439d for v1 ID format` — meaningless
- `b97157c ): improve JSON...` — malformed (from previous session)

The auto-commit daemon continues to create noise in git history.

---

## e) WHAT WE SHOULD IMPROVE 🔧

### Process improvements
1. **Always verify both build modes first.** Before ANY work on this repo, run both `go build ./...` and `GOEXPERIMENT=jsonv2 go build ./...`. This is the smoke test.
2. **Never trust `nix fmt` on build-tagged files.** The goimports component is fundamentally incompatible with the dual-import pattern. Consider excluding v1 files from goimports or adding a post-format repair step.
3. **Add automated guards, not just documentation.** A test that imports the v1 file and asserts the import path would have caught this instantly. Documentation rots; tests don't.
4. **The auto-commit daemon is actively harmful.** It commits broken code with garbage messages. It should either be disabled during active sessions or configured to run tests before committing.

### Code improvements
5. **The v1/v2 file pair is a maintenance liability.** Two files that must stay identical but can't be merged. Consider code generation (one source file, two outputs) or a different approach entirely.
6. **The 79 lint issues are debt.** Most are varnamelen (parameter names too short) and err113 (errors defined as package-level vars). These are style choices, not bugs, but they make the lint output noisy and real issues harder to spot.

### Documentation improvements
7. **AGENTS.md is getting long.** The goimports warning is important but adds to an already dense file. Consider a dedicated `docs/gotchas.md` for hazard documentation.
8. **The previous status report is now misleading.** It claims success but the code was broken. Status reports should be updated when their claims are invalidated.

---

## f) THINGS TO GET DONE NEXT (prioritized)

### P0 — Critical (blocks release)
1. **Add meta-test for v1 import correctness** — Test that verifies `id_json_v1.go` source contains `encoding/json` not `encoding/json/v2`.
2. **Add v1/v2 equivalence test** — Marshal/unmarshal in both modes, assert byte-identical output.
3. **Decide version number** — v0.5.0? v1.0.0? Needed for release and downstream bumps.
4. **Create signed annotated tag** once version is decided.
5. **Verify CI passes** on both modes before tagging (check GitHub Actions runs).

### P1 — High value
6. **Bump 14 downstream ecosystem repos** to new version.
7. **Add `golangci-lint` with `GOEXPERIMENT=jsonv2`** to CI to lint v2 code paths.
8. **Fix website `changelog.mdx` grammar** — "dual-supports" → "dual-support".
9. **Update previous status report** (2026-07-27_11-27) to note the import bug was found and fixed post-report.
10. **Update AGENTS.md "Stale Files" section** — CONTRIBUTING.md is no longer stale.
11. **Disable or fix the auto-commit daemon** — it commits broken code and creates garbage commit messages.
12. **Consider code generation** for v1/v2 file pair to eliminate manual sync risk.
13. **Add pre-commit guard** that runs `go build ./...` (v1 mode) and fails if it doesn't compile.

### P2 — Quality improvements
14. **Run `art-dupl` at threshold 5** to find smaller duplications.
15. **Address varnamelen lint issues** (50 instances) — rename short parameters.
16. **Address err113 lint issues** (16 instances) — consider if errors should be sentinel values.
17. **Address makezero lint issues** (8 instances) — use `make` with initial capacity.
18. **Fix `testableexamples` lint issue** (1 instance).
19. **Fix `tparallel` lint issues** (2 instances).
20. **Fix `goconst` lint issues** (2 instances).
21. **Add benchmark comparing v1 vs v2 JSON performance.**
22. **Review `cmd/namer` for additional test coverage** (currently 80.1%, target 90%+).
23. **Verify website builds** — run `nix run .#build` from `website/`.
24. **Consider adding `goimports` exclusion** for v1 files in formatter config (if possible).
25. **Add `.editorconfig` or formatter config** that excludes build-tagged files from import rewriting.
26. **Review all `//nolint` directives** — ensure each has a current, accurate justification.
27. **Clean up git history** — the 16 unpushed commits include garbage messages from auto-commit daemon; consider squashing before push.

### P3 — Nice to have
28. **Add CONTRIBUTING.md rewrite** — the file was partially fixed but may still have stale references.
29. **Review `flake.nix`** for the dual-mode test app — ensure it's clean and well-documented.
30. **Add architecture decision record (ADR)** for the dual JSON v1/v2 build-tag pattern.
31. **Add ADR for the phantom-types branding pattern.**
32. **Consider adding `omitempty`-style option** for JSON serialization (currently zero → null always).
33. **Review `id_sql.go` Scan method** for additional SQL driver type coverage.
34. **Add fuzz tests** for Text and SQL round-trips (currently only JSON and Binary).
35. **Document the binary serialization endianness** in the public API docs (currently only in AGENTS.md).
36. **Consider adding `fmt.Scanner` / `fmt.ScanState` support** for interactive input.
37. **Review `Compare()` for uint types** — currently a runtime check, could be compile-time with type constraints.
38. **Add `crypto.Hash` support** for IDs that are hash-derived.
39. **Consider `ID[B, V].Ptr()` documentation** — clarify when to use pointer vs value IDs.
40. **Review `Reset()` method** — is it idiomatic? Should it be `Clear()` or `SetZero()`?
41. **Add integration test** with a real database driver (sqlite) for SQL round-trip.
42. **Add integration test** with a real HTTP JSON API for JSON round-trip.
43. **Consider adding `context.Context` support** for any cancellation-aware operations.
44. **Review error message format** — all errors start with `"id: "` prefix; document this convention.
45. **Add `errors.Is` / `errors.As` support** for typed error handling (`ErrNotOrdered`, etc.).
46. **Consider adding `ID[B, V].Validate()` shorthand** that calls `ValidateID`.
47. **Review `BrandNamer` interface** — should it be `BrandNamer[B any]` with a type parameter?
48. **Add example with UUID value type** in docs.
49. **Add example with ULID value type** in docs.
50. **Review module path** — `github.com/larsartmann/go-branded-id` vs potential rename.

---

## g) QUESTIONS (cannot determine myself) ❓

### 1. Version number for the release?
The dual JSON v1/v2 support is a significant feature (the library now works without `GOEXPERIMENT` for the first time). Tags so far: v0.1.0, v0.3.0, v0.3.1, v0.3.2, v0.3.3, v0.4.0. Options:
- **v0.5.0** — minor bump, reflects new dual-mode capability
- **v1.0.0** — major milestone, signals API stability (the core `ID[B,V]` type hasn't changed)
- **v0.4.1** — patch bump, if you consider dual-mode a fix rather than a feature

I cannot decide this — it's a product/positioning decision.

### 2. Should the auto-commit daemon be disabled during active sessions?
It committed broken code (the goimports corruption) and creates garbage commit messages. But it may be useful for crash recovery. This is a workflow preference I can't infer. Options:
- Disable entirely
- Disable during active Crush sessions only
- Configure it to run `go build ./...` before committing
- Leave as-is and just deal with the noise

### 3. Should I squash the 16 unpushed commits before pushing?
The commit history includes garbage messages from the auto-commit daemon (`for v1 ID format`, `): improve JSON...`). Squashing would clean the history but lose granular tracking. Alternatively, an interactive rebase to fix just the bad messages. This is a git-hygiene preference.
