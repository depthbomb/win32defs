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
