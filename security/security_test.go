package security

import "testing"

func TestNTAuthority(t *testing.T) {
	t.Parallel()

	want := SIDIdentifierAuthority{0, 0, 0, 0, 0, 5}
	if SECURITY_NT_AUTHORITY != want {
		t.Fatalf("SECURITY_NT_AUTHORITY = %#v", SECURITY_NT_AUTHORITY)
	}
}
