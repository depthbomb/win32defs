package foundation

import "testing"

func TestInvalidHandleValue(t *testing.T) {
	t.Parallel()

	if InvalidHandleValue() != ^uintptr(0) {
		t.Fatalf("InvalidHandleValue() = %#x", InvalidHandleValue())
	}
}
