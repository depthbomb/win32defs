//go:build windows

package message

import (
	"testing"

	"github.com/depthbomb/win32defs/ntstatus"
	"github.com/depthbomb/win32defs/winerror"
)

func TestFormatWinError(t *testing.T) {
	t.Parallel()

	message, err := FormatWinError(winerror.ERROR_ACCESS_DENIED)
	if err != nil {
		t.Fatal(err)
	}

	if message == "" {
		t.Fatal("FormatWinError returned an empty message")
	}
}

func TestFormatNTStatus(t *testing.T) {
	t.Parallel()

	message, err := FormatNTStatus(ntstatus.STATUS_ACCESS_DENIED)
	if err != nil {
		t.Fatal(err)
	}

	if message == "" {
		t.Fatal("FormatNTStatus returned an empty message")
	}
}

func TestFormatWithOptions(t *testing.T) {
	t.Parallel()

	message, err := FormatWithOptions(uint32(winerror.ERROR_ACCESS_DENIED), Options{SearchSystem: true})
	if err != nil {
		t.Fatal(err)
	}

	if message == "" {
		t.Fatal("FormatWithOptions returned an empty message")
	}
}
