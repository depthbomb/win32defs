package catalog

import "testing"

var benchmarkDefinition Definition
var benchmarkDefinitionOK bool
var benchmarkDefinitionCount int

func BenchmarkDefinitions(b *testing.B) {
	for range b.N {
		count := 0
		for definition := range Definitions() {
			benchmarkDefinition = definition
			count++
		}
		benchmarkDefinitionCount = count
	}
}

func BenchmarkLookup(b *testing.B) {
	cases := [...]struct {
		packageName string
		name        string
	}{
		{packageName: "hresult", name: "S_OK"},
		{packageName: "winerror", name: "ERROR_ACCESS_DENIED"},
		{packageName: "winsock", name: "WSAEWOULDBLOCK"},
		{packageName: "winmsg", name: "WHEEL_DELTA"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		benchmarkDefinition, benchmarkDefinitionOK = Lookup(cases[index&3].packageName, cases[index&3].name)
	}
}

func BenchmarkLookupMiss(b *testing.B) {
	b.ReportAllocs()

	for range b.N {
		benchmarkDefinition, benchmarkDefinitionOK = Lookup("missing", "MISSING_VALUE")
	}
}
