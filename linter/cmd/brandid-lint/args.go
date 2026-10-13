package main

import (
	"path/filepath"
	"strings"
)

// valueFlags are the flags that consume a separate value token when written
// as "-flag value"; every other flag of this CLI is boolean. The table only
// drives arg permutation: an unknown flag stays a parse error from
// flag.FlagSet, exactly as before.
var valueFlags = map[string]bool{
	"format": true,
	"rules":  true,
}

// splitArgs permutes argv into flag tokens and positional paths so flags may
// appear anywhere. Go's flag package would stop at the first path and then
// treat a trailing "-rules" as one, which is how "brandid-lint . -rules ..."
// turned into a stat call. "--" ends flag parsing; "-" stays a path.
func splitArgs(argv []string) (flagArgs, paths []string) {
	for i := 0; i < len(argv); i++ {
		arg := argv[i]

		switch {
		case arg == "--":
			return flagArgs, append(paths, argv[i+1:]...)
		case arg == "-" || !strings.HasPrefix(arg, "-"):
			paths = append(paths, arg)
		default:
			flagArgs = append(flagArgs, arg)

			name := strings.TrimLeft(arg, "-")
			if valueFlags[name] && !strings.Contains(arg, "=") && i+1 < len(argv) {
				i++
				flagArgs = append(flagArgs, argv[i])
			}
		}
	}

	return flagArgs, paths
}

// resolvePaths expands Go-style package patterns, deduplicates in order, and
// defaults to the working directory: an invocation with no path scans ".",
// like golangci-lint, instead of printing a usage lecture.
func resolvePaths(raw []string) []string {
	if len(raw) == 0 {
		return []string{"."}
	}

	seen := make(map[string]bool, len(raw))
	paths := make([]string, 0, len(raw))

	for _, path := range raw {
		for _, resolved := range expandPattern(path) {
			if !seen[resolved] {
				seen[resolved] = true
				paths = append(paths, resolved)
			}
		}
	}

	return paths
}

// expandPattern maps one argument to its scan roots: "./..." and the bare
// "..." become "." and "pkg/..." becomes "pkg" (the scanner already walks
// directory trees recursively and skips vendor/, testdata/, and .git/).
// Anything else is itself.
func expandPattern(path string) []string {
	slash := filepath.ToSlash(path)

	switch slash {
	case "...", "./...":
		return []string{"."}
	}

	if root, ok := strings.CutSuffix(slash, "/..."); ok {
		if root == "" {
			return []string{"/"}
		}

		return []string{filepath.FromSlash(root)}
	}

	return []string{path}
}
