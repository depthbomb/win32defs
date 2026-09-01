package guid

import "testing"

func TestParseRoundTrip(t *testing.T) {
	t.Parallel()

	const text = "00000000-0000-0000-c000-000000000046"

	value, err := Parse("{" + text + "}")
	if err != nil {
		t.Fatal(err)
	}

	if got := value.String(); got != text {
		t.Fatalf("String() = %q, want %q", got, text)
	}
}

func TestGeneratedIUnknown(t *testing.T) {
	t.Parallel()

	value, ok := Lookup("IID_IUnknown")
	if !ok {
		t.Fatal("IID_IUnknown is missing")
	}

	if value.String() != "00000000-0000-0000-c000-000000000046" {
		t.Fatalf("unexpected IID_IUnknown: %s", value)
	}
}

func TestCOMClassAliases(t *testing.T) {
	t.Parallel()

	if CLSID_ShellLink != GUID_ShellLink {
		t.Fatal("CLSID_ShellLink and GUID_ShellLink differ")
	}
}

func TestGeneratedName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value GUID
	}{
		{name: "CLSID_AACMFTEncoder", value: CLSID_AACMFTEncoder},
		{name: "CLSID_GUID_NULL", value: CLSID_GUID_NULL},
		{name: "IID_IUnknown", value: IID_IUnknown},
		{name: "IID__WMPOCXEvents", value: IID__WMPOCXEvents},
	}

	for _, test := range tests {
		got, ok := Name(test.value)
		if !ok {
			t.Fatalf("Name(%s) did not find a name", test.value)
		}

		if got != test.name {
			t.Fatalf("Name(%s) = %q, want %q", test.value, got, test.name)
		}
	}
}

func TestGeneratedNameMiss(t *testing.T) {
	t.Parallel()

	_, ok := Name(GUID{Data1: 0xDEADBEEF})
	if ok {
		t.Fatal("Name returned a match for an unknown GUID")
	}
}
