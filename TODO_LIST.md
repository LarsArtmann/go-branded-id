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

| Task                                                                       | Status | Impact | Evidence                                                                                                                                                  |
| -------------------------------------------------------------------------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Regenerate website `package-lock.json` (`npm install` in `website/`)       | 🔴     | High   | `package.json` overrides are set (`astro` → `^7.1.0` XSS fix, `fast-uri` → `^3.1.4` host-confusion fix) but the lockfile still holds vulnerable versions. |
| Build & verify the website (`npm run build` in `website/`)                 | 🔴     | High   | `guides/error-handling.mdx` and `guides/namer-tool.mdx` were added but never compiled; sidebar links (`astro.config.mjs`) and frontmatter are unverified. |
| Add a CI/release guard that rejects a tracked compiled binary at repo root | 🔴     | High   | Prevents recurrence of the v0.5.0 incident where a tracked `namer` binary inflated release source archives ~10x. Currently relies on `.gitignore` only.   |

## Medium Impact

| Task                                                                           | Status | Impact | Evidence                                                                                                                                                         |
| ------------------------------------------------------------------------------ | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Add `ErrMarshal`/`ErrUnmarshal` test coverage for the remaining delegate paths | 🔴     | Med    | Only `MarshalBinary` currently proves `ErrMarshal`. Untested: JSON marshaler, SQL `Value()` TextMarshaler, BinaryUnmarshaler delegate, TextUnmarshaler delegate. |
| Verify `nix flake check --all-systems` passes locally                          | 🔴     | Med    | The `flake-check` CI job was added in the v0.5.1 cycle but never executed locally (flagged as missed in the 2026-07-28 23:01 status report, section D3).         |

## Ecosystem tracking

| Task                                         | Status       | Impact | Evidence                                                                                                                                                                                                                   |
| -------------------------------------------- | ------------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Bump 14 downstream ecosystem repos to v0.5.1 | 🔵 `BLOCKED` | Med    | Source fixes from the v0.3.x cycle (added `Name()` methods, `.String()` → `.Get()`) are applied and pushed to all repos. The `go.mod` dependency bump is not yet done — requires per-repo access to clone, bump, test, PR. |

---

## Ecosystem detail

This library has 14 downstream repos. The v0.3.0/v0.3.1 source changes (`Name()`
methods added, `.String()` → `.Get()` fixes, test fixes) are **applied and pushed**
to all of them. The **`go.mod` dependency bump to v0.5.1 is not yet done**
in any repo.

Downstream repos: InboxClean, CreditReformBilanzampel, ActaFlow, SEC, storbi,
ChastityAPI, smart-configs, StopTube, universal-workflow, Zlota44, timesheets,
complaints-mcp (archived), cqrs-htmx, emeet-pixyd.

Deliberately **not** changed (correct as-is): go-cqrs-lite (marker types — see
AGENTS.md "Brands That Deliberately Skip `Name()`"), BerryBig (test brands
only), Cyberdom (no brand types).

Pre-existing test failures **not** caused by this library: CreditReformBilanzampel
(BDD undefined step), timesheets (fuzz hours overflow), emeet-pixyd (PipeWire
state file), Zlota44 (internal/discovery).
