# Domain Language

A **Ubiquitous Language** for `go-branded-id` — shared across maintainers, downstream
consumers, and AI agents. Inspired by Domain-Driven Design (DDD).

This is a utility library, not a business domain, so the glossary below is the
authoritative part. The classic DDD sections (Entities, Value Objects, Events,
Commands, Bounded Contexts) do not apply and are intentionally omitted.

Every term below should mean the **same thing** to everyone who reads it.

## Glossary

| Term          | Definition                                                                    | Context                             |
| ------------- | ----------------------------------------------------------------------------- | ----------------------------------- |
| Brand         | A phantom type that distinguishes different ID types at compile time          | `type UserBrand struct{}`           |
| Brand Name    | A human-readable label for a brand, provided via the `BrandNamer` interface   | `"User"` for `UserBrand`            |
| Named Brand   | A brand that implements `Name() string` — enables `"Brand:value"` in String() | Debug-visible IDs                   |
| Unnamed Brand | A brand without `Name()` — String() returns just the value                    | Backward compatible                 |
| Branded ID    | An `ID[B, V]` value — strongly typed, zero-cost identifier                    | `ID[UserBrand, string]`             |
| Value         | The underlying raw data of an ID, accessed via `.Get()`                       | The actual ULID, string, or int     |
| Validation    | Checking an ID is not zero, optionally with custom rules                      | `ValidateID`, `ValidateIDWithValue` |
| Phantom Type  | A type parameter used only at compile time (carries no runtime data)          | `B` in `ID[B, V]`                   |
| Zero Value    | The unset state of an ID; serializes to `null`/`nil` across all formats       | `var empty UserID`                  |

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
