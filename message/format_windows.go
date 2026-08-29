//go:build windows

package message

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/depthbomb/win32defs/ntstatus"
)

const (
	formatMessageIgnoreInserts = 0x00000200
	formatMessageFromModule    = 0x00000800
	formatMessageFromSystem    = 0x00001000
	maximumMessageCharacters   = 1 << 16
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	ntdll              = syscall.NewLazyDLL("ntdll.dll")
	formatMessageWProc = kernel32.NewProc("FormatMessageW")
)

// Format formats a 32-bit message identifier using FormatMessageW.
func Format(code uint32) (string, error) {
	return FormatWithOptions(code, Options{SearchSystem: true})
}

// FormatWithOptions formats a message using explicit module, language, and fallback settings.
func FormatWithOptions(code uint32, options Options) (string, error) {
	flags := uintptr(formatMessageIgnoreInserts)
	if options.Module != 0 {
		flags |= formatMessageFromModule
	}

	if options.SearchSystem || options.Module == 0 {
		flags |= formatMessageFromSystem
	}

	return format(code, options.Module, options.LanguageID, flags)
}

// FormatFromModule formats a message from a loaded module and falls back to the system table.
func FormatFromModule(code uint32, module uintptr) (string, error) {
	return FormatWithOptions(code, Options{Module: module, SearchSystem: true})
}

// FormatNTStatus formats an NTSTATUS using the NTDLL message table.
func FormatNTStatus(code ntstatus.Code) (string, error) {
	if err := ntdll.Load(); err != nil {
		return "", fmt.Errorf("load ntdll.dll: %w", err)
	}

	return FormatWithOptions(uint32(code), Options{Module: uintptr(ntdll.Handle())})
}

func format(code uint32, module uintptr, languageID uint32, flags uintptr) (string, error) {
	for size := uint32(256); size <= maximumMessageCharacters; size *= 2 {
		buffer := make([]uint16, size)
		length, _, callErr := formatMessageWProc.Call(
			flags,
			module,
			uintptr(code),
			uintptr(languageID),
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(size),
			0,
		)
		if length != 0 {
			message := syscall.UTF16ToString(buffer[:length])

			return strings.TrimSpace(message), nil
		}

		if callErr != syscall.ERROR_INSUFFICIENT_BUFFER {
			return "", fmt.Errorf("FormatMessageW(%d): %w", code, callErr)
		}
	}

	return "", fmt.Errorf("FormatMessageW(%d): message exceeds %d characters", code, maximumMessageCharacters)
}
