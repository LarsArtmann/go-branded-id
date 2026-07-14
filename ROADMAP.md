# Roadmap

> Long-term direction and raw ideas not yet refined into actionable tasks.
> For short-term bounded work, see `TODO_LIST.md`.

## Theme 1: Ecosystem v0.3.1 Adoption

The library is stable at v0.3.1 but 14 downstream repos have source fixes applied
without the `go.mod` dependency bump. Full ecosystem migration is the path to v1.0.

- Strategy for batch-bumping 14 repos (automated PRs? migration script?)
- CI integration test: compile representative ecosystem repos against new versions
- Deprecate `go-composable-business-types/id` with a final redirect tag

## Theme 2: Stability & v1.0

Before tagging v1.0, the API surface should be frozen and audited:

- Evaluate whether `encoding/json/v2` is the right long-term choice (currently requires `GOEXPERIMENT=jsonv2`)
- Consider compile-time constraint for `Compare` (currently runtime `ErrNotOrdered`)
- API review: are there methods that should not exist? Are there missing methods users keep asking for?
- Stability guarantee: once v1.0 ships, breaking changes require v2.0

## Theme 3: Ecosystem Tooling

- Expand `cmd/namer` into a full migration toolkit (scan, fix, verify)
- Consider a `go generate` integration for brand type stubs
- Cross-repo compatibility test harness

## Non-Goals

- **ORM integration** — this is a type-safety library, not a data layer
- **ID generation** (UUID, ULID, snowflake) — that belongs in consumer code; this library wraps existing values
- **Database driver abstraction** — `database/sql` interfaces are sufficient
