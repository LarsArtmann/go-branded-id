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
assessment. You will get an acknowledgment within 7 days.

## Scope

This library is stdlib-only (zero third-party dependencies) and ships source
code only — there are no binaries or release artifacts beyond the source
archives. Vulnerabilities are therefore almost always in the library code
itself (e.g. panics on malformed input, incorrect deserialization bounds).
Dependabot-alert noise from the `website/` documentation site is tracked
separately and does not affect library consumers.
