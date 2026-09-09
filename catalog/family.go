package catalog

import (
	"iter"
	"math/bits"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Family identifies a metadata enum or a named prefix within an API namespace.
type Family struct {
	Package   string
	Namespace string
	Name      string
}

type flagDefinition struct {
	name  string
	value uint64
}

type familyGroup struct {
	names   map[string][]string
	flags   []flagDefinition
	exact   map[uint64]string
	isFlags bool
}

type familyDescriptor struct {
	family Family
	start  int
	end    int
}

var familyGroups sync.Map

func compareFamilies(left Family, right Family) int {
	if order := strings.Compare(left.Package, right.Package); order != 0 {
		return order
	}

	if order := strings.Compare(left.Namespace, right.Namespace); order != 0 {
		return order
	}

	return strings.Compare(left.Name, right.Name)
}

func findFamily(family Family) *familyDescriptor {
	index := sort.Search(len(familyDescriptors), func(index int) bool {
		return compareFamilies(familyDescriptors[index].family, family) >= 0
	})
	if index == len(familyDescriptors) || familyDescriptors[index].family != family {
		return nil
	}

	return &familyDescriptors[index]
}

func buildFamilyGroup(descriptor *familyDescriptor) *familyGroup {
	group := &familyGroup{
		names:   make(map[string][]string),
		isFlags: true,
	}
	for _, position := range familyMembers[descriptor.start:descriptor.end] {
		definition := definitions[position]
		value := normalizedFamilyValue(definition.Value)
		group.names[value] = append(group.names[value], definition.Name)
		group.isFlags = group.isFlags && definition.Flags
	}

	if !group.isFlags {
		return group
	}

	group.exact = make(map[uint64]string)
	for _, position := range familyMembers[descriptor.start:descriptor.end] {
		definition := definitions[position]
		value, ok := flagBits(definition)
		if !ok {
			group.isFlags = false
			break
		}

		if _, exists := group.exact[value]; exists {
			continue
		}

		group.exact[value] = definition.Name
		if value != 0 {
			group.flags = append(group.flags, flagDefinition{
				name:  definition.Name,
				value: value,
			})
		}
	}
	sort.SliceStable(group.flags, func(left, right int) bool {
		return bits.OnesCount64(group.flags[left].value) > bits.OnesCount64(group.flags[right].value)
	})

	return group
}

func indexedFamily(family Family) *familyGroup {
	descriptor := findFamily(family)
	if descriptor == nil {
		return nil
	}

	if group, ok := familyGroups.Load(descriptor); ok {
		return group.(*familyGroup)
	}

	group, _ := familyGroups.LoadOrStore(descriptor, buildFamilyGroup(descriptor))

	return group.(*familyGroup)
}
func definitionFamily(definition Definition) Family {
	return Family{
		Package:   definition.Package,
		Namespace: definition.Namespace,
		Name:      definition.Family,
	}
}

func normalizedFamilyValue(value string) string {
	if value == "" || value[0] != '-' && value[0] != '+' && (value[0] < '0' || value[0] > '9') {
		return value
	}

	if number, err := strconv.ParseInt(value, 0, 64); err == nil {
		return strconv.FormatInt(number, 10)
	}

	if number, err := strconv.ParseUint(value, 0, 64); err == nil {
		return strconv.FormatUint(number, 10)
	}

	return value
}

func flagBits(definition Definition) (uint64, bool) {
	if value, err := strconv.ParseUint(definition.Value, 0, 64); err == nil {
		return value, true
	}

	value, err := strconv.ParseInt(definition.Value, 0, 64)
	if err != nil {
		return 0, false
	}

	switch definition.Kind {
	case "int8":
		return uint64(uint8(value)), true
	case "int16":
		return uint64(uint16(value)), true
	case "int32":
		return uint64(uint32(value)), true
	case "int64":
		return uint64(value), true
	}

	return 0, false
}

// Families iterates over a package's families in namespace and name order.
// An empty packageName includes every package.
func Families(packageName string) iter.Seq[Family] {
	return func(yield func(Family) bool) {
		start, end := 0, len(familyDescriptors)
		if packageName != "" {
			start = sort.Search(len(familyDescriptors), func(index int) bool {
				return familyDescriptors[index].family.Package >= packageName
			})
			end = sort.Search(len(familyDescriptors), func(index int) bool {
				return familyDescriptors[index].family.Package > packageName
			})
		}

		for index := start; index < end; index++ {
			if !yield(familyDescriptors[index].family) {
				return
			}
		}
	}
}

// FamilyDefinitions iterates over a family's members in lexical name order.
func FamilyDefinitions(family Family) iter.Seq[Definition] {
	return func(yield func(Definition) bool) {
		descriptor := findFamily(family)
		if descriptor == nil {
			return
		}

		for _, position := range familyMembers[descriptor.start:descriptor.end] {
			if !yield(definitions[position]) {
				return
			}
		}
	}
}

// LookupInFamily returns an exact symbol only when it belongs to family.
func LookupInFamily(family Family, name string) (Definition, bool) {
	definition, ok := Lookup(family.Package, name)
	if !ok || definitionFamily(definition) != family {
		return Definition{}, false
	}

	return definition, true
}

// NamesInFamily returns aliases for a value in lexical order. Integer values
// may use decimal or Go integer-literal notation. Other values must match the
// catalog's Value representation exactly, including quotes for string values.
func NamesInFamily(family Family, value string) []string {
	group := indexedFamily(family)
	if group == nil {
		return nil
	}

	return slices.Clone(group.names[normalizedFamilyValue(value)])
}

// FormatFlags formats a value only for an enum marked as flags by the source
// metadata. Exact aliases take precedence, then masks with more bits, with
// lexical names breaking ties. Unknown bits are retained as a hexadecimal
// remainder. Zero is "0" when the enum has no zero-valued member.
func FormatFlags(family Family, value uint64) (string, bool) {
	group := indexedFamily(family)
	if group == nil || !group.isFlags {
		return "", false
	}

	if name, ok := group.exact[value]; ok {
		return name, true
	}

	var output strings.Builder
	for _, flag := range group.flags {
		if value&flag.value == flag.value {
			if output.Len() != 0 {
				output.WriteString(" | ")
			} else {
				output.Grow(128)
			}

			output.WriteString(flag.name)
			value &^= flag.value
		}
	}

	if value != 0 {
		if output.Len() != 0 {
			output.WriteString(" | ")
		}

		var buffer [18]byte
		output.WriteString("0x")
		output.Write(strconv.AppendUint(buffer[:0], value, 16))
	}

	if output.Len() == 0 {
		return "0", true
	}

	return output.String(), true
}
