// Package ntstatus provides NTSTATUS values and classification helpers.
package ntstatus

import "fmt"

// Code is an unsigned 32-bit NTSTATUS bit pattern.
type Code uint32

// Severity identifies an NTSTATUS severity field.
type Severity uint8

// Error is an unsuccessful NTSTATUS value represented as a Go error.
type Error struct {
	Code Code
}

const (
	SeveritySuccess       Severity = 0
	SeverityInformational Severity = 1
	SeverityWarning       Severity = 2
	SeverityError         Severity = 3
)

// Succeeded reports whether code satisfies the NT_SUCCESS test.
func Succeeded(code Code) bool {
	return int32(code) >= 0
}

// Failed reports whether code does not satisfy the NT_SUCCESS test.
func Failed(code Code) bool {
	return int32(code) < 0
}

// SeverityOf extracts the two-bit severity field from code.
func SeverityOf(code Code) Severity {
	return Severity(code >> 30)
}

// IsInformational reports whether code has informational severity.
func IsInformational(code Code) bool {
	return SeverityOf(code) == SeverityInformational
}

// IsWarning reports whether code has warning severity.
func IsWarning(code Code) bool {
	return SeverityOf(code) == SeverityWarning
}

// IsError reports whether code has error severity.
func IsError(code Code) bool {
	return SeverityOf(code) == SeverityError
}

// Err returns nil for values satisfying NT_SUCCESS and an error otherwise.
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

	return fmt.Sprintf("NTSTATUS(0x%08X)", uint32(code))
}

// Error implements error.
func (err Error) Error() string {
	return err.Code.String()
}

// Is compares NTSTATUS errors by numeric code.
func (err Error) Is(target error) bool {
	other, ok := target.(Error)
	if !ok {
		return false
	}

	return err.Code == other.Code
}

// NT_SUCCESS is the exact-name counterpart of the Win32 NT_SUCCESS macro.
func NT_SUCCESS(code Code) bool {
	return Succeeded(code)
}

// NT_INFORMATION is the exact-name counterpart of the Win32 NT_INFORMATION macro.
func NT_INFORMATION(code Code) bool {
	return IsInformational(code)
}

// NT_WARNING is the exact-name counterpart of the Win32 NT_WARNING macro.
func NT_WARNING(code Code) bool {
	return IsWarning(code)
}

// NT_ERROR is the exact-name counterpart of the Win32 NT_ERROR macro.
func NT_ERROR(code Code) bool {
	return IsError(code)
}
