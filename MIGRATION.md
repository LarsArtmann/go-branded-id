# Migration Guide: id/ → go-branded-id

## What Changed

The `id/` package was extracted into a standalone library: [`github.com/larsartmann/go-branded-id`](https://github.com/larsartmann/go-branded-id).

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

**3. Remove the old dependency (if no other packages are needed)**

```bash
go mod tidy
```

## Nothing Else Changes

All types, functions, and methods are identical — only the import path changed:

```go
id.NewID[UserBrand]("user-123")  // same
id.ID[UserBrand, string]         // same
id.ErrNotOrdered                 // same
```
