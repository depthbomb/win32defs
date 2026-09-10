package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

const bufferAccessorTests = `package foundation

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
	"unsafe"
)

func TestAccessors(t *testing.T) {
	b := NewPACKED()
	if len(b.Bytes()) != int(PACKEDSize) || uintptr(b.Pointer())%uintptr(PACKEDAlignment) != 0 {
		t.Fatal("invalid native allocation")
	}
	for i := range b.Bytes() {
		b.Bytes()[i] = 0xa5
	}
	b.SetValue(0xfedcba9876543210)
	if b.GetValue() != 0xfedcba9876543210 || binary.LittleEndian.Uint64(b.Bytes()[1:9]) != b.GetValue() || b.GetTag() != 0xa5 || b.Bytes()[9] != 0xa5 {
		t.Fatal("unaligned write corrupted field or neighbors")
	}
	bits := uint64(0x7ff8000000000042)
	b.SetReal(math.Float64frombits(bits))
	if math.Float64bits(b.GetReal()) != bits || binary.LittleEndian.Uint64(b.Bytes()[9:17]) != bits {
		t.Fatal("float payload was changed")
	}
	b.SetSigned(-9223372036854775807)
	b.SetPointer(^uintptr(0))
	if b.GetSigned() != -9223372036854775807 || b.GetPointer() != ^uintptr(0) {
		t.Fatal("signed or pointer-width value was changed")
	}
	direct := DIRECT{X: 0x12345678, Y: 0xabcd}
	b.SetDirect(direct)
	copyValue := b.GetDirect()
	if copyValue != direct {
		t.Fatal("direct structure copy failed")
	}
	copyValue.X = 0
	if b.GetDirect() != direct {
		t.Fatal("direct getter shared memory")
	}
	inner := b.GetInner()
	inner.SetValue(123)
	if b.GetInner().GetValue() != 123 || binary.LittleEndian.Uint64(b.Bytes()[PACKEDInnerOffset+INNERValueOffset:]) != 123 {
		t.Fatal("nested buffer did not share native storage")
	}
	standalone := NewINNER()
	standalone.SetTag(3)
	standalone.SetValue(456)
	if uintptr(standalone.Pointer())%8 != 0 {
		t.Fatal("8-byte native alignment was not honored")
	}
	b.SetInner(standalone)
	standalone.SetValue(789)
	if b.GetInner().GetValue() != 456 {
		t.Fatal("nested setter did not copy")
	}
	b.SetItems(1, standalone)
	b.GetItems(0).SetValue(987)
	if b.GetItems(0).GetValue() != 987 || b.GetItems(1).GetValue() != 789 || binary.LittleEndian.Uint64(b.Bytes()[PACKEDItemsOffset+INNERSize+INNERValueOffset:]) != 789 {
		t.Fatal("native array stride failed")
	}
	b.SetNumbers(1, 2, 0xcafe)
	if b.GetNumbers(1, 2) != 0xcafe || binary.LittleEndian.Uint16(b.Bytes()[PACKEDNumbersOffset+10:]) != 0xcafe {
		t.Fatal("multidimensional array failed")
	}
	b.SetFlex(1, 0x12345678)
	if b.GetFlex(1) != 0x12345678 || binary.LittleEndian.Uint32(b.Bytes()[PACKEDFlexOffset+4:]) != 0x12345678 {
		t.Fatal("flexible-array initial extent failed")
	}
	view := ViewPACKED(b.Bytes())
	view.SetTag(42)
	if b.GetTag() != 42 {
		t.Fatal("view did not share storage")
	}
	b.SetInner(b.GetInner())
	if !bytes.Equal(b.GetInner().Bytes(), b.Bytes()[PACKEDInnerOffset:PACKEDInnerOffset+INNERSize]) {
		t.Fatal("overlapping copy failed")
	}
	scalar := NewSCALAR()
	scalar.SetValue(-123)
	if scalar.GetValue() != -123 || binary.LittleEndian.Uint64(scalar.Bytes()) != ^uint64(122) {
		t.Fatal("scalar wrapper failed")
	}
}

func TestInvalidViews(t *testing.T) {
	b := NewPACKED()
	aligned := NewINNER()
	data := make([]byte, int(INNERSize)+8)
	offset := int((8-uintptr(unsafe.Pointer(&data[0]))%8)%8)+1
	unaligned := ViewINNER(data[offset:])
	unaligned.SetValue(7)
	if unaligned.GetValue() != 7 {
		t.Fatal("unaligned view cannot access fields")
	}
	for _, call := range []func(){
		func() { ViewINNER(aligned.Bytes()[:INNERSize-1]) },
		func() { unaligned.Pointer() },
		func() { INNER{}.Bytes() },
		func() { INNER{}.GetValue() },
		func() { b.GetItems(-1) },
		func() { b.SetItems(2, aligned) },
		func() { b.GetNumbers(2, 0) },
		func() { b.SetNumbers(0, 3, 1) },
		func() { b.SetFlex(2, 1) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid access did not panic")
				}
			}()
			call()
		}()
	}
}
`

