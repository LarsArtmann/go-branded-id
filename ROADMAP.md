# Roadmap

> Long-term direction and raw ideas not yet refined into actionable tasks.
> For short-term bounded work, see `TODO_LIST.md`.

## Theme 1: Ecosystem v0.5.0 Adoption

The library is stable at v0.5.0 but 14 downstream repos have source fixes applied
without the `go.mod` dependency bump. Full ecosystem migration is the path to v1.0.

- Strategy for batch-bumping 14 repos (automated PRs? migration script?)
- CI integration test: compile representative ecosystem repos against new versions
- Run `cmd/namer` codemod against downstream repos to find brands missing `Name()`
- Deprecate `go-composable-business-types/id` with a final redirect tag

## Theme 2: Stability & v1.0

Before tagging v1.0, the API surface should be frozen and audited:

- ~~Evaluate whether `encoding/json/v2` is the right long-term choice~~ Done: dual-supports both v1 and v2 via build tags
- **Error taxonomy completion**: all 7 sentinel errors need `errors.Is` test coverage (5 of 7 currently untested); remaining `fmt.Errorf` calls without sentinel wrapping should be audited
- Consider compile-time constraint for `Compare` (currently runtime `ErrNotOrdered` via type switch — a `constraints.Ordered` generic could make it a compile error)
- API review: are there methods that should not exist? Are there missing methods users keep asking for?
- Stability guarantee: once v1.0 ships, breaking changes require v2.0
- Consider whether `ErrNotOrdered` message change in v0.5.0 (dropped "(int, uint, or string)") warrants documentation as a behavioral change

## Theme 3: Ecosystem Tooling

- Expand `cmd/namer` into a full migration toolkit (scan, fix, verify)
- Add `-diff` mode to namer showing what would change per file
- Add JSON output mode to namer for CI/editor integration
- Consider a `go generate` integration for brand type stubs
- Cross-repo compatibility test harness

## Theme 4: Advanced Serialization

- Explore `encoding/json/v2` jsontext API for streaming marshal performance
- Consider `NullID[B, V]` type for nullable SQL support (like `sql.NullString`)
- Add `msgpack` or protobuf serialization support
- Cross-language binary compatibility tests (marshal in Go, verify in Python/TS)
- Benchmark and document the little-endian binary format spec

## Non-Goals

- **ORM integration** — this is a type-safety library, not a data layer
- **ID generation** (UUID, ULID, snowflake) — that belongs in consumer code; this library wraps existing values
- **Database driver abstraction** — `database/sql` interfaces are sufficient
