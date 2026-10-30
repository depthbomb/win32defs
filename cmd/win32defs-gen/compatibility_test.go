package main

import (
	"slices"
	"testing"
)

func TestCompatibilityDefinitions(t *testing.T) {
	t.Parallel()

	retained := missingCompatibilityDefinitions(metadataExport{})
	packages, collisions, skipped, rejected := collectConstants(retained)
	if len(collisions) != 0 || len(skipped) != 0 || len(rejected) != 0 {
		t.Fatalf("retained constants could not be projected: %v, %v, %v", collisions, skipped, rejected)
	}
	for _, test := range []struct {
		pkg, name, expression, family string
	}{
		{"ntstatus", "STATUS_ADMINLESS_ACCESS_DENIED", "0xC000A204", "STATUS_"},
		{"power", "MonitorRequestReasonAusterityBatteryDrain", "0x00000037", "POWER_MONITOR_REQUEST_REASON"},
		{"hresult", "WU_E_STANDBY_ACTIVITY_NOT_ALLOWED", "-2145124249", "WU_"},
	} {
		items := packages[test.pkg]
		if len(items) != 1 || items[0].Name != test.name || items[0].Expression != test.expression || items[0].Family != test.family {
			t.Errorf("%s compatibility constant = %+v", test.pkg, items)
		}
	}

	structures, deferred := collectStructures(retained.Types)
	if len(deferred) != 0 || len(structures["cryptography"]) != 2 {
		t.Fatalf("retained structures = %+v, deferred = %+v", structures, deferred)
	}
	for _, item := range structures["cryptography"] {
		if len(item.Fields) != 3 {
			t.Fatalf("%s field count changed: %d", item.Source.Name, len(item.Fields))
		}
		for _, layout := range item.ABI {
			if layout.Size != 12 || layout.Align != 4 || !slices.Equal(layout.Offsets, []int64{0, 4, 8}) {
				t.Errorf("%s ABI changed: %+v", item.Source.Name, layout)
			}
		}
		for index, name := range []string{"DwMagic", "CbParameterSet", "CbKey"} {
			if item.Fields[index].Name != name || item.Fields[index].Type != "uint32" {
				t.Errorf("%s field %d changed: %+v", item.Source.Name, index, item.Fields[index])
			}
		}
	}

	contents, err := renderStructures("cryptography", structures["cryptography"], sourceLock{})
	if err != nil {
		t.Fatal(err)
	}
	for _, arch := range []string{"386", "amd64", "arm64"} {
		t.Run(arch, func(t *testing.T) {
			checkStructureSource(t, contents, arch, false)
		})
	}
}

func TestCompatibilityOnlySuppliesMissingDefinitions(t *testing.T) {
	t.Parallel()

	original := missingCompatibilityDefinitions(metadataExport{})
	retained := missingCompatibilityDefinitions(original)
	if len(retained.Constants) != 0 || len(retained.Types) != 0 {
		t.Fatalf("upstream definitions were duplicated: %+v", retained)
	}

	partial := metadataExport{
		Constants: original.Constants[:1],
		Types:     original.Types[:1],
	}
	retained = missingCompatibilityDefinitions(partial)
	if len(retained.Constants) != 2 || len(retained.Types) != 1 || retained.Types[0].Name != "BCRYPT_PQDSA_KEY_BLOB" {
		t.Fatalf("missing definitions were not retained: %+v", retained)
	}

	constant := original.Constants[0]
	constant.DeclaringType = "UnrelatedEnum"
	structure := original.Types[0]
	structure.Namespace = "Windows.Win32.Other"
	retained = missingCompatibilityDefinitions(metadataExport{
		Constants: []metadataConstant{constant},
		Types:     []metadataType{structure},
	})
	if len(retained.Constants) != 3 || len(retained.Types) != 2 {
		t.Fatalf("unrelated definitions suppressed compatibility entries: %+v", retained)
	}
}
