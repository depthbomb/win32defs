// Package winmacro provides Go equivalents of common value-oriented Win32
// macros. The functions are platform independent and do not call Windows.
package winmacro

// LoByte returns the low-order byte of value.
func LoByte(value uintptr) byte {
	return byte(value)
}

// HiByte returns the high-order byte of the low-order word of value.
func HiByte(value uintptr) byte {
	return byte(value >> 8)
}

// LoWord returns the low-order word of value.
func LoWord(value uintptr) uint16 {
	return uint16(value & 0xffff)
}

// HiWord returns the high-order word of value.
func HiWord(value uintptr) uint16 {
	return uint16((value >> 16) & 0xffff)
}

// MakeWord combines two bytes into a 16-bit value.
func MakeWord(low byte, high byte) uint16 {
	return uint16(low) | uint16(high)<<8
}

// MakeLong combines two words into a 32-bit value.
func MakeLong(low uint16, high uint16) uint32 {
	return uint32(low) | uint32(high)<<16
}

// MakeWParam combines two words into a WPARAM-compatible value.
func MakeWParam(low uint16, high uint16) uintptr {
	return uintptr(MakeLong(low, high))
}

// MakeLParam combines two words into an LPARAM-compatible value.
func MakeLParam(low uint16, high uint16) uintptr {
	return uintptr(MakeLong(low, high))
}

// MakeLResult combines two words into an LRESULT-compatible value.
func MakeLResult(low uint16, high uint16) uintptr {
	return uintptr(MakeLong(low, high))
}

// GetXParam extracts a signed x-coordinate from an LPARAM-compatible value.
func GetXParam(value uintptr) int16 {
	return int16(LoWord(value))
}

// GetYParam extracts a signed y-coordinate from an LPARAM-compatible value.
func GetYParam(value uintptr) int16 {
	return int16(HiWord(value))
}

// GetKeyStateWParam extracts key-state flags from a WPARAM-compatible value.
func GetKeyStateWParam(value uintptr) uint16 {
	return LoWord(value)
}

// GetNCHitTestWParam extracts a signed nonclient hit-test value from a WPARAM-compatible value.
func GetNCHitTestWParam(value uintptr) int16 {
	return int16(LoWord(value))
}

// GetXButtonWParam extracts an X-button identifier from a WPARAM-compatible value.
func GetXButtonWParam(value uintptr) uint16 {
	return HiWord(value)
}

// GetWheelDeltaWParam extracts a signed mouse-wheel delta from a WPARAM-compatible value.
func GetWheelDeltaWParam(value uintptr) int16 {
	return int16(HiWord(value))
}

// RGB constructs a COLORREF value from red, green, and blue components.
func RGB(red byte, green byte, blue byte) uint32 {
	return uint32(red) | uint32(green)<<8 | uint32(blue)<<16
}

// Red extracts the red component from a COLORREF value.
func Red(value uint32) byte {
	return byte(value)
}

// Green extracts the green component from a COLORREF value.
func Green(value uint32) byte {
	return byte(value >> 8)
}

// Blue extracts the blue component from a COLORREF value.
func Blue(value uint32) byte {
	return byte(value >> 16)
}

// GetRValue is the exact-name counterpart of the Win32 GetRValue macro.
func GetRValue(value uint32) byte {
	return Red(value)
}

// GetGValue is the exact-name counterpart of the Win32 GetGValue macro.
func GetGValue(value uint32) byte {
	return Green(value)
}

// GetBValue is the exact-name counterpart of the Win32 GetBValue macro.
func GetBValue(value uint32) byte {
	return Blue(value)
}

// PaletteRGB constructs a palette-relative COLORREF value.
func PaletteRGB(red byte, green byte, blue byte) uint32 {
	return 0x02000000 | RGB(red, green, blue)
}

// PaletteIndex constructs a palette-index COLORREF value.
func PaletteIndex(index uint16) uint32 {
	return 0x01000000 | uint32(index)
}

// CTLCode constructs an I/O control code.
func CTLCode(deviceType uint32, function uint32, method uint32, access uint32) uint32 {
	return deviceType<<16 | access<<14 | function<<2 | method
}

