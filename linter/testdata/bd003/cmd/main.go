// Command main lives under cmd/ and is excluded from the BD003 scan by the
// ported exclusion table: CLI wiring is presentation code.
package main

// not flagged: cmd/ paths are excluded from BD003.
func main() { render("device-1") }

func render(deviceId string) {}
