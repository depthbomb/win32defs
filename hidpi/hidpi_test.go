package hidpi

import "testing"

func TestContextUintptr(t *testing.T) {
	t.Parallel()

	if got := DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2.Uintptr(); got != ^uintptr(3) {
		t.Fatalf("per-monitor v2 context = %#x, want %#x", got, ^uintptr(3))
	}

	if got := DPI_AWARENESS_CONTEXT_UNAWARE.Uintptr(); got != ^uintptr(0) {
		t.Fatalf("unaware context = %#x, want %#x", got, ^uintptr(0))
	}
}
