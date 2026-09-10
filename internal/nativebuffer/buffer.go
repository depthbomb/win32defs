// Package nativebuffer supports generated pointer-free Windows buffer views.
package nativebuffer

import "unsafe"

// View checks length rather than capacity before returning shared storage.
func View(data []byte, size int) []byte {
	if len(data) < size {
		panic("native buffer is too short")
	}

	return data[:size:size]
}

// New allocates zeroed storage with the supplied native alignment.
// The generator checks that size plus alignment fits a Go int on every target.
func New(size int, alignment uintptr) []byte {
	data := make([]byte, size+int(alignment)-1)
	address := uintptr(unsafe.Pointer(unsafe.SliceData(data)))
	offset := int((alignment - address%alignment) % alignment)

	return data[offset : offset+size : offset+size]
}

// Pointer rejects views that cannot be passed as an aligned native object.
func Pointer(data []byte, alignment uintptr) unsafe.Pointer {
	pointer := unsafe.Pointer(&data[0])
	if uintptr(pointer)%alignment != 0 {
		panic("unaligned native buffer")
	}

	return pointer
}

// Read copies bytes into an aligned Go value without dereferencing unaligned
// native storage. Generated callers use only pointer-free types whose size and
// internal layout match Windows. Windows targets are little-endian, as is Go
// on those targets. Alignment of the source storage need not match Go.
func Read[T any](data []byte) T {
	var value T
	size := unsafe.Sizeof(value)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&value)), size), data[:size])

	return value
}

// Write copies a pointer-free Go value with a verified native internal layout.
func Write[T any](data []byte, value T) {
	size := unsafe.Sizeof(value)
	copy(data[:size], unsafe.Slice((*byte)(unsafe.Pointer(&value)), size))
}
