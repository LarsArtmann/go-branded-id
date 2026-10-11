package linter

import (
	"testing"
)

func TestParseRuleSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		spec    string
		want    RuleSet
		wantErr bool
	}{
		{spec: "BD001", want: RuleSet{BD001: true}},
		{spec: "BD003", want: RuleSet{BD003: true}},
		{spec: "BD001,BD003", want: RuleSet{BD001: true, BD003: true}},
		{spec: "BD001, BD003", want: RuleSet{BD001: true, BD003: true}},
		{spec: "BD001,BD001", want: RuleSet{BD001: true}},
		{spec: "BD001,", wantErr: true}, // empty trailing token: SplitSeq yields it, the switch rejects it
		{spec: "BD002", wantErr: true},
		{spec: "BD999", wantErr: true},
		{spec: "bd001", wantErr: true},
		{spec: "", wantErr: true},
		{spec: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			t.Parallel()

			got, err := ParseRuleSet(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseRuleSet(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
			}

			// On error the rule set may carry partially-applied rules
			// ("BD001," sets BD001 before the empty token errors), so only
			// the success path compares values.
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseRuleSet(%q) = %+v, want %+v", tt.spec, got, tt.want)
			}
		})
	}
}

func TestRuleSetDefaults(t *testing.T) {
	t.Parallel()

	if got := DefaultRuleSet(); got != (RuleSet{BD001: true}) {
		t.Errorf("DefaultRuleSet() = %+v, want BD001 only", got)
	}

	if got := AllRules(); got != (RuleSet{BD001: true, BD003: true}) {
		t.Errorf("AllRules() = %+v, want both rules", got)
	}
}

func TestRuleSetString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rules RuleSet
		want  string
	}{
		{rules: RuleSet{BD001: true}, want: "BD001"},
		{rules: RuleSet{BD003: true}, want: "BD003"},
		{rules: RuleSet{BD001: true, BD003: true}, want: "BD001,BD003"},
		{rules: RuleSet{}, want: "none"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			if got := tt.rules.String(); got != tt.want {
				t.Errorf("RuleSet.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
