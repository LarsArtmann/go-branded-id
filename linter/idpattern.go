package linter

import (
	"strings"
	"unicode"
)

// containsIDPattern checks if a name refers to an identifier: the standalone
// names "id"/"ids", a snake_case "_id"/"_ids" suffix, or a compound name whose
// LAST token (split on "_" and lower->upper camel boundaries) is "id"/"ids".
// CamelCase compounds like "isValid" and English words like "valid"/"paid" do
// NOT match, but single-letter-head all-caps acronyms (PID, UID, RID) do.
//
//nolint:cyclop // ported verbatim from branching-flow strongid; the branch count IS the rule
func containsIDPattern(name string) bool {
	if name == "" {
		return false
	}

	lower := strings.ToLower(name)

	// Standalone: id, ID, ids, IDs.
	if lower == "id" || lower == "ids" {
		return true
	}

	// Snake case suffix: user_id, order_ids.
	if strings.HasSuffix(lower, "_id") || strings.HasSuffix(lower, "_ids") {
		return true
	}

	// All-caps acronyms with a single-letter head: PID, UID, RID, AID (and
	// their plural forms). Longer all-caps words ending in "id" are English
	// words or multi-letter acronyms (VALID, SOLID, ACID), not ID names.
	if name == strings.ToUpper(name) {
		if (len(name) == 3 && strings.HasSuffix(lower, "id")) ||
			(len(name) == 4 && strings.HasSuffix(lower, "ids")) {
			return true
		}
	}

	// Compound names: the last token decides. userId -> [user,Id],
	// isValid -> [is,Valid], validationID -> [validation,ID].
	tokens := splitNameTokens(name)
	if len(tokens) == 0 {
		return false
	}

	last := strings.ToLower(tokens[len(tokens)-1])

	return last == "id" || last == "ids"
}

// splitNameTokens splits an identifier into its snake_case and CamelCase
// tokens. Boundaries are "_" separators and lower->upper (or digit->upper)
// transitions: "userId" -> ["user","Id"], "user_ids" -> ["user","ids"],
// "validationID" -> ["validation","ID"]. Consecutive capitals stay glued, so
// "HTTPServer" is one token ["HTTPServer"] and "userID2" -> ["user","ID2"].
func splitNameTokens(name string) []string {
	var tokens []string

	var current strings.Builder

	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}

	runes := []rune(name)
	for i, r := range runes {
		switch {
		case r == '_':
			flush()
		case unicode.IsUpper(r) && i > 0 && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1])):
			flush()
			current.WriteRune(r)
		default:
			current.WriteRune(r)
		}
	}

	flush()

	return tokens
}
