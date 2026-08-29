package winerror

import (
	"errors"
	"syscall"
	"testing"
)

func TestLookupRoundTrip(t *testing.T) {
	t.Parallel()

	name, ok := Name(ERROR_ACCESS_DENIED)
	if !ok {
		t.Fatal("ERROR_ACCESS_DENIED has no name")
	}

	code, ok := Parse(name)
	if !ok || code != ERROR_ACCESS_DENIED {
		t.Fatalf("Parse(%q) = %d, %v", name, code, ok)
	}
}

func TestErrorInterop(t *testing.T) {
	t.Parallel()

	if err := ERROR_SUCCESS.Err(); err != nil {
		t.Fatalf("ERROR_SUCCESS.Err() = %v, want nil", err)
	}

	err := ERROR_ACCESS_DENIED.Err()
	if !errors.Is(err, syscall.Errno(ERROR_ACCESS_DENIED)) {
		t.Fatalf("errors.Is(%v, ERROR_ACCESS_DENIED) = false", err)
	}
}

func TestCanonicalSuccessName(t *testing.T) {
	t.Parallel()

	name, ok := Name(ERROR_SUCCESS)
	if !ok || name != "ERROR_SUCCESS" {
		t.Fatalf("Name(ERROR_SUCCESS) = %q, %v", name, ok)
	}
}

func TestAppModelErrors(t *testing.T) {
	t.Parallel()

	if APPMODEL_ERROR_NO_PACKAGE != 15700 {
		t.Fatalf("APPMODEL_ERROR_NO_PACKAGE = %d", APPMODEL_ERROR_NO_PACKAGE)
	}
}
