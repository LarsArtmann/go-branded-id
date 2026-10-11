package linter

import (
	"fmt"
	"strings"
)

// RuleSet selects which detection rules a scan runs. BD002 is not selectable:
// it is the directive-integrity companion that runs automatically for every
// active rule whose directives could break.
type RuleSet struct {
	BD001 bool
	BD003 bool
}

// DefaultRuleSet is the rule set every existing consumer gets: BD001 plus its
// BD002 companion. BD003 is opt-in only — brandid-lint runs fleet-wide via
// BuildFlow, and repos that deliberately silenced the old strong-id rule
// (webphone, DiscordSync) must not be flooded by an upgrade.
func DefaultRuleSet() RuleSet {
	return RuleSet{BD001: true, BD003: false}
}

// AllRules is the full rule set: everything brandid-lint can detect.
func AllRules() RuleSet {
	return RuleSet{BD001: true, BD003: true}
}

// String renders the rule set as the comma-separated list the -rules flag
// accepts, for usage messages and diagnostics.
func (r RuleSet) String() string {
	parts := make([]string, 0, 2)
	if r.BD001 {
		parts = append(parts, string(RuleIDBD001))
	}

	if r.BD003 {
		parts = append(parts, string(RuleIDBD003))
	}

	if len(parts) == 0 {
		return "none"
	}

	return strings.Join(parts, ",")
}

// ParseRuleSet parses a comma-separated rule list ("BD001,BD003"). Whitespace
// around tokens is trimmed, tokens are case-sensitive, and BD002 is rejected
// with an explanatory error: it is emitted automatically and cannot be
// selected on its own. An unknown or empty rule is a hard error — a typo in a
// CI flag must fail loudly, not silently scan with the wrong rules.
func ParseRuleSet(spec string) (RuleSet, error) {
	rules := RuleSet{}

	if strings.TrimSpace(spec) == "" {
		return rules, fmt.Errorf("empty rule set: want a comma-separated list of %s and %s",
			RuleIDBD001, RuleIDBD003)
	}

	for _, token := range strings.Split(spec, ",") {
		token = strings.TrimSpace(token)

		switch token {
		case string(RuleIDBD001):
			rules.BD001 = true
		case string(RuleIDBD003):
			rules.BD003 = true
		case string(RuleIDBD002):
			return RuleSet{}, fmt.Errorf(
				"rule %s is not selectable: BD002 findings are emitted automatically for active rules",
				RuleIDBD002,
			)
		default:
			return RuleSet{}, fmt.Errorf("unknown rule %q: want %s or %s",
				token, RuleIDBD001, RuleIDBD003)
		}
	}

	return rules, nil
}
