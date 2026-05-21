# Status Report: go-branded-id — COMPLETE

**Date:** 2026-05-20 14:30
**Status:** ALL DONE. Library released as v0.3.0, ecosystem migrated.

---

## What Was Done

### Library (go-branded-id v0.3.0)

| Change                                                                 | Status  |
| ---------------------------------------------------------------------- | ------- |
| Brand-aware `String()` — named brands show `"Brand:value"`             | ✅ Done |
| `BrandNamer` interface + `BrandName[B]()`                              | ✅ Done |
| `ValidateID` + `ValidateIDWithValue` — shipped (was only example code) | ✅ Done |
| `MustValidateID` — panic version                                       | ✅ Done |
| `GoString()` / `%#v` — meaningful debug output                         | ✅ Done |
| `valueString()` internal — serialization never includes prefix         | ✅ Done |
| CHANGELOG, MIGRATION, package docs, benchmarks, fuzz, examples         | ✅ Done |
| Tagged `v0.3.0`, pushed, CI release triggered                          | ✅ Done |

### Ecosystem Breakage Fix (Critical)

Found that 3 repos already had `Name()` and would break with brand-aware `String()`.
Fixed by replacing `.String()` with `.Get()` for API/storage/key usages:

| Repo                    | Fixes                                                           | Status       |
| ----------------------- | --------------------------------------------------------------- | ------------ |
| InboxClean              | 12 `.String()` → `.Get()` for Gmail API + CQRS                  | ✅ Committed |
| CreditReformBilanzampel | 6 `.String()` → `.Get()` for process keys + correlation headers | ✅ Committed |
| ActaFlow                | 1 `.String()` → `.Get()` for correlation ID matching            | ✅ Committed |
| cqrs-htmx/usermgmt      | 8 `.String()` → `.Get()` for Casbin RBAC                        | ✅ Committed |

### Ecosystem `Name()` Addition

Added `Name()` to brand types across 12 repos (64 new Name() methods):

| Repo               | Brands                                 | Status       |
| ------------------ | -------------------------------------- | ------------ |
| SEC                | 2 (Game, Player)                       | ✅ Committed |
| storbi             | 8 (Item, SKU, Category, etc.)          | ✅ Committed |
| ChastityAPI        | 14 (User, UUID, Device, Command, etc.) | ✅ Committed |
| smart-configs      | 10 (Account, Build, PullRequest, etc.) | ✅ Committed |
| StopTube           | 6 (Schedule, VID, IDID, etc.)          | ✅ Committed |
| universal-workflow | 10 (Workflow, Activity, Session, etc.) | ✅ Committed |
| Zlota44            | 6 (DiscoveredProperty, Property, etc.) | ✅ Committed |
| timesheets         | 6 (Timesheet, NaturalPerson, etc.)     | ✅ Committed |
| complaints-mcp     | 1 (Complaint)                          | ✅ Committed |
| cqrs-htmx/usermgmt | 1 (user)                               | ✅ Committed |
| emeet-pixyd        | 2 (pid, sourceID)                      | ✅ Committed |

### Deliberately Skipped

| Repo         | Reason                                                                          |
| ------------ | ------------------------------------------------------------------------------- |
| go-cqrs-lite | Markers are internal keys — adding Name() would break stream keys, storage keys |
| BerryBig     | Only test brands                                                                |
| Cyberdom     | No brand types found                                                            |

## Final Numbers

- **89 tests** pass with `-race`, **0 lint issues**
- **v0.3.0** tagged and pushed to GitHub
- **64 Name() methods** added across ecosystem
- **27 .String()→.Get() fixes** to prevent breakage
- **4 repos** with pre-existing Name() migrated to safe patterns
