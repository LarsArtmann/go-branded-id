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
