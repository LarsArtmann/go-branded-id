# go-branded-id — Agent Context

## What This Is

A Go library providing branded, strongly-typed identifiers using phantom types (generics). `ID[Brand, Value]` prevents mixing different entity IDs at compile time. Single package (`package id`) at repository root.

The repository is **two independent Go modules**: the zero-dependency root library and `linter/` (`github.com/larsartmann/go-branded-id/linter`, the `brandid-lint` BuildFlow provider/CLI). The modules do NOT import each other, and there is deliberately no `go.work` — run Go commands inside the module you target; `GOWORK=off` is set in the dev shell and flake.

## Essential Commands

No justfile/Makefile — all automation lives in `flake.nix`. Root floor: `go 1.26`. Linter floor: `go 1.27`. CI builds and tests the root module in BOTH JSON modes (v1 default + `GOEXPERIMENT=jsonv2`).

| Command                      | Purpose                                                         |
| ---------------------------- | --------------------------------------------------------------- |
| `nix run .#test`             | Tests for BOTH modules (root in json v1 + v2, linter once)      |
| `nix run .#test-race`        | Race detector (both modules)                                    |
| `nix run .#build`            | Build both modules                                              |
| `nix run .#lint`             | golangci-lint (root v1+v2, then `linter/` with its own config)  |
| `nix run .#brandid-lint`     | Run the linter CLI, e.g. `-- -rules BD003 <path>`               |
| `nix run .#vet`              | `go vet ./...` (both modules)                                   |
| `nix run .#coverage`         | Coverage report (both modules)                                  |
| `nix flake check`            | All flake checks (root build/test + linter build/test + format) |
| `nix fmt`                    | Format everything (gofumpt, goimports, golines, nixfmt)         |
| `go test ./... -count=1`     | Plain Go test in the CURRENT module (root: no Nix needed)       |
| `cd linter && go test ./...` | Linter module tests (needs go ≥ 1.27; dev shell ships it)       |

The dev shell (`nix develop`) sets `GOWORK=off` and ships Go 1.27.x, golangci-lint, gopls, and trash-cli. Flake checks deliberately build the root with `go_1_26` to enforce the root consumer floor.

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
│   ├── idpattern.go   #   BD003 name heuristic (containsIDPattern)
│   ├── suggest_id.go  #   BD003 suggested type/brand names, message, severity ranking
│   ├── finding.go     #   BD001/BD002/BD003 finding construction (go-finding Builder)
│   ├── generated.go   #   "Code generated ... DO NOT EDIT." header detection
│   ├── suggest.go     #   Name() suggestion derivation
│   ├── suppress.go    #   //brandid-lint:ignore(...) directive parsing + BD002
│   ├── repair.go      #   Repair — Name() stub insertion (idempotent, BD001 only)
│   ├── provider/      #   toolsdk self-registration (BuildFlow blank-imports this)
│   ├── cmd/brandid-lint/ # CLI: -rules, -format text|sarif, -fix; exit 0/1/2
│   └── testdata/      #   Parse-only fixtures (never compiled)
├── website/           # Astro + Starlight documentation website (own flake, Firebase Hosting)
└── *_test.go          # Tests, benchmarks, fuzz tests, example tests
```

There is no `internal/` or `pkg/` — intentionally flat, single-package (root module).

## Architecture & Data Flow

- `ID[B any, V comparable]` is a struct wrapping `value V`. Zero value means "unset".
- **Brand types** are empty structs (`type UserBrand struct{}`). They exist only as phantom type parameters.
- **Named brands** implement `BrandNamer` (`Name() string`). This affects `String()`, `GoString()`, `ValidateID` error messages, and `%#v` formatting.
- **Serialization always uses raw values** — never the brand prefix. JSON, SQL, Text, Binary, Gob all call `valueString()` internally.
- `String()` is for **human display**: `"User:abc123"` for named brands, `"abc123"` for unnamed.
- `Get()` is for **programmatic use**: always returns the raw value.

## Naming Conventions

- Source files: `id_<feature>.go`; build-tagged pairs: `id_<feature>_v{1,2}.go` with `//go:build goexperiment.jsonv2` / `!goexperiment.jsonv2`
- Test files: `id_<feature>_test.go` or `id_<scope>_test.go`
- Test brand types: `StringBrand`, `Int64Brand`, etc. (PascalCase; exception: `testUserBrand` in `id_brand_test.go`)
- Generic test helpers: `test<Name>` / `assert<Cmp><Action>` (e.g., `testIDRoundTrip`, `assertCmpEqual`)

## Testing Approach

- **All tests use `t.Parallel()`** — top-level AND inside `t.Run` subtests. Do not omit it (paralleltest linter).
- Subtests via `t.Run("descriptive name", ...)`; generic helpers take brand + value type params.
- Shared assertion: `assertCmpEqual[T comparable](tb testing.TB, got, want T)`.
- Fuzz tests for JSON and Binary round-trips; benchmarks use `b.Loop()`; `Example*` functions double as docs tests.

## Linting & Code Quality

