package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseGUID(t *testing.T) {
	t.Parallel()

	parts, err := parseGUID("00000000-0000-0000-c000-000000000046")
	if err != nil {
		t.Fatal(err)
	}

	if parts.Data4 != [8]byte{0xc0, 0, 0, 0, 0, 0, 0, 0x46} {
		t.Fatalf("unexpected Data4: %#v", parts.Data4)
	}
}

func TestRenderGUIDNameIndexesData1(t *testing.T) {
	t.Parallel()

	items := []generatedGUID{
		{Identifier: "GUID_Alias", Value: "00000001-0000-0000-0000-000000000000", Parts: guidParts{Data1: 1}},
		{Identifier: "GUID_First", Value: "00000001-0000-0000-0000-000000000000", Parts: guidParts{Data1: 1}},
		{Identifier: "GUID_Second", Value: "00000002-0000-0000-0000-000000000000", Parts: guidParts{Data1: 2}},
	}

	contents, err := renderGUIDs(items, sourceLock{})
	if err != nil {
		t.Fatal(err)
	}

	text := string(contents)
	if !strings.Contains(text, "switch value.Data1") || !strings.Contains(text, "case 0x00000001") {
		t.Fatalf("generated Name lacks a Data1 index:\n%s", text)
	}

	if !strings.Contains(text, "return \"GUID_Alias\", true") || strings.Contains(text, "return \"GUID_First\", true") {
		t.Fatalf("generated Name did not preserve the first canonical alias:\n%s", text)
	}
}

func TestRenderCatalogUsesNestedLookup(t *testing.T) {
	t.Parallel()

	packages := map[string][]generatedConstant{
		"hresult": {
			{Name: "S_OK", Expression: "0", Namespace: "Windows.Win32.Foundation", DeclaringType: "Apis", Comment: "<b>Success.</b>"},
		},
	}

	contents, err := renderCatalog(packages, sourceLock{})
	if err != nil {
		t.Fatal(err)
	}

	text := string(contents)
	if !strings.Contains(text, "switch packageName") || !strings.Contains(text, "switch name") {
		t.Fatalf("generated Lookup lacks nested switches:\n%s", text)
	}

	if strings.Contains(text, "packageName +") {
		t.Fatalf("generated Lookup concatenates its inputs:\n%s", text)
	}

	wantDefinition := `{Package: "hresult", Name: "S_OK", Value: "0", Namespace: "Windows.Win32.Foundation", DeclaringType: "Apis", Documentation: "", Comment: "Success."}`
	if !strings.Contains(text, "var definitions = [...]Definition{") || !strings.Contains(text, "range definitions") {
		t.Fatalf("generated Definitions lacks a static table:\n%s", text)
	}

	if !strings.Contains(text, wantDefinition) {
		t.Fatalf("generated Definitions did not preserve every field:\n%s", text)
	}
}

func TestParsePropertyKey(t *testing.T) {
	t.Parallel()

	parts, propertyID, err := parsePropertyKey("{3305783056, 43612, 16967, 184, 48, 214, 166, 248, 234, 163, 16}, 4")
	if err != nil {
		t.Fatal(err)
	}

	if parts.Data1 != 3305783056 || propertyID != 4 {
		t.Fatalf("unexpected property key: %#v, %d", parts, propertyID)
	}
}

func TestWriteGeneratedComment(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	writeGeneratedComment(&output, "VALUE", `Uses <b>official</b> text &amp; preserves the meaning.`)

	got := output.String()
	if got != "\t// VALUE: Uses official text & preserves the meaning.\n" {
		t.Fatalf("comment = %q", got)
	}
}

