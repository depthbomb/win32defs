package winmacro

import "testing"

func TestWords(t *testing.T) {
	t.Parallel()

	value := MakeLong(0x1234, 0xabcd)
	if LoWord(uintptr(value)) != 0x1234 || HiWord(uintptr(value)) != 0xabcd {
		t.Fatalf("word round trip failed for %#x", value)
	}
}

func TestRGB(t *testing.T) {
	t.Parallel()

	value := RGB(0x12, 0x34, 0x56)
	if Red(value) != 0x12 || Green(value) != 0x34 || Blue(value) != 0x56 {
		t.Fatalf("RGB round trip failed for %#x", value)
	}
}

func TestCTLCode(t *testing.T) {
	t.Parallel()

	if got := CTLCode(0x22, 0x900, 0, 0); got != 0x00222400 {
		t.Fatalf("CTLCode() = %#x, want %#x", got, uint32(0x00222400))
	}
}
