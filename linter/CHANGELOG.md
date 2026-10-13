# Changelog

All notable changes to the `brandid-lint` module will be documented in this
file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- **`-version` flag** — prints `brandid-lint <semver>` and exits 0. The
  version was previously only visible inside SARIF reports.
- **Go-style path patterns** — `./...`, `pkg/...`, and `/abs/...` now expand
  to the directory tree the scanner walks. `brandid-lint ./...` used to fail
  with `stat ./...: no such file or directory` because every argument was
  treated as a literal path.
- **Flags may appear anywhere** — `brandid-lint . -rules BD003 -fix` now
  parses. Go's flag package stops flag parsing at the first positional
  argument, so a path before `-rules` turned the flag into a stat target
  (`error: repair -rules: stat -rules: stat -rules: ...`). The CLI now
  permutes arguments into flags and paths before parsing; `--` still ends
  flag parsing and `-` stays a path.
- **`-rules all` alias and BD002 accepted as a no-op** — asking for every
  rule (`-rules BD001,BD002,BD003` or `-rules all`) now just works instead
  of failing with "rule BD002 is not selectable". BD002 needs no selection:
  its directive-integrity findings ride along with every active rule.

### Changed

- **No path means the working directory** — `brandid-lint` and
  `brandid-lint -fix` with no path argument now scan `.` (like
  golangci-lint) instead of exiting 2 with a usage message. Duplicate paths
  are collapsed, so `brandid-lint . .` scans once.

### Fixed

- **Errors name the offending path exactly once** — a missing path now
  reports `error: scan: stat ./nope: no such file or directory` instead of
  the triple-wrapped `stat` chains older versions produced.

## [0.5.1] - 2026-10-11

### Fixed

- **Generated files are skipped** — files carrying the ecosystem-standard
  `Code generated ... DO NOT EDIT.` header (sqlc, templ, protoc-gen-go,
  mockgen, stringer) are now neither linted nor repaired, restoring the
  `excludeGenerated` default of the rule's branching-flow ancestor. The
  v0.4.0 parity corpus contained no generated files, so the gap only surfaced
  on real output: the CV repo's BD003 zero-ratchet gate saw ~90 findings in
  `internal/database/*.sql.go` and `*_templ.go` where the old strong-id rule
  reported zero. The skip lives in `scanFile`, so detection and repair see it
  identically — repair never inserts `Name()` stubs into machine output (the
  next generation run would delete them). Applies to every rule (BD001, BD002,
  BD003); a `Code generated` comment AFTER the package clause does not skip a
  file. Pinned by `TestGeneratedFilesAreSkipped`.

## [0.5.0] - 2026-10-11

### Added

- **`rules` tool option on the BuildFlow provider** — repos can now run
  BD003 through BuildFlow instead of shelling out to the CLI:
  `tool_options: { brandid-lint: { rules: "BD001,BD003" } }`. The value is
  parsed by `ParseRuleSet`, so an invalid entry is a hard Detect error (a
  config typo cannot quietly scan with the wrong rules). Without the
  option the provider runs the default rule set, byte-identical to every
  earlier release. Also adds `linter.DetectWithRules(ctx, rules)` for
  library consumers. Requires go-finding/toolsdk v1.15.0 for the Option
  declaration API (same stack branching-flow dogfoods).

## [0.4.0] - 2026-10-11

### Added

- **Rule BD003: primitive-typed ID detection, ported from
  `branching-flow strong-id`** (finding parity gated by
  `scripts/parity-bd003.sh`). BD003 flags `string`/`int`/`uint`/`bool`
  parameters, named return values, and struct fields whose name marks them
  as an identifier (`userId`, `order_id`, `PID`) and suggests the branded ID
  type to adopt. Every finding carries `suggestedTypeName`,
  `suggestedBrand`, `importPath`, and a complete `typeDefinition` in
  metadata. **BD003 is opt-in** — the default rule set stays BD001+BD002
  (byte-identical behavior for every existing consumer) because
  brandid-lint runs fleet-wide and repos that deliberately silenced the old
  strong-id rule must not be flooded by an upgrade.
- **`-rules` flag** on the CLI: comma-separated rule selection
  (`-rules BD003`, `-rules BD001,BD003`). Unknown rules and BD002 (not
  selectable; it rides along automatically) are hard errors. `-fix` with
  BD001 deselected is rejected — repair applies BD001 stubs only.
- **`RuleSet` API**: `DefaultRuleSet`, `AllRules`, `ParseRuleSet`, and
  `DetectPathWithRules` select rules programmatically. `DetectPath` and the
  BuildFlow provider `Detect` keep the default rule set.
- **BD003 suppression**: `//brandid-lint:ignore(BD003) <reason>` suppresses
  the primitive-ID findings it covers (trailing on the declaration line,
  directly above it, or attached to the enclosing function/type — a doc
  directive covers the whole declaration). Directives that suppress nothing
  are BD002 findings. Directives of inactive rules are neither honored nor
  policed; malformed directives stay loud in every mode. BD003 has no
  repair by design: a real fix rewrites the declaration and every call
  site, which is a codemod's job, not a lint repair.

### Fixed

