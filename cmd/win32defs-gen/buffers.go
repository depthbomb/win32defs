package main

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"
)

func structureSymbols(item projectedType) []string {
	name := item.Source.Name
	symbols := []string{name}
	if item.Buffer {
		symbols = append(symbols, "New"+name, "View"+name, name+"Size", name+"Alignment")
		for _, field := range item.Fields {
			symbols = append(symbols, name+field.Name+"Offset")
		}
	}

	return symbols
}

func renderNativeBuffer(output *bytes.Buffer, packageName string, item projectedType) {
	name := item.Source.Name
	fmt.Fprintf(output, "// %s is a view of native %s.%s storage.\n", name, item.Source.Namespace, name)
	output.WriteString("// Its Go representation is not the native layout. Use New or View to initialize it,\n// and Pointer when passing it to Windows. Copies share storage; the zero value is invalid.\n")
	if item.Source.Documentation != "" {
		fmt.Fprintf(output, "// See %s.\n", item.Source.Documentation)
	}
	fmt.Fprintf(output, "type %s struct { data []byte }\n\n", name)
	output.WriteString("// Native size, alignment, and field offsets for the current pointer width.\nconst (\n")
	fmt.Fprintf(output, "%sSize = %s\n%sAlignment = %s\n", name, abiExpression(item.ABI[0].Size, item.ABI[1].Size), name, abiExpression(item.ABI[0].Align, item.ABI[1].Align))
	for index, field := range item.Fields {
		fmt.Fprintf(output, "%s%sOffset = %s\n", name, field.Name, abiExpression(item.ABI[0].Offsets[index], item.ABI[1].Offsets[index]))
	}
	output.WriteString(")\n\n")
	fmt.Fprintf(output, "// New%s allocates zeroed, aligned native storage.\nfunc New%s() %s {\nreturn %s{data: nativebuffer.New(int(%sSize), uintptr(%sAlignment))}\n}\n\n", name, name, name, name, name, name)
	fmt.Fprintf(output, "// View%s shares data without copying. It panics if data is shorter than %sSize.\n// Unaligned views support field access, but Pointer requires native alignment.\nfunc View%s(data []byte) %s {\nreturn %s{data: nativebuffer.View(data, int(%sSize))}\n}\n\n", name, name, name, name, name, name)
	fmt.Fprintf(output, "// Bytes returns the shared native storage, including padding and the initial flexible-array extent.\nfunc (b %s) Bytes() []byte {\nreturn b.data[:%sSize:%sSize]\n}\n\n", name, name, name)
	fmt.Fprintf(output, "// Pointer returns the native address. It panics for an invalid or unaligned view.\n// Keep the view alive while Windows uses it.\nfunc (b %s) Pointer() unsafe.Pointer {\nreturn nativebuffer.Pointer(b.Bytes(), uintptr(%sAlignment))\n}\n\n", name, name)
	for _, field := range item.Fields {
		resolved := field.Resolved
		offset := name + field.Name + "Offset"
		var parameters []string
		var checks []string
		for dimension := 0; ; dimension++ {
			array, ok := resolved.GoType.(*types.Array)
			if !ok {
				break
			}

			parameter := fmt.Sprintf("index%d", dimension)
			parameters = append(parameters, parameter+" int")
			checks = append(checks, fmt.Sprintf("if %s < 0 || %s >= %d {\npanic(\"native array index out of range\")\n}\n", parameter, parameter, array.Len()))
			resolved = resolved.Element
			offset += fmt.Sprintf(" + uintptr(%s)*(%s)", parameter, abiExpression(resolved.ABI[0].Size, resolved.ABI[1].Size))
		}
		fieldType := types.TypeString(resolved.GoType, packageQualifier(packageName))
		storage := fmt.Sprintf("b.Bytes()[offset:offset+(%s)]", abiExpression(resolved.ABI[0].Size, resolved.ABI[1].Size))
		get := "nativebuffer.Read[" + fieldType + "](" + storage + ")"
		set := "nativebuffer.Write(" + storage + ", value)"
		if resolved.Buffer {
			qualifier := ""
			if pkg := structurePackage(resolved.Source); pkg != packageName {
				qualifier = pkg + "."
			}
			get = qualifier + "View" + resolved.Source.Name + "(" + storage + ")"
			set = "copy(" + storage + ", value.Bytes())"
		}

		if field.Flexible {
			fmt.Fprintf(output, "// Get%s accesses the metadata-declared initial flexible-array extent only.\n", field.Name)
		} else if resolved.Buffer {
			fmt.Fprintf(output, "// Get%s returns a view sharing this buffer's storage.\n", field.Name)
		} else {
			fmt.Fprintf(output, "// Get%s returns a copy of the native field value.\n", field.Name)
		}
		fmt.Fprintf(output, "func (b %s) Get%s(%s) %s {\n%soffset := uintptr(%s)\n\nreturn %s\n}\n\n", name, field.Name, strings.Join(parameters, ", "), fieldType, strings.Join(checks, ""), offset, get)
		parameters = append(parameters, "value "+fieldType)
		fmt.Fprintf(output, "// Set%s copies value into the native field.\nfunc (b %s) Set%s(%s) {\n%soffset := uintptr(%s)\n%s\n}\n\n", field.Name, name, field.Name, strings.Join(parameters, ", "), strings.Join(checks, ""), offset, set)
	}
}
