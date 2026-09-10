package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type metadataType struct {
	Namespace            string          `json:"namespace"`
	Name                 string          `json:"name"`
	Struct               bool            `json:"struct"`
	Sequential           bool            `json:"sequential"`
	Packing              int             `json:"packing"`
	Size                 int             `json:"size"`
	Typedef              bool            `json:"typedef"`
	Enum                 bool            `json:"enum"`
	ArchitectureSpecific bool            `json:"architecture_specific"`
	Documentation        string          `json:"documentation"`
	Fields               []metadataField `json:"fields"`
}

type metadataField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Offset      int    `json:"offset"`
	Flexible    bool   `json:"flexible"`
	Unsupported bool   `json:"unsupported"`
}

type deferredStructure struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type abiLayout struct {
	Size    int64
	Align   int64
	Offsets []int64
}

type projectedType struct {
	GoType       types.Type
	ABI          [2]abiLayout
	Source       metadataType
	Fields       []projectedField
	Base         string
	Dependencies []string
}

type projectedField struct {
	Name     string
	Type     string
	Flexible bool
}

type structureProjection struct {
	Metadata map[string][]metadataType
	Types    map[string]projectedType
	Visiting map[string]bool
}

var structureNamespaces = map[string]string{
	"Windows.Win32.Foundation":                  "foundation",
	"Windows.Win32.Graphics.Gdi":                "gdi",
	"Windows.Win32.Graphics.Dwm":                "dwm",
	"Windows.Win32.Storage.FileSystem":          "filesystem",
	"Windows.Win32.System.Memory":               "memory",
	"Windows.Win32.System.Services":             "service",
	"Windows.Win32.System.Registry":             "registry",
	"Windows.Win32.System.Com":                  "com",
	"Windows.Win32.System.Ole":                  "com",
	"Windows.Win32.System.Variant":              "com",
	"Windows.Win32.System.Console":              "console",
	"Windows.Win32.System.JobObjects":           "jobobject",
	"Windows.Win32.System.Threading":            "process",
	"Windows.Win32.System.Power":                "power",
	"Windows.Win32.System.SystemInformation":    "sysinfo",
	"Windows.Win32.System.Diagnostics.ToolHelp": "toolhelp",
	"Windows.Win32.System.LibraryLoader":        "libraryloader",
	"Windows.Win32.Networking.WinSock":          "winsock",
	"Windows.Win32.Networking.WinHttp":          "winhttp",
	"Windows.Win32.UI.WindowsAndMessaging":      "winmsg",
	"Windows.Win32.UI.Shell":                    "shell",
	"Windows.Win32.UI.Controls":                 "controls",
	"Windows.Win32.UI.Controls.Dialogs":         "dialogs",
	"Windows.Win32.UI.Controls.RichEdit":        "richedit",
	"Windows.Win32.UI.HiDpi":                    "hidpi",
	"Windows.Win32.UI.Input":                    "input",
	"Windows.Win32.UI.Input.Pointer":            "input",
	"Windows.Win32.UI.Input.Touch":              "input",
	"Windows.Win32.System.Pipes":                "pipes",
	"Windows.Win32.Security":                    "security",
	"Windows.Win32.Security.Cryptography":       "cryptography",
	"Windows.Win32.Globalization":               "globalization",
	"Windows.Win32.System.Ioctl":                "ioctl",
}

var structurePrimitives = map[string]types.BasicKind{
	"SByte":   types.Int8,
	"Byte":    types.Uint8,
	"Int16":   types.Int16,
	"UInt16":  types.Uint16,
	"Char":    types.Uint16,
	"Int32":   types.Int32,
	"UInt32":  types.Uint32,
	"Int64":   types.Int64,
	"UInt64":  types.Uint64,
	"Single":  types.Float32,
	"Double":  types.Float64,
	"IntPtr":  types.Int,
	"UIntPtr": types.Uintptr,
}

func newStructureProjection(items []metadataType) *structureProjection {
	projection := &structureProjection{
		Metadata: make(map[string][]metadataType),
		Types:    make(map[string]projectedType),
		Visiting: make(map[string]bool),
	}
	for _, item := range items {
		name := item.Namespace + "." + item.Name
		projection.Metadata[name] = append(projection.Metadata[name], item)
	}

	return projection
}