- **golangci-lint v2**, extremely strict (`.golangci.yml`): `exhaustruct_v5`, `gochecknoglobals`, `paralleltest`, `wrapcheck`, `cyclop`, `funlen`, etc. Tests count toward linting.
- **Linter renames break `nolint` comments silently** — `nolint` must reference the linter's CURRENT name (e.g. `exhaustruct_v5`, not the old `exhaustruct`).
- Cyclop max complexity: 12. Line length: 120 (golines). Formatter: gofumpt + goimports + golines.
- `nolint` comments are common and expected, always with a justification: `//nolint:forcetypeassert // guaranteed by type switch`.
- `exhaustruct_v5` is disabled for test files and generated files; revive `exported`/`package-comments` are disabled.

## Critical Gotchas

### Dual JSON v1/v2 Support — Build Tags

- `id_json_v1.go` (`!goexperiment.jsonv2`) and `id_json_v2.go` (`goexperiment.jsonv2`) plus the `json_helpers_v{1,2}_test.go` pairs must stay in sync. This project deliberately supports BOTH modes.
- **Never rewrite the v1 imports to `"encoding/json/v2"`.** Signature of corruption: `build constraints exclude all Go files in encoding/json/v2` on a plain (no GOEXPERIMENT) build — fix by restoring `"encoding/json"` imports in `id_json_v1.go` / `json_helpers_v1_test.go`. The pre-push hook greps these imports and fails fast.
- **Dev-shell gotcha**: inside `nix develop`, plain root `go vet`/`go test ./...` fails (`json.Marshal requires go1.27 or later (module is go1.26)`) because Go 1.27 default-enables the jsonv2 tag, which is language-version-gated above the root's `go 1.26`. Use `GOEXPERIMENT=jsonv2 go test ./...` in the dev shell — the flake apps pin per-module toolchains and are immune. Do NOT "fix" this by raising the root `go` directive.

### String() vs Get()

`String()` returns `"Brand:value"` for named brands (display/logging only); `Get()` returns the raw value (programmatic use). Serialization never uses `String()`, but user code that parses `String()` output breaks when a brand gains `Name()`.

### `BrandName[B]()` Fallback

Unnamed brands render as `fmt.Sprintf("%T", brand)` — package-qualified (`"main.UserBrand"`). `GoString()` and `%#v` always go through `BrandName[B]()`.

### Zero Value Semantics

- `IsZero()` compares against the zero value of `V`; JSON zero → `null`; SQL `Value()` zero → `nil`; Text/Binary zero → `nil`/empty.
- `Scan(nil)` and `UnmarshalJSON(null)` both reset to the zero value.

### Compare / SQL / Binary

- `Compare()` supports ordered types only (ints, uints, string) via runtime type switch — returns `ErrNotOrdered` otherwise. No compile-time constraint.
- `Scan` accepts `int64`/`int`/`float64` (driver-dependent) and casts to the target integer type; the casts carry `gosec G115` suppressions with justifications.
- Binary marshaling is little-endian for all numeric types; `int` serializes as 8 bytes (uint64). Matters for cross-language compatibility.

### Go Version Pins Must Move Together

Three places per module must agree: the `go` directive, `flake.nix` (`go_1_26` root / `go_1_27` linter), and CI `go-version` (`GOTOOLCHAIN=local` as tripwire).

- `scripts/check-go-pins.sh` asserts the three-way agreement (CI `go-pins` job + first pre-push guard). When it fires: restore the go.mod directive AND check the other two pins — never "fix" it by bumping the flake pin to match the rewrite.
- **Treat any `go`-directive-only diff in an auto-commit as hostile and restore it.** The bump lands as a dirty file the auto-commit daemon sweeps into an unrelated commit — check `git log -p -- go.mod` when CI fails with `go.mod requires go >= ...`. `go-mod-update` is in `skip_steps` in `.buildflow.yml`, but writers have appeared anyway.
- The root `go` directive is a consumer-facing minimum for a library — bump all three places in one deliberate commit, never auto-bump.

### Tool False Positives — Deliberate Non-Fixes

- **go-version-auto-configure** warns "CI pins 1.26 but the module floor is 1.27" on every root-module CI job. It compares each `go-version` against the repo-wide max floor, but `go.yml`/`release.yml` map jobs to per-module pins (root jobs 1.26, linter jobs 1.27). The authoritative guard is `scripts/check-go-pins.sh`. Do not "fix" by bumping root CI jobs.
- **jscpd** would flag `website/dist/**` (gitignored build output mirroring `website/public/`) and the shared YAML in `.golangci.yml` vs `linter/.golangci.yml` (golangci-lint v2 has no config include mechanism). `.jscpd.json` owns the ignore list — keep its defaults in sync with BuildFlow's provider defaults, and re-check after BuildFlow changes its defaults.
- **lychee** 404s on private-repo links (e.g. `stargazers`) and fails transiently with exit 2 when GitHub rate-limits unauthenticated checks — run with `GITHUB_TOKEN=$(gh auth token)` to stabilize. Result-cache replays can resurface stale findings after a fix: bypass with `BUILDFLOW_NO_RESULT_CACHE=1`.

### Two-Module Layout & Layering

