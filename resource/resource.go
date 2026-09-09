package resource

// Uintptr returns the integer identifier representation accepted by Win32
// resource APIs in place of a string pointer. It does not point to string data.
func (id ID) Uintptr() uintptr {
	return uintptr(id)
}
