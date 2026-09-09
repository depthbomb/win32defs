package catalog

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func fileDialogFamily() Family {
	return Family{
		Package:   "shell",
		Namespace: "Windows.Win32.UI.Shell",
		Name:      "FILEOPENDIALOGOPTIONS",
	}
}

func TestFamilyLookupAndAliases(t *testing.T) {
	t.Parallel()

	family := fileDialogFamily()
	definition, ok := LookupInFamily(family, "FOS_PICKFOLDERS")
	if !ok || !definition.Flags {
		t.Fatalf("missing flags provenance: %#v", definition)
	}

	if _, ok := LookupInFamily(family, "NIF_ICON"); ok {
		t.Fatal("lookup crossed an enum boundary")
	}

	for _, value := range []string{"32", "0x20"} {
		got := NamesInFamily(family, value)
		if !reflect.DeepEqual(got, []string{"FOS_PICKFOLDERS"}) {
			t.Fatalf("NamesInFamily(%q) = %v", value, got)
		}
	}

	got := NamesInFamily(family, "32")
	got[0] = "mutated"
	if names := NamesInFamily(family, "32"); names[0] != "FOS_PICKFOLDERS" {
		t.Fatal("caller mutation affected later lookups")
	}

	if !slices.Contains(slices.Collect(Families("shell")), family) {
		t.Fatal("family discovery is missing the file dialog enum")
	}

	for definition := range FamilyDefinitions(family) {
		if !strings.HasPrefix(definition.Name, "FOS_") {
			t.Fatalf("unrelated member: %#v", definition)
		}
		break
	}
}

func TestFlagFormatting(t *testing.T) {
	t.Parallel()

	family := fileDialogFamily()
	got, ok := FormatFlags(family, 0x60)
	if !ok || got != "FOS_FORCEFILESYSTEM | FOS_PICKFOLDERS" {
		t.Fatalf("combined flags = %q, %v", got, ok)
	}

	got, ok = FormatFlags(family, 0x61)
	if !ok || got != "FOS_FORCEFILESYSTEM | FOS_PICKFOLDERS | 0x1" {
		t.Fatalf("unknown bits = %q, %v", got, ok)
	}

	got, ok = FormatFlags(family, 0)
	if !ok || got != "0" {
		t.Fatalf("zero flags = %q, %v", got, ok)
	}

	nonFlags := Family{
		Package:   "dwm",
		Namespace: "Windows.Win32.Graphics.Dwm",
		Name:      "DWMWINDOWATTRIBUTE",
	}
	if _, ok := FormatFlags(nonFlags, 20); ok {
		t.Fatal("ordinary enum was formatted as flags")
	}

	if _, ok := FormatFlags(Family{}, 0); ok {
		t.Fatal("missing family was formatted as flags")
	}
}

func TestFamilyValueKinds(t *testing.T) {
	t.Parallel()

	family := Family{
		Package:   "cryptography",
		Namespace: "Windows.Win32.Security.Cryptography",
		Name:      "BCRYPT_",
	}
	if got := NamesInFamily(family, `"SHA256"`); !slices.Contains(got, "BCRYPT_SHA256_ALGORITHM") {
		t.Fatalf("string family lookup = %v", got)
	}
}
