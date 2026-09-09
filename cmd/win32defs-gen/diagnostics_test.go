package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerationRegressionGate(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	baseline := generationReport{
		Symbols: map[string][]string{
			"winmsg": {"WM_USER"},
		},
		Collisions: []collision{
			{
				Package:     "guid",
				Name:        "KNOWN",
				Definitions: []string{"one", "two"},
			},
		},
	}
	path := filepath.Join(root, "internal", "source", "report.json")
	if err := writeReport(path, baseline); err != nil {
		t.Fatal(err)
	}

	if err := validateGeneration(root, baseline, false); err != nil {
		t.Fatal(err)
	}

	changed := baseline
	changed.Symbols = map[string][]string{}
	changed.Rejected = []rejectedDefinition{
		{
			Package: "winmsg",
			Name:    "BAD_VALUE",
			Value:   "99999999999999999",
			Reason:  "overflow",
		},
	}
	err := validateGeneration(root, changed, false)
	if err == nil || !strings.Contains(err.Error(), "WM_USER") || !strings.Contains(err.Error(), "BAD_VALUE") || !strings.Contains(err.Error(), "overflow") {
		t.Fatalf("incomplete diagnostics: %v", err)
	}

	contents, err := os.ReadFile(path)
	if err != nil || strings.Contains(string(contents), "BAD_VALUE") {
		t.Fatal("validation changed the baseline report")
	}

	if err := validateGeneration(root, changed, true); err != nil {
		t.Fatal(err)
	}

	changed = baseline
	changed.Collisions = []collision{
		{
			Package:     "guid",
			Name:        "NEW",
			Definitions: []string{"one", "two"},
		},
	}
	if err := validateGeneration(root, changed, false); err == nil || !strings.Contains(err.Error(), "new collision") {
		t.Fatalf("new collision was not rejected: %v", err)
	}
}

func TestRejectedConstantDetails(t *testing.T) {
	t.Parallel()

	export := metadataExport{
		Constants: []metadataConstant{
			{
				Name:      "FOS_TEST",
				Namespace: "Windows.Win32.UI.Shell",
				Kind:      "uint64",
				Value:     "4294967296",
			},
		},
	}
	_, _, skipped, rejected := collectConstants(export)
	if skipped["unsupported_value"] != 1 || len(rejected) != 1 || rejected[0].Name != "FOS_TEST" || rejected[0].Value != "4294967296" || rejected[0].Reason == "" {
		t.Fatalf("unexpected rejection details: %#v, %#v", skipped, rejected)
	}
}
