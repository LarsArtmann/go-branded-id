#!/usr/bin/env bash
#
# BD003 parity gate: A/B diff between brandid-lint -rules BD003 (this repo)
# and the legacy branching-flow strong-id analyzer on a frozen corpus.
#
# Both tools render SARIF via the same go-finding library, so findings are
# normalized to "file:startLine" pairs (rule IDs BD003 and STRONG_ID mapped
# onto each other) and diffed. Every divergence must be triaged in
# docs/status/*bd003-parity*.md before branching-flow removal.
#
# Usage:
#   scripts/parity-bd003.sh
#   BRANCHING_FLOW_DIR=... scripts/parity-bd003.sh
#
# Exit codes: 0 = parity (or only documented accepts), 1 = divergence.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LINTER_DIR="$REPO_ROOT/linter"
BF_DIR="${BRANCHING_FLOW_DIR:-$HOME/projects/branching-flow}"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

echo "== building both linters =="
(cd "$LINTER_DIR" && go build -o "$WORK/brandid-lint" ./cmd/brandid-lint)
(cd "$BF_DIR" && go build -o "$WORK/branching-flow" ./cmd/branching-flow)

# The corpus: every non-testdata .go file of this repo (both modules), the
# BD003 fixture corpus (including its cmd/ exclusion case), and the legacy
# analyzer's own sample fixture. Frozen by git on both sides.
mapfile -t CORPUS < <(
	find "$REPO_ROOT" -name '*.go' \
		-not -path '*/.git/*' \
		-not -path '*/testdata/*' \
		-not -path '*/website/*' | sort
)
CORPUS+=(
	"$LINTER_DIR/testdata/bd003/params.go"
	"$LINTER_DIR/testdata/bd003/fields.go"
	"$LINTER_DIR/testdata/bd003/returns.go"
	"$BF_DIR/pkg/strongid/testdata/sample.go"
)

echo "== corpus: ${#CORPUS[@]} files =="

# sarif_rule_lines extracts "file:startLine" pairs for one rule ID from a
# SARIF report on stdin.
sarif_rule_lines() {
	jq -r --arg rule "$1" '
		.runs[0].results[]?
		| select(.ruleId == $rule)
		| "\(.locations[0].physicalLocation.artifactLocation.uri):\(.locations[0].physicalLocation.region.startLine)"
	' | sort
}

status=0

# Documented accepts (file:startLine): findings the new tool emits but the
# old one suppressed via the legacy "//lint:ignore STRONG_ID" comment form.
# BD003 deliberately does not bridge that form (or //nolint:branching-flow):
# verified zero uses outside this fixture, and the replacement directive is
# //brandid-lint:ignore(BD003) <reason>. See the parity report.
ACCEPTS=(
	"$BF_DIR/pkg/strongid/testdata/sample.go:96"
)

# strip_accepts removes documented-accept findings from stdin.
strip_accepts() {
	local line accepted
	while IFS= read -r line; do
		accepted=false
		for accept in "${ACCEPTS[@]}"; do
			if [ "$line" = "$accept" ]; then
				accepted=true
				break
			fi
		done
		if [ "$accepted" = false ]; then
			echo "$line"
		fi
	done
}

: >"$WORK/new.txt"
: >"$WORK/old.txt"

for file in "${CORPUS[@]}"; do
	abs="$(cd "$(dirname "$file")" && pwd)/$(basename "$file")"

	"$WORK/brandid-lint" -rules BD003 -format sarif "$abs" >"$WORK/new.sarif" 2>"$WORK/new.err" || true
	"$WORK/branching-flow" strong-id --format sarif "$abs" >"$WORK/old.sarif" 2>"$WORK/old.err" || true

	sarif_rule_lines BD003 <"$WORK/new.sarif" >>"$WORK/new.txt"
	sarif_rule_lines STRONG_ID <"$WORK/old.sarif" >>"$WORK/old.txt"
done

echo
echo "== per-file corpus diff (old vs new, after documented accepts) =="
strip_accepts <"$WORK/new.txt" >"$WORK/new-accepted.txt"
if diff -u "$WORK/old.txt" "$WORK/new-accepted.txt"; then
	echo "PARITY: per-file corpus identical ($(wc -l <"$WORK/new.txt") findings, $(($(wc -l <"$WORK/new.txt") - $(wc -l <"$WORK/new-accepted.txt"))) accepted)"
else
	echo "DIVERGENCE: per-file corpus differs"
	status=1
fi

# Directory-level run over the fixture corpus: compares walker behavior
# (cmd/ exclusion) end to end. suppressed.go diverges by design (new
# suppression contract) and is stripped from the old side as a documented
# accept.
FIXTURE_DIR="$LINTER_DIR/testdata/bd003"

"$WORK/brandid-lint" -rules BD003 -format sarif "$FIXTURE_DIR" >"$WORK/new-dir.sarif" 2>/dev/null || true
"$WORK/branching-flow" strong-id --format sarif "$FIXTURE_DIR" >"$WORK/old-dir.sarif" 2>/dev/null || true

sarif_rule_lines BD003 <"$WORK/new-dir.sarif" >"$WORK/new-dir.txt"
sarif_rule_lines STRONG_ID <"$WORK/old-dir.sarif" >"$WORK/old-dir.txt"

grep -v "suppressed.go" "$WORK/old-dir.txt" >"$WORK/old-dir-accepted.txt" || true

echo
echo "== fixture-dir diff (documented accept: suppressed.go on old side only) =="
if diff -u "$WORK/old-dir-accepted.txt" "$WORK/new-dir.txt"; then
	echo "PARITY: fixture dir identical after accepts ($(wc -l <"$WORK/new-dir.txt") new-side findings)"
else
	echo "DIVERGENCE: fixture dir differs beyond documented accepts"
	status=1
fi

echo
echo "== suppression fixture counts (expected divergence, for the report) =="
echo "new side findings in suppressed.go: $(grep -c "suppressed.go" "$WORK/new-dir.txt" || true)"
echo "old side findings in suppressed.go: $(grep -c "suppressed.go" "$WORK/old-dir.txt" || true)"

exit "$status"
