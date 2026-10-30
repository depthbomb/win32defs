package main

import "slices"

const compatibilitySource = "https://www.nuget.org/packages/Microsoft.Windows.SDK.Win32Metadata/71.0.30-preview"

// These definitions are absent from Win32 metadata 71.0.37-preview. Keep the
// published values and layouts, including their catalog and lookup entries.
var compatibilityConstants = []metadataConstant{
	{
		Namespace:     "Windows.Win32.Foundation",
		DeclaringType: "Apis",
		Name:          "STATUS_ADMINLESS_ACCESS_DENIED",
		ManagedType:   "Windows.Win32.Foundation.NTSTATUS",
		Documentation: compatibilitySource,
		Comment:       "Retained from Win32 metadata 71.0.30-preview for compatibility.",
		Kind:          "int32",
		Value:         "-1073700348",
	},
	{
		Namespace:     "Windows.Win32.System.Power",
		DeclaringType: "POWER_MONITOR_REQUEST_REASON",
		Name:          "MonitorRequestReasonAusterityBatteryDrain",
		ManagedType:   "Windows.Win32.System.Power.POWER_MONITOR_REQUEST_REASON",
		Documentation: compatibilitySource,
		Comment:       "Retained from Win32 metadata 71.0.30-preview for compatibility.",
		Kind:          "int32",
		Value:         "55",
		Enum:          true,
	},
	{
		Namespace:     "Windows.Win32.System.UpdateAgent",
		DeclaringType: "Apis",
		Name:          "WU_E_STANDBY_ACTIVITY_NOT_ALLOWED",
		ManagedType:   "Windows.Win32.Foundation.HRESULT",
		Documentation: compatibilitySource,
		Comment:       "Retained from Win32 metadata 71.0.30-preview for compatibility.",
		Kind:          "int32",
		Value:         "-2145124249",
	},
}

var compatibilityStructures = []metadataType{
	{
		Namespace:     "Windows.Win32.Security.Cryptography",
		Name:          "BCRYPT_MLKEM_KEY_BLOB",
		Struct:        true,
		Sequential:    true,
		Documentation: compatibilitySource,
		Fields: []metadataField{
			{Name: "dwMagic", Type: "UInt32", Offset: -1},
			{Name: "cbParameterSet", Type: "UInt32", Offset: -1},
			{Name: "cbKey", Type: "UInt32", Offset: -1},
		},
	},
	{
		Namespace:     "Windows.Win32.Security.Cryptography",
		Name:          "BCRYPT_PQDSA_KEY_BLOB",
		Struct:        true,
		Sequential:    true,
		Documentation: compatibilitySource,
		Fields: []metadataField{
			{Name: "dwMagic", Type: "UInt32", Offset: -1},
			{Name: "cbParameterSet", Type: "UInt32", Offset: -1},
			{Name: "cbKey", Type: "UInt32", Offset: -1},
		},
	},
}

func missingCompatibilityDefinitions(export metadataExport) metadataExport {
	var retained metadataExport
	for _, item := range compatibilityConstants {
		present := slices.ContainsFunc(export.Constants, func(candidate metadataConstant) bool {
			return candidate.Namespace == item.Namespace && candidate.DeclaringType == item.DeclaringType && candidate.Name == item.Name
		})
		if !present {
			retained.Constants = append(retained.Constants, item)
		}
	}

	for _, item := range compatibilityStructures {
		present := slices.ContainsFunc(export.Types, func(candidate metadataType) bool {
			return candidate.Namespace == item.Namespace && candidate.Name == item.Name
		})
		if !present {
			retained.Types = append(retained.Types, item)
		}
	}

	return retained
}
