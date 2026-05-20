# Domain Language

A **Unified Language** for `.` — shared across Customer, Product Owner, Developer, and AI.
Inspired by Domain-Driven Design (DDD) Ubiquitous Language.

Every term below should mean the **same thing** to everyone who reads it.
If a word means something different to a developer than to a customer, define it here.

## Glossary

| Term         | Definition               | Context                        |
| ------------ | ------------------------ | ------------------------------ |
| Brand         | A phantom type that distinguishes different ID types at compile time | `type UserBrand struct{}` |
| Brand Name    | A human-readable label for a brand, provided via the `BrandNamer` interface | `"User"` for `UserBrand` |
| Named Brand   | A brand that implements `Name() string` — enables `"Brand:value"` in String() | Debug-visible IDs |
| Unnamed Brand | A brand without `Name()` — String() returns just the value | Backward compatible |
| Branded ID    | An `ID[B, V]` value — strongly typed, zero-cost identifier | `ID[UserBrand, string]` |
| Value         | The underlying raw data of an ID, accessed via `.Get()` | The actual ULID, string, or int |
| Validation    | Checking an ID is not zero, optionally with custom rules | `ValidateID`, `ValidateIDWithValue` |

## Entities

Objects with identity and lifecycle (e.g., User, Order, Account).

<!-- Add your entities here:
| Term | Definition | Context |
|------|-----------|---------|
| User | A person who interacts with the system | Customer-facing |
-->

## Value Objects

Immutable objects defined by attributes (e.g., Email, Money, Address).

<!-- Add your value objects here:
| Term | Definition | Context |
|------|-----------|---------|
| Email | A validated email address | Unique identifier for users |
-->

## Events

Things that happen in the domain (e.g., UserRegistered, PaymentProcessed).

<!-- Add your events here:
| Term | Definition | Context |
|------|-----------|---------|
| UserRegistered | A new user completed signup | Triggers welcome email |
-->

## Commands

Actions the system can perform (e.g., CreateUser, ProcessPayment).

<!-- Add your commands here:
| Term | Definition | Context |
|------|-----------|---------|
| CreateUser | Registers a new user account | Admin action |
-->

## Bounded Contexts

Subsystems with distinct vocabulary (e.g., Billing vs. Shipping).

<!-- Define contexts where the same word means different things:
| Context | Description |
|---------|------------|
| Billing | Handles payments and invoices |
-->

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
