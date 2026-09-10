package main

import (
	"fmt"
	"go/token"
	"math"
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
		Name: "foundation", Doc: "common Win32 boolean, path, handle duplication, and invalid-handle sentinel values.", TypeName: "Value",
		Underlying: "int64", Width: 64, Signed: true, Declare: true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.Foundation" &&
				(hasExactName(item.Name, "FALSE", "TRUE", "INVALID_HANDLE_VALUE", "MAX_PATH") ||
					(item.Enum && item.DeclaringType == "DUPLICATE_HANDLE_OPTIONS"))
		},
		CanonicalRank: func(name string) int {
			if strings.HasPrefix(name, "DUPLICATE_") {
				return 10
			}

			return 0
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
		Name:       "memory",
		Doc:        "virtual memory, section, mapping, protection, and allocation flags.",
		TypeName:   "Value",
		Underlying: "uint64",
		Width:      64,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return isExistingMemoryConstant(item.Name) ||
				(hasExactName(item.Namespace, "Windows.Win32.System.Memory", "Windows.Win32.System.SystemServices", "Windows.Win32.System.WindowsProgramming") &&
					hasAnyPrefix(item.Name, "GMEM_", "LMEM_", "HEAP_"))
		},
		CanonicalRank: memoryCanonicalRank,
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
		Name:       "com",
		Doc:        "COM initialization, activation, storage, marshaling, variant types, and data-transfer flags.",
		TypeName:   "Value",
		Underlying: "uint32",
		Width:      32,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return isExistingCOMConstant(item.Name) ||
				(item.Namespace == "Windows.Win32.System.Ole" && hasAnyPrefix(item.Name, "DROPEFFECT_")) ||
				(item.Namespace == "Windows.Win32.System.Variant" && hasAnyPrefix(item.Name, "VT_"))
		},
		CanonicalRank: comCanonicalRank,
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
		Name:       "winmsg",
		Doc:        "window and control messages, styles, show states, message-box flags, and virtual keys.",
		TypeName:   "Value",
		Underlying: "int64",
		Width:      64,
		Signed:     true,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			if item.Namespace == "Windows.Win32.System.Ole" && item.Name == "OCM__BASE" {
				return true
			}

			if item.Namespace == "Windows.Win32.System.SystemServices" && strings.HasPrefix(item.Name, "MK_") {
				return true
			}

			if hasExactName(item.Namespace, "Windows.Win32.UI.WindowsAndMessaging", "Windows.Win32.UI.Controls") && isBasicControlConstant(item.Name) {
				return true
			}

			if item.Namespace == "Windows.Win32.System.SystemServices" && strings.HasPrefix(item.Name, "SS_") {
				return true
			}

			if item.Namespace == "Windows.Win32.UI.WindowsAndMessaging" && (isWindowOperationConstant(item.Name) || isWindowEventConstant(item.Name) || item.DeclaringType == "ACCEL_VIRT_FLAGS" || item.Name == "DC_HASDEFID") {
				return true
			}

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
		Name:       "shell",
		Doc:        "Shell notification icons, dialogs, execution, file operations, attributes, and known-folder flags.",
		TypeName:   "Value",
		Underlying: "uint32",
		Width:      32,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			if item.Namespace == "Windows.Win32.System.SystemServices" && strings.HasPrefix(item.Name, "SFGAO_") {
				return true
			}

			return item.Namespace == "Windows.Win32.UI.Shell" &&
				(isExistingShellConstant(item.Name) ||
					hasAnyPrefix(item.Name, "BIF_", "BFFM_", "SHGFI_", "SEE_MASK_", "FO_", "FOF_", "FOFX_", "SIGDN_", "KF_FLAG_", "SFGAO_", "SHCNE_", "SHCNF_", "SHCONTF_", "TBPF_", "THBN_", "THBF_", "THB_", "NINF_"))
		},
		CanonicalRank: shellCanonicalRank,
	},
	{
		Name:       "resource",
		Doc:        "integer resource types and manifest resource identifiers.",
		TypeName:   "ID",
		Underlying: "uint16",
		Width:      16,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			if item.Namespace == "Windows.Win32.Media.KernelStreaming" {
				return hasExactName(item.Name, "RT_RCDATA", "RT_STRING")
			}

			return item.Namespace == "Windows.Win32.UI.WindowsAndMessaging" &&
				(strings.HasPrefix(item.Name, "RT_") || strings.HasSuffix(item.Name, "_MANIFEST_RESOURCE_ID"))
		},
		CanonicalRank: func(name string) int {
			if strings.HasPrefix(name, "RT_") {
				return 0
			}

			return 10
		},
	},
	{
		Name:       "versioninfo",
		Doc:        "file version signatures, flags, operating systems, types, and query flags.",
		TypeName:   "Value",
		Underlying: "uint32",
		Width:      32,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.Storage.FileSystem" &&
				hasAnyPrefix(item.Name, "VS_", "VOS_", "VFT_", "VFT2_", "FILE_VER_")
		},
	},
	{
		Name:       "pe",
		Doc:        "Portable Executable and COFF image signatures, directories, machines, sections, and relocations.",
		TypeName:   "Value",
		Underlying: "uint64",
		Width:      64,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return hasExactName(item.Namespace, "Windows.Win32.System.Diagnostics.Debug", "Windows.Win32.System.SystemInformation", "Windows.Win32.System.SystemServices") &&
				strings.HasPrefix(item.Name, "IMAGE_")
		},
	},
	{
		Name:       "controls",
		Doc:        "common control messages, styles, notifications, class names, and message bases.",
		TypeName:   "Value",
		Underlying: "int64",
		Width:      64,
		Signed:     true,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.UI.Controls" && !isBasicControlConstant(item.Name)
		},
	},
	{
		Name:       "dialogs",
		Doc:        "common file, color, font, print, and find dialog flags and messages.",
		TypeName:   "Value",
		Underlying: "int64",
		Width:      64,
		Signed:     true,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.UI.Controls.Dialogs"
		},
	},
	{
		Name:       "gdi",
		Doc:        "GDI drawing, text, font, bitmap, monitor, and redraw constants.",
		TypeName:   "Value",
		Underlying: "int64",
		Width:      64,
		Signed:     true,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.Graphics.Gdi"
		},
	},
	{
		Name:       "dwm",
		Doc:        "Desktop Window Manager attributes, policies, and flags.",
		TypeName:   "Value",
		Underlying: "int64",
		Width:      64,
		Signed:     true,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.Graphics.Dwm"
		},
	},
	{
		Name:       "hidpi",
		Doc:        "DPI awareness, hosting, and scaling behavior constants.",
		TypeName:   "Value",
		Underlying: "int64",
		Width:      64,
		Signed:     true,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.UI.HiDpi"
		},
	},
	{
		Name:       "clipboard",
		Doc:        "standard clipboard formats.",
		TypeName:   "Value",
		Underlying: "uint32",
		Width:      32,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.System.Ole" && strings.HasPrefix(item.Name, "CF_")
		},
	},
	{
		Name:       "pipes",
		Doc:        "named-pipe access, modes, and operation flags.",
		TypeName:   "Value",
		Underlying: "uint32",
		Width:      32,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.System.Pipes" ||
				(item.Namespace == "Windows.Win32.Storage.FileSystem" && strings.HasPrefix(item.Name, "PIPE_"))
		},
	},
	{
		Name:       "richedit",
		Doc:        "Rich Edit messages, notifications, formatting, and stream flags.",
		TypeName:   "Value",
		Underlying: "int64",
		Width:      64,
		Signed:     true,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.UI.Controls.RichEdit"
		},
	},
	{
		Name:       "input",
		Doc:        "raw input, mouse, pointer, and touch flags and identifiers.",
		TypeName:   "Value",
		Underlying: "int64",
		Width:      64,
		Signed:     true,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return hasExactName(item.Namespace, "Windows.Win32.UI.Input", "Windows.Win32.UI.Input.Pointer", "Windows.Win32.UI.Input.Touch") ||
				(item.Namespace == "Windows.Win32.UI.WindowsAndMessaging" && hasAnyPrefix(item.Name, "RI_", "RIM_", "RIDI_"))
		},
	},
	{
		Name:       "security",
		Doc:        "security access rights, token values, privileges, and SID authorities.",
		TypeName:   "Value",
		Underlying: "uint64",
		Width:      64,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.Security"
		},
	},
	{
		Name:       "globalization",
		Doc:        "code pages, locales, text conversion, and character classification constants.",
		TypeName:   "Value",
		Underlying: "int64",
		Width:      64,
		Signed:     true,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.Globalization"
		},
	},
	{
		Name:       "winhttp",
		Doc:        "WinHTTP options, request flags, status codes, and authentication values.",
		TypeName:   "Value",
		Underlying: "int64",
		Width:      64,
		Signed:     true,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.Networking.WinHttp"
		},
	},
	{
		Name:       "cryptography",
		Doc:        "cryptographic algorithm names, options, certificate values, and data-protection flags.",
		TypeName:   "Value",
		Underlying: "int64",
		Width:      64,
		Signed:     true,
		Declare:    true,
		Match: func(item metadataConstant) bool {
			return item.Namespace == "Windows.Win32.Security.Cryptography"
		},
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
	if hasAnyPrefix(name, "CBN_") || hasExactName(name, "FVIRTKEY", "FNOINVERT", "FSHIFT", "FCONTROL", "FALT", "DC_HASDEFID") {
		return 50
	}

	if isWindowEventConstant(name) {
		return 40
	}

	if hasAnyPrefix(name, "WM_", "WS_", "SW_", "MB_", "VK_", "MOD_", "HOTKEYF_", "GWL_", "GWLP_") {
		return 0
	}

	if isControlConstant(name) {
		return 20
	}

	if isBasicControlConstant(name) || isWindowOperationConstant(name) {
		return 30
	}

	return 10
}

