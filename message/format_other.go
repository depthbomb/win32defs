//go:build !windows

package message

import "github.com/depthbomb/win32defs/ntstatus"

// Format returns ErrUnsupported outside Windows.
func Format(code uint32) (string, error) {
	return "", ErrUnsupported
}

// FormatWithOptions returns ErrUnsupported outside Windows.
func FormatWithOptions(code uint32, options Options) (string, error) {
	return "", ErrUnsupported
}

// FormatFromModule returns ErrUnsupported outside Windows.
func FormatFromModule(code uint32, module uintptr) (string, error) {
	return "", ErrUnsupported
}

// FormatNTStatus returns ErrUnsupported outside Windows.
func FormatNTStatus(code ntstatus.Code) (string, error) {
	return "", ErrUnsupported
}
