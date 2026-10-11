# go-branded-id — Agent Context

## What This Is

A Go library providing branded, strongly-typed identifiers using phantom types (generics). `ID[Brand, Value]` prevents mixing different entity IDs at compile time. Single package (`package id`) at repository root.

The repository is **two independent Go modules**: the zero-dependency root library and `linter/` (`github.com/larsartmann/go-branded-id/linter`, the `brandid-lint` BuildFlow provider/CLI). The linter sub-module does NOT import the root module, so there is no `go.work` — each module builds standalone with `GOWORK=off`.

## Essential Commands

All build/test tasks go through the Nix flake. There is no `justfile` and no `Makefile` — all automation is in `flake.nix`.

Standard `go build`/`go test` commands work without any special environment variables — in the ROOT module. The linter module has its own floor (`go 1.27.1`, via go-finding) and its own `.golangci.yml`; `cd linter` before running plain `go` commands there. The library supports both `encoding/json` (v1, default) and `encoding/json/v2` (when `GOEXPERIMENT=jsonv2` is set) via build tags — see "Dual JSON v1/v2 Support" in Critical Gotchas below.

| Command                      | Purpose                                                             |
| ---------------------------- | ------------------------------------------------------------------- |
| `nix run .#test`             | Run tests for BOTH modules (root in json v1 + v2, linter once)      |
| `nix run .#test-race`        | Run with race detector (both modules)                               |
| `nix run .#build`            | Build both modules                                                  |
| `nix run .#lint`             | Run golangci-lint (root v1+v2, then `linter/` with its own config)  |
| `nix run .#vet`              | Run `go vet ./...` (both modules)                                   |
| `nix run .#coverage`         | Generate and display coverage report (both modules)                 |
| `nix run .#clean`            | Clean test cache and coverage.out                                   |
| `nix flake check`            | Run all flake checks (root build/test + linter build/test + format) |
| `nix fmt`                    | Format everything (gofumpt, goimports, golines, nixfmt)             |
| `go test ./... -count=1`     | Plain Go test in the CURRENT module (root: no Nix needed)           |
| `cd linter && go test ./...` | Linter module tests (needs go ≥ 1.27.1; dev shell ships it)         |

The dev shell (`nix develop`) sets `GOWORK=off` and provides Go 1.27.1 (satisfies both module floors), golangci-lint, gopls, and trash-cli. The root checks in `nix flake check` deliberately build with `go_1_26` to enforce the root consumer floor.

## Code Organization

```
.
├── go.mod             # module github.com/larsartmann/go-branded-id (go 1.26, ZERO dependencies)
├── id.go              # Core ID type, NewID, Get, IsZero, Equal, Compare, Or, String, GoString, Format
├── errors.go          # Sentinel errors (ErrInvalidID, ErrNotOrdered, ErrUnsupportedType, etc.)
├── id_brand.go        # BrandNamer interface, BrandName, ValidateID, ValidateIDWithValue, MustValidateID
├── id_ptr.go          # Ptr(), FromPtr() for optional ID fields
├── id_json_v1.go      # MarshalJSON / UnmarshalJSON (encoding/json, default)
├── id_json_v2.go      # MarshalJSON / UnmarshalJSON (encoding/json/v2, build-tagged)
├── id_sql.go          # Scan / Value for database/sql (all int/uint/string types)
├── id_text.go         # MarshalText / UnmarshalText (XML, TOML)
├── id_binary.go       # MarshalBinary / UnmarshalBinary (little-endian)
├── id_gob.go          # GobEncode / GobDecode (delegates to binary)
├── linter/            # SEPARATE Go module: brandid-lint (see linter/go.mod, linter/CHANGELOG.md)
│   ├── detect.go      #   Detect/DetectWithRules — shared fileScan walk (skips testdata/, vendor/, generated)
│   ├── rules.go       #   RuleSet (BD001 default, BD003 opt-in) + ParseRuleSet for -rules
│   ├── brands.go      #   Syntactic brand detection (empty struct + id.ID[...] type arg)
│   ├── primitives.go  #   BD003: primitive-typed ID params/returns/fields, path exclusions, suppression
│   ├── idpattern.go   #   BD003 name heuristic (containsIDPattern: id/ids, _id suffix, PID-class acronyms)
│   ├── suggest_id.go  #   BD003 suggested type/brand names, message, severity ranking tables
│   ├── finding.go     #   BD001/BD002/BD003 finding construction (go-finding Builder)
│   ├── generated.go   #   "Code generated ... DO NOT EDIT." header detection (skip lint + repair)
│   ├── suggest.go     #   Name() suggestion derivation
│   ├── suppress.go    #   //brandid-lint:ignore(...) directive parsing + BD002
│   ├── repair.go      #   Repair(ctx)/RepairPath(path) — Name() stub insertion (idempotent, BD001 only)
│   ├── provider/      #   toolsdk self-registration (BuildFlow blank-imports this)
│   ├── cmd/brandid-lint/ # CLI: -rules, -format text|sarif, -fix; exit 0=clean/1=findings/2=error
│   └── testdata/      #   Parse-only fixtures (never compiled)
├── website/           # Astro + Starlight documentation website (deployed to Firebase Hosting)
└── *_test.go          # Tests, benchmarks, fuzz tests, example tests
```