func TestNativeLayoutsAndBufferSelection(t *testing.T) {
	t.Parallel()

	scalar := structureFixture("SCALAR", fieldFixture("value", "UInt64"))
	scalar.Typedef = true
	inner := structureFixture("INNER", fieldFixture("prefix", "Byte"), fieldFixture("value", "UInt64"))
	packed := inner
	packed.Name = "PACKED"
	packed.Packing = 1
	outer := structureFixture("OUTER", fieldFixture("tag", "Byte"), fieldFixture("items", "Windows.Win32.Foundation.INNER[2]"), fieldFixture("pointer", "UIntPtr"))
	items := []metadataType{scalar, inner, packed, outer}
	projection := newStructureProjection(items)
	tests := []struct {
		name string
		abi  [2]abiLayout
	}{
		{"SCALAR", [2]abiLayout{{8, 8, []int64{0}}, {8, 8, []int64{0}}}},
		{"INNER", [2]abiLayout{{16, 8, []int64{0, 8}}, {16, 8, []int64{0, 8}}}},
		{"PACKED", [2]abiLayout{{9, 1, []int64{0, 1}}, {9, 1, []int64{0, 1}}}},
		{"OUTER", [2]abiLayout{{48, 8, []int64{0, 8, 40}}, {48, 8, []int64{0, 8, 40}}}},
	}
	for _, test := range tests {
		result, err := projection.resolve("Windows.Win32.Foundation." + test.name)
		if err != nil || !result.Buffer || !reflect.DeepEqual(result.ABI, test.abi) {
			t.Fatalf("%s native layout = %+v, buffer = %v, error = %v", test.name, result.ABI, result.Buffer, err)
		}
	}

	packages, deferred := collectStructures(items)
	before, err := renderStructures("foundation", packages["foundation"], sourceLock{})
	if err != nil || len(deferred) != 0 || len(packages["foundation"]) != len(items) {
		t.Fatalf("uniform buffer generation: %v, %+v", err, deferred)
	}
	slices.Reverse(items)
	packages, _ = collectStructures(items)
	after, err := renderStructures("foundation", packages["foundation"], sourceLock{})
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("buffer generation depends on metadata order: %v", err)
	}
}

func TestBufferSymbolCollisions(t *testing.T) {
	t.Parallel()

	buffer := structureFixture("BUFFER", fieldFixture("value", "UInt64"))
	parent := structureFixture("PARENT", fieldFixture("value", "Windows.Win32.Foundation.BUFFER"))
	for _, symbol := range []string{"NewBUFFER", "ViewBUFFER", "BUFFERSize", "BUFFERAlignment", "BUFFERValueOffset"} {
		packages, _ := collectStructures([]metadataType{buffer, parent})
		deferred := filterStructureCollisions(packages, nil, map[string]bool{"foundation." + symbol: true})
		if len(packages) != 0 || len(deferred) != 2 {
			t.Fatalf("%s collision did not exclude dependents: %+v, %+v", symbol, packages, deferred)
		}
	}

	collision := structureFixture("BUFFERSize", fieldFixture("value", "UInt32"))
	packages, _ := collectStructures([]metadataType{buffer, parent, collision})
	deferred := filterStructureCollisions(packages, nil, map[string]bool{})
	if len(packages) != 0 || len(deferred) != 3 {
		t.Fatalf("generated symbol collision: %+v, %+v", packages, deferred)
	}
}

func TestGeneratedBufferAccessors(t *testing.T) {
	t.Parallel()

	inner := structureFixture("INNER", fieldFixture("tag", "Byte"), fieldFixture("value", "UInt64"))
	direct := structureFixture("DIRECT", fieldFixture("x", "UInt32"), fieldFixture("y", "UInt16"))
	flexible := fieldFixture("flex", "UInt32[2]")
	flexible.Flexible = true
	packed := structureFixture("PACKED", fieldFixture("tag", "Byte"), fieldFixture("value", "UInt64"), fieldFixture("real", "Double"), fieldFixture("signed", "Int64"), fieldFixture("pointer", "UIntPtr"), fieldFixture("direct", "Windows.Win32.Foundation.DIRECT"), fieldFixture("inner", "Windows.Win32.Foundation.INNER"), fieldFixture("items", "Windows.Win32.Foundation.INNER[2]"), fieldFixture("numbers", "UInt16[3][2]"), flexible)
	packed.Packing = 1
	scalar := structureFixture("SCALAR", fieldFixture("value", "Int64"))
	scalar.Typedef = true
	packages, deferred := collectStructures([]metadataType{inner, direct, packed, scalar})
	contents, err := renderStructures("foundation", packages["foundation"], sourceLock{})
	if err != nil || len(deferred) != 0 {
		t.Fatalf("render fixture: %v, %+v", err, deferred)
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	files := map[string][]byte{
		"go.mod":            fmt.Appendf(nil, "module github.com/depthbomb/win32defs/buffertest\n\ngo 1.26\n\nrequire github.com/depthbomb/win32defs v0.0.0\nreplace github.com/depthbomb/win32defs => %q\n", filepath.ToSlash(root)),
		"generated.go":      contents,
		"generated_test.go": []byte(bufferAccessorTests),
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(directory, name), contents, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.CommandContext(t.Context(), "go", "test", "-count=1", ".")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("generated buffer accessors: %v\n%s", err, output)
	}
}
