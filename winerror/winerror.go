// Package winerror provides Win32 system error codes.
package winerror

import (
	"fmt"
	"syscall"
)

// Code is an unsigned 32-bit Win32 system error code.
type Code uint32

// Error is a nonzero Win32 system error represented as a Go error.
type Error struct {
	Code Code
}

// Err returns nil for ERROR_SUCCESS and a Go error for any other value.
func (code Code) Err() error {
	if code == ERROR_SUCCESS {
		return nil
	}

	return Error{Code: code}
}

// Errno converts code to the standard library's syscall representation.
func (code Code) Errno() syscall.Errno {
	return syscall.Errno(code)
}

// String returns the canonical symbolic name or a decimal representation.
func (code Code) String() string {
	if name, ok := Name(code); ok {
		return name
	}

	return fmt.Sprintf("Win32Error(%d)", uint32(code))
}

// Error implements error.
func (err Error) Error() string {
	return err.Code.String()
}

// Unwrap exposes the equivalent syscall.Errno for errors.Is interoperability.
func (err Error) Unwrap() error {
	return err.Code.Errno()
}

// Is compares Win32 errors by numeric code.
func (err Error) Is(target error) bool {
	switch other := target.(type) {
	case Error:
		return err.Code == other.Code
	case syscall.Errno:
		return err.Code == Code(other)
	default:
		return false
	}
}
