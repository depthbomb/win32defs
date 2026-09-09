package hidpi

// Uintptr returns the pointer-sized bit representation of a DPI awareness
// context, including negative pseudo-handle constants. It is intended for
// passing DPI_AWARENESS_CONTEXT_* values to Windows syscall bindings.
func (value Value) Uintptr() uintptr {
	return uintptr(value)
}
