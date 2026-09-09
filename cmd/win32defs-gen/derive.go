package main

import (
	"fmt"
	"strconv"
)

// deriveConstants supplies SDK expressions absent from the pinned metadata.
func deriveConstants(constants []metadataConstant) ([]metadataConstant, error) {
	const namespace = "Windows.Win32.UI.Shell"
	values := make(map[string]uint64, 3)
	for _, item := range constants {
		if item.Namespace != namespace || !hasExactName(item.Name, "NIN_SELECT", "NINF_KEY", "NIN_KEYSELECT") {
			continue
		}

		literal, err := normalizeIntegerText(item.Value, item.Kind, specByName("shell"))
		if err != nil {
			return nil, fmt.Errorf("derive NIN_KEYSELECT: invalid %s: %w", item.Name, err)
		}

		value, err := strconv.ParseUint(literal, 0, 32)
		if err != nil {
			return nil, fmt.Errorf("derive NIN_KEYSELECT: parse %s: %w", item.Name, err)
		}

		previous, exists := values[item.Name]
		if exists && previous != value {
			return nil, fmt.Errorf("derive NIN_KEYSELECT: conflicting %s values", item.Name)
		}

		values[item.Name] = value
	}

	for _, name := range []string{"NIN_SELECT", "NINF_KEY"} {
		_, exists := values[name]
		if !exists {
			return nil, fmt.Errorf("derive NIN_KEYSELECT: missing %s.%s", namespace, name)
		}
	}

	value := values["NIN_SELECT"] | values["NINF_KEY"]
	existing, exists := values["NIN_KEYSELECT"]
	if exists {
		if existing != value {
			return nil, fmt.Errorf("NIN_KEYSELECT metadata value %#x conflicts with SDK expression NIN_SELECT | NINF_KEY (%#x)", existing, value)
		}

		return nil, nil
	}

	return []metadataConstant{
		{
			Namespace:     namespace,
			DeclaringType: "Apis",
			Name:          "NIN_KEYSELECT",
			ManagedType:   "UInt32",
			Documentation: "https://github.com/microsoft/win32metadata/blob/29896383c51d9dd6a2ea0ec6304d095baca9418c/generation/WinSDK/RecompiledIdlHeaders/um/shellapi.h#L1048",
			Comment:       "Derived from NIN_SELECT | NINF_KEY, as defined in the Windows SDK shellapi.h header. The operands come from the pinned metadata.",
			Kind:          "uint32",
			Value:         strconv.FormatUint(value, 10),
		},
	}, nil
}
