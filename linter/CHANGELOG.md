# Changelog

All notable changes to the `brandid-lint` module will be documented in this
file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

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

[0.1.0]: https://github.com/larsartmann/go-branded-id/releases/tag/linter%2Fv0.1.0
