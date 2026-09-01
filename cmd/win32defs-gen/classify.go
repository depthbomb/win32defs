package main

import (
	"fmt"
	"go/token"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

var packageSpecs = []packageSpec{
	{
		Name: "hresult", Doc: "HRESULT values and classification helpers.", TypeName: "Code",
		Underlying: "int32", Width: 32, Signed: true,
		Match: func(item metadataConstant) bool {
			return strings.HasSuffix(item.ManagedType, ".HRESULT") || item.NativeType == "HRESULT"
		},
		CanonicalRank: hresultCanonicalRank,
	},
	{
		Name: "winerror", Doc: "Win32 system error codes.", TypeName: "Code",
		Underlying: "uint32", Width: 32,
		Match: func(item metadataConstant) bool {
			return item.DeclaringType == "WIN32_ERROR" ||
				hasAnyPrefix(
					item.Name,
					"ERROR_", "APPMODEL_ERROR_", "DNS_ERROR_", "DNS_INFO_", "DNS_REQUEST_", "DNS_STATUS_", "DNS_WARNING_",
					"EPT_S_", "FRS_ERR_", "NERR_", "OR_", "PEERDIST_ERROR_", "RPC_S_", "RPC_X_", "STORE_ERROR_", "WARNING_",
				) || item.Name == "NO_ERROR"
		},
		CanonicalRank: winerrorCanonicalRank,
	},
	{
		Name: "ntstatus", Doc: "NTSTATUS values and classification helpers.", TypeName: "Code",
		Underlying: "uint32", Width: 32,
		Match: func(item metadataConstant) bool {
			return isNTStatusConstant(item) ||
				hasAnyPrefix(item.Name, "RPC_NT_") ||
				(strings.HasPrefix(item.Name, "STATUS_") &&
					hasExactName(item.Namespace, "Windows.Win32.Foundation", "Windows.Win32.System.VirtualDosMachines"))
		},
		CanonicalRank: ntstatusCanonicalRank,
	},
	{
		Name: "exitcode", Doc: "process exit codes defined by Windows.", TypeName: "Code",
		Underlying: "uint32", Width: 32,
		Match: func(item metadataConstant) bool {
			return item.Name == "STILL_ACTIVE" || item.Name == "CONTROL_C_EXIT"
		},
	},
	{
		Name: "facility", Doc: "HRESULT facility identifiers.", TypeName: "Code",
		Underlying: "uint16", Width: 16,
		Match: func(item metadataConstant) bool {
			return strings.HasPrefix(item.Name, "FACILITY_")
		},
	},
	{
		Name: "wait", Doc: "wait and synchronization result values.", TypeName: "Value",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			return strings.HasPrefix(item.Name, "WAIT_") || item.Name == "INFINITE" || item.Name == "MAXIMUM_WAIT_OBJECTS"
		},
	},
	{
		Name: "exception", Doc: "structured exception and debugger status codes.", TypeName: "Code",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			return strings.HasPrefix(item.Name, "EXCEPTION_") ||
				(strings.HasPrefix(item.Name, "DBG_") && isNTStatusConstant(item))
		},
	},
	{
		Name: "access", Doc: "common Win32 access masks and standard rights.", TypeName: "Mask",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			return hasAnyPrefix(item.Name, "GENERIC_", "STANDARD_RIGHTS_", "SPECIFIC_RIGHTS_") ||
				hasExactName(item.Name, "ACCESS_SYSTEM_SECURITY", "MAXIMUM_ALLOWED", "DELETE", "READ_CONTROL", "WRITE_DAC", "WRITE_OWNER", "SYNCHRONIZE")
		},
	},
	{
		Name: "foundation", Doc: "common Win32 boolean, path, and invalid-handle sentinel values.", TypeName: "Value",
		Underlying: "int64", Width: 64, Signed: true, Declare: true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.Foundation" &&
				hasExactName(item.Name, "FALSE", "TRUE", "INVALID_HANDLE_VALUE", "MAX_PATH")
		},
	},
	{
		Name: "filesystem", Doc: "filesystem attributes, sharing modes, dispositions, and operation flags.", TypeName: "Value",
		Underlying: "uint64", Width: 64, Declare: true,
		Match: func(item metadataConstant) bool {
			return hasAnyPrefix(item.Name, "FILE_", "CREATE_", "OPEN_", "TRUNCATE_", "COPY_FILE_", "MOVEFILE_", "REPLACEFILE_") ||
				(item.Namespace == "Windows.Win32.Storage.FileSystem" &&
					hasExactName(item.Name, "INVALID_FILE_ATTRIBUTES", "INVALID_FILE_SIZE", "INVALID_SET_FILE_POINTER"))
		},
	},
	{
		Name: "memory", Doc: "virtual memory, section, mapping, and protection flags.", TypeName: "Value",
		Underlying: "uint64", Width: 64, Declare: true,
		Match: func(item metadataConstant) bool {
			return hasAnyPrefix(item.Name, "PAGE_", "MEM_", "SECTION_", "SEC_", "FILE_MAP_")
		},
	},
	{
		Name: "service", Doc: "Service Control Manager states, controls, access rights, and flags.", TypeName: "Value",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			return hasAnyPrefix(item.Name, "SERVICE_", "SC_MANAGER_")
		},
	},
	{
		Name: "registry", Doc: "registry value types, access masks, options, and retrieval flags.", TypeName: "Value",
		Underlying: "uint64", Width: 64, Declare: true,
		Match: func(item metadataConstant) bool {
			return hasAnyPrefix(item.Name, "REG_", "KEY_", "RRF_", "HKEY_")
		},
	},
	{
		Name: "com", Doc: "COM initialization, activation, storage, marshaling, and data-transfer flags.", TypeName: "Value",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			return hasAnyPrefix(item.Name, "CLSCTX_", "COINIT_", "STGM_", "DVASPECT_", "TYMED_", "MSHCTX_", "MSHLFLAGS_")
		},
	},
	{
		Name: "console", Doc: "console modes, control events, standard handles, colors, and flags.", TypeName: "Value",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			if !strings.Contains(item.Namespace, ".Console") {
				return false
			}

			return hasAnyPrefix(item.Name, "CONSOLE_", "ENABLE_", "CTRL_", "STD_", "FOREGROUND_", "BACKGROUND_", "COMMON_LVB_") ||
				hasExactName(item.Name, "ATTACH_PARENT_PROCESS", "DISABLE_NEWLINE_AUTO_RETURN")
		},
	},
	{
		Name: "process", Doc: "process and thread creation, startup, debugging, priority flags, and pseudo-handles.", TypeName: "Value",
		Underlying: "uint64", Width: 64, Declare: true,
		Match: func(item metadataConstant) bool {
			if !strings.Contains(item.Namespace, ".Threading") {
				return false
			}

			return hasAnyPrefix(
				item.Name,
				"CREATE_", "STARTF_", "PROCESS_", "THREAD_", "PROC_THREAD_ATTRIBUTE_", "WT_", "TLS_", "FLS_",
				"EVENT_", "MUTEX_", "SEMAPHORE_", "TIMER_",
			) ||
				hasAnySuffix(item.Name, "_PRIORITY_CLASS") ||
				hasExactName(item.Name, "DEBUG_PROCESS", "DEBUG_ONLY_THIS_PROCESS", "DETACHED_PROCESS") ||
				hasExactName(item.DeclaringType, "PROCESS_MITIGATION_POLICY", "PROC_THREAD_ATTRIBUTE_NUM", "SYNCHRONIZATION_ACCESS_RIGHTS")
		},
		CanonicalRank: processCanonicalRank,
	},
	{
		Name: "jobobject", Doc: "job object limits, controls, information classes, and rate-control flags.", TypeName: "Value",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			return isEnumInNamespace(item, "Windows.Win32.System.JobObjects")
		},
	},
	{
		Name: "power", Doc: "power management states, actions, policies, and notification values.", TypeName: "Value",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			return isEnumInNamespace(item, "Windows.Win32.System.Power")
		},
	},
	{
		Name: "sysinfo", Doc: "system architecture, product, version, and processor information values.", TypeName: "Value",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			return isEnumInNamespace(item, "Windows.Win32.System.SystemInformation")
		},
	},
	{
		Name: "toolhelp", Doc: "Tool Help snapshot and heap-entry flags.", TypeName: "Value",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			return isEnumInNamespace(item, "Windows.Win32.System.Diagnostics.ToolHelp")
		},
	},
	{
		Name: "libraryloader", Doc: "library-loading and module-handle flags.", TypeName: "Value",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			return isEnumInNamespace(item, "Windows.Win32.System.LibraryLoader")
		},
	},
	{
		Name: "winsock", Doc: "Winsock errors, address families, protocols, options, and message flags.", TypeName: "Value",
		Underlying: "int64", Width: 64, Signed: true, Declare: true,
		Match: func(item metadataConstant) bool {
			if !strings.Contains(item.Namespace, ".WinSock") {
				return false
			}

			return hasAnyPrefix(
				item.Name,
				"WSA", "AF_", "PF_", "SOCK_", "IPPROTO_", "SO_", "MSG_", "AI_", "NI_", "FD_", "POLL", "SD_",
				"SIO_", "IOC_", "IP_", "IPV6_", "TCP_", "UDP_",
			) ||
				hasExactName(item.Name, "SOL_SOCKET", "SOMAXCONN", "SOCKET_ERROR")
		},
		CanonicalRank: winsockCanonicalRank,
	},
	{
		Name: "winmsg", Doc: "window messages, styles, show states, message-box flags, and virtual keys.", TypeName: "Value",
		Underlying: "int64", Width: 64, Signed: true, Declare: true,
		Match: func(item metadataConstant) bool {
			if !strings.Contains(item.Namespace, ".WindowsAndMessaging") && !strings.Contains(item.Namespace, ".Input.KeyboardAndMouse") {
				return false
			}

			return hasAnyPrefix(
				item.Name,
				"WM_", "WS_", "SW_", "MB_", "VK_", "MOD_", "HOTKEYF_", "GWL_", "GWLP_", "SPI_", "SM_", "SWP_",
				"TPM_", "MF_", "CS_", "WH_", "QS_", "SB_", "OCR_", "IDC_", "IDI_", "GCL_", "GCLP_", "HWND_",
				"MOUSEEVENTF_", "KEYEVENTF_", "KLF_", "MAPVK_", "TME_", "INPUT_",
			) || hasExactName(item.Name, "CW_USEDEFAULT", "WHEEL_DELTA")
		},
		CanonicalRank: winmsgCanonicalRank,
	},
	{
		Name: "ioctl", Doc: "I/O control device types, methods, access flags, and control codes.", TypeName: "Value",
		Underlying: "uint32", Width: 32, Declare: true,
		Match: func(item metadataConstant) bool {
			return hasAnyPrefix(item.Name, "FILE_DEVICE_", "METHOD_", "IOCTL_", "FSCTL_")
		},
	},
}

