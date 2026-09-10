package catalog

import "testing"

func TestDuplicateHandleFamily(t *testing.T) {
	t.Parallel()

	family := Family{
		Package:   "foundation",
		Namespace: "Windows.Win32.Foundation",
		Name:      "DUPLICATE_HANDLE_OPTIONS",
	}
	for _, name := range []string{"DUPLICATE_CLOSE_SOURCE", "DUPLICATE_SAME_ACCESS"} {
		definition, ok := LookupInFamily(family, name)
		if !ok || !definition.Flags || definition.Kind != "uint32" {
			t.Fatalf("missing duplicate-handle flag metadata: %+v", definition)
		}
	}

	text, ok := FormatFlags(family, 3)
	if !ok || text != "DUPLICATE_CLOSE_SOURCE | DUPLICATE_SAME_ACCESS" {
		t.Fatalf("FormatFlags = %q, %v", text, ok)
	}
}
