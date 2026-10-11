// Package testdata exercises named return values: a primitive-typed ID named return
// is flagged; unnamed returns are skipped (no name to match).
package testdata

// flagged: named return orderId.
func nextOrder() (orderId string) { return }

// flagged: named return alongside other named returns.
func splitUser() (userId int64, name string, err error) { return }

// not flagged: unnamed return has no name to match.
func shadow() string { return "" }