There is no `internal/` or `pkg/` — this is intentionally a flat, single-package library (root module).

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

- **golangci-lint v2** with an extremely strict config (`.golangci.yml`). Many linters enabled including `exhaustruct_v5`, `gochecknoglobals`, `paralleltest`, `wrapcheck`, `cyclop`, `funlen`, etc.
- **Linter renames break `nolint` comments silently**: current golangci-lint ships `exhaustruct_v5` (renamed from `exhaustruct`); a `//nolint:exhaustruct` comment no longer suppresses anything and the finding reappears. `nolint` comments must reference the linter's CURRENT name (2026-09-29: all `exhaustruct` nolints and the `.golangci.yml` test-file exclusion entry were updated to `exhaustruct_v5`).
- Cyclop max complexity: 12.
- Line length: 120 (golines).
- Formatter: gofumpt (stricter than gofmt) + goimports + golines.
- `nolint` comments are common and expected. Typical patterns:
  - `//nolint:forcetypeassert // guaranteed by type switch`
  - `//nolint:gosec,forcetypeassert // G115: ... safe for serialization; guaranteed by type switch`
  - `//nolint:cyclop,funlen // exhaustive type switch over numeric types`
- `exhaustruct_v5` is disabled for test files (`_test\.go`) and generated files.
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

**Root cause and current state (updated 2026-09-22):** the corruption was caused by BuildFlow's `go-auto-upgrade` step (NOT goimports, which is build-tag-aware and innocent). The step was skipped via `.buildflow.yml` from 2026-08-02; on 2026-09-22 it was **re-enabled** after retesting against gau v0.6.2, whose dual-mode guard now detects `goexperiment.jsonv2` file pairs and skips them, and whose module go-version gate blocks remaining v1→v2 swaps at `go 1.26`. A run is a green no-op emitting info-level notes. This project deliberately supports BOTH v1 and v2 via build tags.

**If the v1 build ever breaks again** with `build constraints exclude all Go files in encoding/json/v2`, the imports in `id_json_v1.go` / `json_helpers_v1_test.go` have been rewritten to `"encoding/json/v2"` — change them back to `"encoding/json"` and confirm `go build ./...` (no GOEXPERIMENT) passes. The pre-push hook greps these imports before running tests, so a push with corrupted imports fails fast with a clear message.

**Dev-shell gotcha (Go 1.27.1 + root module floor `go 1.26`)**: inside `nix develop`, plain `go vet`/`go test ./...` on the ROOT module fails with `json.Marshal requires go1.27 or later (module is go1.26)` — in Go 1.27 the `goexperiment.jsonv2` tag is default-true, so the v2 files compile, but their API is language-version-gated above the module's `go 1.26` directive. Use `GOEXPERIMENT=jsonv2 go test ./...` in the dev shell (or run root tests outside it with ambient Go 1.26.x). `go build ./...` is unaffected. Since v0.7.0 the flake apps (`nix run .#test` etc.) pin per-module toolchains (root legs go1.26, linter legs go1.27) and are immune to this gotcha — before that they ran the dev-shell Go for root legs, under which the "v1" leg either failed the language gate or silently tested v2 twice. Do NOT "fix" this by raising the root `go` directive.