func isExistingMemoryConstant(name string) bool {
	return hasAnyPrefix(name, "PAGE_", "MEM_", "SECTION_", "SEC_", "FILE_MAP_")
}

func memoryCanonicalRank(name string) int {
	if isExistingMemoryConstant(name) {
		return 0
	}

	return 10
}

func isExistingCOMConstant(name string) bool {
	return hasAnyPrefix(name, "CLSCTX_", "COINIT_", "STGM_", "DVASPECT_", "TYMED_", "MSHCTX_", "MSHLFLAGS_")
}

func comCanonicalRank(name string) int {
	if isExistingCOMConstant(name) {
		return 0
	}

	return 10
}

func isExistingShellConstant(name string) bool {
	return hasAnyPrefix(name, "NIM_", "NIF_", "NIS_", "NIIF_", "NIN_", "NOTIFYICON_", "FOS_")
}

func shellCanonicalRank(name string) int {
	if strings.HasPrefix(name, "NINF_") {
		return 30
	}

	if hasAnyPrefix(name, "SHCNE_", "SHCNF_", "SHCONTF_", "TBPF_", "THBN_", "THBF_", "THB_") {
		return 20
	}

	if isExistingShellConstant(name) {
		return 0
	}

	return 10
}

func isBasicControlConstant(name string) bool {
	return isControlConstant(name) || hasAnyPrefix(name, "ES_", "EM_", "EN_", "SS_", "STM_", "STN_", "LB_", "LBS_", "LBN_")
}

