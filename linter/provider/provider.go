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
// CLI and BuildFlow findings can never diverge. The `rules` option selects
// the rule set per repo (tool_options in .buildflow.yml); without it Detect
// runs the default rule set, byte-identical to every earlier release.
// Repair inserts the suggested Name() stub for every unsuppressed BD001
// finding; BuildFlow re-runs Detect to measure the delta, and suppressed
// brands are never modified.
package provider

import (
	"context"
	"fmt"

	"github.com/larsartmann/go-branded-id/linter"
	gofinding "github.com/larsartmann/go-finding"
	gofstoolsdk "github.com/larsartmann/go-finding/toolsdk"
)

// optionRules is the Spec-declared knob for per-repo rule selection. The
// value rides the per-run context (toolsdk.WithOptions, injected by
// BuildFlow's tool_options config); the toolsdk supports scalar kinds only,
// so the rule list is a comma-separated string parsed by linter.ParseRuleSet.
const optionRules = "rules"

// detect runs the rule set selected by the context options, defaulting to
// the fleet default (BD001+BD002). An invalid value is a hard Detect error:
// a config typo must fail loudly, not silently scan with the wrong rules.
func detect(ctx context.Context) ([]gofinding.Finding, error) {
	rules := linter.DefaultRuleSet()

	if values, ok := gofstoolsdk.OptionsFromContext(ctx); ok {
		if raw, ok := values[optionRules].(string); ok && raw != "" {
			parsed, err := linter.ParseRuleSet(raw)
			if err != nil {
				return nil, fmt.Errorf("brandid-lint option %s: %w", optionRules, err)
			}

			rules = parsed
		}
	}

	return linter.DetectWithRules(ctx, rules)
}

// Provider registers the brandid-lint Spec in the process-global toolsdk
// registry. Blank-import this package to activate it:
//
//	import _ "github.com/larsartmann/go-branded-id/linter/provider"
//
//nolint:gochecknoglobals // self-registration at import time is the toolsdk contract
var Provider = gofstoolsdk.Register(
	gofstoolsdk.Spec{
		Name:        linter.ToolName,
		Description: "BD001: brand types used with id.ID but missing their Name() string method; BD002: broken suppression directives; BD003 (opt-in via a RuleSet): primitive-typed ID parameters, returns, and fields that should become branded ID types",
		Trigger:     gofstoolsdk.OnGoModule(),
		DependsOn:   nil,
		// Brands are a per-Go-module concept: fan out so findings paths are
		// module-relative and monorepos get one run per module.
		ModuleFanOut: true,
		Inputs:       []string{"**/*.go"},
		Detect:       gofinding.NamedDetectorFunc(linter.ToolName, detect),
		Options: []gofstoolsdk.Option{
			{
				Name:        optionRules,
				Kind:        gofstoolsdk.OptionKindString,
				Default:     linter.DefaultRuleSet().String(),
				Description: "Comma-separated rule IDs to run: BD001, BD003 (BD002 directive findings ride along automatically). Default BD001.",
			},
		},
		Repair: gofstoolsdk.RepairerFunc(
			func(ctx context.Context) (gofstoolsdk.RepairResult, error) {
				inserted, err := linter.Repair(ctx)
				if err != nil {
					return gofstoolsdk.RepairResult{}, fmt.Errorf("brandid-lint repair: %w", err)
				}

				return gofstoolsdk.RepairResult{
					Description: fmt.Sprintf("inserted %d Name() stub(s)", inserted),
				}, nil
			},
		),
		HealthCheck: nil,
	},
)
