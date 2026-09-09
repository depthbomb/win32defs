package facility

import "testing"

func TestNTBitIsNotAFacilityIdentifier(t *testing.T) {
	t.Parallel()

	if FACILITY_NT_BIT != uint32(0x10000000) {
		t.Fatalf("FACILITY_NT_BIT = %#x", FACILITY_NT_BIT)
	}

	if _, ok := Parse("FACILITY_NT_BIT"); ok {
		t.Fatal("32-bit HRESULT mask was parsed as a 16-bit facility ID")
	}
}
