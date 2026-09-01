package guid

import "testing"

var benchmarkGUID GUID
var benchmarkGUIDName string
var benchmarkGUIDOK bool

func BenchmarkLookup(b *testing.B) {
	names := [...]string{
		"CLSID_AACMFTEncoder",
		"IID_IUnknown",
		"IID__WMPOCXEvents",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		benchmarkGUID, benchmarkGUIDOK = Lookup(names[index%len(names)])
	}
}

func BenchmarkName(b *testing.B) {
	values := [...]GUID{
		CLSID_AACMFTEncoder,
		IID_IUnknown,
		IID__WMPOCXEvents,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		benchmarkGUIDName, benchmarkGUIDOK = Name(values[index%len(values)])
	}
}

func BenchmarkNameMiss(b *testing.B) {
	value := GUID{
		Data1: 0xDEADBEEF,
		Data2: 0xCAFE,
		Data3: 0xBABE,
		Data4: [8]byte{0, 1, 2, 3, 4, 5, 6, 7},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		benchmarkGUIDName, benchmarkGUIDOK = Name(value)
	}
}
