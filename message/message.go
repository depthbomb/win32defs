// Package message formats Windows status and error messages using the local
// operating system message tables when running on Windows.
package message

import (
	"errors"

	"github.com/depthbomb/win32defs/hresult"
	"github.com/depthbomb/win32defs/winerror"
)

// ErrUnsupported indicates that native Windows message formatting is not
// available on the current platform.
var ErrUnsupported = errors.New("Windows message formatting is unsupported on this platform")

// Options controls native Windows message-table lookup.
type Options struct {
	Module       uintptr
	LanguageID   uint32
	SearchSystem bool
}

// FormatHRESULT formats an HRESULT using the local Windows message tables.
func FormatHRESULT(code hresult.Code) (string, error) {
	return Format(uint32(code))
}

// FormatWinError formats a Win32 system error using the local Windows message tables.
func FormatWinError(code winerror.Code) (string, error) {
	return Format(uint32(code))
}