func isWindowOperationConstant(name string) bool {
	return hasAnyPrefix(name, "HT", "SC_", "PM_", "MIIM_", "MFT_", "MFS_", "MIM_", "GMDI_", "GW_", "GA_", "DCX_", "DI_", "LR_", "IMAGE_", "ICON_", "SIF_", "SBS_", "DLGC_", "DS_", "DM_") ||
		hasExactName(name, "IDOK", "IDCANCEL", "IDABORT", "IDRETRY", "IDIGNORE", "IDYES", "IDNO", "IDCLOSE", "IDHELP", "IDTRYAGAIN", "IDCONTINUE", "IDTIMEOUT")
}

func isWindowEventConstant(name string) bool {
	return hasAnyPrefix(name, "MK_", "XBUTTON", "SIZE_", "WA_", "WMSZ_", "WVR_") || name == "OCM__BASE"
}

func isControlConstant(name string) bool {
	return hasAnyPrefix(name, "CB_", "CBS_", "CBN_", "BS_", "BM_", "BN_", "BST_")
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
	return hasExactName(kind, "int8", "uint8", "int16", "uint16", "char", "int32", "uint32", "int64", "uint64")
}

func normalizeConstant(item metadataConstant, spec packageSpec) (generatedConstant, error) {
	result := generatedConstant{
		Name:          item.Name,
		Namespace:     item.Namespace,
		DeclaringType: item.DeclaringType,
		Documentation: item.Documentation,
		Comment:       item.Comment,
		Family:        constantFamily(item),
		Flags:         item.Enum && item.Flags,
		Kind:          item.Kind,
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

	if item.Kind == "float32" || item.Kind == "float64" {
		bits := int(sourceWidth(item.Kind))
		value, err := strconv.ParseFloat(item.Value, bits)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			return generatedConstant{}, fmt.Errorf("invalid finite %s constant %q", item.Kind, item.Value)
		}

		result.Type = item.Kind
		result.Expression = strconv.FormatFloat(value, 'g', -1, bits)
		if !strings.ContainsAny(result.Expression, ".eE") {
			result.Expression += ".0"
		}

		return result, nil
	}

	if !isNumericKind(item.Kind) {
		return generatedConstant{}, fmt.Errorf("unsupported constant kind %s", item.Kind)
	}

	isFacilityMask := spec.Name == "facility" && item.Name == "FACILITY_NT_BIT"
	if isFacilityMask {
		spec.TypeName = "uint32"
		spec.Underlying = "uint32"
		spec.Width = 32
	}

	normalized, err := normalizeIntegerText(item.Value, item.Kind, spec)
	if err != nil {
		return generatedConstant{}, err
	}

	result.Type = spec.TypeName
	result.Expression = normalized
	result.LookupKey = normalized
	result.Numeric = !isFacilityMask

	return result, nil
}