func TestWrapComment(t *testing.T) {
	t.Parallel()

	lines := wrapComment("VALUE: one two three four", 16)
	if strings.Join(lines, "|") != "VALUE: one two|three four" {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestMeasureCoverage(t *testing.T) {
	t.Parallel()

	export := metadataExport{Constants: []metadataConstant{
		{Name: "ERROR_TEST", Namespace: "Windows.Win32.Foundation", DeclaringType: "WIN32_ERROR"},
		{Name: "UNCLASSIFIED_TEST", Namespace: "Windows.Win32.Test", DeclaringType: "Apis"},
	}}

	coverage := measureCoverage(export)
	if coverage.TotalRows != 2 || coverage.MatchedRows != 1 || coverage.UnclassifiedRows != 1 {
		t.Fatalf("unexpected coverage: %#v", coverage)
	}

	if coverage.UnclassifiedByNamespace["Windows.Win32.Test"] != 1 {
		t.Fatalf("unexpected namespace coverage: %#v", coverage.UnclassifiedByNamespace)
	}
}

func TestEnumPackageClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		packageName string
		namespace   string
	}{
		{packageName: "jobobject", namespace: "Windows.Win32.System.JobObjects"},
		{packageName: "power", namespace: "Windows.Win32.System.Power"},
		{packageName: "sysinfo", namespace: "Windows.Win32.System.SystemInformation"},
		{packageName: "toolhelp", namespace: "Windows.Win32.System.Diagnostics.ToolHelp"},
		{packageName: "libraryloader", namespace: "Windows.Win32.System.LibraryLoader"},
	}

	for _, test := range tests {
		t.Run(test.packageName, func(t *testing.T) {
			t.Parallel()

			item := metadataConstant{Namespace: test.namespace, Enum: true}
			if !specByName(test.packageName).Match(item) {
				t.Fatal("enum member was not classified")
			}

			item = metadataConstant{Namespace: test.namespace}
			if specByName(test.packageName).Match(item) {
				t.Fatal("loose constant was classified")
			}

			item = metadataConstant{Namespace: test.namespace + ".Other", Enum: true}
			if specByName(test.packageName).Match(item) {
				t.Fatal("enum member from another namespace was classified")
			}
		})
	}
}

func TestSecondBatchClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		packageName   string
		namespace     string
		declaringType string
		name          string
	}{
		{packageName: "foundation", namespace: "Windows.Win32.Foundation", name: "INVALID_HANDLE_VALUE"},
		{packageName: "filesystem", namespace: "Windows.Win32.Storage.FileSystem", name: "INVALID_FILE_ATTRIBUTES"},
		{packageName: "console", namespace: "Windows.Win32.System.Console", name: "DISABLE_NEWLINE_AUTO_RETURN"},
		{packageName: "process", namespace: "Windows.Win32.System.Threading", declaringType: "PROCESS_MITIGATION_POLICY", name: "ProcessDEPPolicy"},
		{packageName: "winsock", namespace: "Windows.Win32.Networking.WinSock", name: "FD_READ"},
		{packageName: "winmsg", namespace: "Windows.Win32.UI.WindowsAndMessaging", name: "SPI_GETWORKAREA"},
		{packageName: "winmsg", namespace: "Windows.Win32.UI.Input.KeyboardAndMouse", name: "MOUSEEVENTF_MOVE"},
	}

	for _, test := range tests {
		t.Run(test.packageName+"/"+test.name, func(t *testing.T) {
			t.Parallel()

			item := metadataConstant{Namespace: test.namespace, DeclaringType: test.declaringType, Name: test.name}
			if !specByName(test.packageName).Match(item) {
				t.Fatal("constant was not classified")
			}

			item.Namespace = "Windows.Win32.Other"
			if specByName(test.packageName).Match(item) {
				t.Fatal("constant from another namespace was classified")
			}
		})
	}
}

func TestCollectConstantMethods(t *testing.T) {
	t.Parallel()

	methods, skipped := collectConstantMethods([]metadataConstantMethod{
		{
			Namespace:  "Windows.Win32.System.Threading",
			Name:       "GetCurrentProcessToken",
			ReturnType: "Windows.Win32.Foundation.HANDLE",
			Value:      "-4",
		},
	})
	if skipped != 0 || len(methods) != 1 {
		t.Fatalf("methods = %#v, skipped = %d", methods, skipped)
	}

	if methods[0].Expression != "^uintptr(3)" {
		t.Fatalf("expression = %q", methods[0].Expression)
	}
}
