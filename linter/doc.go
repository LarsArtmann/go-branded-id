// Package linter implements brandid-lint: a static analyzer for
// github.com/larsartmann/go-branded-id with two semantically distinct rule
// families:
//
//   - Brand hygiene (default): BD001 flags brand types used with
//     id.ID[Brand, Value] that are missing their Name() string method;
//     BD002 flags suppression directives that cannot do their job.
//   - Brand adoption (opt-in): BD003 flags primitive-typed identifiers —
//     string/int/uint/bool parameters, named returns, and struct fields
//     whose name marks them as an ID ("userId", "order_id", "PID") — that
//     should become branded ID types.
//
// A brand type is identified by two signals:
//   - It is declared as an empty struct.
//   - It is used as the first type argument of id.ID[...] (qualified or
//     same-package).
//
// Both signals must hold, so empty structs that never appear in an id.ID
// instantiation are never flagged. Findings are emitted as go-finding
// Finding values positioned at the type declaration, where the fix lands.
//
// # Rule sets: BD003 is opt-in
//
// The default rule set is BD001 plus its BD002 companion; every existing
// consumer (BuildFlow, the CLI without -rules) keeps byte-identical behavior
// across upgrades. BD003 joins only when a scan selects it — the CLI's
// -rules flag ("brandid-lint -rules BD001,BD003 ."), or RuleSet-based
// library calls (DetectPathWithRules). The opt-in default is deliberate:
// brandid-lint runs fleet-wide, and repos that deliberately silenced the old
// branching-flow strong-id rule must not be flooded by an upgrade. BD002 has
// no entry in the rule set: directive-integrity findings are emitted
// automatically for every active rule, and naming BD002 in -rules is an
// accepted no-op, as is the "all" alias for BD001+BD003.
//
// # BD003 findings
//
// BD003 is detection only. A real fix rewrites the declaration and every
// call site, which is a codemod's job, not a lint repair, so Repair never
// touches BD003 findings. Every finding carries its actionable payload in
// metadata instead: suggestedTypeName, suggestedBrand, importPath, and a
// complete typeDefinition to paste:
//
//	type UserID = id.ID[UserBrand, string]
//
//	type UserBrand struct{}
//
// Severities mirror the old strong-id ranking: high-priority domains
// (user, order, product, ...) in signatures map to error, medium to warning,
// the rest to info. The BD003 scan skips files under cmd/, ui/, output,
// test-helper, and tooling paths plus _test.go files (ported from the
// strong-id exclusion table so finding parity holds); BD001 keeps scanning
// every file it always scanned. Method declarations named String or Error
// are exempt: interface contracts must stay primitive.
//
// # Suppression (rule BD002)
//
// A brand that deliberately skips Name() (for example when String() output
// is a storage key), or a primitive ID that must stay primitive, is
// suppressed in source, right where the decision lives:
//
//	//brandid-lint:ignore(BD001) event stream marker: String() is the stream name
//	type StreamMarker struct{}
//
//	//brandid-lint:ignore(BD003) external API contract, name is fixed
//	func keep(userId string)
//
// A directive names its rule (BD001 or BD003 are suppressible) and requires
// a reason. BD001 directives sit on the brand declaration line or as the
// last line of the comment group directly above it. BD003 directives sit on
// the primitive-ID declaration line, directly above it, or on the enclosing
// function or type declaration (a doc directive covers all of that
// declaration's ID params, returns, and fields). Suppressed declarations
// produce no finding and are never repaired. A directive that cannot do its
// job — malformed syntax, missing reason, unknown rule, stale, duplicate,
// or suppressing nothing — is itself reported as a BD002 finding, because a
// silent no-op directive is worse than none. Directives of a rule that is
// not active in the current scan are neither honored nor policed; malformed
// directives stay loud in every mode.
//
// # Repair
//
// Repair (and the CLI's -fix flag) inserts the suggested Name() stub after
// the type declaration of every unsuppressed BD001 finding, as byte-level
// insertion edits applied back to front so nothing shifts. It is idempotent
// (a second run finds nothing to insert), never inserts a stub next to an
// existing Name() method regardless of what that method returns, preserves
// file permissions, never touches suppressed brands, and never repairs any
// file under a testdata path — parse-only fixtures must never be rewritten,
// no matter how the path reaches the tool.
//
// Detection is purely syntactic (go/parser): files that fail to parse are
// skipped, and the scanner never resolves imports across packages. The
// directories .git, vendor, node_modules, and testdata are never scanned
// (testdata holds intentionally incomplete fixtures, by the same convention
// go vet, staticcheck, and golangci-lint follow).
package linter
