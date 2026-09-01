package catalog

import "testing"

func TestLookup(t *testing.T) {
	t.Parallel()

	definition, ok := Lookup("winerror", "ERROR_ACCESS_DENIED")
	if !ok {
		t.Fatal("ERROR_ACCESS_DENIED is missing")
	}

	if definition.Namespace == "" || definition.Value == "" {
		t.Fatalf("incomplete definition: %#v", definition)
	}
}

func TestLookupCoverage(t *testing.T) {
	t.Parallel()

	for definition := range Definitions() {
		got, ok := Lookup(definition.Package, definition.Name)
		if !ok {
			t.Fatalf("Lookup(%q, %q) did not find a definition", definition.Package, definition.Name)
		}

		if got != definition {
			t.Fatalf("Lookup(%q, %q) = %#v, want %#v", definition.Package, definition.Name, got, definition)
		}
	}
}
