#!/usr/bin/env bash
# Asserts the three-way Go version pin agreement documented in AGENTS.md
# ("Go Version Pins Must Move Together"): each module's go.mod `go` directive
# must not exceed the flake.nix Go pin, and every CI go-version must match one
# of the flake pins (major.minor). Guards against the auto-commit daemon's
# go-mod-update bumps that broke builds three times (2026-08, 2026-09-17,
# 2026-09-29). Run locally or via the CI `go-pins` job.
set -euo pipefail

cd "$(dirname "$0")/.."

fail() {
	echo "::error::$1" >&2
	exit 1
}

flake_pin() { # $1 = variable name in flake.nix (goPkg | linterGoPkg)
	grep -oE "(^|[[:space:]])$1 = pkgs\.go_1_[0-9]+" flake.nix |
		grep -oE 'go_1_[0-9]+$' | head -n1 | sed 's/^go_//; s/_/./'
}

go_directive() { # $1 = path to go.mod
	awk '$1 == "go" { print $2; exit }' "$1"
}

major_minor() { echo "$1" | cut -d. -f1-2; }

assert_within_pin() { # $1 = label, $2 = go.mod directive, $3 = flake pin
	local directive_mm
	directive_mm=$(major_minor "$2")
	if [[ "$(printf '%s\n' "$directive_mm" "$3" | sort -V | tail -n1)" != "$3" ]]; then
		fail "$1: go.mod directive $2 exceeds flake.nix pin $3. The go directive is a \
consumer-facing minimum and must move in lockstep with flake.nix and the CI go-version."
	fi
}

root_pin=$(flake_pin goPkg)
linter_pin=$(flake_pin linterGoPkg)
[[ -n "$root_pin" ]] || fail "flake.nix: goPkg = pkgs.go_1_X pin not found"
[[ -n "$linter_pin" ]] || fail "flake.nix: linterGoPkg = pkgs.go_1_X pin not found"

assert_within_pin "root module" "$(go_directive go.mod)" "$root_pin"
assert_within_pin "linter module" "$(go_directive linter/go.mod)" "$linter_pin"

for workflow in .github/workflows/go.yml .github/workflows/release.yml .github/workflows/validate-docs.yml; do
	while IFS= read -r version; do
		case "$version" in
		"$root_pin" | "$linter_pin") ;;
		*)
			fail "$workflow: go-version $version matches neither flake pin ($root_pin / $linter_pin)"
			;;
		esac
	done < <(grep -oE 'go-version: "[0-9.]+"' "$workflow" | sed -E 's/go-version: "([0-9.]+)"/\1/')
done

echo "Go pins consistent: root $(go_directive go.mod) <= go_$root_pin, \
linter $(go_directive linter/go.mod) <= go_$linter_pin, CI go-versions match."