### String() vs Get() — Know the Difference

`String()` changed behavior in v0.3.0. For named brands it now returns `"Brand:value"`. **Serialization never uses String()** — it always uses `valueString()` internally. But if _user code_ was parsing `String()` output, it will break after adding `Name()` to a brand.

**Rule**: Use `Get()` for any programmatic value extraction. Use `String()` only for display/logging.

### BrandName[B]() Fallback

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

### Go Version Pins Must Move Together (go.mod / flake.nix / CI)

The Go toolchain version is pinned in **three places per module that must
agree**: the module's `go` directive, `flake.nix` (`goPkg = pkgs.go_1_26` for
the root, `linterGoPkg = pkgs.go_1_27` for the linter), and
`.github/workflows/go.yml` (`go-version`, with `GOTOOLCHAIN=local` as
tripwire). Auto-upgraders sometimes bump only `go.mod`, which breaks the Nix
sandbox build (`could not create module cache: mkdir /homeless-shelter`),
local shells with `GOTOOLCHAIN=local`, and gopls/golangci-lint (`go.mod
requires go >= ...`). If `go build ./...` fails with a version error while no
source changed, compare the module's `go.mod` against `flake.nix`. Bump all
three in one deliberate commit — the `go` directive is a consumer-facing
minimum for a library, not something to auto-bump. The root floor is `go
1.26`; the linter floor is `go 1.27.1` (go-finding requires it).

**Machine tripwire**: `scripts/check-go-pins.sh` asserts the three-way
agreement for BOTH modules (go.mod directive ≤ flake `go_1_2x` pin on
major.minor, and every CI `go-version` matches one of the pins). It runs as
the CI `go-pins` job in `go.yml` and as the first guard in the pre-push
hook. When it fires, restore the directive AND check the other two pins —
never "fix" it by bumping the flake pin to match the daemon's rewrite.

**The recurring writer is BuildFlow's `go-mod-update` step** ("Updates Go
toolchain and dependencies"), NOT `go-auto-upgrade` (that one rewrites JSON
imports, and was itself fixed upstream in gau v0.6.2). It bumped `go 1.26` →
`1.27.1` twice (documented pre-v0.6.0, and again on 2026-09-17 via daemon
commit `1acc808`, caught by CI within minutes). A third instance landed via
daemon commit `d2e3118` (2026-09-29, `1.27.1` → `1.27`) and broke root builds
until it was restored to `1.26`. `go-mod-update` is in `skip_steps` in
`.buildflow.yml` (2026-09-22: `go-auto-upgrade` was un-skipped after its
upstream fix, so only `go-mod-update` remains skipped). The bump lands as a
dirty file that the auto-commit daemon later sweeps into an unrelated commit
— check `git log -p -- go.mod` when CI fails with `go.mod requires go >= ...`.
CI is the reliable tripwire (root and linter jobs set `GOTOOLCHAIN=local`).
**Fourth instance (release day, 2026-09-29)**: unpushed daemon commit
`8262d3c` bumped root `1.26` → `1.27`, and pushed `7a970f5` re-swept
`linter/go.mod` `1.27.1` → `1.27` — despite `go-mod-update` still being in
`skip_steps`, so the writer is unidentified (remaining suspect:
`go-auto-upgrade`'s directive bump; check the BuildFlow audit-log if it
recurs). Both restored before tagging; `scripts/check-go-pins.sh` caught the
root bump locally pre-push — the tripwire's first in-anger catch.
**Fifth instance (2026-10-08, caught 2026-10-11)**: daemon commit `9e5190d`
bumped BOTH go.mods at once (root `1.26` → `1.27`, linter `1.27.1` →
`1.27`) in the same commit that re-corrupted the linter testdata fixtures;
`check-go-pins.sh` caught it and both directives were restored. The writer
is still unidentified — treat any `go`-directive-only diff in an
auto-commit as hostile and restore it.

### Nix Sandbox Build Cache (GOCACHE)

`nix flake check` builds the `checks.build` derivation in a sandbox where `HOME=/homeless-shelter` (read-only). Go's build cache cannot initialize at `$HOME/.cache/go-build`. The flake's `checks.build` sets `GOCACHE=$TMPDIR/go-cache` to work around this. Any Go-based Nix check derivation needs this.

