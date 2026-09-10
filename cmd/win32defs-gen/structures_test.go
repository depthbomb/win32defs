package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func structureFixture(name string, fields ...metadataField) metadataType {
	return metadataType{
		Namespace:  "Windows.Win32.Foundation",
		Name:       name,
		Struct:     true,
		Sequential: true,
		Fields:     fields,
	}
}

func fieldFixture(name string, fieldType string) metadataField {
	return metadataField{
		Name:   name,
		Type:   fieldType,
		Offset: -1,
	}
}

func TestStructureProjectionAndAssertions(t *testing.T) {
	t.Parallel()

	handle := structureFixture("TEST_HANDLE", fieldFixture("Value", "Void*"))
	handle.Typedef = true
	inner := structureFixture("INNER", fieldFixture("tag", "Byte"), fieldFixture("handle", "Windows.Win32.Foundation.TEST_HANDLE"))
	outer := structureFixture("OUTER", fieldFixture("prefix", "UInt16"), fieldFixture("inner", "Windows.Win32.Foundation.INNER"), fieldFixture("tail", "UInt32[3]"))
	projection := newStructureProjection([]metadataType{outer, handle, inner})
	result, err := projection.resolve("Windows.Win32.Foundation.OUTER")
	if err != nil {
		t.Fatal(err)
	}

	if len(projection.Types) != 3 || result.Fields[2].Type != "[3]uint32" {
		t.Fatalf("dependency closure or array projection lost: %+v", projection.Types)
	}

	var items []projectedType
	for _, item := range projection.Types {
		items = append(items, item)
	}
	slices.SortFunc(items, func(a, b projectedType) int {
		return strings.Compare(a.Source.Name, b.Source.Name)
	})
	contents, err := renderStructures("foundation", items, sourceLock{})
	if err != nil {
		t.Fatal(err)
	}

	for _, arch := range []string{"386", "amd64", "arm64"} {
		t.Run(arch, func(t *testing.T) {
			checkStructureSource(t, contents, arch, false)
			broken := bytes.Replace(contents, []byte("[3]uint32"), []byte("[5]uint32"), 1)
			checkStructureSource(t, broken, arch, true)
		})
	}
}

func checkStructureSource(t *testing.T, contents []byte, arch string, wantError bool) {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generated.go", contents, 0)
	if err != nil {
		t.Fatal(err)
	}

	config := types.Config{
		Importer: importer.Default(),
		Sizes:    types.SizesFor("gc", arch),
	}
	_, err = config.Check("foundation", fset, []*ast.File{file}, nil)
	if (err != nil) != wantError {
		t.Fatalf("ABI assertions: error = %v, want error = %v", err, wantError)
	}
}

func TestStructureRejections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		change func(*metadataType)
		want   string
	}{
		{
			name: "packing",
			change: func(item *metadataType) {
				item.Packing = 1
			},
			want: "Windows 386",
		},
		{
			name: "explicit layout",
			change: func(item *metadataType) {
				item.Sequential = false
			},
			want: "layout",
		},
		{
			name: "architecture",
			change: func(item *metadataType) {
				item.ArchitectureSpecific = true
			},
			want: "architecture",
		},
		{
			name: "explicit size",
			change: func(item *metadataType) {
				item.Size = 16
			},
			want: "layout",
		},
		{
			name: "64 bit alignment",
			change: func(item *metadataType) {
				item.Fields[0].Type = "UInt64"
			},
			want: "Go provides 8/4",
		},
		{
			name: "field semantics",
			change: func(item *metadataType) {
				item.Fields[0].Unsupported = true
			},
			want: "field metadata",
		},
		{
			name: "explicit offset",
			change: func(item *metadataType) {
				item.Fields[0].Offset = 0
			},
			want: "field metadata",
		},
		{
			name: "unknown array length",
			change: func(item *metadataType) {
				item.Fields[0].Type = "UInt32[]"
			},
			want: "array shape",
		},
		{
			name: "pointer field",
			change: func(item *metadataType) {
				item.Fields[0].Type = "Void*"
			},
			want: "missing, ambiguous, or recursive",
		},
		{
			name: "recursive field",
			change: func(item *metadataType) {
				item.Fields[0].Type = "Windows.Win32.Foundation.TEST"
			},
			want: "recursive",
		},
		{
			name: "field name collision",
			change: func(item *metadataType) {
				item.Fields = append(item.Fields, fieldFixture("Value", "UInt32"))
			},
			want: "field name collision",
		},
		{
			name: "nonterminal flexible array",
			change: func(item *metadataType) {
				item.Fields[0].Type = "UInt32[1]"
				item.Fields[0].Flexible = true
				item.Fields = append(item.Fields, fieldFixture("tail", "UInt32"))
			},
			want: "field metadata",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := structureFixture("TEST", fieldFixture("value", "UInt32"))
			test.change(&item)
			_, err := newStructureProjection([]metadataType{item}).resolve("Windows.Win32.Foundation.TEST")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}

	item := structureFixture("TEST", fieldFixture("value", "UInt32"))
	_, err := newStructureProjection([]metadataType{item, item}).resolve("Windows.Win32.Foundation.TEST")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("duplicate metadata types: %v", err)
	}
}

