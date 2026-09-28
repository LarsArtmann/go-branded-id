package linter

import "strings"

// suggestName derives a suggested Name() return value from a brand type name
// by repeatedly stripping the "Brand" and "ID" suffixes and a leading "T"
// generic-type prefix (only when followed by an uppercase letter, so
// "TProduct" becomes "Product" but a real word like "Tenant" is left intact).
func suggestName(brandName string) string {
	name := brandName

	for {
		prev := name
		name = strings.TrimSuffix(name, "Brand")
		name = strings.TrimSuffix(name, "ID")

		if name == prev {
			break
		}
	}

	if len(name) > 1 && name[0] == 'T' && name[1] >= 'A' && name[1] <= 'Z' {
		name = name[1:]
	}

	if name == "" {
		return brandName
	}

	return name
}
