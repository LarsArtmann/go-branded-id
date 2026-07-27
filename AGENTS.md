# go-branded-id — Agent Context

## What This Is

A Go library providing branded, strongly-typed identifiers using phantom types (generics). `ID[Brand, Value]` prevents mixing different entity IDs at compile time. Single package (`package id`) at repository root.

## Essential Commands

All build/test tasks go through the Nix flake. There is no `justfile` and no `Makefile` — all automation is in `flake.nix`.

Standard `go build`/`go test` commands work without any special environment variables. The library supports both `encoding/json` (v1, default) and `encoding/json/v2` (when `GOEXPERIMENT=jsonv2` is set) via build tags — see "Dual JSON v1/v2 Support" in Critical Gotchas below.

| Command                        | Purpose                                                 |
| ------------------------------ | ------------------------------------------------------- |
| `nix run .#test`               | Run tests (`go test ./... -count=1`)                    |
| `nix run .#test-race`          | Run with race detector                                  |
| `nix run .#build`              | Build (`go build ./...`)                                |
| `nix run .#lint`               | Run golangci-lint                                       |
| `nix run .#vet`                | Run `go vet ./...`                                      |
| `nix run .#coverage`           | Generate and display coverage report                    |
| `nix run .#clean`              | Clean test cache and coverage.out                       |
| `nix flake check`              | Run all flake checks (includes build)                   |
| `nix fmt`                      | Format everything (gofumpt, goimports, golines, nixfmt) |
| `go test ./... -count=1`       | Plain Go test (no Nix needed)                           |
| `go test ./... -race -count=1` | Plain race test                                         |

The dev shell (`nix develop`) sets `GOWORK=off` and provides Go 1.26, golangci-lint, gopls, and trash-cli.

## Code Organization

```
.
├── id.go              # Core ID type, NewID, Get, IsZero, Equal, Compare, Or, String, GoString, Format
├── id_brand.go        # BrandNamer interface, BrandName, ValidateID, ValidateIDWithValue, MustValidateID
├── id_ptr.go          # Ptr(), FromPtr() for optional ID fields
├── id_json_v1.go      # MarshalJSON / UnmarshalJSON (encoding/json, default)
├── id_json_v2.go      # MarshalJSON / UnmarshalJSON (encoding/json/v2, build-tagged)
├── id_sql.go          # Scan / Value for database/sql (all int/uint/string types)
├── id_text.go         # MarshalText / UnmarshalText (XML, TOML)
├── id_binary.go       # MarshalBinary / UnmarshalBinary (little-endian)
├── id_gob.go          # GobEncode / GobDecode (delegates to binary)
├── cmd/namer/         # Standalone codemod tool: scans Go files for brand types missing Name() method
├── website/           # Astro + Starlight documentation website (deployed to Firebase Hosting)
└── *_test.go          # Tests, benchmarks, fuzz tests, example tests
```

There is no `internal/` or `pkg/` — this is intentionally a flat, single-package library.

## Architecture & Data Flow

- `ID[B any, V comparable]` is a struct wrapping `value V`. Zero value means "unset".
- **Brand types** are empty structs (`type UserBrand struct{}`). They exist only as phantom type parameters.
- **Named brands** implement `BrandNamer` (`Name() string`). This affects `String()`, `GoString()`, `ValidateID` error messages, and `%#v` formatting.
- **Serialization always uses raw values** — never the brand prefix. JSON, SQL, Text, Binary, Gob all call `valueString()` internally.
- `String()` is for **human display**: `"User:abc123"` for named brands, `"abc123"` for unnamed.
- `Get()` is for **programmatic use**: always returns the raw value.

## Naming Conventions

- Source files: `id_<feature>.go` (e.g., `id_json_v1.go`, `id_binary.go`)
- Build-tagged pairs: `id_<feature>_v{1,2}.go` with `//go:build goexperiment.jsonv2` / `!goexperiment.jsonv2`
- Test files: `id_<feature>_test.go` or `id_<scope>_test.go` (e.g., `id_brand_test.go`, `id_alltypes_test.go`)
- Test brand types in `_test.go` files: `StringBrand`, `Int64Brand`, etc. (PascalCase, no `test` prefix except `testUserBrand` in `id_brand_test.go`)
- Generic test helpers: `test<Name>` or `assert<Cmp><Action>` (e.g., `testIDRoundTrip`, `assertCmpEqual`)

## Testing Approach

