# go-branded-id

[![Go Version](https://img.shields.io/github/go-mod/go-version/larsartmann/go-branded-id?logo=go&logoColor=white)](go.mod)
[![CI](https://github.com/larsartmann/go-branded-id/actions/workflows/go.yml/badge.svg)](https://github.com/larsartmann/go-branded-id/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/go-branded-id.svg)](https://pkg.go.dev/github.com/larsartmann/go-branded-id)
[![GitHub stars](https://img.shields.io/github/stars/larsartmann/go-branded-id?style=flat)](https://github.com/larsartmann/go-branded-id/stargazers)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-branded--id.lars.software-8b5cf6)](https://branded-id.lars.software)

**Branded, strongly-typed identifiers for Go.** Phantom types prevent mixing different entity IDs at compile time. Zero-allocation. Stdlib-only. Full serialization.

[Documentation](https://branded-id.lars.so) &middot; [Quick Start](https://branded-id.lars.so/getting-started/quick-start/) &middot; [API Reference](https://pkg.go.dev/github.com/larsartmann/go-branded-id)

## Why?

In Go, `string` and `int64` provide no compile-time safety against mixing IDs:

```go
func GetUser(id string) error { return nil }
func GetOrder(id string) error { return nil }

GetOrder(userID) // Compiles. Silent bug.
```

With branded IDs, the compiler catches it:

```go
type UserBrand struct{}
type OrderBrand struct{}

type UserID = id.ID[UserBrand, string]
type OrderID = id.ID[OrderBrand, string]

func GetUser(id UserID) error   { return nil }
func GetOrder(id OrderID) error { return nil }

GetOrder(userID) // COMPILE ERROR: cannot use UserID as OrderID
```

The brand types (`UserBrand`, `OrderBrand`) are empty structs that exist only as phantom type parameters. They add zero runtime cost — the underlying value is still just a `string` or `int64`.

## Installation

**Prerequisite:** Go 1.26+ with `GOEXPERIMENT=jsonv2` enabled (this library uses `encoding/json/v2`).

```bash
GOEXPERIMENT=jsonv2 go get github.com/larsartmann/go-branded-id
```

> Set `GOEXPERIMENT=jsonv2` for all Go commands. For convenience, add it to your `go.env` or shell profile.

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/larsartmann/go-branded-id"
)

type UserBrand struct{}

func (UserBrand) Name() string { return "User" }

type UserID = id.ID[UserBrand, string]

func main() {
    userID := id.NewID[UserBrand]("user-123")
    other  := id.NewID[UserBrand]("user-123")

    fmt.Println(userID)              // User:user-123  (named brand shows prefix)
    fmt.Println(userID.Equal(other)) // true
    fmt.Println(userID.Get())        // user-123      (raw value for programmatic use)

    var empty UserID
    fmt.Println(empty.IsZero())      // true

    // Provide a default for zero values
    fmt.Println(empty.Or(id.NewID[UserBrand]("unknown")).Get()) // unknown
}
```

## Features

| Feature                      | Description                                                                                 |
| ---------------------------- | ------------------------------------------------------------------------------------------- |
| **Compile-time type safety** | Phantom types prevent mixing `UserID` with `OrderID` at the compiler level                  |
| **Zero allocations**         | Core operations (`NewID`, `Get`, `Equal`, `Compare`, `IsZero`) allocate nothing             |
| **Stdlib-only**              | No third-party dependencies. Uses `encoding/json/v2` from the Go standard library           |
| **Full serialization**       | JSON, SQL, Text (XML/TOML), Binary, Gob — all implemented                                   |
| **Named brands**             | Optional `Name()` method enables `"User:abc123"` display strings and brand-aware validation |
| **Any comparable type**      | `ID[Brand, V comparable]` works with strings, ints, and any comparable type                 |
| **Zero value semantics**     | Zero value means "unset" — serializes to `null` in JSON, `nil` in SQL                       |
| **SQL scanner/valuer**       | `Scan` accepts all driver types; `Value` returns the correct type                           |

## Named Brand Types

Adding a `Name()` method to your brand type enables debug-visible IDs, brand-aware validation errors, and runtime introspection:

```go
type UserBrand struct{}
func (UserBrand) Name() string { return "User" }

userID := id.NewID[UserBrand]("abc123")
fmt.Println(userID)               // User:abc123
fmt.Printf("%#v\n", userID)       // id.User(abc123)

var empty ID[UserBrand, string]
fmt.Println(id.ValidateID(empty)) // id: invalid: User: empty

fmt.Println(id.BrandName[UserBrand]()) // User
```

IDs without `Name()` work exactly as before — `String()` returns just the value.

## Supported Value Types

The generic type is `ID[Brand, V comparable]` — any comparable type works as `V`.

Full serialization support (JSON, SQL, Text, Binary, Gob): `string`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`.

Other comparable types (structs, arrays, etc.) work for core operations (`Get`, `Equal`, `IsZero`, `Reset`) but lack specialized serialization.

## Serialization

All standard Go interfaces are implemented. Serialization always uses the raw value (no brand prefix):

```go
// JSON — zero values serialize to null
data, _ := json.Marshal(id.NewID[UserBrand]("user-123"))
fmt.Println(string(data)) // "user-123"

var empty UserID
data, _ = json.Marshal(empty)
fmt.Println(string(data)) // null

// SQL — works with database/sql directly
var userID UserID
row.Scan(&userID)
db.Exec("INSERT INTO users (id) VALUES (?)", userID)
```

## API Reference

| Method                           | Description                                 |
| -------------------------------- | ------------------------------------------- |
| `NewID[Brand](value)`            | Create a new ID (type inferred for strings) |
| `Get() V`                        | Returns the underlying value                |
| `IsZero() bool`                  | True if ID has its zero value               |
| `Equal(other ID) bool`           | True if IDs are equal                       |
| `Compare(other ID) (int, error)` | -1, 0, or 1 for ordered types               |
| `Or(default ID) ID`              | Returns self if not zero, otherwise default |
| `String() string`                | `"Brand:value"` if named, else value only   |
| `Ptr() *ID` / `FromPtr(*ID) ID`  | For optional fields                         |

**Brand utilities:** `BrandName[B]()`, `ValidateID(id)`, `ValidateIDWithValue(id, fn)`, `MustValidateID(id)`

Full API: [pkg.go.dev](https://pkg.go.dev/github.com/larsartmann/go-branded-id)

## Performance

Stdlib-only, allocation-conscious (benchmarked on Go 1.26.4):

| Operation           | Latency | Allocations |
| ------------------- | ------- | ----------- |
| `NewID`             | ~0.4 ns | 0           |
| `Get`               | ~1 ns   | 0           |
| `Equal`             | ~0.3 ns | 0           |
| `IsZero`            | ~1.4 ns | 0           |
| `String` (no brand) | ~5 ns   | 0           |
| `MarshalJSON`       | ~200 ns | 3           |

Core operations are zero-allocation. Named-brand `String()` requires one allocation for the `"Brand:value"` concatenation.

## Contributing

Contributions are welcome. Ensure all tests pass and lint is clean:

```bash
GOEXPERIMENT=jsonv2 go test ./... -race
GOEXPERIMENT=jsonv2 golangci-lint run
```

## License

MIT — Copyright (c) 2026 Lars Artmann. See [LICENSE](LICENSE).