func (projection *structureProjection) resolve(name string) (projectedType, error) {
	kind, primitive := structurePrimitives[name]
	if primitive {
		goType := types.Typ[kind]
		result := projectedType{
			GoType: goType,
		}
		for index, pointerSize := range []int64{4, 8} {
			size := (&types.StdSizes{WordSize: pointerSize, MaxAlign: 8}).Sizeof(goType)
			result.ABI[index] = abiLayout{
				Size:  size,
				Align: size,
			}
		}

		return result, nil
	}

	array := strings.LastIndexByte(name, '[')
	if array >= 0 && strings.HasSuffix(name, "]") {
		count, err := strconv.ParseInt(name[array+1:len(name)-1], 10, 32)
		if err != nil || count <= 0 {
			return projectedType{}, fmt.Errorf("unsupported array shape %s", name)
		}

		element, err := projection.resolve(name[:array])
		if err != nil {
			return projectedType{}, err
		}

		result := projectedType{
			GoType: types.NewArray(element.GoType, count),
		}
		for index, layout := range element.ABI {
			if layout.Size > math.MaxInt32/count {
				return projectedType{}, fmt.Errorf("array exceeds portable Go object size: %s", name)
			}

			result.ABI[index] = abiLayout{
				Size:  layout.Size * count,
				Align: layout.Align,
			}
		}

		return result, nil
	}

	result, found := projection.Types[name]
	if found {
		return result, nil
	}

	definitions := projection.Metadata[name]
	if len(definitions) != 1 || projection.Visiting[name] {
		return projectedType{}, fmt.Errorf("missing, ambiguous, or recursive type %s", name)
	}

	item := definitions[0]
	packageName := structurePackage(item)
	if packageName == "" || !token.IsIdentifier(item.Name) || !token.IsExported(item.Name) {
		return projectedType{}, fmt.Errorf("unsupported type name or namespace %s", name)
	}

	invalidPacking := item.Packing < 0 || item.Packing > 128 || (item.Packing != 0 && item.Packing&(item.Packing-1) != 0)
	if (!item.Struct && !item.Enum) || item.ArchitectureSpecific || (!item.Sequential && !item.Enum) || invalidPacking || item.Size != 0 || len(item.Fields) == 0 {
		return projectedType{}, fmt.Errorf("unsupported architecture, layout, packing, or empty type %s", name)
	}

	projection.Visiting[name] = true
	defer delete(projection.Visiting, name)
	result.Source = item
	var fields []*types.Var
	for index, field := range item.Fields {
		if field.Unsupported || field.Offset != -1 || (field.Flexible && (index != len(item.Fields)-1 || !strings.Contains(field.Type, "["))) {
			return projectedType{}, fmt.Errorf("unsupported field metadata %s.%s", name, field.Name)
		}

		fieldType := field.Type
		if item.Typedef && len(item.Fields) == 1 && fieldType == "Void*" {
			fieldType = "UIntPtr"
		}

		resolved, err := projection.resolve(fieldType)
		if err != nil {
			return projectedType{}, fmt.Errorf("%s.%s: %w", name, field.Name, err)
		}
		result.Dependencies = append(result.Dependencies, structureDependencies(resolved.GoType)...)

		fieldName := exportFieldName(field.Name)
		if !token.IsIdentifier(fieldName) || !token.IsExported(fieldName) {
			return projectedType{}, fmt.Errorf("invalid field name %s.%s", name, field.Name)
		}

		for _, existing := range fields {
			if existing.Name() == fieldName {
				return projectedType{}, fmt.Errorf("field name collision %s.%s", name, fieldName)
			}
		}
		fields = append(fields, types.NewVar(token.NoPos, nil, fieldName, resolved.GoType))
		result.Fields = append(result.Fields, projectedField{
			Name:     fieldName,
			Type:     types.TypeString(resolved.GoType, packageQualifier(packageName)),
			Flexible: field.Flexible,
		})
		for arch, layout := range resolved.ABI {
			current := &result.ABI[arch]
			alignment := layout.Align
			if item.Packing != 0 {
				alignment = min(alignment, int64(item.Packing))
			}

			offset := alignSize(current.Size, alignment)
			if offset > math.MaxInt32-layout.Size {
				return projectedType{}, fmt.Errorf("structure exceeds portable Go object size: %s", name)
			}

			current.Offsets = append(current.Offsets, offset)
			current.Size = offset + layout.Size
			current.Align = max(current.Align, alignment)
		}
	}

	var underlying types.Type = types.NewStruct(fields, nil)
	if item.Typedef || item.Enum {
		if len(fields) != 1 || (item.Enum && item.Fields[0].Name != "value__") {
			return projectedType{}, fmt.Errorf("invalid scalar wrapper %s", name)
		}

		if _, ok := fields[0].Type().Underlying().(*types.Basic); !ok {
			return projectedType{}, fmt.Errorf("non-scalar wrapper %s", name)
		}

		underlying = fields[0].Type().Underlying()
		result.Base = result.Fields[0].Type
		result.Fields = nil
	}

	for arch, goarch := range []string{"386", "amd64"} {
		layout := &result.ABI[arch]
		layout.Size = alignSize(layout.Size, layout.Align)
		sizes := types.SizesFor("gc", goarch)
		if sizes.Sizeof(underlying) != layout.Size || sizes.Alignof(underlying) != layout.Align {
			return projectedType{}, fmt.Errorf("%s: Windows %s requires size/alignment %d/%d; Go provides %d/%d", name, goarch, layout.Size, layout.Align, sizes.Sizeof(underlying), sizes.Alignof(underlying))
		}

		if result.Base != "" {
			continue
		}

		for index, offset := range sizes.Offsetsof(fields) {
			if offset != layout.Offsets[index] {
				return projectedType{}, fmt.Errorf("%s.%s: Windows %s field offset differs from Go", name, fields[index].Name(), goarch)
			}
		}
	}

	goPackage := types.NewPackage("github.com/depthbomb/win32defs/"+packageName, packageName)
	result.GoType = types.NewNamed(types.NewTypeName(token.NoPos, goPackage, item.Name, nil), underlying, nil)
	projection.Types[name] = result

	return result, nil
}

