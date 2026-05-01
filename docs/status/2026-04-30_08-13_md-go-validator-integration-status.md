# Status Report: md-go-validator Integration into go-branded-id

**Generated:** 2026-04-30 08:13  
**Date:** Thursday, April 30, 2026  
**Status:** ✅ IN PROGRESS — Integration partially complete

---

## Executive Summary

Integrated `md-go-validator` into `go-branded-id` repository. All 9 Go code blocks in markdown files now pass validation with valid Go syntax. CI workflow created. **Blocker: md-go-validator module is not published to a Git tag, so CI `go install` will fail until first release.**

---

## Work Items

### 1. Build Fix: md-go-validator duplicate `:=` declarations

- **Status:** ✅ FULLY DONE
- **File:** `md-go-validator/pkg/output/output.go` lines 209, 216
- **Problem:** `err := writeCSVRows(...)` and `err := csvWriter.Error()` redeclared `err` in same scope — invalid Go.
- **Fix:** Renamed to `writeErr` and `flushErr` respectively.
- **Verification:** `GOWORK=off go build -o /tmp/md-go-validator ./cmd/md-go-validator` succeeds.

### 2. go-branded-id README.md Code Block Fixes

- **Status:** ✅ FULLY DONE
- **File:** `go-branded-id/README.md`
- **Changes:**
  - **"Why?" Block 1 (line 12):** Restructured from `{ ... }` pseudo-code to valid `package main` with `func main()` and `var` declarations. Added `// Compiles! Runtime bug.` comment.
  - **"Why?" Block 2 (line 28):** Same restructuring. `GetOrder(userID)` correctly produces type mismatch. Added `// Compile error: type mismatch` comment.
  - **Quick Start Block (line 56):** Added `fmt.Println(orderID.Get())` to use previously-unused `orderID` variable.
  - **ValidateID Block (line 105):** Added `import "fmt"` since `fmt.Errorf` was used without an import.
- **Verification:** All 9 code blocks pass (`md-go-validator -v` — 0 errors, 0 skipped).
- **Note:** Removed `<!-- skip-validate -->` HTML comment directives entirely — all blocks are now valid Go.

### 3. GitHub Actions CI Workflow

- **Status:** ✅ FULLY DONE
- **File:** `.github/workflows/validate-docs.yml` (new)
- **Trigger:** Every push to `master`/`main` and all pull requests.
- **Steps:** checkout → setup-go 1.26 → `go install github.com/larsartmann/md-go-validator@latest` → run validator with JSON output → upload artifact.
- **⚠️ Blocker:** The `go install` step requires the module to be published with a git tag. Currently returns `404 Not Found` because `md-go-validator` has no releases yet.

### 4. go-branded-id Test Files (Unrelated Changes)

- **Status:** ⚠️ PARTIALLY DONE
- **Files:** `id_encoding_test.go`, `id_json_test.go`
- **Problem:** These files have uncommitted changes from a previous session (refactoring to `testIDRoundTrip` shared helper). These were NOT made during this session and need review.
- **Action needed:** Determine if these changes should be committed, reverted, or reviewed separately.

---

## Validation Results

```
$ md-go-validator -v /home/lars/projects/go-branded-id

📁 Processing 6 markdown files with 4 workers
  ✅ Block 1 (line 12): OK    -- "Why?" before (package main, GetUser/GetOrder)
  ✅ Block 1 (line 32): OK    -- MIGRATION.md diff block
  ✅ Block 2 (line 28): OK    -- "Why?" after (package main, branded types)
  ✅ Block 3 (line 56): OK    -- Quick Start (fmt used, orderID used)
  ✅ Block 4 (line 95): OK    -- Named Brand Types
  ✅ Block 5 (line 105): OK   -- ValidateID (import "fmt" added)
  ✅ Block 6 (line 148): OK   -- JSON Example
  ✅ Block 7 (line 161): OK   -- SQL Example
  ✅ Block 8 (line 172): OK   -- Comparison & Sorting

Valid: 9 | Skipped: 0 | Errors: 0
```

---

## Known Issues & Blockers

| #   | Issue                                        | Severity     | Notes                                                                     |
| --- | -------------------------------------------- | ------------ | ------------------------------------------------------------------------- |
| 1   | md-go-validator has no Git tags/releases     | **Critical** | CI `go install @latest` fails with 404. Need first release.               |
| 2   | Unrelated test file changes in go-branded-id | Medium       | `id_encoding_test.go` and `id_json_test.go` have uncommitted refactoring. |
| 3   | md-go-validator not in parent `go.work`      | Low          | Build requires `GOWORK=off`. Not a runtime issue.                         |
| 4   | go-branded-id not in parent `go.work`        | Low          | Currently not in `/home/lars/projects/go.work`.                           |

