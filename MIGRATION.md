# Migration Guide: id/ → go-branded-id

## What Changed

The `id/` package was extracted into a standalone library: [`github.com/larsartmann/go-branded-id`](https://github.com/larsartmann/go-branded-id).

The `id/` directory no longer exists in `go-composable-business-types` — migration is required.

## Prerequisites

- **Go 1.26+** (see [`go.mod`](go.mod))
- **`GOEXPERIMENT=jsonv2`** — this library uses `encoding/json/v2`. Set it for all Go commands: `export GOEXPERIMENT=jsonv2`

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

Set `GOEXPERIMENT=jsonv2`:

```bash
export GOEXPERIMENT=jsonv2
go build ./...
```

This is required because the library uses `encoding/json/v2` (Go's new JSON v2 package).

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
