// Package winmacro provides Go equivalents of common value-oriented Win32
// macros. The functions are platform independent and do not call Windows.
package winmacro

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

// MakeLParam combines two words into an LPARAM-compatible value.
func MakeLParam(low uint16, high uint16) uintptr {
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

// CTLCode constructs an I/O control code.
func CTLCode(deviceType uint32, function uint32, method uint32, access uint32) uint32 {
	return deviceType<<16 | access<<14 | function<<2 | method
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

// CTL_CODE is the exact-name counterpart of the Win32 CTL_CODE macro.
func CTL_CODE(deviceType uint32, function uint32, method uint32, access uint32) uint32 {
	return CTLCode(deviceType, function, method, access)
}