---

## What We Should Improve

1. **Release md-go-validator** — Tag and release v1.0.0 so CI `go install @latest` works.
2. **Add md-go-validator to parent `go.work`** — Include in workspace so builds work without `GOWORK=off`.
3. **Review/handle unrelated test file changes** — `id_encoding_test.go` and `id_json_test.go` have uncommitted work from a previous session.
4. **Add go-branded-id to CI** — Once md-go-validator is released, CI will work. Consider also adding lint + test checks.
5. **Extend md-go-validator parser** — Add Strategy 6: try `package main + var` wrapper for code that declares variables at package level (e.g., `userID := ...`). Currently Strategy 2 fails on `:=` at package level because it's not valid outside functions.
6. **Tree-sitter validator bug** — The treesitter validator seems to report actual errors instead of a "parser unavailable" message. Investigate why it's not gracefully skipping.
7. **Add `md-go-validator` to go-branded-id as a dev dependency** — Track the version being used.
8. **Coverage** — Verify go-branded-id has adequate test coverage after the test refactoring.
9. **golanci-lint run** — Run the full linter on go-branded-id to ensure no style issues remain.
10. **go mod tidy** — Run in both repos to clean up dependencies.

---

## Top #25 Things to Get Done Next

1. **Release md-go-validator** (git tag v1.0.0) — unblocks CI
2. **Verify CI workflow passes** after first release
3. **Review uncommitted test changes** (`id_encoding_test.go`, `id_json_test.go`)
4. **Add md-go-validator to parent `go.work`** workspace
5. **Run `go mod tidy`** in both repos
6. **Run `golangci-lint run`** on go-branded-id
7. **Run `go test ./...`** on go-branded-id to verify tests still pass
8. **Add `md-go-validator` as a tool directive** in go-branded-id (`//go:build tool` or `tools.go`)
9. **Add CI step for tests** — `go test ./... -race` in validate-docs workflow
10. **Add CI step for lint** — `golangci-lint run` in validate-docs workflow
11. **Improve md-go-validator parser** — Strategy for `package main + var` declarations (`:=`)
12. **Add `--fix` flag to md-go-validator** — Auto-add `<!-- skip-validate -->` or auto-fix common patterns
13. **Add markdownlint to CI** — Validate markdown formatting/style
14. **Add `go.sum` entries** for md-go-validator dependencies in go-branded-id
15. **Write tests for md-go-validator** — If not already present, add comprehensive test coverage
16. **Add `--diff` flag** — Show what changed in markdown files
17. **CI: fail-fast vs collect all** — Decide if CI should fail on first error or collect all results
18. **Add `--fail-on` flag** — Control which severity levels cause CI failure
19. **Improve treesitter error reporting** — Make it clearer when a language parser is unavailable
20. **Add Python/Java support** to md-go-validator
21. **Add `--ignore-paths` flag** — Skip certain directories/files from validation
22. **Add `--max-line-length` flag** — Warn on overly long code blocks
23. **Cache parsed results** — Avoid re-parsing unchanged files in CI
24. **Add `--git-diff` mode** — Only validate code blocks changed in the current commit/PR
25. **Add pre-commit hook** — Integrate md-go-validator as a pre-commit hook for go-branded-id

---

## Top #1 Question I Cannot Figure Out

**Why does `md-go-validator`'s parser.go Strategy 2 (`package main\n\n` + code) fail for Block 1 and Block 2 after my first round of fixes (where I replaced `{ ... }` with `return nil` but kept `:=` at what appeared to be package level)?**

Specifically:

- Strategy 1: `func GetUser...` → "expected 'package', found 'func'" ✅ makes sense
- Strategy 2: `package main\nfunc GetUser...` → `expected declaration, found userID` ❌ at the `userID := "user-123"` line
- But `userID := "user-123"` with `package main` should be a valid package-level short variable declaration in Go 1.26...?

I spent significant time debugging this before discovering the actual issue: after replacing `{ ... }` with `return nil`, the restructured code still had `:=` declarations (like `userID := "user-123"`) appearing AFTER the `func main() { GetOrder(userID) }` block. In Go, short variable declarations (`:=`) are NOT valid at package level — only `var` declarations are. So Strategy 2 was correctly failing because `userID :=` in the package scope is a syntax error.

**The fix** was to restructure both blocks as a complete `package main` file where everything that can't use `:=` at package level is properly declared with `var`, and the executable code goes inside `func main()`.

But I still don't fully understand WHY Strategy 3 (func main wrapper) failed — I never fully isolated whether the failure was from the `:=` issue or from the nested function declaration issue. Further investigation needed.
