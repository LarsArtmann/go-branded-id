package id

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// The dual-JSON architecture relies on two build-tagged file pairs that are
// structurally identical except for the build constraint and the json import
// path. A recurring hazard (documented in commit history as the "goimports v1
// corruption hazard") is that goimports, when re-adding the json import from
// scratch, picks encoding/json/v2 for the v1 file — silently breaking the
// default build mode. These tests lock down the contract so any drift fails
// loudly in both CI modes.

const (
	jsonV1ImplFile      = "id_json_v1.go"
	jsonV2ImplFile      = "id_json_v2.go"
	jsonV1HelperTest    = "json_helpers_v1_test.go"
	jsonV2HelperTest    = "json_helpers_v2_test.go"
	jsonV1ImportPath    = "\"encoding/json\""
	jsonV2ImportPath    = "\"encoding/json/v2\""
	jsonV1BuildConstrnt = "//go:build !goexperiment.jsonv2"
	jsonV2BuildConstrnt = "//go:build goexperiment.jsonv2"
)

// TestDualJSONContract_Imports locks the import split: the v1 files MUST import
// encoding/json (never v2) and the v2 files MUST import encoding/json/v2.
// This is the direct guard against the goimports corruption hazard.
func TestDualJSONContract_Imports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		filename    string
		mustContain string
		mustOmit    string
	}{
		{"v1 impl uses json v1 import", jsonV1ImplFile, jsonV1ImportPath, jsonV2ImportPath},
		{"v1 helper uses json v1 import", jsonV1HelperTest, jsonV1ImportPath, jsonV2ImportPath},
		{"v2 impl uses json v2 import", jsonV2ImplFile, jsonV2ImportPath, jsonV1ImportPath},
		{"v2 helper uses json v2 import", jsonV2HelperTest, jsonV2ImportPath, jsonV1ImportPath},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			content, err := readSource(tc.filename)
			if err != nil {
				t.Fatalf("read %s: %v", tc.filename, err)
			}

			if !strings.Contains(content, tc.mustContain) {
				t.Errorf("%s must contain %s", tc.filename, tc.mustContain)
			}

			if strings.Contains(content, tc.mustOmit) {
				t.Errorf(
					"%s must NOT contain %s (goimports corruption hazard)",
					tc.filename,
					tc.mustOmit,
				)
			}
		})
	}
}

// TestDualJSONContract_BuildTags locks the build-constraint split so the two
// files of each pair are never compiled together.
func TestDualJSONContract_BuildTags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		filename string
		want     string
	}{
		{"v1 impl build tag", jsonV1ImplFile, jsonV1BuildConstrnt},
		{"v1 helper build tag", jsonV1HelperTest, jsonV1BuildConstrnt},
		{"v2 impl build tag", jsonV2ImplFile, jsonV2BuildConstrnt},
		{"v2 helper build tag", jsonV2HelperTest, jsonV2BuildConstrnt},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			content, err := readSource(tc.filename)
			if err != nil {
				t.Fatalf("read %s: %v", tc.filename, err)
			}

			if !strings.Contains(content, tc.want) {
				t.Errorf("%s must contain build constraint %s", tc.filename, tc.want)
			}
		})
	}
}

// TestDualJSONContract_StructuralParity asserts each v1/v2 pair is byte-identical
// after normalizing away the two intentional differences (build constraint line
// and json import path). If a future edit changes one half of a pair without the
// other, this fails — preventing the two modes from silently diverging.
func TestDualJSONContract_StructuralParity(t *testing.T) {
	t.Parallel()

	pairs := []struct {
		name         string
		v1, v2       string
		v1Tag, v2Tag string
	}{
		{
			"implementation files",
			jsonV1ImplFile,
			jsonV2ImplFile,
			jsonV1BuildConstrnt,
			jsonV2BuildConstrnt,
		},
		{
			"test helper files",
			jsonV1HelperTest,
			jsonV2HelperTest,
			jsonV1BuildConstrnt,
			jsonV2BuildConstrnt,
		},
	}

	for _, pair := range pairs {
		t.Run(pair.name, func(t *testing.T) {
			t.Parallel()

			v1, err := readSource(pair.v1)
			if err != nil {
				t.Fatalf("read %s: %v", pair.v1, err)
			}

			v2, err := readSource(pair.v2)
			if err != nil {
				t.Fatalf("read %s: %v", pair.v2, err)
			}

			normalizedV1 := normalizeForParity(v1, pair.v1Tag)
			normalizedV2 := normalizeForParity(v2, pair.v2Tag)

			if normalizedV1 != normalizedV2 {
				t.Errorf(
					"%s and %s diverged after normalization (build tag + json import):\n--- v1 ---\n%s\n--- v2 ---\n%s",
					pair.v1,
					pair.v2,
					normalizedV1,
					normalizedV2,
				)
			}
		})
	}
}

// normalizeForParity removes the build-constraint line and collapses the json
// import path to a single canonical form so only the two pairs can be compared.
func normalizeForParity(content, buildConstraint string) string {
	lines := strings.Split(content, "\n")

	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == buildConstraint {
			continue
		}

		filtered = append(filtered, line)
	}

	joined := strings.Join(filtered, "\n")
	joined = strings.ReplaceAll(joined, jsonV2ImportPath, jsonV1ImportPath)

	return joined
}

// TestJSONByteEquivalence asserts the EXACT byte output of MarshalJSON for
// representative values. Because CI runs the full suite in BOTH json modes
// (default and GOEXPERIMENT=jsonv2), this test effectively verifies that v1 and
// v2 produce byte-identical output — if either mode diverges, that mode's run
// fails. This is the practical equivalence guarantee for the dual-mode design.
func TestJSONByteEquivalence(t *testing.T) {
	t.Parallel()

	t.Run("string non-zero", func(t *testing.T) {
		t.Parallel()

		id := NewID[StringBrand, string]("abc123")

		got, err := id.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}

		assertBytes(t, got, []byte(`"abc123"`))
	})

	t.Run("string zero is null", func(t *testing.T) {
		t.Parallel()

		var id ID[StringBrand, string]

		got, err := id.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}

		assertBytes(t, got, []byte("null"))
	})

	t.Run("int64 non-zero", func(t *testing.T) {
		t.Parallel()

		id := NewID[Int64Brand, int64](42)

		got, err := id.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}

		assertBytes(t, got, []byte("42"))
	})

	t.Run("int64 zero is null", func(t *testing.T) {
		t.Parallel()

		var id ID[Int64Brand, int64]

		got, err := id.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}

		assertBytes(t, got, []byte("null"))
	})

	t.Run("uint64 non-zero", func(t *testing.T) {
		t.Parallel()

		id := NewID[Uint64Brand, uint64](42)

		got, err := id.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}

		assertBytes(t, got, []byte("42"))
	})

	t.Run("int32 non-zero", func(t *testing.T) {
		t.Parallel()

		id := NewID[Int32Brand, int32](42)

		got, err := id.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}

		assertBytes(t, got, []byte("42"))
	})
}

func assertBytes(t *testing.T, got, want []byte) {
	t.Helper()

	if !bytes.Equal(got, want) {
		t.Errorf("expected %q, got %q", want, got)
	}
}

// readSource loads a source file as a string for content assertions. Tests run
// from the package directory, so relative paths resolve correctly.
func readSource(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("read source %s: %w", filename, err)
	}

	return string(content), nil
}
