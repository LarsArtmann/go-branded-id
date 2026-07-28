# Deduplication Acceptance

Rationale for clone groups intentionally left in place after deduplication review.

## Accepted Clones

### id_json_v1.go ↔ id_json_v2.go (full-file duplication)

**Reason:** Build-tag constraint. The two files must remain separate because one
imports `encoding/json` (v1, default) and the other imports `encoding/json/v2`
(when `GOEXPERIMENT=jsonv2` is set). The logic is intentionally identical — both
provide `MarshalJSON`/`UnmarshalJSON` with the same semantics through their
respective json package. Merging them into one file would defeat the build-tag
architecture.

### id_binary.go:135-141 ↔ id_text.go:29-35 (empty-data early return)

**Reason:** 4-line idiomatic guard: `if len(data) == 0 { id.Reset(); return nil }`.
Extracting a helper like `resetIfEmpty(id, data) bool` would add 5+ lines and
obscure the early-return control flow — a net loss in readability for no
maintenance benefit. Both are standard Go empty-input guards.

## Decision: No Code Generation for the JSON File Pairs

**Considered:** generating `id_json_v{1,2}.go` and `json_helpers_v{1,2}_test.go`
from a single template to eliminate the duplication mechanically.

**Decision:** keep the files hand-written. The duplication is accepted.

**Rationale:**

1. Each pair is ~25-58 lines. A generator (templating tool, `go generate`,
   build-time script) adds a build step, a template file, and generated-file
   hygiene rules (`exhaustruct`/`nolint` markers, formatter exceptions) — more
   moving parts than the code it would replace.
2. `TestDualJSONContract_StructuralParity` in `id_json_contract_test.go` now
   **enforces** that each v1/v2 pair is byte-identical after normalizing the two
   intentional differences (build constraint + json import path). Any edit that
   diverges the pair fails CI in both modes. The duplication is therefore
   _guarded_, not _unguarded_.
3. `TestDualJSONContract_Imports` additionally locks the import split, catching
   the goimports corruption hazard that previously broke the default build mode.

**Revisit if:** a third json variant is ever needed (e.g. a v3), or the per-file
logic grows past ~100 lines. At two small files, manual + parity-test is the
lower-complexity choice.