- **All tests use `t.Parallel()`** — both at the top-level function and inside `t.Run` subtests. Do not omit this.
- Subtests via `t.Run("descriptive name", func(t *testing.T) { ... })` for scenarios.
- Generic helpers take brand and value type parameters: `testIDRoundTrip[B any, V comparable](t *testing.T, ...)`.
- Shared assertion: `assertCmpEqual[T comparable](tb testing.TB, got, want T)`.
- Test brands are empty structs declared in test files.
- Fuzz tests for JSON and Binary round-trips (string, int64, uint64).
- Benchmarks use `b.Loop()` (Go 1.24+ pattern).
- Example functions (`Example*`) are used for documentation-driven testing.

## Linting & Code Quality

- **golangci-lint v2** with an extremely strict config (`.golangci.yml`). Many linters enabled including `exhaustruct`, `gochecknoglobals`, `paralleltest`, `wrapcheck`, `cyclop`, `funlen`, etc.
- Cyclop max complexity: 12.
- Line length: 120 (golines).
- Formatter: gofumpt (stricter than gofmt) + goimports + golines.
- `nolint` comments are common and expected. Typical patterns:
  - `//nolint:forcetypeassert // guaranteed by type switch`
  - `//nolint:gosec,forcetypeassert // G115: ... safe for serialization; guaranteed by type switch`
  - `//nolint:cyclop,funlen // exhaustive type switch over numeric types`
- `exhaustruct` is disabled for test files (`_test\.go`) and generated files.
- `exported` and `package-comments` rules in revive are **disabled** — no package comment required on every file.
- Tests count toward linting (`tests: true` in config).

## Critical Gotchas

### Dual JSON v1/v2 Support — Build Tags

The library supports both `encoding/json` (v1) and `encoding/json/v2` via Go build tags:

- `id_json_v1.go` (`//go:build !goexperiment.jsonv2`) — uses `encoding/json`, compiled by default
- `id_json_v2.go` (`//go:build goexperiment.jsonv2`) — uses `encoding/json/v2`, compiled when consumer sets `GOEXPERIMENT=jsonv2`
- `json_helpers_v{1,2}_test.go` — same pattern for test helpers

The `goexperiment.jsonv2` build tag is set automatically by the Go toolchain when `GOEXPERIMENT=jsonv2` is active (Go 1.25–1.26). In Go 1.27+, json/v2 becomes the default and the tag is always true.

**CI tests both modes** — every workflow runs build + tests in both v1 (no flag) and v2 (`GOEXPERIMENT=jsonv2`) modes.