func TestStructureFlexibleExtentAndDependencyFailure(t *testing.T) {
	t.Parallel()

	field := fieldFixture("data", "UIntPtr[7]")
	field.Flexible = true
	item := structureFixture("FLEX", fieldFixture("count", "UInt32"), field)
	projection := newStructureProjection([]metadataType{item})
	result, err := projection.resolve("Windows.Win32.Foundation.FLEX")
	if err != nil || result.Fields[1].Type != "[7]uintptr" || !result.Fields[1].Flexible {
		t.Fatalf("metadata flexible extent not preserved: %+v, %v", result, err)
	}

	dependency := structureFixture("ALIGNED", fieldFixture("value", "Int64"))
	parent := structureFixture("PARENT", fieldFixture("value", "Windows.Win32.Foundation.ALIGNED"))
	_, err = newStructureProjection([]metadataType{parent, dependency}).resolve("Windows.Win32.Foundation.PARENT")
	if err == nil || !strings.Contains(err.Error(), "PARENT.value") || !strings.Contains(err.Error(), "Windows 386") {
		t.Fatalf("dependency ABI failure not propagated: %v", err)
	}
}

func TestStructurePackingAndPointerScalars(t *testing.T) {
	t.Parallel()

	item := structureFixture("PACKED", fieldFixture("bytes", "Byte[3]"))
	item.Packing = 1
	_, err := newStructureProjection([]metadataType{item}).resolve("Windows.Win32.Foundation.PACKED")
	if err != nil {
		t.Fatalf("naturally compatible packing was rejected: %v", err)
	}

	signed := structureFixture("SIGNED_POINTER", fieldFixture("value", "IntPtr"))
	signed.Typedef = true
	result, err := newStructureProjection([]metadataType{signed}).resolve("Windows.Win32.Foundation.SIGNED_POINTER")
	if err != nil || result.Base != "int" || result.ABI[0].Size != 4 || result.ABI[1].Size != 8 {
		t.Fatalf("pointer-sized signed scalar = %+v, %v", result, err)
	}

	_, err = newStructureProjection(nil).resolve("UInt32[2147483647]")
	if err == nil || !strings.Contains(err.Error(), "portable Go object size") {
		t.Fatalf("oversized array was not rejected: %v", err)
	}
}

