# Changelog

All notable changes to the `brandid-lint` module will be documented in this
file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

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
