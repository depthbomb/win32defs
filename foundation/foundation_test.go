package foundation

import (
	"slices"
	"testing"
)

func TestDuplicateHandleOptions(t *testing.T) {
	t.Parallel()

	for name, want := range map[string]Value{
		"DUPLICATE_CLOSE_SOURCE": 1,
		"DUPLICATE_SAME_ACCESS":  2,
	} {
		value, ok := Parse(name)
		if !ok || value != want || !slices.Contains(Names(value), name) {
			t.Fatalf("duplicate option %s: value = %d, found = %v", name, value, ok)
		}
	}

	name, ok := Name(TRUE)
	if !ok || name != "TRUE" {
		t.Fatalf("existing canonical name changed: %q, %v", name, ok)
	}

	if DUPLICATE_CLOSE_SOURCE|DUPLICATE_SAME_ACCESS != 3 {
		t.Fatal("duplicate-handle options do not combine as independent flags")
	}
}

func TestInvalidHandleValue(t *testing.T) {
	t.Parallel()

	if InvalidHandleValue() != ^uintptr(0) {
		t.Fatalf("InvalidHandleValue() = %#x", InvalidHandleValue())
	}
}