// DeviceTypeFromCTLCode extracts the device-type field from an I/O control code.
func DeviceTypeFromCTLCode(controlCode uint32) uint32 {
	return controlCode >> 16
}

// AccessFromCTLCode extracts the access field from an I/O control code.
func AccessFromCTLCode(controlCode uint32) uint32 {
	return controlCode >> 14 & 0x3
}

// FunctionFromCTLCode extracts the function field from an I/O control code.
func FunctionFromCTLCode(controlCode uint32) uint32 {
	return controlCode >> 2 & 0xfff
}

// MethodFromCTLCode extracts the method field from an I/O control code.
func MethodFromCTLCode(controlCode uint32) uint32 {
	return controlCode & 0x3
}

// MakeIntResource converts an integer identifier to a resource-compatible value.
func MakeIntResource(identifier uint16) uintptr {
	return uintptr(identifier)
}

// IsIntResource reports whether value contains an integer resource identifier.
func IsIntResource(value uintptr) bool {
	return value>>16 == 0
}

// MakeLangID constructs a legacy Windows language identifier.
func MakeLangID(primaryLanguage uint16, sublanguage uint16) uint16 {
	return sublanguage<<10 | primaryLanguage
}

// PrimaryLangID extracts the primary-language field from a language identifier.
func PrimaryLangID(languageID uint16) uint16 {
	return languageID & 0x3ff
}

// SubLangID extracts the sublanguage field from a language identifier.
func SubLangID(languageID uint16) uint16 {
	return languageID >> 10
}

// MakeLCID constructs a legacy Windows locale identifier.
func MakeLCID(languageID uint16, sortID uint16) uint32 {
	return uint32(sortID)<<16 | uint32(languageID)
}

// LangIDFromLCID extracts the language identifier from a locale identifier.
func LangIDFromLCID(localeID uint32) uint16 {
	return uint16(localeID)
}

// SortIDFromLCID extracts the four-bit sort identifier from a locale identifier.
func SortIDFromLCID(localeID uint32) uint16 {
	return uint16(localeID >> 16 & 0xf)
}

// LOBYTE is the exact-name counterpart of the Win32 LOBYTE macro.
func LOBYTE(value uintptr) byte {
	return LoByte(value)
}

// HIBYTE is the exact-name counterpart of the Win32 HIBYTE macro.
func HIBYTE(value uintptr) byte {
	return HiByte(value)
}

// LOWORD is the exact-name counterpart of the Win32 LOWORD macro.
func LOWORD(value uintptr) uint16 {
	return LoWord(value)
}

// HIWORD is the exact-name counterpart of the Win32 HIWORD macro.
func HIWORD(value uintptr) uint16 {
	return HiWord(value)
}

// MAKEWORD is the exact-name counterpart of the Win32 MAKEWORD macro.
func MAKEWORD(low byte, high byte) uint16 {
	return MakeWord(low, high)
}

// MAKELONG is the exact-name counterpart of the Win32 MAKELONG macro.
func MAKELONG(low uint16, high uint16) uint32 {
	return MakeLong(low, high)
}

// MAKEWPARAM is the exact-name counterpart of the Win32 MAKEWPARAM macro.
func MAKEWPARAM(low uint16, high uint16) uintptr {
	return MakeWParam(low, high)
}

// MAKELPARAM is the exact-name counterpart of the Win32 MAKELPARAM macro.
func MAKELPARAM(low uint16, high uint16) uintptr {
	return MakeLParam(low, high)
}

// MAKELRESULT is the exact-name counterpart of the Win32 MAKELRESULT macro.
func MAKELRESULT(low uint16, high uint16) uintptr {
	return MakeLResult(low, high)
}

// GET_X_LPARAM is the exact-name counterpart of the Win32 GET_X_LPARAM macro.
func GET_X_LPARAM(value uintptr) int16 {
	return GetXParam(value)
}

// GET_Y_LPARAM is the exact-name counterpart of the Win32 GET_Y_LPARAM macro.
func GET_Y_LPARAM(value uintptr) int16 {
	return GetYParam(value)
}