func alignSize(size int64, alignment int64) int64 {
	return (size + alignment - 1) / alignment * alignment
}

func exportFieldName(name string) string {
	characters := []rune(name)
	if len(characters) > 0 {
		characters[0] = unicode.ToUpper(characters[0])
	}

	return string(characters)
}

func packageQualifier(current string) types.Qualifier {
	return func(pkg *types.Package) string {
		if pkg.Name() == current {
			return ""
		}

		return pkg.Name()
	}
}

func structurePackage(item metadataType) string {
	packageName := structureNamespaces[item.Namespace]
	if packageName != "" {
		return packageName
	}

	// Use the library's existing family rules for namespaces shared by several
	// domains, such as PE and exception definitions in diagnostics metadata.
	candidate := metadataConstant{
		Namespace: item.Namespace,
		Name:      item.Name,
	}
	for _, spec := range packageSpecs {
		if spec.Match(candidate) {
			return spec.Name
		}
	}

	return ""
}

func collectStructures(items []metadataType) (map[string][]projectedType, []deferredStructure) {
	packages := make(map[string][]projectedType)
	var deferred []deferredStructure
	seen := make(map[string]bool)
	projection := newStructureProjection(items)
	for _, candidate := range items {
		if !candidate.Struct || structurePackage(candidate) == "" {
			continue
		}

		root := candidate.Namespace + "." + candidate.Name
		if seen[root] {
			continue
		}

		seen[root] = true
		_, err := projection.resolve(root)
		if err != nil {
			deferred = append(deferred, deferredStructure{
				Name:   root,
				Reason: err.Error(),
			})
			continue
		}
	}
	for _, item := range projection.Types {
		packageName := structurePackage(item.Source)
		packages[packageName] = append(packages[packageName], item)
	}
	for name := range packages {
		sort.Slice(packages[name], func(i, j int) bool {
			return packages[name][i].Source.Name < packages[name][j].Source.Name
		})
	}
	sort.Slice(deferred, func(i, j int) bool {
		return deferred[i].Name < deferred[j].Name
	})

	return packages, deferred
}

func structureDependencies(item types.Type) []string {
	switch value := item.(type) {
	case *types.Named:
		return []string{value.Obj().Pkg().Name() + "." + value.Obj().Name()}
	case *types.Array:
		return structureDependencies(value.Elem())
	default:
		return nil
	}
}

func filterStructureCollisions(packages map[string][]projectedType, constants map[string][]generatedConstant, reserved map[string]bool) []deferredStructure {
	for _, spec := range packageSpecs {
		reserved[spec.Name+"."+spec.TypeName] = true
	}
	for packageName, items := range constants {
		for _, item := range items {
			reserved[packageName+"."+item.Name] = true
		}
	}

	blocked := make(map[string]string)
	counts := make(map[string]int)
	for packageName, items := range packages {
		for _, item := range items {
			key := packageName + "." + item.Source.Name
			counts[key]++
			if reserved[key] {
				blocked[key] = "collides with existing symbol " + key
			}

			if counts[key] > 1 {
				blocked[key] = "ambiguous projected type name " + key
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for packageName, items := range packages {
			for _, item := range items {
				key := packageName + "." + item.Source.Name
				if blocked[key] != "" {
					continue
				}

				for _, dependency := range item.Dependencies {
					if blocked[dependency] != "" {
						blocked[key] = "unavailable dependency"
						changed = true
						break
					}
				}
			}
		}
	}

	var deferred []deferredStructure
	for packageName, items := range packages {
		var supported []projectedType
		for _, item := range items {
			reason := blocked[packageName+"."+item.Source.Name]
			if reason != "" {
				if reason == "unavailable dependency" {
					for _, dependency := range item.Dependencies {
						if blocked[dependency] != "" {
							reason += " " + dependency
							break
						}
					}
				}

				deferred = append(deferred, deferredStructure{
					Name:   item.Source.Namespace + "." + item.Source.Name,
					Reason: reason,
				})
				continue
			}

			supported = append(supported, item)
		}
		if len(supported) == 0 {
			delete(packages, packageName)
		} else {
			packages[packageName] = supported
		}
	}

	return deferred
}

func existingStructureSymbols(root string, packages map[string][]projectedType) (map[string]bool, error) {
	reserved := make(map[string]bool)
	for packageName := range packages {
		directory := filepath.Join(root, packageName)
		entries, err := os.ReadDir(directory)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || name == "zz_generated_structures.go" {
				continue
			}

			file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(directory, name), nil, 0)
			if err != nil {
				return nil, err
			}

			for _, declaration := range file.Decls {
				switch item := declaration.(type) {
				case *ast.FuncDecl:
					if item.Recv == nil {
						reserved[packageName+"."+item.Name.Name] = true
					}
				case *ast.GenDecl:
					for _, spec := range item.Specs {
						switch value := spec.(type) {
						case *ast.TypeSpec:
							reserved[packageName+"."+value.Name.Name] = true
						case *ast.ValueSpec:
							for _, identifier := range value.Names {
								reserved[packageName+"."+identifier.Name] = true
							}
						}
					}
				}
			}
		}
	}

	return reserved, nil
}

