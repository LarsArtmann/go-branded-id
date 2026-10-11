package linter

import (
	"errors"
	"fmt"
	"strings"
)

// Rule-set parse errors. Sentinels so callers can match on them; dynamic
// detail (the offending token) rides along via %w wrapping.
var (
	// errEmptyRuleSet rejects a blank -rules value, which is usually a shell
	// quoting mistake rather than "run nothing".
	errEmptyRuleSet = errors.New("empty rule set: want a comma-separated list of BD001 and BD003")
	// errUnknownRule rejects a rule ID this tool never had.
	errUnknownRule = errors.New("unknown rule")
	// errBD002NotSelectable rejects selecting the directive-integrity rule on
	// its own: BD002 rides along with every active rule automatically.
	errBD002NotSelectable = errors.New(
		"rule BD002 is not selectable: BD002 findings are emitted automatically for active rules",
	)
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
		return rules, errEmptyRuleSet
	}

	for token := range strings.SplitSeq(spec, ",") {
		token = strings.TrimSpace(token)

		switch token {
		case string(RuleIDBD001):
			rules.BD001 = true
		case string(RuleIDBD003):
			rules.BD003 = true
		case string(RuleIDBD002):
			return rules, errBD002NotSelectable
		default:
			return rules, fmt.Errorf("%w %q: want %s or %s",
				errUnknownRule, token, RuleIDBD001, RuleIDBD003)
		}
	}

	return rules, nil
}
