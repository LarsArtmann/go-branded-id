# go-branded-id

A Go library providing branded, strongly-typed identifiers that prevent mixing different entity IDs at compile time using phantom types.

## Why?

In Go, regular types like `string` or `int64` provide no compile-time safety:

```go
func GetUser(id string) error { ... }
func GetOrder(id string) error { ... }

userID := "user-123"
orderID := "order-456"

GetOrder(userID)  // Compiles! Runtime bug.
```

With this package, the compiler catches these errors:

```go
type UserBrand struct{}
type OrderBrand struct{}

type UserID = id.ID[UserBrand, string]
type OrderID = id.ID[OrderBrand, string]

func GetUser(id UserID) error { ... }
func GetOrder(id OrderID) error { ... }

userID := id.NewID[UserBrand]("user-123")
orderID := id.NewID[OrderBrand]("order-456")

GetOrder(userID)  // Compile error: type mismatch
```

## Installation

```bash
go get github.com/larsartmann/go-branded-id
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/larsartmann/go-branded-id"
)

type UserBrand struct{}
type OrderBrand struct{}

type UserID = id.ID[UserBrand, string]
type OrderID = id.ID[OrderBrand, string]

func main() {
    userID := id.NewID[UserBrand]("user-123")
    orderID := id.NewID[OrderBrand]("order-456")

    fmt.Println(userID.Get())  // user-123

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

## Best Practice: Named Brand Types

For better debugging and error messages, add a `Name()` method to your brand types:

```go
type UserBrand struct{}

func (UserBrand) Name() string { return "User" }

type UserID = id.ID[UserBrand, string]
```

This enables runtime introspection for logging, error messages, and debugging:

```go
func ValidateID[B interface{ Name() string }, V comparable](id id.ID[B, V]) error {
    if id.IsZero() {
        var brand B
        return fmt.Errorf("invalid %s ID: empty", brand.Name())
    }
    return nil
}
// Output: "invalid User ID: empty"
```

**Note:** This is optional. The phantom type pattern works perfectly without any methods on the brand type.

## Supported Value Types

| Type     | Alias               |
| -------- | ------------------- |
| `string` | `ID[Brand, string]` |
| `int`    | `ID[Brand, int]`    |
| `int8`   | `ID[Brand, int8]`   |
| `int16`  | `ID[Brand, int16]`  |
| `int32`  | `ID[Brand, int32]`  |
| `int64`  | `ID[Brand, int64]`  |
| `uint`   | `ID[Brand, uint]`   |
| `uint8`  | `ID[Brand, uint8]`  |
| `uint16` | `ID[Brand, uint16]` |
| `uint32` | `ID[Brand, uint32]` |
| `uint64` | `ID[Brand, uint64]` |

## Serialization

The package implements all standard Go interfaces for seamless serialization:

- **JSON**: `json.Marshaler` / `json.Unmarshaler` (zero values → `null`)
- **SQL**: `sql.Scanner` / `driver.Valuer` (string, int64, int32, uint64, []byte, nil)
- **Binary**: `encoding.BinaryMarshaler` / `BinaryUnmarshaler`
- **Text**: `encoding.TextMarshaler` / `TextUnmarshaler` (XML, TOML)
- **Gob**: `gob.GobEncoder` / `gob.GobDecoder`

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
| `String() string`                 | String representation                       |
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

## Performance

Zero-allocation, stdlib-only implementation:

- `NewID`: ~1-2 ns/op
- `Get`: ~1 ns/op
- `MarshalJSON`: ~50-100 ns/op
- `Scan` (string): ~30-50 ns/op

## License

Proprietary — Copyright (c) 2026 Lars Artmann. All rights reserved.
