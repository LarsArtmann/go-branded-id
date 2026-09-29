// Command brandid-lint scans Go files for brand types used with
// id.ID[Brand, Value] that are missing their Name() string method (rule
// BD001) and prints one go-finding finding per violation.
//
// Usage:
//
//	brandid-lint [flags] <path>...
//
// Exit codes: 0 = no findings, 1 = findings remain, 2 = invocation or
// runtime error.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/larsartmann/go-branded-id/linter"
	gofinding "github.com/larsartmann/go-finding"
)

// errUnknownFormat is the sentinel for an unsupported -format value.
var errUnknownFormat = errors.New("unknown format: want text or sarif")

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is the testable entry point. It parses flags, scans the given paths,
// prints findings in the requested format, and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	flagSet := flag.NewFlagSet(linter.ToolName, flag.ContinueOnError)
	flagSet.SetOutput(stderr)

	format := flagSet.String("format", "text", "output format: text or sarif")
	fix := flagSet.Bool(
		"fix",
		false,
		"apply repairs: insert the suggested Name() stub for every unsuppressed BD001 finding (dry-run by default)",
	)

	if err := flagSet.Parse(args); err != nil {
		return 2
	}

	paths := flagSet.Args()
	if len(paths) < 1 {
		printUsage(stderr, flagSet)

		return 2
	}

	if *fix {
		inserted, fixErr := repairAll(paths)
		if fixErr != nil {
			_, _ = fmt.Fprintf(stderr, "error: %v\n", fixErr)

			return 2
		}

		if inserted > 0 {
			_, _ = fmt.Fprintf(
				stderr,
				"%s: inserted %d Name() stub(s)\n",
				linter.ToolName,
				inserted,
			)
		}
	}

	findings, runErr := detectAll(paths)
	if runErr != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", runErr)

		return 2
	}

	if printErr := printFindings(stdout, findings, *format); printErr != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", printErr)

		return 2
	}

	if len(findings) > 0 {
		return 1
	}

	return 0
}

// detectAll scans every path and merges the findings.
func detectAll(paths []string) ([]gofinding.Finding, error) {
	var findings []gofinding.Finding

	for _, path := range paths {
		pathFindings, err := linter.DetectPath(path)
		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", path, err)
		}

		findings = append(findings, pathFindings...)
	}

	return findings, nil
}

// repairAll applies the BD001 repair to every path and returns the number of
// inserted Name() stubs.
func repairAll(paths []string) (int, error) {
	inserted := 0

	for _, path := range paths {
		count, err := linter.RepairPath(path)
		if err != nil {
			return inserted, fmt.Errorf("repair %s: %w", path, err)
		}

		inserted += count
	}

	return inserted, nil
}

// printFindings writes findings in the requested format.
func printFindings(w io.Writer, findings []gofinding.Finding, format string) error {
	switch format {
	case "text":
		if err := gofinding.FormatTextRich(w, findings); err != nil {
			return fmt.Errorf("format text: %w", err)
		}

		return nil
	case "sarif":
		report := gofinding.NewReportFromFindings(
			gofinding.ToolInfo{Name: linter.ToolName, Version: linter.Version},
			findings,
		)

		sarif, err := report.ToSARIF()
		if err != nil {
			return fmt.Errorf("render sarif: %w", err)
		}

		if _, err := w.Write(sarif); err != nil {
			return fmt.Errorf("write sarif: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("%w: %q", errUnknownFormat, format)
	}
}

func printUsage(w io.Writer, flagSet *flag.FlagSet) {
	_, _ = fmt.Fprintf(w, "Usage: %s [flags] <path>...\n", linter.ToolName)
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(
		w,
		"Scans Go files for brand types missing their Name() string method (BD001).",
	)
	_, _ = fmt.Fprintln(w, "Exit codes: 0 = clean, 1 = findings, 2 = error.")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Flags:")

	flagSet.PrintDefaults()
}
