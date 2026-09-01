package winmacro

import "testing"

func TestBytes(t *testing.T) {
	t.Parallel()

	const value = uintptr(0xabcd)

	if LoByte(value) != 0xcd || HiByte(value) != 0xab || LOBYTE(value) != 0xcd || HIBYTE(value) != 0xab {
		t.Fatalf("byte extraction failed for %#x", value)
	}
}

func TestWords(t *testing.T) {
	t.Parallel()

	value := MakeLong(0x1234, 0xabcd)
	if LoWord(uintptr(value)) != 0x1234 || HiWord(uintptr(value)) != 0xabcd {
		t.Fatalf("word round trip failed for %#x", value)
	}
}

func TestMessageParameters(t *testing.T) {
	t.Parallel()

	value := MakeWParam(0xffff, 0xff88)
	if value != MakeLParam(0xffff, 0xff88) || value != MakeLResult(0xffff, 0xff88) {
		t.Fatalf("packed parameter values differ: %#x", value)
	}

	if GetKeyStateWParam(value) != 0xffff || GetNCHitTestWParam(value) != -1 {
		t.Fatalf("low-word extraction failed for %#x", value)
	}

	if GetXButtonWParam(value) != 0xff88 || GetWheelDeltaWParam(value) != -120 {
		t.Fatalf("high-word extraction failed for %#x", value)
	}

	if GET_X_LPARAM(value) != -1 || GET_Y_LPARAM(value) != -120 {
		t.Fatalf("coordinate extraction failed for %#x", value)
	}

	if MAKEWPARAM(0xffff, 0xff88) != value || MAKELPARAM(0xffff, 0xff88) != value || MAKELRESULT(0xffff, 0xff88) != value {
		t.Fatalf("exact-name packing failed for %#x", value)
	}
}

func TestRGB(t *testing.T) {
	t.Parallel()

	value := RGB(0x12, 0x34, 0x56)
	if Red(value) != 0x12 || Green(value) != 0x34 || Blue(value) != 0x56 {
		t.Fatalf("RGB round trip failed for %#x", value)
	}

	if GetRValue(value) != 0x12 || GetGValue(value) != 0x34 || GetBValue(value) != 0x56 {
		t.Fatalf("native-name RGB extraction failed for %#x", value)
	}

	if PaletteRGB(0x12, 0x34, 0x56) != 0x02563412 || PaletteIndex(0x1234) != 0x01001234 {
		t.Fatal("palette COLORREF construction failed")
	}
}

func TestCTLCode(t *testing.T) {
	t.Parallel()

	got := CTLCode(0x22, 0x900, 2, 1)
	if got != 0x00226402 {
		t.Fatalf("CTLCode() = %#x, want %#x", got, uint32(0x00226402))
	}

	if DeviceTypeFromCTLCode(got) != 0x22 || AccessFromCTLCode(got) != 1 || FunctionFromCTLCode(got) != 0x900 || MethodFromCTLCode(got) != 2 {
		t.Fatalf("CTL code decomposition failed for %#x", got)
	}

	if DEVICE_TYPE_FROM_CTL_CODE(got) != 0x22 {
		t.Fatalf("DEVICE_TYPE_FROM_CTL_CODE() failed for %#x", got)
	}
}

func TestIntegerResources(t *testing.T) {
	t.Parallel()

	resource := MakeIntResource(32512)
	if resource != 32512 || !IsIntResource(resource) || !IS_INTRESOURCE(resource) {
		t.Fatalf("integer resource conversion failed for %#x", resource)
	}

	if IsIntResource(0x10000) {
		t.Fatal("pointer-like value was classified as an integer resource")
	}

	if MAKEINTRESOURCE(1) != 1 || MAKEINTRESOURCEA(2) != 2 || MAKEINTRESOURCEW(3) != 3 {
		t.Fatal("exact-name integer resource conversion failed")
	}
}

func TestLocaleIdentifiers(t *testing.T) {
	t.Parallel()

	languageID := MakeLangID(0x09, 0x01)
	if languageID != 0x0409 || PrimaryLangID(languageID) != 0x09 || SubLangID(languageID) != 0x01 {
		t.Fatalf("language identifier round trip failed for %#x", languageID)
	}

	localeID := MakeLCID(languageID, 0x02)
	if localeID != 0x00020409 || LangIDFromLCID(localeID) != languageID || SortIDFromLCID(localeID) != 0x02 {
		t.Fatalf("locale identifier round trip failed for %#x", localeID)
	}

	if MAKELANGID(0x09, 0x01) != languageID || MAKELCID(languageID, 0x02) != localeID {
		t.Fatal("exact-name locale construction failed")
	}
}
