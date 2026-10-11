# Security Policy

## Supported Versions

Only the latest tagged release receives security fixes:

| Version             | Supported |
| ------------------- | --------- |
| latest `v0.x.y` tag | Yes       |
| older tags          | No        |

## Reporting a Vulnerability

**Do not open a public issue for security reports.**

Use GitHub's private vulnerability reporting:
[Security Advisories → Report a vulnerability](https://github.com/larsartmann/go-branded-id/security/advisories/new).

Include: affected version/tag, a minimal reproduction, and the impact
assessment. You will get an initial response within 7 days.

## Security Practices

- The root library is stdlib-only (zero third-party Go dependencies), keeping
  the supply-chain surface for consumers minimal by design.
- Automated quality gates run on every change: linting, vetting, tests in
  both JSON v1/v2 modes, and vulnerability scans (govulncheck, vulnix, and
  pnpm-audit for the website).
- Release tags are signed, annotated SSH tags, and CI pins the Go toolchain
  (`GOTOOLCHAIN=local`) so releases cannot silently build with a different
  Go version than the one pinned in `flake.nix`.
- Dependency updates are automated and reviewed (Dependabot plus the
  BuildFlow update pipeline); the website's security version pins are
  centralized in `website/pnpm-workspace.yaml`.

## Scope

This library is stdlib-only (zero third-party dependencies) and ships source
code only — there are no binaries or release artifacts beyond the source
archives. Vulnerabilities are therefore almost always in the library code
itself (e.g. panics on malformed input, incorrect deserialization bounds).
Dependabot-alert noise from the `website/` documentation site is tracked
separately and does not affect library consumers.