func isNTStatusConstant(item metadataConstant) bool {
	return strings.HasSuffix(item.ManagedType, ".NTSTATUS") || item.NativeType == "NTSTATUS"
}

func isEnumInNamespace(item metadataConstant, namespace string) bool {
	return item.Enum && item.Namespace == namespace
}

func hresultCanonicalRank(name string) int {
	if name == "S_OK" {
		return 0
	}

	if strings.HasPrefix(name, "S_") || strings.HasPrefix(name, "E_") {
		return 10
	}

	return 100
}

func winerrorCanonicalRank(name string) int {
	if name == "ERROR_SUCCESS" {
		return 0
	}

	if strings.HasPrefix(name, "ERROR_") {
		return 10
	}

	return 100
}

func ntstatusCanonicalRank(name string) int {
	if name == "STATUS_SUCCESS" {
		return 0
	}

	if strings.HasPrefix(name, "STATUS_") {
		return 10
	}

	if strings.HasPrefix(name, "RPC_NT_") {
		return 20
	}

	return 100
}

func processCanonicalRank(name string) int {
	if hasAnyPrefix(name, "CREATE_", "STARTF_", "PROCESS_", "THREAD_") ||
		hasAnySuffix(name, "_PRIORITY_CLASS") ||
		hasExactName(name, "DEBUG_PROCESS", "DEBUG_ONLY_THIS_PROCESS", "DETACHED_PROCESS") {
		return 0
	}

	return 10
}