- **Repair never touches files under a testdata path.** Walk-level dir
  skipping did not apply to testdata paths passed directly or to
  single-file invocations, and an orchestrator replaying stale per-file
  findings corrupted this linter's own parse-only fixtures through that
  hole. Repair now skips every file under a `testdata/` path segment
  (detection keeps scanning explicitly-passed testdata roots so fixture
  tests still work).

### Migration from branching-flow strong-id

- Replace `branching-flow strong-id <path>` with
  `brandid-lint -rules BD003 <path>` (add BD001 if you also want brand
  hygiene; SARIF and text output available via `-format`).
- `//nolint:branching-flow` and `//lint:ignore STRONG_ID` suppressions are
  not bridged (verified: zero uses outside branching-flow's own fixtures).
  Use `//brandid-lint:ignore(BD003) <reason>`.
- Parity evidence: `docs/status/2026-10-11_11-50_bd003-parity-report.md` —
  26/26 findings identical on the frozen corpus; both documented
  divergences are the intended suppression-contract change.

## [0.3.0] - 2026-10-11

### Fixed

- **BD001 no longer fires on brands whose `Name()` method exists but does
  not return a string literal** (issue #10). Detection asked "does the
  method return a statically parseable literal?" while the shipped rule
  contract asks "does a `Name() string` method exist?" — a `Name()`
  returning a package-level constant was treated as missing, so Detect
  produced a false BD001 on a compliant type and Repair inserted a duplicate
  method that did not compile (go-sse saw five duplicate-method commits in
  one day before inlining the literal as a workaround). Detection is now
  existence-based, matching the rule contract: any `Name() string` method
  counts, whatever its body returns. Repair can no longer insert a stub next
  to an existing same-file `Name()` method, which makes the idempotence
  guarantee unconditional. A suppression directive on a brand with a
  constant-returning `Name()` is now correctly reported as stale (BD002).
- **`testdata/` directories are never scanned**, by the same convention as
  go vet, staticcheck, and golangci-lint: fixture files hold intentionally
  incomplete brands, and repo-wide scans were rewriting them. This
  repository's own `linter/testdata` fixtures were corrupted by repair
  passes twice (restored 2026-10-05, re-corrupted 2026-10-08). A directory
  passed as the scan root is still visited, so scanning fixtures directly
  keeps working.
- The reported tool `Version` tracks releases again (v0.2.0 shipped with
  SARIF `ToolInfo` still saying 0.1.0).

## [0.2.0] - 2026-09-29

### Added

- **BD002 — broken suppression directives stay loud**: brands that
  deliberately skip `Name()` (e.g. CQRS marker types whose `String()` output
  is a storage key) are suppressed in source with
  `//brandid-lint:ignore(BD001) <reason>`, placed on the declaration line or
  as the last line of the comment group directly above it. The reason is
  mandatory and only BD001 is suppressible. A directive that cannot do its
  job — malformed syntax, missing reason, unknown rule, stale (the brand has
  `Name()` by now), duplicate, or annotating no declaration — is itself
  reported as a BD002 finding, because a silent no-op directive is worse
  than none. Suppressed brands produce no BD001 finding and are never
  repaired.
- **Repair capability**: `linter.Repair(ctx)` / `RepairPath(path)` insert the
  suggested `Name()` stub after the type declaration of every unsuppressed
  BD001 finding, as byte-level insertions spliced in ascending offset order
  (grouped `type (...)` blocks keep stub order aligned with declaration
  order). Repair is idempotent, preserves file permissions, keeps gofmt-clean
  files gofmt-clean (trailing-newline handling for EOF, blank-line, and
  adjacent-declaration cases), and never touches suppressed or already-named
  brands. BD001 findings now carry `FixStrategy: direct` with
  `BeforeCode`/`AfterCode`.
- **CLI `-fix` flag**: applies the repair, then re-detects so the exit code
  reflects what remains (dry-run remains the default; `-fix` is the only
  mode that writes files).
- **BuildFlow provider Repairer**: the toolsdk spec now registers a real
  `Repair` (BuildFlow re-runs Detect to measure the delta, as designed).

## [0.1.0] - 2026-09-29

### Added

- **BD001 — unnamed brand detection**: flags brand types (empty structs used
  as the first type argument of `id.ID[...]`) that have no `Name() string`
  method, as `go-finding` findings positioned at the type declaration, with a
  concrete `Name()` stub as the suggestion. Detection is purely syntactic
  (`go/parser`); files that fail to parse are skipped.
- **BuildFlow provider** (`linter/provider`): self-registers the tool Spec
  (`Name: brandid-lint`, `ModuleFanOut`, go-module trigger) in the
  process-global toolsdk registry. Blank-import it to activate.
- **CLI** (`linter/cmd/brandid-lint`): scans files or directory trees, prints
  findings as text or SARIF. Exit codes: 0 clean, 1 findings, 2 error.
- Skips `.git`, `vendor`, and `node_modules` directories; scans test files.

[0.3.0]: https://github.com/larsartmann/go-branded-id/releases/tag/linter%2Fv0.3.0
[0.2.0]: https://github.com/larsartmann/go-branded-id/releases/tag/linter%2Fv0.2.0
[0.1.0]: https://github.com/larsartmann/go-branded-id/releases/tag/linter%2Fv0.1.0
