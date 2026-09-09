package catalog

import "testing"

var benchmarkFamilyNames []string
var benchmarkFlags string
var benchmarkGroup *familyGroup

func BenchmarkBuildFamilyGroup(b *testing.B) {
	descriptor := findFamily(fileDialogFamily())
	b.ReportAllocs()

	for range b.N {
		benchmarkGroup = buildFamilyGroup(descriptor)
	}
}

func BenchmarkFamilyNames(b *testing.B) {
	family := fileDialogFamily()
	b.ReportAllocs()

	for range b.N {
		benchmarkFamilyNames = NamesInFamily(family, "0x20")
	}
}

func BenchmarkFamilyDefinitions(b *testing.B) {
	family := fileDialogFamily()
	b.ReportAllocs()

	for range b.N {
		for definition := range FamilyDefinitions(family) {
			benchmarkDefinition = definition
		}
	}
}

func BenchmarkFormatFlags(b *testing.B) {
	family := fileDialogFamily()
	b.ReportAllocs()

	for range b.N {
		benchmarkFlags, benchmarkDefinitionOK = FormatFlags(family, 0x60)
	}
}

func BenchmarkFamilies(b *testing.B) {
	b.ReportAllocs()

	for range b.N {
		count := 0
		for range Families("shell") {
			count++
		}
		benchmarkDefinitionCount = count
	}
}
