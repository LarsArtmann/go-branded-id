package linter

import (
	"fmt"
	"strings"

	gofinding "github.com/larsartmann/go-finding"
)

// primitiveIDImportPath is the import path of the go-branded-id library the
// suggested type definitions reference.
const primitiveIDImportPath = "github.com/larsartmann/go-branded-id"

// capitalizeFirst returns s with its first byte uppercased.
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

// extractIDRoot extracts the domain root from an ID parameter name:
// "userId" -> "User", "order_id" -> "Order", "requestID" -> "Request",
// "product_id_v2" -> "Product".
func extractIDRoot(name string) string {
	lower := strings.ToLower(name)

	for _, suffix := range []string{"_id", "_ids", "_id_v2", "_id_v3"} {
		if strings.HasSuffix(lower, suffix) {
			return capitalizeFirst(strings.TrimSuffix(name, suffix))
		}
	}

	if strings.HasPrefix(lower, "id_") {
		return capitalizeFirst(strings.TrimPrefix(name, "id_"))
	}

	// Suffix "Id" or "ID": use the first case-insensitive "id" occurrence.
	if idx := strings.Index(lower, "id"); idx > 0 {
		return capitalizeFirst(name[:idx])
	}

	return capitalizeFirst(name)
}

// commonVerbPrefixes are verb prefixes commonly used in function names. When a
// parameter is just "id", these are stripped to extract the domain noun.
//
//nolint:gochecknoglobals // Immutable lookup table.
var commonVerbPrefixes = []string{
	// CRUD operations.
	"Get", "Add", "Create", "Delete", "Update", "Remove", "Upsert",
	"Insert", "Select", "Replace", "Modify", "Edit", "Write",

	// Query/Find operations.
	"Find", "List", "Search", "Query", "Lookup", "Fetch", "Load", "Read",

	// Persistence operations.
	"Save", "Store", "Persist", "Flush", "Commit", "Rollback",

	// HTTP methods.
	"Head", "Post", "Put", "Patch",

	// Lifecycle operations.
	"Start", "Stop", "Begin", "End", "Init", "New", "Make", "Build",
	"Initialize", "Finalize", "Setup", "Teardown", "Open", "Close",

	// Validation operations.
	"Validate", "Check", "Test", "Verify", "Assert", "Ensure",

	// Processing operations.
	"Process", "Handle", "Execute", "Run", "Perform", "Apply",
	"Parse", "Format", "Convert", "Transform", "Map", "Reduce",
	"Filter", "Sort", "Group", "Aggregate", "Merge", "Split",

	// Communication operations.
	"Send", "Receive", "Push", "Pull", "Publish",
	"Emit", "Broadcast", "Notify", "Signal",

	// State operations.
	"Enable", "Disable", "Activate", "Deactivate", "Toggle",
	"Lock", "Unlock", "Acquire", "Release", "Reserve",

	// Connection operations.
	"Connect", "Disconnect", "Attach", "Detach", "Bind", "Unbind",
	"Register", "Unregister", "Subscribe", "Unsubscribe",

	// Other common verbs.
	"Render", "Display", "Print", "Log", "Dump",
	"Clone", "Copy", "Duplicate", "Move", "Rename",
	"Reset", "Clear", "Purge", "Archive",
	"Generate", "Produce", "Compute", "Calculate",
	"Resolve", "Evaluate", "Analyze", "Inspect", "Examine",
	"Prepare", "Configure", "Set", "Assign",
	"Accept", "Reject", "Approve", "Deny",
	"Export", "Import", "Serialize", "Deserialize",
	"Encode", "Decode", "Compress", "Decompress",
	"Encrypt", "Decrypt", "Hash", "Sign",
}

// verbOnlyNames are function names that are purely verbs with no domain noun;
// they would produce meaningless suggestions like "CloneID".
//
//nolint:gochecknoglobals // Immutable lookup table.
var verbOnlyNames = map[string]bool{
	"Clone": true, "Copy": true, "Reset": true, "Clear": true,
	"Flush": true, "Commit": true, "Close": true, "Open": true,
	"Run": true, "Start": true, "Stop": true, "Send": true,
	"Receive": true, "Parse": true, "Format": true, "Set": true,
}

// extractDomainFromFunction extracts the domain noun from a function name by
// stripping common verb prefixes: "AddRule" -> "Rule", "GetUser" -> "User".
// A name that is only a verb ("Clone") yields "" (GenericID fallback); a name
// with no verb prefix is returned as-is.
func extractDomainFromFunction(funcName string) string {
	if funcName == "" {
		return ""
	}

	for _, verb := range commonVerbPrefixes {
		if after, ok := strings.CutPrefix(funcName, verb); ok {
			// Only accept the remainder as a domain when it starts with an
			// uppercase letter, so "Finding" is not parsed as "Find" + "ing".
			if after != "" && after[0] >= 'A' && after[0] <= 'Z' {
				return after
			}
		}
	}

	if verbOnlyNames[funcName] {
		return ""
	}

	return funcName
}

// appendSuffixDedup appends suffix to base unless base already ends with it
// (case-insensitive); an empty base becomes "Generic"+suffix.
func appendSuffixDedup(base, suffix, lowerSuffix string) string {
	if base == "" {
		return "Generic" + suffix
	}

	if strings.HasSuffix(strings.ToLower(base), lowerSuffix) {
		return base
	}

	return base + suffix
}

