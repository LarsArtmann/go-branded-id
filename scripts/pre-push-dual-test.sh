#!/usr/bin/env bash
# Dual-mode test hook: runs go test in both v1 and v2 JSON modes.
# Install: cp this file to .git/hooks/pre-push && chmod +x
# This catches single-mode blind spots where code passes v1 but fails v2.
set -e

echo "Running dual-mode go tests..."

echo "  → json v1..."
go test ./... -count=1 -race
echo "  → json v2..."
GOEXPERIMENT=jsonv2 go test ./... -count=1 -race

echo "Dual-mode tests passed."
