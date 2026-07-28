# Migration Guide: id/ → go-branded-id

## What Changed

The `id/` package was extracted into a standalone library: [`github.com/larsartmann/go-branded-id`](https://github.com/larsartmann/go-branded-id).

The `id/` directory no longer exists in `go-composable-business-types` — migration is required.

## Prerequisites

- **Go 1.26+** (see [`go.mod`](go.mod))
- No special environment variables required. The library uses `encoding/json` by default.
  Set `GOEXPERIMENT=jsonv2` only if you want to use `encoding/json/v2` internally.

## Migration Steps

**1. Add the dependency**

```bash
go get github.com/larsartmann/go-branded-id
```

**2. Replace the import path**

```diff
- "github.com/larsartmann/go-composable-business-types/id"
+ "github.com/larsartmann/go-branded-id"
```

**3. Remove the old dependency**

```bash
go mod tidy
```

> **Note:** If you use other packages from `go-composable-business-types`, the old dependency remains in `go.mod` — that's expected. `go mod tidy` only removes unused imports.

**4. Verify**

```bash
go build ./...
go test ./...
```

If both pass, the migration is complete.

## Nothing Else Changes

All types, functions, and methods are identical — only the import path changed:

```go
id.NewID[UserBrand]("user-123")  // same
id.ID[UserBrand, string]         // same
id.ErrNotOrdered                 // same
```

## Bonus Features

After migrating, you also gain access to new APIs not present in the original `id/` package:

- `id.Ptr()` — returns `*ID[B, V]` for optional fields
- `id.FromPtr()` — dereferences a pointer, returns zero value if nil
- `Format` method — implements `fmt.Formatter` (`%s`, `%d`, `%v`, `%#v`, `%q`)

## v0.3.0: Brand-Aware String()

`String()` now includes the brand name prefix for brands that implement `BrandNamer`:

```go
// Before v0.3.0 (all brands)
fmt.Println(userID) // "abc123"

// After v0.3.0 (named brands)
fmt.Println(userID) // "User:abc123"

// After v0.3.0 (unnamed brands — unchanged)
fmt.Println(orderID) // "abc123"
```

### Is this a breaking change?

**No**, for brands without a `Name()` method, `String()` returns the same value as before.

For brands with `Name()` (ActaFlow, CreditReformBilanzampel, InboxClean), `String()` now includes the prefix. If you parse `String()` output, use `Get()` instead:

```go
// Before: parsed String() output
value := id.String()

// After: use Get() for the raw value
value := id.Get() // always returns "abc123"
```

Serialization (JSON, SQL, Text, Binary, Gob) is unaffected — always uses the raw value.

### New APIs

- `BrandNamer` interface — add `Name() string` to your brand types
- `BrandName[B]()` — returns brand name for logging
- `ValidateID(id)` — returns brand-aware error for zero IDs
- `ValidateIDWithValue(id, fn)` — validates ID and value
- `GoString()` — returns `id.BrandName(value)` for debugging

## Troubleshooting

### `go get` fails with Go version error

Ensure your project's `go.mod` has `go 1.26` or later. This library uses modern Go generics features.

### `go get` or `go build` fails with "build constraints exclude all Go files in encoding/json/v2"

This should no longer happen — the library uses `encoding/json` (v1) by default. If you previously set `GOEXPERIMENT=jsonv2`, remove it:

```bash
unset GOEXPERIMENT
go build ./...
```

If you **want** v2 semantics, set `GOEXPERIMENT=jsonv2` and the library will automatically use `encoding/json/v2` internally via build tags.

### `replace` directive pointing at the old package

If your `go.mod` has a `replace` directive for the old `id/` path, remove it:

```diff
- replace github.com/larsartmann/go-branded-id => ../go-branded-id
```

Then run `go mod tidy`.

### Import still resolves to the old path

Run `go mod tidy` and ensure no files still import the old path:

```bash
grep -r "go-composable-business-types/id" --include="*.go" .
```

## v0.4.0: Dual JSON v1/v2 Support

v0.4.0 introduces a **dual-mode JSON architecture**. The library now works with
both `encoding/json` (v1, default) and `encoding/json/v2` (experimental) using
build tags. **No special environment variables are required.**

### What changed for consumers

**Nothing.** If you were on v0.3.x, upgrading to v0.4.0 is a drop-in change:

```bash
go get github.com/larsartmann/go-branded-id@v0.4.0
go mod tidy
```

The public API is identical. JSON, SQL, Text, Binary, and Gob serialization
all produce byte-identical output to v0.3.x.

### Why dual-mode?

v0.3.1 hard-required `GOEXPERIMENT=jsonv2`, which broke every consumer that
didn't set the flag. v0.4.0 eliminates this requirement entirely:

| Mode         | When                               | `GOEXPERIMENT` needed? |
| ------------ | ---------------------------------- | ---------------------- |
| v1 (default) | Always                             | No                     |
| v2           | When you set `GOEXPERIMENT=jsonv2` | Yes (opt-in)           |

In Go 1.27+, json/v2 becomes the default and the v2 code path is used
automatically — no action needed.

### If you want v2 semantics

Set the environment variable in your dev shell, CI, or `.envrc`:

```bash
export GOEXPERIMENT=jsonv2
```

The library detects this via build tags and uses `encoding/json/v2` internally.
This is entirely optional — v1 mode is fully functional.

### Verification

The library includes contract tests (`id_json_contract_test.go`) that verify:

1. **Import split** — v1 files import `encoding/json`, v2 files import `encoding/json/v2`
2. **Build tags** — correct `//go:build` constraints on each file pair
3. **Structural parity** — the v1 and v2 files are byte-identical after normalization
4. **Byte equivalence** — `MarshalJSON` produces identical bytes in both modes

CI runs the full test suite in **both** modes (matrix strategy), so any
divergence between v1 and v2 behavior is caught immediately.

### Troubleshooting

#### Build error: "build constraints exclude all Go files in encoding/json/v2"

This means you have `GOEXPERIMENT=jsonv2` set but your Go toolchain doesn't
support it (Go < 1.24), OR you don't have it set but something is importing
`encoding/json/v2` directly. Since v0.4.0, the library never requires the flag.
Remove it:

```bash
unset GOEXPERIMENT
go build ./...
```

#### goimports corrupted the import

If `goimports` rewrote `"encoding/json"` to `"encoding/json/v2"` in a v1-tagged
file, the contract test will catch it. Run:

```bash
go test -run TestDualJSONContract ./...
```

## Bumping Downstream Repositories

The library has 14 downstream repos in the ecosystem. After a new release, each
needs to be bumped. The process is mechanical and identical for each repo.

### Affected Repos

InboxClean, CreditReformBilanzampel, ActaFlow, SEC, storbi, ChastityAPI,
smart-configs, StopTube, universal-workflow, Zlota44, timesheets,
complaints-mcp, cqrs-htmx, emeet-pixyd

### Bump Procedure (per repo)

```bash
# 1. Bump the dependency
go get github.com/larsartmann/go-branded-id@latest
go mod tidy

# 2. Build and test (both json modes)
go build ./...
go test ./... -race -count=1

# 3. Run the namer tool to check for brands missing Name()
go run github.com/larsartmann/go-branded-id/cmd/namer@latest ./...

# 4. If namer flags brands that SHOULD have Name(), add the method:
#    func (YourBrand) Name() string { return "YourBrandName" }

# 5. Verify String() output hasn't changed for any named brands
#    (only relevant if you parse String() output — use Get() instead)
```

### Brands That Deliberately Skip Name()

Not all brands should implement `Name()`. If a brand's `String()` output is
used as a data key (storage key, stream name, routing key), do NOT add
`Name()`. The namer tool will flag these as false positives.

Known examples:

- **go-cqrs-lite marker types** — `String()` output is used directly as
  storage/stream keys
- **BerryBig** — test brands only
- **Cyberdom** — no brand types at all
