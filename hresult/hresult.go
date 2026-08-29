// Package hresult provides HRESULT values and helpers corresponding to common
// Win32 error-handling macros.
package hresult

import "fmt"

// Code is a signed 32-bit HRESULT value.
type Code int32

// Error is a failed HRESULT represented as a Go error.
type Error struct {
	Code Code
}

const facilityWin32 uint32 = 7

// Succeeded reports whether code represents success.
func Succeeded(code Code) bool {
	return code >= 0
}

// Failed reports whether code represents failure.
func Failed(code Code) bool {
	return code < 0
}

// Make constructs an HRESULT from its severity, facility, and code fields.
func Make(severity uint32, facility uint32, code uint32) Code {
	value := (severity << 31) | (facility << 16) | code

	return Code(int32(value))
}

// FromWin32 maps a Win32 system error code to an HRESULT.
func FromWin32(code uint32) Code {
	signed := int32(code)
	if signed <= 0 {
		return Code(signed)
	}

	value := (code & 0xffff) | (facilityWin32 << 16) | 0x80000000

	return Code(int32(value))
}

// CodeOf extracts the low 16-bit status code from an HRESULT.
func CodeOf(code Code) uint16 {
	return uint16(uint32(code))
}

// FacilityOf extracts the facility field from an HRESULT.
func FacilityOf(code Code) uint16 {
	return uint16((uint32(code) >> 16) & 0x1fff)
}

// SeverityOf extracts the severity bit from an HRESULT.
func SeverityOf(code Code) uint8 {
	return uint8(uint32(code) >> 31)
}

// FromNTStatus maps an NTSTATUS bit pattern to an HRESULT.
func FromNTStatus(code uint32) Code {
	return Code(int32(code | 0x10000000))
}

// Win32Code extracts a Win32 system error when code uses FACILITY_WIN32.
func Win32Code(code Code) (uint32, bool) {
	if FacilityOf(code) != uint16(facilityWin32) {
		return 0, false
	}

	return uint32(CodeOf(code)), true
}

// Err returns nil for successful values and a Go error for failed values.
func (code Code) Err() error {
	if Succeeded(code) {
		return nil
	}

	return Error{Code: code}
}

// String returns the canonical symbolic name or a hexadecimal representation.
func (code Code) String() string {
	if name, ok := Name(code); ok {
		return name
	}

	return fmt.Sprintf("HRESULT(0x%08X)", uint32(code))
}

// Error implements error.
func (err Error) Error() string {
	return err.Code.String()
}

// Is compares HRESULT errors by numeric code.
func (err Error) Is(target error) bool {
	other, ok := target.(Error)
	if !ok {
		return false
	}

	return err.Code == other.Code
}

// SUCCEEDED is the exact-name counterpart of the Win32 SUCCEEDED macro.
func SUCCEEDED(code Code) bool {
	return Succeeded(code)
}

// FAILED is the exact-name counterpart of the Win32 FAILED macro.
func FAILED(code Code) bool {
	return Failed(code)
}

// MAKE_HRESULT is the exact-name counterpart of the Win32 MAKE_HRESULT macro.
func MAKE_HRESULT(severity uint32, facility uint32, code uint32) Code {
	return Make(severity, facility, code)
}

// MAKE_SCODE is the exact-name counterpart of the Win32 MAKE_SCODE macro.
func MAKE_SCODE(severity uint32, facility uint32, code uint32) Code {
	return Make(severity, facility, code)
}

// HRESULT_CODE is the exact-name counterpart of the Win32 HRESULT_CODE macro.
func HRESULT_CODE(code Code) uint16 {
	return CodeOf(code)
}

// HRESULT_FACILITY is the exact-name counterpart of the Win32 HRESULT_FACILITY macro.
func HRESULT_FACILITY(code Code) uint16 {
	return FacilityOf(code)
}

// HRESULT_SEVERITY is the exact-name counterpart of the Win32 HRESULT_SEVERITY macro.
func HRESULT_SEVERITY(code Code) uint8 {
	return SeverityOf(code)
}

// HRESULT_FROM_WIN32 is the exact-name counterpart of the Win32 HRESULT_FROM_WIN32 helper.
func HRESULT_FROM_WIN32(code uint32) Code {
	return FromWin32(code)
}

// HRESULT_FROM_NT is the exact-name counterpart of the Win32 HRESULT_FROM_NT macro.
func HRESULT_FROM_NT(code uint32) Code {
	return FromNTStatus(code)
}

// SCODE_CODE is the exact-name counterpart of the Win32 SCODE_CODE macro.
func SCODE_CODE(code Code) uint16 {
	return CodeOf(code)
}

// SCODE_FACILITY is the exact-name counterpart of the Win32 SCODE_FACILITY macro.
func SCODE_FACILITY(code Code) uint16 {
	return FacilityOf(code)
}

// SCODE_SEVERITY is the exact-name counterpart of the Win32 SCODE_SEVERITY macro.
func SCODE_SEVERITY(code Code) uint8 {
	return SeverityOf(code)
}

// IS_ERROR reports whether the high-order severity bit of status is set.
func IS_ERROR(status uint32) bool {
	return status>>31 != 0
}
