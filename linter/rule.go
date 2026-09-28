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
