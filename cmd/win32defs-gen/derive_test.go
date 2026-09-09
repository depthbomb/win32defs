package main

import (
	"slices"
	"strings"
	"testing"
)

func TestDeriveShellKeySelect(t *testing.T) {
	t.Parallel()

	operands := []metadataConstant{
		{
			Namespace: "Windows.Win32.UI.Shell",
			Name:      "NIN_SELECT",
			Kind:      "uint32",
			Value:     "1024",
		},
		{
			Namespace: "Windows.Win32.UI.Shell",
			Name:      "NINF_KEY",
			Kind:      "uint32",
			Value:     "1",
		},
	}
	derived, err := deriveConstants(operands)
	if err != nil || len(derived) != 1 {
		t.Fatalf("deriveConstants = %#v, %v", derived, err)
	}

	if derived[0].Name != "NIN_KEYSELECT" || derived[0].Value != "1025" || derived[0].Documentation == "" || !strings.Contains(derived[0].Comment, "NIN_SELECT | NINF_KEY") {
		t.Fatalf("incorrect value or missing provenance: %#v", derived[0])
	}

	for _, test := range []struct {
		name  string
		edit  func([]metadataConstant) []metadataConstant
		error string
	}{
		{
			name: "upstream definition",
			edit: func(items []metadataConstant) []metadataConstant {
				return append(items, derived[0])
			},
		},
		{
			name: "conflicting upstream definition",
			edit: func(items []metadataConstant) []metadataConstant {
				item := derived[0]
				item.Value = "1026"

				return append(items, item)
			},
			error: "conflicts with SDK expression",
		},
		{
			name: "missing operand",
			edit: func(items []metadataConstant) []metadataConstant {
				return items[:1]
			},
			error: "missing Windows.Win32.UI.Shell.NINF_KEY",
		},
		{
			name: "wrong namespace",
			edit: func(items []metadataConstant) []metadataConstant {
				items[1].Namespace = "Windows.Win32.Other"

				return items
			},
			error: "missing Windows.Win32.UI.Shell.NINF_KEY",
		},
		{
			name: "overflow",
			edit: func(items []metadataConstant) []metadataConstant {
				items[1].Value = "4294967296"

				return items
			},
			error: "invalid NINF_KEY",
		},
		{
			name: "conflicting operand",
			edit: func(items []metadataConstant) []metadataConstant {
				item := items[1]
				item.Value = "2"

				return append(items, item)
			},
			error: "conflicting NINF_KEY",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := deriveConstants(test.edit(slices.Clone(operands)))
			if test.error != "" {
				if err == nil || !strings.Contains(err.Error(), test.error) {
					t.Fatalf("error = %v, want %q", err, test.error)
				}

				return
			}

			if err != nil || len(got) != 0 {
				t.Fatalf("upstream constant should not be duplicated: %#v, %v", got, err)
			}
		})
	}
}
