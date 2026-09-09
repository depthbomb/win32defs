package main

import "testing"

var benchmarkConstant generatedConstant

func BenchmarkNormalizeConstants(b *testing.B) {
	items := []metadataConstant{
		{
			Name:  "FOS_PICKFOLDERS",
			Kind:  "uint32",
			Value: "32",
		},
		{
			Name:  "FOS_HIGH_BIT",
			Kind:  "uint32",
			Value: "2147483648",
		},
		{
			Name:  "FOS_SIGNED",
			Kind:  "int32",
			Value: "-2147123200",
		},
		{
			Name:  "FOS_ZERO",
			Kind:  "uint32",
			Value: "0",
		},
	}
	spec := specByName("shell")
	b.ReportAllocs()
	b.ResetTimer()

	for index := range b.N {
		var err error
		benchmarkConstant, err = normalizeConstant(items[index&3], spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}
