package catalog

import (
	"iter"
	"math/bits"
	"sort"
	"strconv"
	"strings"
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

func definitionFamily(definition Definition) Family {
	return Family{
		Package:   definition.Package,
		Namespace: definition.Namespace,
		Name:      definition.Family,
	}
}

func normalizedFamilyValue(value string) string {
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
		seen := make(map[Family]bool)
		var families []Family
		for definition := range Definitions() {
			if packageName != "" && definition.Package != packageName {
				continue
			}

			family := definitionFamily(definition)
			if !seen[family] {
				seen[family] = true
				families = append(families, family)
			}
		}
		sort.Slice(families, func(left, right int) bool {
			if families[left].Package != families[right].Package {
				return families[left].Package < families[right].Package
			}

			if families[left].Namespace != families[right].Namespace {
				return families[left].Namespace < families[right].Namespace
			}

			return families[left].Name < families[right].Name
		})
		for _, family := range families {
			if !yield(family) {
				return
			}
		}
	}
}

// FamilyDefinitions iterates over a family's members in lexical name order.
func FamilyDefinitions(family Family) iter.Seq[Definition] {
	return func(yield func(Definition) bool) {
		for definition := range Definitions() {
			if definitionFamily(definition) == family && !yield(definition) {
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
	value = normalizedFamilyValue(value)
	var names []string
	for definition := range FamilyDefinitions(family) {
		if normalizedFamilyValue(definition.Value) == value {
			names = append(names, definition.Name)
		}
	}

	return names
}

// FormatFlags formats a value only for an enum marked as flags by the source
// metadata. Exact aliases take precedence, then masks with more bits, with
// lexical names breaking ties. Unknown bits are retained as a hexadecimal
// remainder. Zero is "0" when the enum has no zero-valued member.
func FormatFlags(family Family, value uint64) (string, bool) {
	var flags []flagDefinition
	found := false
	for definition := range FamilyDefinitions(family) {
		found = true
		if !definition.Flags {
			return "", false
		}

		mask, ok := flagBits(definition)
		if !ok {
			return "", false
		}

		if mask == value {
			return definition.Name, true
		}

		if mask != 0 {
			flags = append(flags, flagDefinition{
				name:  definition.Name,
				value: mask,
			})
		}
	}

	if !found {
		return "", false
	}

	sort.SliceStable(flags, func(left, right int) bool {
		return bits.OnesCount64(flags[left].value) > bits.OnesCount64(flags[right].value)
	})
	var names []string
	for _, flag := range flags {
		if value&flag.value == flag.value {
			names = append(names, flag.name)
			value &^= flag.value
		}
	}

	if value != 0 {
		names = append(names, "0x"+strconv.FormatUint(value, 16))
	}

	if len(names) == 0 {
		return "0", true
	}

	return strings.Join(names, " | "), true
}