**Historical note:** Previously the library hard-required `GOEXPERIMENT=jsonv2`, which caused the v0.3.1 release to fail (CI didn't set it). The dual-mode approach eliminates this class of problem entirely.

**CRITICAL: goimports corrupts v1 files.** Running `nix fmt` or any `goimports`-based formatter will silently rewrite `"encoding/json"` to `"encoding/json/v2"` in `id_json_v1.go` and `json_helpers_v1_test.go` (because goimports sees `json.Marshal` calls and picks the v2 package). This breaks the v1 build entirely — `go build ./...` without `GOEXPERIMENT` fails with `build constraints exclude all Go files in encoding/json/v2`. **Always run `go build ./...` (no GOEXPERIMENT) after any formatting pass** and manually fix the imports back if corrupted. This has happened multiple times.

### String() vs Get() — Know the Difference

`String()` changed behavior in v0.3.0. For named brands it now returns `"Brand:value"`. **Serialization never uses String()** — it always uses `valueString()` internally. But if _user code_ was parsing `String()` output, it will break after adding `Name()` to a brand.

**Rule**: Use `Get()` for any programmatic value extraction. Use `String()` only for display/logging.

### BrandName[B](<>) Fallback

For unnamed brands (no `Name()` method), `BrandName[B]()` returns `fmt.Sprintf("%T", brand)` — this includes the package path (e.g., `"id.Int64Brand"` or `"main.UserBrand"`). `GoString()` and `%#v` always call `BrandName[B]()`, so unnamed brands show full package-qualified names in debug output.

### Zero Value Semantics

- `IsZero()` compares against the zero value of `V`.
- JSON: zero value serializes to `null`.
- SQL: `Value()` returns `nil` for zero values.
- Text/Binary: zero value returns `nil` / empty bytes.
- `Scan(nil)` and `UnmarshalJSON(null)` both reset to zero value.

### Compare Limitations

`Compare()` only supports ordered types: `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `string`. Returns `ErrNotOrdered` for anything else. There is no generic constraint enforcing this at compile time — it's a runtime check via type switch.

### SQL Scan Type Coercion

`Scan` accepts `int64`, `int`, and `float64` from database drivers and casts them to the target integer type. This is necessary because different SQL drivers return different Go types. The casts have `gosec G115` suppressions with justifications about safe serialization boundaries.

### Binary Serialization Endianness

Binary marshaling uses **little-endian** for all numeric types. `int` is serialized as 8 bytes (uint64). This is an implementation detail but matters for cross-language compatibility.

### Nix Sandbox Build Cache (GOCACHE)

`nix flake check` builds the `checks.build` derivation in a sandbox where `HOME=/homeless-shelter` (read-only). Go's build cache cannot initialize at `$HOME/.cache/go-build`. The flake's `checks.build` sets `GOCACHE=$TMPDIR/go-cache` to work around this. Any Go-based Nix check derivation needs this.

### No go.work / GOWORK=off

The flake explicitly sets `GOWORK=off`. This library is not part of a Go workspace.

### Lint Action Version Mismatch (Fixed)

The release workflow (`release.yml`) used `golangci-lint-action@v6` while `go.yml` used `@v7`. Fixed in this session — both now use `@v7`.

## Ecosystem Context

This library was extracted from `go-composable-business-types/id`. It has 14 downstream repos in the ecosystem. The `cmd/namer` tool was created to help migrate those repos by identifying brand types missing `Name()` methods.

When making breaking changes, consider the migration impact across:

- InboxClean, CreditReformBilanzampel, ActaFlow, SEC, storbi, ChastityAPI, smart-configs, StopTube, universal-workflow, Zlota44, timesheets, complaints-mcp, cqrs-htmx, emeet-pixyd

### Brands That Deliberately Skip `Name()`

Not all brand types should implement `Name()`. The `cmd/namer` tool may flag these, but they are correct as-is:

- **go-cqrs-lite marker types** — These brands serve as event/stream type identifiers in the CQRS framework. Their `String()` output is used directly as storage keys and stream names. Adding `Name()` would change `String()` from `"TypeName"` to `"HumanReadable:TypeName"`, breaking the key format in event stores and breaking existing data.
- **BerryBig** — Test brands only, no production impact.
- **Cyberdom** — No brand types at all.

**Rule**: If a brand's `String()` output is used as a data key (storage, stream name, routing key), do NOT add `Name()`. The `cmd/namer` tool flags them, but the flag is a false positive in this context.

## Release Process

- CI creates a GitHub Release automatically on semver tags (`v*.*.*`) — pattern: `v[0-9]+.[0-9]+.[0-9]+*`.
- Release workflow (`.github/workflows/release.yml`) runs tests with race detector + golangci-lint before creating the release.
- Tags must be signed (SSH) and annotated (`git tag -a`).
- **To release**: update CHANGELOG, commit, tag, push tag: `git push origin v0.3.1`.
- `git-town.toml` configures `master` as the main branch.
- BuildFlow pre-commit hook runs 34 checks (Go mode) including golangci-lint, gofumpt, goimports, statix, gitleaks, doc-files-age-check (max 3w freshness), and nix-flake-check.
- `doc-files-age-check` requires README.md and TODO_LIST.md to be updated within 3 weeks of code changes — SARIF format reveals the specific stale file (`buildflow --step doc-files-age-check --format sarif`).

## Website

The `website/` directory contains an Astro + Starlight documentation site deployed to Firebase Hosting.

- **Live URL**: `https://branded-id.lars.software` (custom domain, DNS pending `terraform apply` in `domains/` repo)
- **Temporary URL**: `https://brandedid.web.app` (Firebase default, works now)
- **Firebase project**: `lars-software`
- **Hosting target**: `brandedid`
- **Color theme**: Violet (#a855f7)
- **DNS**: CNAME `branded-id.lars.software` → `brandedid.web.app` (in `domains/lars.software.tf`, needs `terraform apply`)
- **Build**: `nix run .#build` (from `website/`) or `npm run build`
- **Dev**: `nix run .#dev` (from `website/`) or `npm run dev`
- **Deploy**: `nix run .#deploy` (from `website/`) — builds and runs `firebase deploy --only hosting`
- The website has its own `flake.nix`, `package.json`, and `firebase.json` — independent from the Go library's flake