### Flake `outputs` Must Include `...` for Unnamed Inputs

The `outputs` function in `flake.nix` destructures `inputs@{ self, flake-parts, treefmt-nix, ... }`. The `...` is **required** because not all inputs are named in the pattern. Without it, Nix fails with an error about extra arguments. Always include `...` when adding new flake inputs to this project.

### Flake Systems Are Inlined (No `nix-systems` Input)

The supported systems are declared inline in `flake.nix` (`x86_64-linux`, `aarch64-linux`, `aarch64-darwin`) — the `nix-systems/default` input was removed in v0.6.0 prep because nixpkgs unstable 26.11 dropped `x86_64-darwin` and that input's default list still contained it, breaking `nix flake check`. Related: CI's `flake-check` job runs two passes — `nix flake check --all-systems --no-build` (eval-only across all systems) plus `nix flake check` (builds current-system checks). Newer Nix builds checks for ALL systems under `--all-systems`, which fails with a platform mismatch for foreign systems on a single-arch runner.

### No go.work / GOWORK=off

The flake explicitly sets `GOWORK=off`. There is deliberately no `go.work`:
the two modules are independent (the linter does not import the root module),
so a workspace would add nothing. Run Go commands from inside the module you
target; `./...` from the root never sees `linter/` (own go.mod) and vice
versa.

### Two-Module Layout (root stays zero-dependency)

`linter/` is a fully separate Go module (`github.com/larsartmann/go-branded-id/linter`)
with its own go.mod, go.sum, .golangci.yml (`run.go: 1.27`, no experiment
build-tags), CHANGELOG.md, and version tags (`linter/vX.Y.Z` — the `linter/`
prefix is REQUIRED for sub-module tags). It depends on go-finding +
go-finding/toolsdk. The dependency direction is load-bearing:

- **The root module must NEVER depend on go-finding or on `linter/`.** The
  root library is the zero-dependency foundation for 14+ downstream repos,
  and the go-finding CLI itself pins go-branded-id — importing go-finding
  from the root would invert the layering and create a module cycle.
- The linter does NOT import the root module either (detection is purely
  syntactic), so the modules can version independently.
- BuildFlow integration: `linter/provider` self-registers a toolsdk Spec
  (Detect + Repair); BuildFlow consumes it via a blank import in
  `tools/providers/sdk_imports.go` plus a go.mod require. Tool name:
  `brandid-lint`, rule IDs: `BD001`, `BD002`, `BD003` (stable forever).
  Repair inserts `Name()` stubs (BD001 only); BuildFlow re-runs Detect to
  measure the delta.
- In-source suppression: `//brandid-lint:ignore(BD001) <reason>` on the decl
  line or the last line of the comment group directly above it suppresses
  BD001; a broken directive (missing reason, unknown rule, stale, duplicate,
  unplaced) is reported as BD002. Suppressed brands are never repaired.
- Known false-positive class resolved by suppression: brands whose
  `String()` is a data key (go-cqrs-lite `StreamMarker`/`TimerMarker`) now
  carry the suppression directive in their own repos —
  see "Brands That Deliberately Skip Name()" below.
- **BD001 detection is existence-based** (issue #10, linter v0.3.0): a brand
  counts as named when ANY `Name() string` method exists — literal,
  constant-identifier, or computed return. Do NOT reintroduce return-value
  parsing into `HasName`: treating a non-literal method as "missing" makes
  Detect lie ("has no Name() method") and Repair insert a duplicate method
  that does not compile (go-sse lost a day to that loop).
- **`testdata/` directories are never scanned** (`skipDir` in
  `linter/detect.go`, same convention as go vet/staticcheck/golangci-lint).
  Before this skip, repo-wide repair passes rewrote the linter's own
  parse-only fixtures TWICE (restored 2026-10-05 in `79ef83d`, re-corrupted
  by daemon `9e5190d` on 2026-10-08), leaving the linter's own tests red.
  A testdata dir passed as the scan ROOT is still visited, so the fixture
  tests keep working. If fixture tests fail with findings-count mismatches,
  check `git log -p -- linter/testdata` for stray stub insertions first.
- The root repo's own test brands (`StringBrand`, `Int64Brand`, … in
  `*_test.go`) gained `Name()` stubs in a deliberate 2026-10-01 cleanup
  (kept in `79ef83d`) so the linter runs clean on this repo. If they are
  ever stripped again, repo-wide scans will report them — that is the
  linter working as designed, not a bug.

