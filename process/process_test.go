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

func TestExpandedAliasesPreserveCanonicalNames(t *testing.T) {
	t.Parallel()

	name, ok := Name(Value(0xffffffff))
	if !ok || name != "THREAD_PRIORITY_BELOW_NORMAL" {
		t.Fatalf("Name(0xffffffff) = %q, %v", name, ok)
	}
}
