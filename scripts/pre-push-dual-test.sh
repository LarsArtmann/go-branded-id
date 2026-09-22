#!/usr/bin/env bash
# Dual-mode test hook: runs go test in both v1 and v2 JSON modes.
# Install: cp this file to .git/hooks/pre-push && chmod +x
# This catches single-mode blind spots where code passes v1 but fails v2.
#
# The import guard runs BEFORE go test: when the v1 files import
# encoding/json/v2, the v1 package fails to compile and no test can run, so a
# Go-level contract test is structurally blind to that corruption
# (docs/status/2026-08-02_16-11, section e.1). A grep catches it in
# milliseconds with a clear message.

set -u

fail=0

echo "Guarding JSON v1 imports..."
if grep -n 'encoding/json/v2' id_json_v1.go json_helpers_v1_test.go; then
	echo "FAIL: v1 JSON files must import \"encoding/json\", not \"encoding/json/v2\"."
	echo "      See AGENTS.md 'Dual JSON v1/v2 Support' before re-formatting these files."
	fail=1
else
	echo "  → imports ok"
fi

echo "Running dual-mode go tests..."

echo "  → json v1..."
if go test ./... -count=1 -race; then
	echo "  → json v1 ok"
else
	echo "  → json v1 FAILED"
	fail=1
fi

echo "  → json v2..."
if GOEXPERIMENT=jsonv2 go test ./... -count=1 -race; then
	echo "  → json v2 ok"
else
	echo "  → json v2 FAILED"
	fail=1
fi

if [ "$fail" -ne 0 ]; then
	echo "Dual-mode checks FAILED (see above)."
	exit 1
fi

echo "Dual-mode tests passed."