### BD003 — Opt-In Primitive-ID Rule (ported from branching-flow strong-id)

BD003 (linter v0.4.0+) flags primitive-typed identifiers — `string`, `bool`,
all `int`/`uint` family params, named returns, and struct fields whose name
marks them as an ID — that should become branded `id.ID` types. Detection is
purely syntactic (`primitives.go` + `idpattern.go`): pointers, slices, and
named types never match; `String`/`Error` method decls are skipped (their
signatures must stay primitive). The name heuristic (`containsIDPattern`)
matches standalone `id`/`ids`, `_id`/`_ids` suffixes, compound names whose
LAST token is id/ids (`userId`, `validationID`), and single-letter-head
all-caps acronyms (PID, UID, RID) — but NOT English words (`isValid`,
`valid`, `paid`, `VALID`).

- **Opt-in only.** `DefaultRuleSet()` is BD001 alone: brandid-lint runs
  fleet-wide, and repos that deliberately silenced the old strong-id rule
  (webphone, DiscordSync) must not be flooded by an upgrade. Select via CLI
  `-rules BD001,BD003`, API `DetectWithRules`/`DetectPathWithRules`, or
  BuildFlow `tool_options: brandid-lint: {rules: "BD001,BD003"}`. BD002 is
  not selectable (rides along automatically); a typo'd or empty `-rules`
  token is a hard error, never a silent wrong-rules scan.
- **Detect-only.** `-fix` applies BD001 repairs only — a real BD003 fix
  rewrites the declaration and every call site (codemod territory, not a
  lint repair). Findings carry FixStrategy Suggest plus metadata:
  `suggestedTypeName`, `suggestedBrand`, `importPath`, and a paste-ready
  `typeDefinition`.
- **Severity model** (`suggest_id.go` prefix tables, ported for parity):
  high-priority domains (user, order, product, customer, account,
  transaction, payment, session, token) → error in signatures / warning in
  struct fields; medium-priority (entity, resource, item, record, entry,
  document, file, config, setting) → warning / info; everything else info.
  BuildFlow guards use `error_rules: brandid-lint: [BD003]`.
- **Path exclusions, BD003 only** (ported from strong-id; BD001 keeps
  scanning): `ui/`, `internal/output/`, `cmd/`, `testhelpers/`, `testutil/`,
  `docs/examples/`, `tools/` substrings, plus `_test.go`-family suffixes.
  Presentation/CLI wiring primitives are not domain-model debt.
