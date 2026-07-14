# Contributing to go-branded-id

> **Thank you for contributing!** This guide covers everything you need to know.

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/LarsArtmann/go-branded-id.git
cd go-branded-id

# 2. Enter the Nix dev shell (provides Go 1.26, golangci-lint, gopls)
nix develop

# 3. Verify everything works
nix run .#test
nix run .#lint
```

> **CRITICAL:** This library uses `encoding/json/v2`, which requires `GOEXPERIMENT=jsonv2`.
> The Nix flake sets this automatically. If you run `go` commands outside the Nix shell,
> always set it: `export GOEXPERIMENT=jsonv2`.

## Development Setup

All build and task automation goes through the Nix flake:

| Command               | Purpose                                                    |
| --------------------- | ---------------------------------------------------------- |
| `nix develop`         | Enter dev shell (Go 1.26, golangci-lint, gopls, trash-cli) |
| `nix run .#test`      | Run tests                                                  |
| `nix run .#test-race` | Run with race detector                                     |
| `nix run .#build`     | Build                                                      |
| `nix run .#lint`      | Run golangci-lint                                          |
| `nix run .#vet`       | Run `go vet`                                               |
| `nix run .#coverage`  | Generate coverage report                                   |
| `nix flake check`     | Run all flake checks (includes sandbox build)              |
| `nix fmt`             | Format everything (gofumpt, goimports, golines, nixfmt)    |

There is no `justfile`, no `Makefile`, and no `CONTRIBUTING-setup.sh`. If you see
references to these in older documentation, they are stale.

## Code Standards

- **Formatter:** gofumpt (stricter than gofmt) + goimports + golines (120 char max)
- **Linter:** golangci-lint v2 with an extremely strict config (`.golangci.yml`)
- `nolint` comments are common and expected — include a justification
- All tests use `t.Parallel()` — both at the function level and inside `t.Run` subtests
- Cyclop max complexity: 12

## Testing

- All tests run via `nix run .#test` (or `GOEXPERIMENT=jsonv2 go test ./... -count=1`)
- Tests use generic helpers: `testIDRoundTrip[B any, V comparable](t *testing.T, ...)`
- Benchmarks use `b.Loop()` (Go 1.24+ pattern)
- Fuzz tests cover JSON and Binary round-trips
- Example functions (`Example*`) provide documentation-driven testing

## Architecture

This is a flat, single-package library (`package id`) at the repository root.
There is no `internal/`, `pkg/`, or `cmd/` subdirectory structure beyond `cmd/namer/`.

See `AGENTS.md` for detailed architecture context, gotchas, and conventions.

## Pull Request Process

1. Create a branch from `master`
2. Ensure `nix run .#test` and `nix run .#lint` pass
3. Ensure `nix fmt` has been run (no formatting diff)
4. Write tests for any new functionality
5. Update `CHANGELOG.md` under an `[Unreleased]` section
6. Submit a PR with a clear description of what and why

## Commit Messages

Follow conventional commits:

```
feat: add new serialization format
fix: prevent nil pointer in Scan
docs: update README performance table
refactor: simplify type switch
```

Tags must be signed (SSH) and annotated (`git tag -a`). CI creates a GitHub
Release automatically on semver tags matching `v*.*.*`.
