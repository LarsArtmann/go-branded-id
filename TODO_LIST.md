# TODO List — go-branded-id

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.
> Completed work is removed from this list and recorded in `CHANGELOG.md`.

## Status legend

| Status           | Meaning                                                 |
| ---------------- | ------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                               |
| 🟡 `IN_PROGRESS` | Actively being worked on.                               |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed. |

## High Impact

| Task                                                                              | Status | Impact | Evidence                                                                                                                                                                                                                   |
| --------------------------------------------------------------------------------- | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Refresh `website/pnpm-lock.yaml` (`pnpm install` in `website/`)                   | 🔴     | High   | The website migrated from npm to pnpm (`pnpm-lock.yaml`, commit `f1f2f42`). `package.json` holds the overrides (`astro` `^7.3.3`, `fast-uri` `^3.1.4`, `brace-expansion` `5.0.6`) but the lockfile still resolves vulnerable versions — 10 open Dependabot alerts on the default branch (astro AVIF RCE + auth bypass, 4× fast-uri SSRF/host-confusion, sharp, 2× svgo, js-yaml), verified 2026-09-17 via the GitHub Dependabot API. |
| Build & verify the website (`pnpm run build` in `website/`)                       | 🔴     | High   | `guides/error-handling.mdx` and `guides/namer-tool.mdx` were added but never compiled; sidebar links (`astro.config.mjs`) and frontmatter are unverified.                                                                 |
| Add a CI/release guard that rejects a tracked compiled binary at repo root        | 🔴     | High   | Prevents recurrence of the v0.5.0 incident where a tracked `namer` binary inflated release source archives ~10x. Currently relies on `.gitignore` only — no workflow checks for build artifacts.                           |
| Decide the next release version (v0.5.2 additive vs v0.6.0) and date `[Unreleased]` | 🔴   | High   | `CHANGELOG.md` `[Unreleased]` has no version header. The `ErrNotOrdered` message restoration is a behavioral change for message-parsing consumers, so the semver call gates the release and all 14 downstream `go.mod` bumps. Harvested from `docs/status/2026-07-28_23-22` (Q2, f.3). |

## Medium Impact

| Task                                                                            | Status | Impact | Evidence                                                                                                                                                       |
| ------------------------------------------------------------------------------- | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Add `ErrMarshal`/`ErrUnmarshal` test coverage for the remaining delegate paths   | 🔴     | Med    | `id_errors_test.go` proves `ErrMarshal` only via `MarshalBinary` (one subtest). Untested delegates: JSON marshaler, SQL `Value()` TextMarshaler, `UnmarshalBinary` custom-type delegate, `unmarshalTextDefault` TextUnmarshaler path. |
| Guard the JSON v1 imports at build time (the contract test cannot catch this)    | 🔴     | Med    | `TestDualJSONContract_Imports` runs inside the package: when v1 imports are corrupted, the package fails to compile and the test never executes (`docs/status/2026-08-02_16-11`, e1). Add a grep-based import check to the pre-push hook (before `go test`) and/or CI, and make the hook report both modes instead of dying on v1 (`set -e`). |
| Verify `nix flake check --all-systems` passes locally                            | 🔴     | Med    | The CI `flake-check` job runs it (`go.yml`), but it has never been verified locally. A local plain `nix flake check` failed on 2026-09-17 due to the `go.mod` auto-bump to 1.27.1 (since repaired to 1.26) — re-verify after resolving the toolchain-pin policy. |

## Ecosystem tracking

| Task                                         | Status       | Impact | Evidence                                                                                                                                                                                                                   |
| -------------------------------------------- | ------------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Bump 14 downstream ecosystem repos to the next release | 🔵 `BLOCKED` | Med    | Source fixes from the v0.3.x cycle (added `Name()` methods, `.String()` → `.Get()`) are applied and pushed to all repos. The `go.mod` dependency bump is not yet done — requires per-repo access to clone, bump, test, PR. Blocked on the version decision above. |

---

## Ecosystem detail

Downstream repos: InboxClean, CreditReformBilanzampel, ActaFlow, SEC, storbi,
ChastityAPI, smart-configs, StopTube, universal-workflow, Zlota44, timesheets,
complaints-mcp (archived), cqrs-htmx, emeet-pixyd.

Brands deliberately **not** changed (correct as-is — go-cqrs-lite marker types,
BerryBig, Cyberdom): see AGENTS.md "Brands That Deliberately Skip `Name()`".

Pre-existing test failures **not** caused by this library: CreditReformBilanzampel
(BDD undefined step), timesheets (fuzz hours overflow), emeet-pixyd (PipeWire
state file), Zlota44 (internal/discovery).