- **The root module must NEVER depend on go-finding or `linter/`.** The root is the zero-dependency foundation for 14+ downstream repos, and the go-finding CLI itself pins go-branded-id — importing it would invert the layering and create a module cycle.
- The linter does NOT import the root module (detection is purely syntactic), so the modules version independently.
- Linter tags REQUIRE the `linter/` prefix (`linter/vX.Y.Z`) — sub-module tagging.

### brandid-lint Behavior

- **BD001 detection is existence-based**: a brand counts as named when ANY `Name() string` method exists (literal, constant, or computed return). Never reintroduce return-value parsing into `HasName` — it makes Detect lie and Repair insert duplicate methods that do not compile.
- **`testdata/` directories are never scanned** (same convention as go vet/staticcheck), but a testdata dir passed as the scan ROOT is still visited — fixture tests rely on this. If fixture tests fail with findings-count mismatches, check `git log -p -- linter/testdata` for stray stub insertions.
- The root repo's own test brands carry `Name()` stubs deliberately so the linter runs clean on this repo; stripping them re-fires BD001 repo-wide (by design, not a bug).
- Suppression: `//brandid-lint:ignore(BD001) <reason>` on the brand declaration line or as the last line of the comment group directly above. BD002 polices broken/stale/duplicate/unplaced directives. **Top-level types only** — a directive on a function-local type is reported as unplaced ("suppresses nothing"); local types are never flagged, so delete such directives instead of repositioning them.
- BD003 (primitive-typed identifier rule) is opt-in via the RuleSet (`-rules`).
- End-to-end verification through BuildFlow: set `tool_options: {brandid-lint: {rules: BD003}}` in a repo's `.buildflow.yml` and run `buildflow -s brandid-lint --format finding` — validated 2026-10-12 through buildflow 8a34928 (finding + suggestion render, findings gate exit 69, negative control exit 0). Known gap: the summary shows `filesScanned: 0` until the provider adopts go-finding v1.15.0's `DetectResult` (TODO row).

### Brands That Deliberately Skip Name()

If a brand's `String()` output is used as a data key (storage key, stream name, routing key), do NOT add `Name()` — write the suppression directive instead, with a reason (a reasonless directive is itself a BD002 finding). Examples: go-cqrs-lite marker types (`StreamMarker`/`TimerMarker`), BerryBig test brands. Cyberdom has no brand types at all.

### Nix Gotchas

- Nix sandbox builds run with `HOME=/homeless-shelter` (read-only): any Go-based check derivation needs `GOCACHE=$TMPDIR/go-cache` (see `checks.build`).
- The flake `outputs` destructuring must include `...` — unnamed inputs make the pattern otherwise invalid.
- Supported systems are declared inline in `flake.nix` (no `nix-systems` input; nixpkgs dropped `x86_64-darwin`).

### Pre-Push Hook

`scripts/pre-push-dual-test.sh` runs the go-pins check, greps the v1 JSON imports, then tests both JSON modes. Reinstall after editing: `cp scripts/pre-push-dual-test.sh .git/hooks/pre-push && chmod +x .git/hooks/pre-push`.

## Ecosystem Context

Extracted from `go-composable-business-types/id`; 14+ downstream repos (InboxClean, CreditReformBilanzampel, ActaFlow, SEC, storbi, ChastityAPI, smart-configs, StopTube, universal-workflow, Zlota44, timesheets, complaints-mcp, cqrs-htmx, emeet-pixyd, plus go-output, cmdguard, and the go-finding CLI in the linter-stack tooling graph). Breaking changes must consider migration impact across all of them.

## Release Process

- CI creates a GitHub Release on semver tags; the body is extracted from the `## [X.Y.Z]` section of `CHANGELOG.md` — **cut the CHANGELOG section BEFORE tagging**, or the release has empty notes.
- Tags must be signed (SSH) and annotated: `git tag -a vX.Y.Z`, then `git push origin vX.Y.Z`.
- The linter releases separately: cut its `linter/CHANGELOG.md` section, tag `linter/vX.Y.Z`. `release.yml` matches `linter/v*` and creates the release from that section.
- CI hygiene rejects tracked binaries ANYWHERE in the repo (`linter/brandid-lint` is gitignored).
- The website changelog page is manually maintained — add the new version section and `nix run .#deploy` (from `website/`) as part of each release, or the live site goes stale.
- The BuildFlow pre-commit hook runs Go-mode checks including `doc-files-age-check`: README.md and TODO_LIST.md must be touched within 3 weeks of code changes.

## Website

Astro + Starlight docs site in `website/` (independent flake + pnpm), deployed to Firebase Hosting: `https://branded-id.lars.software` (project `lars-software`, hosting target `brandedid`).

- pnpm 11 reads `overrides` and `allowBuilds` from `pnpm-workspace.yaml`, NOT `package.json` — security version pins and build-script approvals live there. A missing `allowBuilds: esbuild: true` breaks `astro build`.
- `pnpm update <pkg>` only touches direct dependencies — transitive bumps (e.g. audit fixes) need `overrides` entries.
- Build/dev/deploy: `nix run .#build` / `.#dev` / `.#deploy` from `website/`.
