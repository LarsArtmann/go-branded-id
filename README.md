# go-branded-id

[![Go Version](https://img.shields.io/github/go-mod/go-version/larsartmann/go-branded-id)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

A Go library providing branded, strongly-typed identifiers that prevent mixing different entity IDs at compile time using phantom types.

## Why?

In Go, regular types like `string` or `int64` provide no compile-time safety:

```go
package main

func GetUser(id string) error { return nil }
func GetOrder(id string) error { return nil }

func main() {
	GetOrder(userID) // Compiles! Runtime bug.
}

var userID = "user-123"
var orderID = "order-456"
```

With this package, the compiler catches these errors:

```go
package main

type UserBrand struct{}
type OrderBrand struct{}

type UserID = id.ID[UserBrand, string]
type OrderID = id.ID[OrderBrand, string]

func GetUser(id UserID) error { return nil }
func GetOrder(id OrderID) error { return nil }

func main() {
	GetOrder(userID) // Compile error: type mismatch
}

var userID = id.NewID[UserBrand]("user-123")
var orderID = id.NewID[OrderBrand]("order-456")
```

## Installation

**Prerequisite:** Go 1.26+ with `GOEXPERIMENT=jsonv2` enabled (this library uses `encoding/json/v2`).

```bash
GOEXPERIMENT=jsonv2 go get github.com/larsartmann/go-branded-id
```

You must set `GOEXPERIMENT=jsonv2` for all Go commands (`build`, `test`, `run`, etc.).
For convenience, add it to your environment or `go.env`.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/larsartmann/go-branded-id"
)

type UserBrand struct{}

func (UserBrand) Name() string { return "User" }

type OrderBrand struct{}

type UserID = id.ID[UserBrand, string]
type OrderID = id.ID[OrderBrand, string]

func main() {
    userID := id.NewID[UserBrand]("user-123")
    orderID := id.NewID[OrderBrand]("order-456")

    fmt.Println(userID)       // User:user-123  (named brand shows prefix)
    fmt.Println(orderID)      // order-456      (unnamed brand, value only)
    fmt.Printf("%#v\n", userID) // id.User(user-123) — display only, not valid Go syntax

    // Type-safe comparison
    otherUserID := id.NewID[UserBrand]("user-123")
    fmt.Println(userID.Equal(otherUserID))  // true

    // Zero value check
    var emptyUserID UserID
    fmt.Println(emptyUserID.IsZero())  // true

    // Default value with Or
    defaultID := id.NewID[UserBrand]("unknown")
    fmt.Println(emptyUserID.Or(defaultID).Get())  // unknown
}
```

## Named Brand Types

Adding a `Name()` method to your brand type enables:

- **Debug-visible IDs**: `String()` returns `"User:abc123"` instead of just `"abc123"`
- **Brand-aware validation**: `ValidateID` errors include the brand name
- **Runtime introspection**: `BrandName[T]()` for logging and error messages

```go
type UserBrand struct{}

func (UserBrand) Name() string { return "User" }

// String output includes the brand name
userID := id.NewID[UserBrand]("abc123")
fmt.Println(userID) // User:abc123

// Debug format shows full type info
fmt.Printf("%#v\n", userID) // id.User(abc123) — display format, not valid Go syntax

// Validation errors identify the brand
var empty ID[UserBrand, string]
err := id.ValidateID(empty)
fmt.Println(err) // id: invalid: User: empty

// Introspection for logging
fmt.Println(id.BrandName[UserBrand]()) // User
```

IDs without `Name()` work exactly as before — `String()` returns just the value.

## Supported Value Types

The generic type is `ID[Brand, V comparable]` — any comparable type works as `V`.

Full serialization support (JSON, SQL, Text, Binary, Gob): `string`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`.

Other comparable types (structs, arrays, etc.) work for the core operations (`Get`, `Equal`, `IsZero`, `Reset`) but lack specialized serialization.

## Serialization

The package implements all standard Go interfaces for seamless serialization:

- **JSON**: `json.Marshaler` / `json.Unmarshaler` (zero values → `null`)
- **SQL**: `sql.Scanner` / `driver.Valuer` (string, all int/uint types, nil)
- **Binary**: `encoding.BinaryMarshaler` / `BinaryUnmarshaler`
- **Text**: `encoding.TextMarshaler` / `TextUnmarshaler` (XML, TOML)
- **Gob**: `gob.GobEncoder` / `gob.GobDecoder`

Serialization always uses the raw value (no brand prefix), so `"user-123"` not `"User:user-123"`.

### JSON Example

```go
id := id.NewID[UserBrand]("user-123")
data, _ := json.Marshal(id)
fmt.Println(string(data))  // "user-123"

// Zero values serialize to null
var empty UserID
data, _ = json.Marshal(empty)
fmt.Println(string(data))  // null
```

### SQL Example

```go
// Scan from database
var userID UserID
row.Scan(&userID)

// Save to database
_, err := db.Exec("INSERT INTO users (id, name) VALUES (?, ?)", userID, "John")
```

## Comparison & Sorting

```go
id1 := id.NewID[UserBrand, int64](100)
id2 := id.NewID[UserBrand, int64](200)

id1.Compare(id2)  // -1 (less)
id2.Compare(id1)  //  1 (greater)

sort.Slice(ids, func(i, j int) bool {
    cmp, _ := ids[i].Compare(ids[j])
    return cmp < 0
})
```

## API Reference

| Method                            | Description                                 |
| --------------------------------- | ------------------------------------------- |
| `NewID[Brand](value)`             | Create a new ID (type inferred for strings) |
| `NewID[Brand, V](value)`          | Create a new ID with explicit type          |
| `Get() V`                         | Returns the underlying value                |
| `IsZero() bool`                   | True if ID has its zero value               |
| `Reset()`                         | Sets ID to its zero value                   |
| `Equal(other ID) bool`            | True if IDs are equal                       |
| `Compare(other ID) (int, error)`  | -1, 0, or 1 for less/equal/greater          |
| `Or(default ID) ID`               | Returns self if not zero, otherwise default |
| `String() string`                 | `"Brand:value"` if named, else value only   |
| `GoString() string`               | `id.Brand(value)` for debugging             |
| `Format(fmt.State, rune)`         | Custom formatting (%s, %d, %v, %#v, %q)     |
| `MarshalJSON() ([]byte, error)`   | JSON serialization                          |
| `UnmarshalJSON([]byte) error`     | JSON deserialization                        |
| `MarshalText() ([]byte, error)`   | Text serialization (XML/TOML)               |
| `UnmarshalText([]byte) error`     | Text deserialization                        |
| `MarshalBinary() ([]byte, error)` | Binary serialization                        |
| `UnmarshalBinary([]byte) error`   | Binary deserialization                      |
| `GobEncode() ([]byte, error)`     | Gob encoding                                |
| `GobDecode([]byte) error`         | Gob decoding                                |
| `Scan(any) error`                 | SQL scan                                    |
| `Value() (driver.Value, error)`   | SQL value                                   |
| `Ptr() *ID[B, V]`                 | Returns pointer to ID (for optional fields) |
| `FromPtr(*ID[B, V]) ID[B, V]`     | Dereferences pointer, returns zero if nil   |

### Brand Utilities

| Function                          | Description                                    |
| --------------------------------- | ---------------------------------------------- |
| `BrandName[B]() string`           | Returns brand name (or type name if no Name()) |
| `ValidateID(id ID[B, V]) error`   | Returns error if ID is zero (brand-aware)      |
| `ValidateIDWithValue(id, fn) err` | Validates ID and optionally validates value    |
| `MustValidateID(id)`              | Panics if ID is zero                           |

## Performance

Stdlib-only, allocation-conscious implementation (benchmarked on Go 1.26.4 with `GOEXPERIMENT=jsonv2`):

| Operation           | Typical Latency | Allocations |
| ------------------- | --------------- | ----------- |
| `NewID`             | ~0.4 ns/op      | 0           |
| `Get`               | ~1 ns/op        | 0           |
| `Equal`             | ~0.3 ns/op      | 0           |
| `Compare`           | ~3 ns/op        | 0           |
| `IsZero`            | ~1.4 ns/op      | 0           |
| `String` (no brand) | ~5 ns/op        | 0           |
| `String` (named)    | ~30 ns/op       | 1           |
| `MarshalJSON`       | ~200 ns/op      | 3           |
| `MarshalBinary`     | ~22 ns/op       | 1           |
| `Scan` (string)     | ~44 ns/op       | 1           |

Core operations (`NewID`, `Get`, `Equal`, `Compare`, `IsZero`) and unbranded `String()` are zero-allocation. Named-brand `String()` requires one allocation for the `"Brand:value"` concatenation. `Format()` (via `fmt.Sprintf`) benefits from direct `io.Writer` writes introduced in v0.3.1.

> Numbers from `go test -bench=. -benchmem` on Go 1.26.4 linux/amd64. Rerun with `GOEXPERIMENT=jsonv2 go test -bench=. -benchmem ./...`.

## Contributing

Contributions are welcome. Please ensure all tests pass (`GOEXPERIMENT=jsonv2 go test ./... -race`) and lint is clean (`GOEXPERIMENT=jsonv2 golangci-lint run`) before submitting changes.

## License

MIT — Copyright (c) 2026 Lars Artmann. See [LICENSE](LICENSE).