func constantFamily(item metadataConstant) string {
	if item.Enum {
		return item.DeclaringType
	}

	if index := strings.IndexByte(item.Name, '_'); index > 0 {
		return item.Name[:index+1]
	}

	return item.DeclaringType
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

func collectConstants(export metadataExport) (map[string][]generatedConstant, []collision, map[string]int, []rejectedDefinition) {
	packages := make(map[string][]generatedConstant, len(packageSpecs))
	skipped := make(map[string]int)
	allCollisions := make([]collision, 0)
	var rejected []rejectedDefinition

	for _, spec := range packageSpecs {
		byName := make(map[string][]generatedConstant)

		for _, item := range export.Constants {
			if !spec.Match(item) {
				continue
			}

			if !token.IsIdentifier(item.Name) {
				skipped["invalid_identifier"]++
				rejected = append(rejected, rejectConstant(spec.Name, item, "invalid Go identifier"))

				continue
			}

			constant, err := normalizeConstant(item, spec)
			if err != nil {
				skipped["unsupported_value"]++
				rejected = append(rejected, rejectConstant(spec.Name, item, err.Error()))

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

	return packages, allCollisions, skipped, rejected
}

func rejectConstant(packageName string, item metadataConstant, reason string) rejectedDefinition {
	return rejectedDefinition{
		Package:   packageName,
		Namespace: item.Namespace,
		Name:      item.Name,
		Value:     item.Value,
		Reason:    reason,
	}
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