- **Generated files are skipped** (v0.5.1): the standard `Code generated ...
  DO NOT EDIT.` header before the package clause (sqlc/templ/protoc/mockgen/
  stringer output) short-circuits `scanFile`, so detection AND repair stay
  out of machine-written code. The v0.4.0–v0.5.0 parity hole here (~90 false
  findings on CV's first real scan) is why the frozen parity corpus must
  always contain generated-shaped fixtures.
- **Suppression**: `//brandid-lint:ignore(BD003) <reason>` on the
  primitive-ID declaration line, the line directly above it, or the
  enclosing function/type declaration. A directive that covers nothing
  (bad placement) or is redundant with another directive is itself a BD002
  finding.
- **The legacy rule is gone**: branching-flow removed strong-id in v0.6.7;
  `branching-flow strong-id` remains reachable only through the parity
  script's v0.6.6 worktree. Consumers use `brandid-lint -rules BD003`.

### Linter vendorHash — Nix FOD Content-Addressing Trap

The linter module's deps are vendored through a fixed-output derivation
pinned by `vendorHash` in `flake.nix`. A stale hash does NOT fail with a
hash-mismatch error: FOD outputs are content-addressed, so the stale hash
resolves to an already-present store path and nix silently injects the OLD
vendor tree — surfacing later as misleading "inconsistent vendoring" build
errors (or worse, a green build of stale code).

- **Procedure after any linter dependency change**:
  `cd linter && go mod vendor && nix hash path linter/vendor`, paste the SRI
  hash into `flake.nix` `vendorHash`, then verify with `nix flake check`.
- **Escape hatch** when already confused: paste any valid-length dummy SRI
  hash to force the real `got:` printout, then use that.
- **BuildFlow side of the same coin**: BuildFlow's nix build embeds
  LarsArtmann deps via preparedSrc from **flake.lock input revs**, NOT
  go.mod versions — bumping go.mod alone changes nothing in the embedded
  provider. Re-embedding a linter release into BuildFlow requires the
  flake.lock input bump AND a profile reinstall; the installed binary is
  whatever the profile last installed, not what master says.

### Parity Gate (scripts/parity-bd003.sh) + Testdata-Corruption Postmortem

`scripts/parity-bd003.sh` A/B-diffs `brandid-lint -rules BD003` against the
legacy branching-flow strong-id analyzer (both render SARIF via go-finding;
normalized to `file:startLine` pairs) over a frozen corpus. Exit 0 = parity
or documented accepts, 1 = divergence (must be triaged in
`docs/status/*bd003-parity*.md`). Run:
`nix develop -c bash -c "bash scripts/parity-bd003.sh"`.

- **Self-healing legacy side** (post-v0.6.7-removal): when the
  branching-flow checkout no longer carries `pkg/strongid`, the script
  materializes a detached worktree of `BRANCHING_FLOW_TAG` (default v0.6.6,
  the last strong-id release) and fails LOUD if it cannot. Before this
  existed, the gate silently built an old-side binary WITHOUT the strong-id
  subcommand and compared against empty — a vacuous verdict. General rule:
  an A/B gate must fail loud when one side cannot run; an empty-side diff is
  not parity.
- **Testdata-corruption postmortem**: the recurring `linter/testdata`
  rewrites (2026-10-05 `79ef83d` restore, 2026-10-08 daemon `9e5190d`
  re-corruption) were caused by the daemon-side BuildFlow binary embedding
  linter v0.2.0, whose provider "repaired" parse-only fixtures during
  repo-wide runs. Fixed in three layers: `skipDir`/`underTestDataPath`
  guards in `detect.go` (scan + repair), the v0.5.0 re-embed into the
  BuildFlow profile binary (the corruption source itself), and the parity
  gate as regression harness. If fixtures corrupt again, check BOTH
  `git log -p -- linter/testdata` and which buildflow binary the daemon
  runs (`buildflow --version`).

### Dual-Mode Pre-Push Hook

A pre-push git hook (`scripts/pre-push-dual-test.sh`) first runs `scripts/check-go-pins.sh` (the three-way Go-pin guard), then greps the v1 JSON files for `encoding/json/v2` imports, then runs `go test` in both v1 and v2 JSON modes, reporting all results. It's installed at `.git/hooks/pre-push` (reinstall after editing the script: `cp scripts/pre-push-dual-test.sh .git/hooks/pre-push && chmod +x .git/hooks/pre-push`). Plain `go test` only exercises v1; the hook catches code that passes v1 but breaks v2 (build tag issues, import corruption) — and the grep guard catches import corruption that would prevent the v1 package from compiling at all, which no in-package test can observe.

## Ecosystem Context

This library was extracted from `go-composable-business-types/id`. It has 14 downstream repos in the ecosystem. The former `cmd/namer` codemod was created to help migrate those repos and lives on as the `brandid-lint` CLI + BuildFlow provider in the `linter/` sub-module.

When making breaking changes, consider the migration impact across:

- InboxClean, CreditReformBilanzampel, ActaFlow, SEC, storbi, ChastityAPI, smart-configs, StopTube, universal-workflow, Zlota44, timesheets, complaints-mcp, cqrs-htmx, emeet-pixyd (plus go-output, cmdguard, and the go-finding CLI, which pin the library in the linter-stack tooling graph)

### Brands That Deliberately Skip `Name()`

Not all brand types should implement `Name()`. Suppress the BD001 finding in source with `//brandid-lint:ignore(BD001) <reason>` — the reason documents the decision where it lives:

- **go-cqrs-lite marker types** — These brands serve as event/stream type identifiers in the CQRS framework. Their `String()` output is used directly as storage keys and stream names. Adding `Name()` would change `String()` from `"TypeName"` to `"HumanReadable:TypeName"`, breaking the key format in event stores and breaking existing data.
- **BerryBig** — Test brands only, no production impact.
- **Cyberdom** — No brand types at all.

**Rule**: If a brand's `String()` output is used as a data key (storage, stream name, routing key), do NOT add `Name()` — write the suppression directive instead. A directive without a reason is itself a BD002 finding.

## Release Process

- CI creates a GitHub Release automatically on semver tags (`v*.*.*`) — pattern: `v[0-9]+.[0-9]+.[0-9]+*`.
- Release workflow (`.github/workflows/release.yml`) runs tests with race detector + golangci-lint before creating the release.
- **Release notes come from the CHANGELOG**: the workflow extracts the `## [X.Y.Z]` section of `CHANGELOG.md` into the release body (`body_path`), with GitHub's `generate_release_notes` appending the compare link. GitHub's generated notes alone are PR-based and nearly empty in this direct-push repo — that's why the CHANGELOG section must be cut **before** tagging. A release with no notes means the tag was pushed without a dated CHANGELOG section.
- Tags must be signed (SSH) and annotated (`git tag -a`).
- **The linter sub-module releases separately**: tag `linter/vX.Y.Z` (the `linter/` prefix is mandatory for sub-module tags). Cut its section in `linter/CHANGELOG.md` before tagging. Root releases (`vX.Y.Z`) and linter releases do not have to coincide. Since `linter/v0.2.0`, `release.yml` matches `linter/v*` tags, verifies the sub-module (`linter-verify` job, Go 1.27 + `GOTOOLCHAIN=local`) and creates the GitHub Release from the `linter/CHANGELOG.md` section automatically — `linter/v0.1.0` predates this and has no release page. The CI hygiene job rejects tracked binaries ANYWHERE in the repo (a 7.6 MB `linter/brandid-lint` binary slipped past the old repo-root-only scan on 2026-09-29; `linter/brandid-lint` is gitignored).
- **To release**: update CHANGELOG, commit, tag, push the tag: `git push origin vX.Y.Z`.
- `git-town.toml` configures `master` as the main branch.
- BuildFlow pre-commit hook runs the configured Go-mode checks (golangci-lint, gofumpt, goimports, statix, gitleaks, doc-files-age-check (max 3w freshness), nix-flake-check); the exact count varies with `.buildflow.yml` — run `buildflow --dry-run` for the current list.
- `doc-files-age-check` requires README.md and TODO_LIST.md to be updated within 3 weeks of code changes — SARIF format reveals the specific stale file (`buildflow --step doc-files-age-check --format sarif`).

## Website

The `website/` directory contains an Astro + Starlight documentation site deployed to Firebase Hosting.

- **URL**: `https://branded-id.lars.software` (custom domain via CNAME → `brandedid.web.app`, configured in `domains/lars.software.tf`; the Firebase default `https://brandedid.web.app` serves the same site)
- **Firebase**: project `lars-software`, hosting target `brandedid`
- **Color theme**: Violet (#a855f7)
- **Build**: `nix run .#build` (from `website/`) or `pnpm run build`
- **Dev**: `nix run .#dev` (from `website/`) or `pnpm run dev`
- **Deploy**: `nix run .#deploy` (from `website/`) — builds and runs `firebase deploy --only hosting`
- The website uses **pnpm** (`pnpm-lock.yaml`, `packageManager` in `package.json`) and has its own `flake.nix` and `firebase.json` — independent from the Go library's flake
- **pnpm 11 reads `overrides` from `pnpm-workspace.yaml`, NOT `package.json`.** A top-level `overrides` block in `package.json` is npm syntax and is silently ignored by pnpm (and `pnpm.overrides` is deprecated in v11 — install warns "no longer read"). Security version pins live in `pnpm-workspace.yaml` under `overrides:`. Also: `pnpm update <pkg>` only touches direct dependencies — transitive deps need an override. Same class: build-script approvals live in `pnpm-workspace.yaml` under `allowBuilds:` (`esbuild: true`) — a missing or placeholder entry makes `astro build` fail on a missing esbuild binary (cmdguard incident, fixed 2026-09-19).
- The website's changelog page (`website/src/content/docs/changelog.mdx`) is a manually maintained summary of `CHANGELOG.md` — add the new version section there and run `nix run .#deploy` (from `website/`) as part of each release, or the live site goes stale.
