// Package provider registers brandid-lint as a BuildFlow tool provider via
// the go-finding toolsdk contract.
//
// The package-level Provider var self-registers at import time. BuildFlow
// (or any orchestrator) discovers it via toolsdk.All() after a blank import
// and converts it into its internal tool representation, so adding
// brandid-lint to a pipeline is a one-line import with zero orchestrator-side
// glue to drift out of date.
//
// Detect delegates to linter.Detect, the same entry point the CLI uses, so
// CLI and BuildFlow findings can never diverge. The spec is detector-only by
// design: writing Name() stubs into arbitrary files requires AST-based
// insertion that does not exist yet, so no Repair capability is claimed.
package provider

import (
	gofinding "github.com/larsartmann/go-finding"
	gofstoolsdk "github.com/larsartmann/go-finding/toolsdk"

	"github.com/larsartmann/go-branded-id/linter"
)

// Provider registers the brandid-lint Spec in the process-global toolsdk
// registry. Blank-import this package to activate it:
//
//	import _ "github.com/larsartmann/go-branded-id/linter/provider"
var Provider = gofstoolsdk.Register(
	gofstoolsdk.Spec{ //nolint:gochecknoglobals // self-registration at import time is the toolsdk contract
		Name:        linter.ToolName,
		Description: "BD001: brand types used with id.ID but missing their Name() string method",
		Trigger:     gofstoolsdk.OnGoModule(),
		DependsOn:   nil,
		// Brands are a per-Go-module concept: fan out so findings paths are
		// module-relative and monorepos get one run per module.
		ModuleFanOut: true,
		Inputs:       []string{"**/*.go"},
		Detect:       gofinding.NamedDetectorFunc(linter.ToolName, linter.Detect),
		Repair:       nil,
		HealthCheck:  nil,
	},
)