// suggestIDSuffixFromParent derives a suffixed name from the parent function
// or type context.
func suggestIDSuffixFromParent(parentName, suffix, lowerSuffix string) string {
	if parentName != "" {
		if domain := extractDomainFromFunction(parentName); domain != "" {
			return appendSuffixDedup(domain, suffix, lowerSuffix)
		}
	}

	return "Generic" + suffix
}

// suggestIDSuffix suggests a name suffix for an ID based on the parameter
// name: a standalone "id" extracts the domain from the parent function name
// ("AddRule" -> "Rule"+suffix, not "AddRule"+suffix).
func suggestIDSuffix(paramName, parentName, suffix string) string {
	lowerSuffix := strings.ToLower(suffix)

	if strings.ToLower(paramName) == "id" {
		return suggestIDSuffixFromParent(parentName, suffix, lowerSuffix)
	}

	root := extractIDRoot(paramName)
	if root == "" {
		return suggestIDSuffixFromParent(parentName, suffix, lowerSuffix)
	}

	return appendSuffixDedup(root, suffix, lowerSuffix)
}

// suggestIDTypeName suggests a strong ID type name for a primitive-ID name:
// "userId" -> "UserID", "product_id" -> "ProductID".
func suggestIDTypeName(paramName, parentName string) string {
	return suggestIDSuffix(paramName, parentName, "ID")
}

// suggestIDBrand suggests the brand type name for a primitive-ID name:
// "userId" -> "UserBrand".
func suggestIDBrand(paramName, parentName string) string {
	return suggestIDSuffix(paramName, parentName, "Brand")
}

// buildTypeDefinition creates the complete suggested type definition: a type
// alias to id.ID plus the empty brand struct it keys on.
func buildTypeDefinition(typeName, brandName, underlyingType string) string {
	return fmt.Sprintf(
		"type %s = id.ID[%s, %s]\n\ntype %s struct{}",
		typeName, brandName, underlyingType, brandName,
	)
}

// buildCodeSnippet renders a before/after code line for a violation, shaped
// by where the primitive ID was found.
func buildCodeSnippet(name, typeStr string, kind primitiveLocation, parentName string) string {
	if parentName == "" && kind != locationReturn {
		return fmt.Sprintf("%s %s", name, typeStr)
	}

	switch kind {
	case locationField:
		return fmt.Sprintf("type %s struct { %s %s }", parentName, name, typeStr)
	case locationReturn:
		if parentName == "" {
			return fmt.Sprintf("(%s %s)", name, typeStr)
		}

		return fmt.Sprintf("func %s() (%s %s, ...)", parentName, name, typeStr)
	case locationParam:
		return fmt.Sprintf("func %s(%s %s)", parentName, name, typeStr)
	default:
		return fmt.Sprintf("%s %s", name, typeStr)
	}
}

// severityRank orders BD003 findings by domain importance, mirroring the
// strong-id ranking this rule was ported from.
type severityRank int

const (
	rankLow severityRank = iota
	rankMedium
	rankHigh
)

// severity maps the rank onto go-finding severities the same way the
// branching-flow STRONG_ID findings were converted: high -> error,
// medium -> warning, low -> info.
func (r severityRank) severity() gofinding.Severity {
	switch r {
	case rankHigh:
		return gofinding.SeverityError
	case rankMedium:
		return gofinding.SeverityWarning
	case rankLow:
		return gofinding.SeverityInfo
	default:
		return gofinding.SeverityInfo
	}
}

// highPriorityDomains are ID domains whose primitive use is most likely a
// real correctness risk, so their findings rank highest.
//
//nolint:gochecknoglobals // Immutable lookup table.
var highPriorityDomains = []string{
	"user", "order", "product", "customer", "account",
	"transaction", "payment", "session", "token",
}

// mediumPriorityDomains are ID domains whose primitive use is a moderate
// adoption gap.
//
//nolint:gochecknoglobals // Immutable lookup table.
var mediumPriorityDomains = []string{
	"entity", "resource", "item", "record", "entry",
	"document", "file", "config", "setting",
}

// primitiveIDRank ranks a primitive-ID violation: high-priority domains rank
// one step lower when found in a struct field (data shape, less risk than a
// function signature), same for medium-priority domains.
func primitiveIDRank(name string, kind primitiveLocation) severityRank {
	lower := strings.ToLower(name)

	for _, domain := range highPriorityDomains {
		if strings.HasPrefix(lower, domain) {
			if kind == locationField {
				return rankMedium
			}

			return rankHigh
		}
	}

	for _, domain := range mediumPriorityDomains {
		if strings.HasPrefix(lower, domain) {
			if kind == locationField {
				return rankLow
			}

			return rankMedium
		}
	}

	return rankLow
}

// createPrimitiveIDMessage builds the user-facing BD003 message, adapted from
// the strong-id wording.
func createPrimitiveIDMessage(
	name, currentType, suggestedType, parentName string,
	kind primitiveLocation,
) string {
	kindStr := "parameter"

	switch kind {
	case locationParam:
		kindStr = "parameter"
	case locationField:
		kindStr = "field"
	case locationReturn:
		kindStr = "return value"
	}

	var msg string
	if strings.ToLower(name) == "id" {
		msg = "Change type of " + kindStr + " '" + name + "' from " + currentType +
			" to strong ID type " + suggestedType
	} else {
		msg = "Replace " + kindStr + " '" + name + "' (" + currentType +
			") with strong ID type " + suggestedType
	}

	if parentName != "" {
		return msg + " in " + parentName
	}

	return msg
}
