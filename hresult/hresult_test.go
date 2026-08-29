package hresult

import "testing"

func TestClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		code      Code
		succeeded bool
	}{
		{name: "S_OK", code: S_OK, succeeded: true},
		{name: "S_FALSE", code: S_FALSE, succeeded: true},
		{name: "E_FAIL", code: E_FAIL, succeeded: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := Succeeded(test.code); got != test.succeeded {
				t.Fatalf("Succeeded(%#x) = %v, want %v", uint32(test.code), got, test.succeeded)
			}

			if got := Failed(test.code); got == test.succeeded {
				t.Fatalf("Failed(%#x) = %v, want %v", uint32(test.code), got, !test.succeeded)
			}
		})
	}
}

func TestFromWin32(t *testing.T) {
	t.Parallel()

	if got := FromWin32(0); got != S_OK {
		t.Fatalf("FromWin32(0) = %#x, want %#x", uint32(got), uint32(S_OK))
	}

	if got := FromWin32(5); got != E_ACCESSDENIED {
		want := E_ACCESSDENIED

		t.Fatalf("FromWin32(5) = %#x, want %#x", uint32(got), uint32(want))
	}
}

func TestFields(t *testing.T) {
	t.Parallel()

	if got := CodeOf(E_ACCESSDENIED); got != 5 {
		t.Fatalf("CodeOf(E_ACCESSDENIED) = %d, want 5", got)
	}

	if got := FacilityOf(E_ACCESSDENIED); got != uint16(facilityWin32) {
		t.Fatalf("FacilityOf(E_ACCESSDENIED) = %d, want %d", got, facilityWin32)
	}

	if got := SeverityOf(E_ACCESSDENIED); got != 1 {
		t.Fatalf("SeverityOf(E_ACCESSDENIED) = %d, want 1", got)
	}
}

func TestLookupRoundTrip(t *testing.T) {
	t.Parallel()

	name, ok := Name(E_ACCESSDENIED)
	if !ok {
		t.Fatal("E_ACCESSDENIED has no name")
	}

	code, ok := Parse(name)
	if !ok || code != E_ACCESSDENIED {
		t.Fatalf("Parse(%q) = %s, %v", name, code, ok)
	}
}

func TestCanonicalSuccessName(t *testing.T) {
	t.Parallel()

	name, ok := Name(S_OK)
	if !ok || name != "S_OK" {
		t.Fatalf("Name(S_OK) = %q, %v", name, ok)
	}
}

func TestExactMacroHelpers(t *testing.T) {
	t.Parallel()

	if got := HRESULT_FROM_WIN32(5); got != E_ACCESSDENIED {
		t.Fatalf("HRESULT_FROM_WIN32(5) = %#x", uint32(got))
	}

	if got := MAKE_HRESULT(1, 7, 5); got != E_ACCESSDENIED {
		t.Fatalf("MAKE_HRESULT(1, 7, 5) = %#x", uint32(got))
	}

	failed := E_FAIL
	succeeded := S_OK
	if !IS_ERROR(uint32(failed)) || IS_ERROR(uint32(succeeded)) {
		t.Fatal("IS_ERROR classified an HRESULT incorrectly")
	}
}

func TestErrorConversion(t *testing.T) {
	t.Parallel()

	if err := S_FALSE.Err(); err != nil {
		t.Fatalf("S_FALSE.Err() = %v, want nil", err)
	}

	if err := E_FAIL.Err(); err == nil {
		t.Fatal("E_FAIL.Err() = nil")
	}
}