func TestWriteStructuresClearsStaleOutput(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	stage := t.TempDir()
	directory := filepath.Join(source, "foundation")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(directory, "zz_generated_structures.go")
	if err := os.WriteFile(path, []byte("package foundation; type STALE int"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := writeStructures(stage, source, nil, sourceLock{}); err != nil {
		t.Fatal(err)
	}

	contents, err := os.ReadFile(filepath.Join(stage, "foundation", "zz_generated_structures.go"))
	if err != nil || bytes.Contains(contents, []byte("STALE")) {
		t.Fatalf("stale output survived: %s, %v", contents, err)
	}
	checkStructureSource(t, contents, "386", false)
}

func TestCollectStructuresUniformScope(t *testing.T) {
	t.Parallel()

	first := structureFixture("NEW_STRUCTURE", fieldFixture("value", "UInt16"))
	second := structureFixture("ANOTHER_STRUCTURE", fieldFixture("value", "Windows.Win32.Foundation.NEW_STRUCTURE"))
	unsupported := structureFixture("UNSUPPORTED_STRUCTURE", fieldFixture("value", "Int64"))
	outside := structureFixture("OUTSIDE_STRUCTURE", fieldFixture("value", "UInt32"))
	outside.Namespace = "Windows.Win32.Other"
	items := []metadataType{first, second, unsupported, outside}
	packages, deferred := collectStructures(items)
	if len(packages["foundation"]) != 2 || len(packages) != 1 || len(deferred) != 1 {
		t.Fatalf("uniform scope: packages = %+v, deferred = %+v", packages, deferred)
	}

	before, err := renderStructures("foundation", packages["foundation"], sourceLock{})
	if err != nil {
		t.Fatal(err)
	}

	slices.Reverse(items)
	again, otherDeferred := collectStructures(items)
	after, err := renderStructures("foundation", again["foundation"], sourceLock{})
	if err != nil || !bytes.Equal(before, after) || !slices.Equal(deferred, otherDeferred) {
		t.Fatalf("metadata ordering changed generation: %v", err)
	}

	reserved := map[string]bool{
		"foundation.NEW_STRUCTURE": true,
	}
	collisions := filterStructureCollisions(packages, nil, reserved)
	if len(packages) != 0 || len(collisions) != 2 {
		t.Fatalf("collision did not exclude transitive dependents: %+v, %+v", packages, collisions)
	}
}

func TestExistingStructureSymbols(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	directory := filepath.Join(root, "foundation")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}

	for name, contents := range map[string]string{
		"existing.go":                "package foundation; type EXISTING int; const CONSTANT = 1; var Variable int; func Function() {}; func (EXISTING) Method() {}",
		"zz_generated_structures.go": "package foundation; type REGENERATED int",
		"existing_test.go":           "package foundation; type TEST_ONLY int",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	reserved, err := existingStructureSymbols(root, map[string][]projectedType{
		"foundation": nil,
	})
	if err != nil || len(reserved) != 4 {
		t.Fatalf("existing declarations = %+v, %v", reserved, err)
	}

	for _, name := range []string{"EXISTING", "CONSTANT", "Variable", "Function"} {
		if !reserved["foundation."+name] {
			t.Errorf("missing existing declaration %s", name)
		}
	}
}

func TestStructureRemovalFailsGeneration(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "internal", "source")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}

	previous := generationReport{
		Symbols: map[string][]string{
			"foundation": {"PREVIOUS_STRUCTURE"},
		},
	}
	contents, err := json.Marshal(previous)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(path, "report.json"), contents, 0o600); err != nil {
		t.Fatal(err)
	}

	err = validateGeneration(root, generationReport{}, false)
	if err == nil || !strings.Contains(err.Error(), "removed symbol: foundation.PREVIOUS_STRUCTURE") {
		t.Fatalf("structure coverage regression was not rejected: %v", err)
	}
}

