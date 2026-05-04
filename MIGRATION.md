# Migration Guide: id/ → go-branded-id

## What Changed

The `id/` package was extracted into a standalone library: [`github.com/larsartmann/go-branded-id`](https://github.com/larsartmann/go-branded-id).

The `id/` directory no longer exists in `go-composable-business-types` — migration is required.

## Prerequisites

- **Go 1.26.2 or later** (see [`go.mod`](go.mod))

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

## Troubleshooting

### `go get` fails with Go version error

Ensure your project's `go.mod` has `go 1.26.2` or later. This library uses modern Go generics features.

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
