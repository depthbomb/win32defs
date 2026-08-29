package propertykey

import "testing"

func TestParseRoundTrip(t *testing.T) {
	t.Parallel()

	text := PKEY_Title.String()
	parsed, err := Parse(text)
	if err != nil {
		t.Fatal(err)
	}

	if parsed != PKEY_Title {
		t.Fatalf("Parse(%q) = %v, want %v", text, parsed, PKEY_Title)
	}
}
