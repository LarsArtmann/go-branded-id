package main

import (
	"slices"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argv     []string
		wantFlag []string
		wantPath []string
	}{
		{
			name:     "flags before paths",
			argv:     []string{"-fix", "-rules", "BD001,BD003", "."},
			wantFlag: []string{"-fix", "-rules", "BD001,BD003"},
			wantPath: []string{"."},
		},
		{
			name:     "flags after paths keep working",
			argv:     []string{".", "-rules", "BD001", "-fix", "."},
			wantFlag: []string{"-rules", "BD001", "-fix"},
			wantPath: []string{".", "."},
		},
		{
			name:     "attached value flag takes no next token",
			argv:     []string{"-rules=BD003", "x.go"},
			wantFlag: []string{"-rules=BD003"},
			wantPath: []string{"x.go"},
		},
		{
			name:     "double dash ends flag parsing",
			argv:     []string{"--", "-rules"},
			wantFlag: []string{},
			wantPath: []string{"-rules"},
		},
		{
			name:     "single dash is a path",
			argv:     []string{"-"},
			wantFlag: []string{},
			wantPath: []string{"-"},
		},
		{
			name:     "value flag at the end stays a flag",
			argv:     []string{"x.go", "-rules"},
			wantFlag: []string{"-rules"},
			wantPath: []string{"x.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotFlag, gotPath := splitArgs(tt.argv)
			if !slices.Equal(gotFlag, tt.wantFlag) {
				t.Errorf("splitArgs() flags = %v, want %v", gotFlag, tt.wantFlag)
			}

			if !slices.Equal(gotPath, tt.wantPath) {
				t.Errorf("splitArgs() paths = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestResolvePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  []string
		want []string
	}{
		{name: "no path defaults to the working directory", raw: nil, want: []string{"."}},
		{name: "empty explicit list also defaults", raw: []string{}, want: []string{"."}},
		{name: "bare pattern", raw: []string{"..."}, want: []string{"."}},
		{name: "dot pattern", raw: []string{"./..."}, want: []string{"."}},
		{name: "subtree pattern keeps its root", raw: []string{"pkg/..."}, want: []string{"pkg"}},
		{
			name: "absolute pattern keeps its root",
			raw:  []string{"/abs/..."},
			want: []string{"/abs"},
		},
		{
			name: "root pattern becomes the filesystem root",
			raw:  []string{"/..."},
			want: []string{"/"},
		},
		{
			name: "plain paths stay themselves",
			raw:  []string{"x.go", "pkg"},
			want: []string{"x.go", "pkg"},
		},
		{
			name: "duplicates collapse in order",
			raw:  []string{".", "./...", "pkg", "pkg"},
			want: []string{".", "pkg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := resolvePaths(tt.raw); !slices.Equal(got, tt.want) {
				t.Errorf("resolvePaths(%v) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}