func winsockCanonicalRank(name string) int {
	if hasAnyPrefix(name, "WSA", "AF_", "PF_", "SOCK_", "IPPROTO_", "SO_", "MSG_", "AI_", "NI_") ||
		hasExactName(name, "SOL_SOCKET", "SOMAXCONN", "SOCKET_ERROR") {
		return 0
	}

	return 10
}

func winmsgCanonicalRank(name string) int {
	if hasAnyPrefix(name, "WM_", "WS_", "SW_", "MB_", "VK_", "MOD_", "HOTKEYF_", "GWL_", "GWLP_") {
		return 0
	}

	return 10
}

func hasAnyPrefix(value string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}

	return false
}

func hasAnySuffix(value string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(value, suffix) {
			return true
		}
	}

	return false
}

func hasExactName(value string, names ...string) bool {
	for _, name := range names {
		if value == name {
			return true
		}
	}

	return false
}

func isNumericKind(kind string) bool {
	return kind != "string" && kind != "bool" && kind != "null"
}

func normalizeConstant(item metadataConstant, spec packageSpec) (generatedConstant, error) {
	result := generatedConstant{
		Name:          item.Name,
		Namespace:     item.Namespace,
		DeclaringType: item.DeclaringType,
		Documentation: item.Documentation,
		Comment:       item.Comment,
	}

	if item.Kind == "string" {
		result.Type = "string"
		result.Expression = strconv.Quote(item.Value)

		return result, nil
	}

	if item.Kind == "bool" {
		result.Type = "bool"
		result.Expression = item.Value

		return result, nil
	}

	if !isNumericKind(item.Kind) {
		return generatedConstant{}, fmt.Errorf("unsupported constant kind %s", item.Kind)
	}

	value := new(big.Int)
	if _, ok := value.SetString(item.Value, 10); !ok {
		return generatedConstant{}, fmt.Errorf("parse numeric value %q", item.Value)
	}

	normalized, err := normalizeInteger(value, item.Kind, spec)
	if err != nil {
		return generatedConstant{}, err
	}

	result.Type = spec.TypeName
	result.Expression = normalized
	result.LookupKey = normalized
	result.Numeric = true

	return result, nil
}