// TestWindowsSDKStructures is an opt-in integration check with the pinned
// metadata and a native MSVC compiler. Set WIN32DEFS_ABI_CC to cl.exe and
// WIN32DEFS_ABI_ARCH to 386, amd64, or arm64 in a matching SDK environment.
func TestWindowsSDKStructures(t *testing.T) {
	compiler := os.Getenv("WIN32DEFS_ABI_CC")
	if compiler == "" {
		t.Skip("set WIN32DEFS_ABI_CC and WIN32DEFS_ABI_ARCH to verify against the Windows SDK")
	}

	arch := os.Getenv("WIN32DEFS_ABI_ARCH")
	index := 1
	if arch == "386" {
		index = 0
	} else if arch != "amd64" && arch != "arm64" {
		t.Fatal("WIN32DEFS_ABI_ARCH must be 386, amd64, or arm64")
	}

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}

	_, metadata, documentation, err := resolveSource(t.Context(), root, generationOptions{
		Offline:  true,
		CacheDir: filepath.Join(cache, "win32defs"),
	})
	if err != nil {
		t.Fatal(err)
	}

	winmd, err := extractWinMD(metadata)
	if err != nil {
		t.Fatal(err)
	}

	docs, err := extractDocs(documentation)
	if err != nil {
		t.Fatal(err)
	}

	export, err := exportMetadata(t.Context(), root, winmd, docs, true)
	if err != nil {
		t.Fatal(err)
	}

	packages, _ := collectStructures(export.Types)
	reserved, err := existingStructureSymbols(root, packages)
	if err != nil {
		t.Fatal(err)
	}

	constants, _, _, _ := collectConstants(export)
	filterStructureCollisions(packages, constants, reserved)

	var code strings.Builder
	code.WriteString("#include <windows.h>\n#include <stddef.h>\n#include <stdint.h>\n")
	fmt.Fprintf(&code, "static_assert(sizeof(uintptr_t) == %d, \"compiler architecture\");\n", 4+index*4)
	var names []string
	all := make(map[string]projectedType)
	for name := range packages {
		names = append(names, name)
		for _, item := range packages[name] {
			all[item.Source.Namespace+"."+item.Source.Name] = item
		}
	}
	slices.Sort(names)
	count := 0
	checks := 0
	for _, packageName := range names {
		for _, item := range packages[packageName] {
			// Core SDK headers provide an independent native check without requiring
			// optional SDK components. Select headers, never individual structures.
			coreHeader := strings.Contains(item.Source.Documentation, "/api/winnt/") ||
				strings.Contains(item.Source.Documentation, "/api/winuser/") ||
				strings.Contains(item.Source.Documentation, "/api/wingdi/")
			if item.Source.Enum || !coreHeader {
				continue
			}

			name := item.Source.Name
			layout := item.ABI[index]
			fmt.Fprintf(&code, "static_assert(sizeof(%s) == %d, \"%s size\");\n", name, layout.Size, name)
			fmt.Fprintf(&code, "static_assert(__alignof(%s) == %d, \"%s alignment\");\n", name, layout.Align, name)
			for fieldIndex := range item.Fields {
				field := item.Source.Fields[fieldIndex].Name
				check := nativeFieldCheck(&code, item, fieldIndex, layout.Offsets[fieldIndex], all, arch, index, &checks)
				fmt.Fprintf(&code, "static_assert(%s<%s>(), \"%s.%s layout\");\n", check, name, name, field)
			}
			count++
		}
	}
	if count == 0 {
		t.Fatal("no structures were checked against the SDK")
	}

	for _, constant := range export.Constants {
		if constant.Namespace == "Windows.Win32.Foundation" && constant.DeclaringType == "DUPLICATE_HANDLE_OPTIONS" {
			value, err := strconv.ParseUint(constant.Value, 10, 32)
			if err != nil {
				t.Fatal(err)
			}

			fmt.Fprintf(&code, "static_assert(%s == %d, \"%s value\");\n", constant.Name, value, constant.Name)
		}
	}

	directory := t.TempDir()
	source := filepath.Join(directory, "abi.cpp")
	if err := os.WriteFile(source, []byte(code.String()), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.CommandContext(t.Context(), compiler, "/nologo", "/std:c++20", "/c", source, "/Fo"+filepath.Join(directory, "abi.obj"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Windows SDK ABI validation: %v\n%s", err, output)
	}
	t.Logf("Windows %s SDK verified %d metadata-projected types and duplicate-handle flags", arch, count)
}

func nativeFieldCheck(code *strings.Builder, item projectedType, fieldIndex int, offset int64, all map[string]projectedType, arch string, abiIndex int, checks *int) string {
	field := item.Source.Fields[fieldIndex]
	nested := all[field.Type]
	var children []string
	for index := range nested.Fields {
		check := nativeFieldCheck(code, nested, index, offset+nested.ABI[abiIndex].Offsets[index], all, arch, abiIndex, checks)
		children = append(children, check+"<T>()")
	}
	*checks++
	name := fmt.Sprintf("check%d", *checks)
	goField := item.GoType.Underlying().(*types.Struct).Field(fieldIndex)
	size := types.SizesFor("gc", arch).Sizeof(goField.Type())
	fmt.Fprintf(code, "template<class T> consteval bool %s() {\n", name)
	fmt.Fprintf(code, "if constexpr (requires (T value) { value.%s; }) {\n", field.Name)
	fmt.Fprintf(code, "return offsetof(T, %s) == %d && sizeof(decltype(T::%s)) == %d;\n", field.Name, offset, field.Name, size)
	code.WriteString("} else {\n")
	if len(children) == 0 {
		code.WriteString("return false;\n")
	} else {
		// Metadata may group native inherited/flattened fields into a structure.
		// Verify those fields directly when no named native member exists.
		fmt.Fprintf(code, "return %s;\n", strings.Join(children, " && "))
	}
	code.WriteString("}\n}\n")

	return name
}
