package winsock

import "testing"

func TestExpandedAliasesPreserveCanonicalNames(t *testing.T) {
	t.Parallel()

	name, ok := Name(-1)
	if !ok || name != "SOCKET_ERROR" {
		t.Fatalf("Name(-1) = %q, %v", name, ok)
	}
}
