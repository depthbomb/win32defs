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