func normalizeInteger(value *big.Int, sourceKind string, spec packageSpec) (string, error) {
	normalized := new(big.Int).Set(value)

	if spec.Signed {
		if spec.Name == "hresult" {
			normalized = reinterpretInteger(normalized, sourceWidth(sourceKind), 32, true)
		}

		if !normalized.IsInt64() {
			return "", fmt.Errorf("value %s does not fit %s", value, spec.Underlying)
		}

		return normalized.String(), nil
	}

	normalized = reinterpretInteger(normalized, sourceWidth(sourceKind), spec.Width, false)
	if normalized.Sign() < 0 || normalized.BitLen() > int(spec.Width) {
		return "", fmt.Errorf("value %s does not fit %s", value, spec.Underlying)
	}

	width := int((spec.Width + 3) / 4)

	return fmt.Sprintf("0x%0*X", width, normalized), nil
}

func sourceWidth(kind string) uint {
	switch kind {
	case "int8", "uint8":
		return 8
	case "int16", "uint16", "char":
		return 16
	case "int32", "uint32", "float32":
		return 32
	case "int64", "uint64", "float64":
		return 64
	default:
		return 64
	}
}

func reinterpretInteger(value *big.Int, fromWidth uint, toWidth uint, signed bool) *big.Int {
	result := new(big.Int).Set(value)
	if result.Sign() < 0 {
		modulus := new(big.Int).Lsh(big.NewInt(1), fromWidth)

		result.Add(result, modulus)
	}

	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), toWidth), big.NewInt(1))
	result.And(result, mask)

	if signed && result.Bit(int(toWidth-1)) == 1 {
		modulus := new(big.Int).Lsh(big.NewInt(1), toWidth)

		result.Sub(result, modulus)
	}

	return result
}

