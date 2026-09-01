package foundation

// InvalidHandleValue returns the pointer-sized representation of INVALID_HANDLE_VALUE.
func InvalidHandleValue() uintptr {
	return ^uintptr(0)
}
