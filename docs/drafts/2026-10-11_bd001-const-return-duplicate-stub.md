# brandid-lint: a `Name()` returning a package-level const is treated as "missing", so BD001 fires on a compliant type and repair inserts a duplicate method

## Problem

`parseNameReturnValue` (`linter/brands.go:228` at `baa989b`) only accepts a
string literal — `ret.Results[0].(*ast.BasicLit)` with
`lit.Kind == token.STRING`. A `Name()` method whose body is a single
`return <constIdent>` (an `*ast.Ident`) returns `""`, so `hasNameMethod`
(`linter/brands.go:183`) treats the brand as if it had no `Name()` at all:

- **Detect** reports BD001 on a type that visibly satisfies the contract.
- **Repair** then inserts its stub next to the existing method — two
  `Name()` methods on one type, which does not compile.

The repair doc comment claims idempotence because "brands that already have
Name() produce no finding" (`linter/repair.go:38-40`). That holds only for
literal-returning methods; a const-returning method still produces a
finding, so the idempotence argument does not cover it.

## Incident (the cost of this, live)

go-sse, 2026-10-03: its brand type returned a constant from a shared
const block. Every `buildflow --fix` re-inserted the stub next to the
existing `Name()` — five duplicate-method commits in one day
(`8b4c859`, `75476e6`, `6a62140`, `346fd2c` are the removals), each one
reddening go-sse's Lint/Vet/Nix jobs until the root cause was found. The
workaround was to inline the literal into the return statement
(`return "SSEEvent"`), which satisfies `parseNameReturnValue` and stops
the loop — the const had to be given up.

## Proposed fix (two layers)

1. **Resolve const identifiers** in `parseNameReturnValue`: when the
   return value is an `*ast.Ident`, look it up in the file's `const`
   declarations (and accept a `token.STRING` value). A package-level
   `const sseEventName = "SSEEvent"` + `return sseEventName` is exactly as
   static as a literal.
2. **Defense in depth in repair**: `repairFile` should skip a brand that
   already declares ANY `Name() string` method, regardless of what the
   return value parses to. Detect may still flag the non-literal form as a
   style note if you want literal-only, but repair inserting a duplicate
   method is never the right action.

Layer 2 alone stops the compile-error loop; layer 1 removes the false
BD001 on types that are already correct.

## Reproduction

```go
package foo

const brandName = "widget"

type WidgetID struct{ id.ID } // or whatever brand shape you lint

func (WidgetID) Name() string { return brandName } // BD001 fires; repair duplicates this method
```

`brandid-lint` v0.7.0 (`baa989b`), Go 1.27. Found via BuildFlow's
provider integration (BuildFlow imports
`github.com/larsartmann/go-branded-id/linter/provider`).

💘 Generated with Crush