// GET_KEYSTATE_WPARAM is the exact-name counterpart of the Win32 GET_KEYSTATE_WPARAM macro.
func GET_KEYSTATE_WPARAM(value uintptr) uint16 {
	return GetKeyStateWParam(value)
}

// GET_NCHITTEST_WPARAM is the exact-name counterpart of the Win32 GET_NCHITTEST_WPARAM macro.
func GET_NCHITTEST_WPARAM(value uintptr) int16 {
	return GetNCHitTestWParam(value)
}

// GET_XBUTTON_WPARAM is the exact-name counterpart of the Win32 GET_XBUTTON_WPARAM macro.
func GET_XBUTTON_WPARAM(value uintptr) uint16 {
	return GetXButtonWParam(value)
}

// GET_WHEEL_DELTA_WPARAM is the exact-name counterpart of the Win32 GET_WHEEL_DELTA_WPARAM macro.
func GET_WHEEL_DELTA_WPARAM(value uintptr) int16 {
	return GetWheelDeltaWParam(value)
}

// PALETTERGB is the exact-name counterpart of the Win32 PALETTERGB macro.
func PALETTERGB(red byte, green byte, blue byte) uint32 {
	return PaletteRGB(red, green, blue)
}

// PALETTEINDEX is the exact-name counterpart of the Win32 PALETTEINDEX macro.
func PALETTEINDEX(index uint16) uint32 {
	return PaletteIndex(index)
}

// CTL_CODE is the exact-name counterpart of the Win32 CTL_CODE macro.
func CTL_CODE(deviceType uint32, function uint32, method uint32, access uint32) uint32 {
	return CTLCode(deviceType, function, method, access)
}

// DEVICE_TYPE_FROM_CTL_CODE is the exact-name counterpart of the Win32 DEVICE_TYPE_FROM_CTL_CODE macro.
func DEVICE_TYPE_FROM_CTL_CODE(controlCode uint32) uint32 {
	return DeviceTypeFromCTLCode(controlCode)
}

// MAKEINTRESOURCE is the exact-name counterpart of the Win32 MAKEINTRESOURCE macro.
func MAKEINTRESOURCE(identifier uint16) uintptr {
	return MakeIntResource(identifier)
}

// MAKEINTRESOURCEA is the exact-name counterpart of the Win32 MAKEINTRESOURCEA macro.
func MAKEINTRESOURCEA(identifier uint16) uintptr {
	return MakeIntResource(identifier)
}

// MAKEINTRESOURCEW is the exact-name counterpart of the Win32 MAKEINTRESOURCEW macro.
func MAKEINTRESOURCEW(identifier uint16) uintptr {
	return MakeIntResource(identifier)
}

// IS_INTRESOURCE is the exact-name counterpart of the Win32 IS_INTRESOURCE macro.
func IS_INTRESOURCE(value uintptr) bool {
	return IsIntResource(value)
}

// MAKELANGID is the exact-name counterpart of the Win32 MAKELANGID macro.
func MAKELANGID(primaryLanguage uint16, sublanguage uint16) uint16 {
	return MakeLangID(primaryLanguage, sublanguage)
}

// PRIMARYLANGID is the exact-name counterpart of the Win32 PRIMARYLANGID macro.
func PRIMARYLANGID(languageID uint16) uint16 {
	return PrimaryLangID(languageID)
}

// SUBLANGID is the exact-name counterpart of the Win32 SUBLANGID macro.
func SUBLANGID(languageID uint16) uint16 {
	return SubLangID(languageID)
}

// MAKELCID is the exact-name counterpart of the Win32 MAKELCID macro.
func MAKELCID(languageID uint16, sortID uint16) uint32 {
	return MakeLCID(languageID, sortID)
}

// LANGIDFROMLCID is the exact-name counterpart of the Win32 LANGIDFROMLCID macro.
func LANGIDFROMLCID(localeID uint32) uint16 {
	return LangIDFromLCID(localeID)
}

// SORTIDFROMLCID is the exact-name counterpart of the Win32 SORTIDFROMLCID macro.
func SORTIDFROMLCID(localeID uint32) uint16 {
	return SortIDFromLCID(localeID)
}
