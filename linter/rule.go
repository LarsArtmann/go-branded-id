package linter

import (
	gofinding "github.com/larsartmann/go-finding"
)

// ToolName is the stable identity of this linter across the linter stack: it
// is the BuildFlow tool name, the finding tool attribution, and the CLI
// binary name. Chosen once; never changed.
const ToolName = "brandid-lint"

// Version is the brandid-lint tool version, independent of the library
// version of the go-branded-id root module. It is bumped only when the
// linter module itself is released (tag linter/vX.Y.Z).
const Version = "0.1.0"

// RuleIDBD001 flags brand types that are used with id.ID but have no Name()
// string method. The ID is stable forever: suppressions and configs key on
// it, so its meaning must never change once shipped.
const RuleIDBD001 gofinding.RuleName = "BD001"

// RuleIDBD002 flags brandid-lint suppression directives that cannot do their
// job: malformed syntax, a missing reason, an unknown rule ID, a directive on
// a brand that already has Name() (stale), a duplicate, or one that
// annotates no brand declaration at all. A silent no-op directive is worse
// than none, so every broken directive stays loud. BD002 itself is not
// suppressible. The ID is stable forever.
const RuleIDBD002 gofinding.RuleName = "BD002"