func renderStructures(packageName string, items []projectedType, source sourceLock) ([]byte, error) {
	var output bytes.Buffer
	output.WriteString(generatedHeader)
	fmt.Fprintf(&output, "// Source: %s %s (%s).\n\npackage %s\n\n", source.Metadata.Package, source.Metadata.Version, source.Metadata.SHA256, packageName)
	if len(items) == 0 {
		return format.Source(output.Bytes())
	}

	imports := make(map[string]bool)
	for _, item := range items {
		types.TypeString(item.GoType.Underlying(), func(pkg *types.Package) string {
			if pkg.Name() != packageName {
				imports[pkg.Path()] = true
			}

			return pkg.Name()
		})
	}
	var paths []string
	for path := range imports {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	output.WriteString("import (\n\"unsafe\"\n\n")
	for _, path := range paths {
		fmt.Fprintf(&output, "%q\n", path)
	}
	output.WriteString(")\n\n")
	for _, item := range items {
		name := item.Source.Name
		fmt.Fprintf(&output, "// %s projects %s.%s.\n", name, item.Source.Namespace, name)
		if item.Source.Documentation != "" {
			fmt.Fprintf(&output, "// See %s.\n", item.Source.Documentation)
		}

		if item.Base != "" {
			fmt.Fprintf(&output, "type %s %s\n\n", name, item.Base)
			continue
		}

		fmt.Fprintf(&output, "type %s struct {\n", name)
		for _, field := range item.Fields {
			if field.Flexible {
				output.WriteString("// Flexible array: this is the metadata-declared initial extent.\n// Additional elements require a larger native allocation.\n")
			}
			fmt.Fprintf(&output, "%s %s\n", field.Name, field.Type)
		}
		output.WriteString("}\n\n")
	}
	output.WriteString("// Verify metadata-derived Windows ABI sizes, alignments, and field offsets.\n")
	output.WriteString("var (\n")
	for _, item := range items {
		name := item.Source.Name
		value := "*new(" + name + ")"
		writeABIAssertion(&output, "unsafe.Sizeof("+value+")", item.ABI[0].Size, item.ABI[1].Size)
		writeABIAssertion(&output, "unsafe.Alignof("+value+")", item.ABI[0].Align, item.ABI[1].Align)
		for index, field := range item.Fields {
			writeABIAssertion(&output, "unsafe.Offsetof("+name+"{}."+field.Name+")", item.ABI[0].Offsets[index], item.ABI[1].Offsets[index])
		}
	}
	output.WriteString(")\n")

	return format.Source(output.Bytes())
}

func writeABIAssertion(output *bytes.Buffer, expression string, size32 int64, size64 int64) {
	expected := strconv.FormatInt(size32, 10)
	if size32 != size64 {
		expected += fmt.Sprintf(" + (unsafe.Sizeof(uintptr(0))-4)/4*%d", size64-size32)
	}
	fmt.Fprintf(output, "_ [%s]byte = [%s]byte{}\n", expected, expression)
}

func writeStructures(root string, sourceRoot string, packages map[string][]projectedType, source sourceLock) error {
	for _, spec := range packageSpecs {
		name := spec.Name
		items := packages[name]
		if len(items) == 0 {
			_, err := os.Stat(filepath.Join(sourceRoot, name, "zz_generated_structures.go"))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			if err != nil {
				return err
			}
		}

		path := filepath.Join(root, name, "zz_generated_structures.go")
		contents, err := renderStructures(name, items, source)
		if err != nil {
			return err
		}

		if err := writeFormatted(path, contents); err != nil {
			return err
		}
	}

	return nil
}
