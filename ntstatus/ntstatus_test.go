package ntstatus

import "testing"

func TestClassification(t *testing.T) {
	t.Parallel()

	if !Succeeded(STATUS_SUCCESS) {
		t.Fatal("STATUS_SUCCESS must satisfy Succeeded")
	}

	if !Failed(STATUS_UNSUCCESSFUL) {
		t.Fatal("STATUS_UNSUCCESSFUL must satisfy Failed")
	}

	if got := SeverityOf(STATUS_UNSUCCESSFUL); got != SeverityError {
		t.Fatalf("SeverityOf(STATUS_UNSUCCESSFUL) = %d, want %d", got, SeverityError)
	}
}

func TestSeverityPredicates(t *testing.T) {
	t.Parallel()

	if !NT_INFORMATION(STATUS_OBJECT_NAME_EXISTS) {
		t.Fatal("STATUS_OBJECT_NAME_EXISTS must be informational")
	}

	if !NT_WARNING(STATUS_BUFFER_OVERFLOW) {
		t.Fatal("STATUS_BUFFER_OVERFLOW must be a warning")
	}

	if !NT_ERROR(STATUS_UNSUCCESSFUL) {
		t.Fatal("STATUS_UNSUCCESSFUL must be an error")
	}
}

func TestCanonicalSuccessName(t *testing.T) {
	t.Parallel()

	name, ok := Name(STATUS_SUCCESS)
	if !ok || name != "STATUS_SUCCESS" {
		t.Fatalf("Name(STATUS_SUCCESS) = %q, %v", name, ok)
	}

	if _, ok := Parse("DBG_FRAME_DEFAULT"); ok {
		t.Fatal("debugger formatting flag leaked into NTSTATUS definitions")
	}
}
