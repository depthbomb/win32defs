package hresult

import "testing"

var benchmarkCode Code
var benchmarkName string
var benchmarkOK bool

func BenchmarkName(b *testing.B) {
	values := [...]Code{S_OK, E_FAIL, E_ACCESSDENIED}

	b.ReportAllocs()
	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		benchmarkName, benchmarkOK = Name(values[index%len(values)])
	}
}

func BenchmarkParse(b *testing.B) {
	names := [...]string{"S_OK", "E_FAIL", "E_ACCESSDENIED"}

	b.ReportAllocs()
	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		benchmarkCode, benchmarkOK = Parse(names[index%len(names)])
	}
}