func collectConstants(export metadataExport) (map[string][]generatedConstant, []collision, map[string]int) {
	packages := make(map[string][]generatedConstant, len(packageSpecs))
	skipped := make(map[string]int)
	allCollisions := make([]collision, 0)

	for _, spec := range packageSpecs {
		byName := make(map[string][]generatedConstant)

		for _, item := range export.Constants {
			if !spec.Match(item) {
				continue
			}

			if !token.IsIdentifier(item.Name) {
				skipped["invalid_identifier"]++

				continue
			}

			constant, err := normalizeConstant(item, spec)
			if err != nil {
				skipped["unsupported_value"]++

				continue
			}

			byName[constant.Name] = append(byName[constant.Name], constant)
		}

		for name, definitions := range byName {
			first := definitions[0]
			conflicting := false
			for _, candidate := range definitions[1:] {
				if candidate.Type != first.Type || candidate.Expression != first.Expression {
					conflicting = true

					break
				}
			}

			if conflicting {
				details := make([]string, 0, len(definitions))
				for _, definition := range definitions {
					details = append(details, definition.Namespace+"."+definition.DeclaringType+"="+definition.Expression)
				}

				sort.Strings(details)
				allCollisions = append(allCollisions, collision{Package: spec.Name, Name: name, Definitions: details})

				continue
			}

			packages[spec.Name] = append(packages[spec.Name], first)
		}

		sort.Slice(packages[spec.Name], func(left int, right int) bool {
			return packages[spec.Name][left].Name < packages[spec.Name][right].Name
		})
	}

	sort.Slice(allCollisions, func(left int, right int) bool {
		if allCollisions[left].Package != allCollisions[right].Package {
			return allCollisions[left].Package < allCollisions[right].Package
		}

		return allCollisions[left].Name < allCollisions[right].Name
	})

	return packages, allCollisions, skipped
}

func measureCoverage(export metadataExport) constantCoverage {
	coverage := constantCoverage{
		TotalRows:               len(export.Constants),
		UnclassifiedByNamespace: make(map[string]int),
	}

	for _, item := range export.Constants {
		matched := false
		for _, spec := range packageSpecs {
			if spec.Match(item) {
				matched = true

				break
			}
		}

		if matched {
			coverage.MatchedRows++

			continue
		}

		coverage.UnclassifiedRows++
		coverage.UnclassifiedByNamespace[item.Namespace]++
	}

	return coverage
}

func collectConstantMethods(items []metadataConstantMethod) ([]generatedConstantMethod, int) {
	result := make([]generatedConstantMethod, 0, len(items))
	skipped := 0

	for _, item := range items {
		if item.Namespace != "Windows.Win32.System.Threading" || !strings.HasSuffix(item.ReturnType, ".HANDLE") {
			skipped++

			continue
		}

		value := new(big.Int)
		if _, ok := value.SetString(item.Value, 10); !ok || value.Sign() >= 0 || !value.IsInt64() {
			skipped++

			continue
		}

		complement := new(big.Int).Sub(new(big.Int).Neg(value), big.NewInt(1))
		result = append(result, generatedConstantMethod{
			Name:          item.Name,
			Expression:    "^uintptr(" + complement.String() + ")",
			Documentation: item.Documentation,
			Comment:       item.Comment,
		})
	}

	sort.Slice(result, func(left int, right int) bool {
		return result[left].Name < result[right].Name
	})

	return result, skipped
}
