package process

import "testing"

func TestTokenPseudoHandles(t *testing.T) {
	t.Parallel()

	if GetCurrentProcessToken() != ^uintptr(3) {
		t.Fatal("unexpected current-process token pseudo-handle")
	}

	if GetCurrentThreadToken() != ^uintptr(4) {
		t.Fatal("unexpected current-thread token pseudo-handle")
	}

	if GetCurrentThreadEffectiveToken() != ^uintptr(5) {
		t.Fatal("unexpected effective-thread token pseudo-handle")
	}
}
