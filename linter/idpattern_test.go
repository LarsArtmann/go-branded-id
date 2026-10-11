package linter

import (
	"slices"
	"testing"
)

func TestSplitNameTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want []string
	}{
		{name: "userId", want: []string{"user", "Id"}},
		{name: "user_id", want: []string{"user", "id"}},
		{name: "user_ids", want: []string{"user", "ids"}},
		{name: "PID", want: []string{"PID"}},
		{name: "UIDs", want: []string{"UIDs"}},
		{name: "isValid", want: []string{"is", "Valid"}},
		{name: "valid", want: []string{"valid"}},
		{name: "paid", want: []string{"paid"}},
		{name: "validationID", want: []string{"validation", "ID"}},
		{name: "HTTPServer", want: []string{"HTTPServer"}},
		{name: "userID2", want: []string{"user", "ID2"}},
		{name: "", want: nil},
		{name: "_", want: nil},
		{name: "id", want: []string{"id"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := splitNameTokens(tt.name)
			if !slices.Equal(got, tt.want) {
				t.Errorf("splitNameTokens(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestContainsIDPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		// Standalone names.
		{name: "id", want: true},
		{name: "ID", want: true},
		{name: "ids", want: true},
		{name: "IDs", want: true},
		// Snake_case suffix.
		{name: "user_id", want: true},
		{name: "order_ids", want: true},
		{name: "product_id_v2", want: false}, // last token is v2, not id
		// Single-letter-head all-caps acronyms.
		{name: "PID", want: true},
		{name: "UID", want: true},
		{name: "RIDS", want: true},
		{name: "AIDS", want: true},
		{name: "RIDs", want: false}, // mixed case: not an all-caps acronym
		{name: "AIDs", want: false},
		// Compound names ending in an ID token.
		{name: "userId", want: true},
		{name: "userID", want: true},
		{name: "validationID", want: true},
		{name: "parentIDs", want: true},
		{name: "id_parent", want: false}, // prefix, not suffix: last token is "parent"
		// English words and non-ID names must NOT match.
		{name: "isValid", want: false},
		{name: "valid", want: false},
		{name: "paid", want: false},
		{name: "overpaid", want: false},
		{name: "solid", want: false},
		{name: "acid", want: false},
		{name: "rapid", want: false},
		{name: "candid", want: false},
		{name: "vivid", want: false},
		{name: "fluid", want: false},
		{name: "UUID", want: false}, // multi-letter acronym head
		{name: "VALID", want: false},
		{name: "SOLID", want: false},
		{name: "HTTPServer", want: false},
		{name: "https", want: false},
		{name: "", want: false},
		{name: "name", want: false},
		{name: "identity", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := containsIDPattern(tt.name); got != tt.want {
				t.Errorf("containsIDPattern(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestContainsIDPattern_EnglishWordSweep sweeps a broad list of lowercase
// English words (and their CamelCase uses) that merely end in "id": none of
// them may match, or every findings flood would be full of false positives.
func TestContainsIDPattern_EnglishWordSweep(t *testing.T) {
	t.Parallel()

	words := []string{
		"solid", "acid", "rapid", "candid", "vivid", "fluid", "squalid",
		"fetid", "torrid", "fervid", "florid", "frigid", "humid", "livid",
		"lucid", "morbid", "pallid", "placid", "rabid", "rigid", "rancid",
		"staid", "turgid", "valid", "tepid", "acrid", "arid", "avid",
		"avoid", "grid", "skid", "squid", "kid", "did", "hid", "lid",
	}

	for _, word := range words {
		t.Run(word, func(t *testing.T) {
			t.Parallel()

			camel := "is" + string(word[0]-32) + word[1:]

			for _, variant := range []string{word, camel} {
				if containsIDPattern(variant) {
					t.Errorf("containsIDPattern(%q) = true, want false", variant)
				}
			}
		})
	}
}
