package catalog_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/depthbomb/win32defs/catalog"
	"github.com/depthbomb/win32defs/clipboard"
	"github.com/depthbomb/win32defs/com"
	"github.com/depthbomb/win32defs/controls"
	"github.com/depthbomb/win32defs/cryptography"
	"github.com/depthbomb/win32defs/dialogs"
	"github.com/depthbomb/win32defs/dwm"
	"github.com/depthbomb/win32defs/gdi"
	"github.com/depthbomb/win32defs/globalization"
	"github.com/depthbomb/win32defs/hidpi"
	"github.com/depthbomb/win32defs/input"
	"github.com/depthbomb/win32defs/memory"
	"github.com/depthbomb/win32defs/pe"
	"github.com/depthbomb/win32defs/pipes"
	"github.com/depthbomb/win32defs/resource"
	"github.com/depthbomb/win32defs/richedit"
	"github.com/depthbomb/win32defs/security"
	"github.com/depthbomb/win32defs/shell"
	"github.com/depthbomb/win32defs/versioninfo"
	"github.com/depthbomb/win32defs/winhttp"
	"github.com/depthbomb/win32defs/winmsg"
)

func checkDesktopConstant[T ~int64 | ~uint16 | ~uint32 | ~uint64](t *testing.T, packageName, name string, value, want T, parse func(string) (T, bool), names func(T) []string) {
	t.Helper()

	t.Run(packageName+"/"+name, func(t *testing.T) {
		t.Parallel()

		if value != want {
			t.Fatalf("constant = %d, want %d", value, want)
		}

		parsed, ok := parse(name)
		if !ok || parsed != want {
			t.Fatalf("Parse(%q) = %d, %v", name, parsed, ok)
		}

		if !slices.Contains(names(want), name) {
			t.Fatalf("Names(%d) is missing %q", want, name)
		}

		definition, ok := catalog.Lookup(packageName, name)
		if !ok || definition.Namespace == "" || definition.DeclaringType == "" {
			t.Fatalf("missing catalog provenance: %#v", definition)
		}

		catalogValue, err := strconv.ParseInt(definition.Value, 0, 64)
		if err != nil || catalogValue != int64(want) {
			t.Fatalf("catalog value = %q, want %d", definition.Value, want)
		}
	})
}

func TestDesktopConstants(t *testing.T) {
	t.Parallel()

	checkDesktopConstant(t, "winmsg", "ES_MULTILINE", winmsg.ES_MULTILINE, 4, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "EM_SETSEL", winmsg.EM_SETSEL, 0xB1, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "SS_LEFT", winmsg.SS_LEFT, 0, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "LB_ERR", winmsg.LB_ERR, -1, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "HTTRANSPARENT", winmsg.HTTRANSPARENT, -1, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "SC_CLOSE", winmsg.SC_CLOSE, 0xF060, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "PM_REMOVE", winmsg.PM_REMOVE, 1, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "MIIM_STRING", winmsg.MIIM_STRING, 0x40, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "controls", "LVM_FIRST", controls.LVM_FIRST, 0x1000, controls.Parse, controls.Names)
	checkDesktopConstant(t, "controls", "TV_FIRST", controls.TV_FIRST, 0x1100, controls.Parse, controls.Names)
	checkDesktopConstant(t, "controls", "TCM_FIRST", controls.TCM_FIRST, 0x1300, controls.Parse, controls.Names)
	checkDesktopConstant(t, "controls", "NM_FIRST", controls.NM_FIRST, 0, controls.Parse, controls.Names)
	checkDesktopConstant(t, "controls", "NM_CLICK", controls.NM_CLICK, 0xFFFFFFFE, controls.Parse, controls.Names)
	checkDesktopConstant(t, "controls", "PBM_SETPOS", controls.PBM_SETPOS, 0x402, controls.Parse, controls.Names)
	checkDesktopConstant(t, "controls", "ICC_WIN95_CLASSES", controls.ICC_WIN95_CLASSES, 0xFF, controls.Parse, controls.Names)
	checkDesktopConstant(t, "dialogs", "OFN_FILEMUSTEXIST", dialogs.OFN_FILEMUSTEXIST, 0x1000, dialogs.Parse, dialogs.Names)
	checkDesktopConstant(t, "dialogs", "CC_RGBINIT", dialogs.CC_RGBINIT, 1, dialogs.Parse, dialogs.Names)
	checkDesktopConstant(t, "dialogs", "CF_SCREENFONTS", dialogs.CF_SCREENFONTS, 1, dialogs.Parse, dialogs.Names)
	checkDesktopConstant(t, "shell", "SIGDN_FILESYSPATH", shell.SIGDN_FILESYSPATH, 0x80058000, shell.Parse, shell.Names)
	checkDesktopConstant(t, "shell", "SFGAO_FOLDER", shell.SFGAO_FOLDER, 0x20000000, shell.Parse, shell.Names)
	checkDesktopConstant(t, "shell", "SHGFI_ICON", shell.SHGFI_ICON, 0x100, shell.Parse, shell.Names)
	checkDesktopConstant(t, "shell", "SEE_MASK_NOCLOSEPROCESS", shell.SEE_MASK_NOCLOSEPROCESS, 0x40, shell.Parse, shell.Names)
	checkDesktopConstant(t, "shell", "FOF_NOCONFIRMATION", shell.FOF_NOCONFIRMATION, 0x10, shell.Parse, shell.Names)
	checkDesktopConstant(t, "shell", "BIF_RETURNONLYFSDIRS", shell.BIF_RETURNONLYFSDIRS, 1, shell.Parse, shell.Names)
	checkDesktopConstant(t, "shell", "KF_FLAG_DEFAULT", shell.KF_FLAG_DEFAULT, 0, shell.Parse, shell.Names)
	checkDesktopConstant(t, "gdi", "SRCCOPY", gdi.SRCCOPY, 0xCC0020, gdi.Parse, gdi.Names)
	checkDesktopConstant(t, "gdi", "GDI_ERROR", gdi.GDI_ERROR, -1, gdi.Parse, gdi.Names)
	checkDesktopConstant(t, "gdi", "DT_CENTER", gdi.DT_CENTER, 1, gdi.Parse, gdi.Names)
	checkDesktopConstant(t, "gdi", "RDW_INVALIDATE", gdi.RDW_INVALIDATE, 1, gdi.Parse, gdi.Names)
	checkDesktopConstant(t, "dwm", "DWMWA_USE_IMMERSIVE_DARK_MODE", dwm.DWMWA_USE_IMMERSIVE_DARK_MODE, 20, dwm.Parse, dwm.Names)
	checkDesktopConstant(t, "hidpi", "DPI_AWARENESS_PER_MONITOR_AWARE", hidpi.DPI_AWARENESS_PER_MONITOR_AWARE, 2, hidpi.Parse, hidpi.Names)
	checkDesktopConstant(t, "hidpi", "DPI_AWARENESS_INVALID", hidpi.DPI_AWARENESS_INVALID, -1, hidpi.Parse, hidpi.Names)
	checkDesktopConstant(t, "clipboard", "CF_UNICODETEXT", clipboard.CF_UNICODETEXT, 13, clipboard.Parse, clipboard.Names)
	checkDesktopConstant(t, "com", "DROPEFFECT_COPY", com.DROPEFFECT_COPY, 1, com.Parse, com.Names)
	checkDesktopConstant(t, "com", "VT_BSTR", com.VT_BSTR, 8, com.Parse, com.Names)
	checkDesktopConstant(t, "memory", "GMEM_MOVEABLE", memory.GMEM_MOVEABLE, 2, memory.Parse, memory.Names)
	checkDesktopConstant(t, "memory", "LMEM_ZEROINIT", memory.LMEM_ZEROINIT, 0x40, memory.Parse, memory.Names)
	checkDesktopConstant(t, "memory", "HEAP_ZERO_MEMORY", memory.HEAP_ZERO_MEMORY, 8, memory.Parse, memory.Names)
	checkDesktopConstant(t, "pipes", "PIPE_ACCESS_DUPLEX", pipes.PIPE_ACCESS_DUPLEX, 3, pipes.Parse, pipes.Names)
	checkDesktopConstant(t, "pipes", "PIPE_TYPE_MESSAGE", pipes.PIPE_TYPE_MESSAGE, 4, pipes.Parse, pipes.Names)
	checkDesktopConstant(t, "winmsg", "OCM__BASE", winmsg.OCM__BASE, 0x2000, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "MK_LBUTTON", winmsg.MK_LBUTTON, 1, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "WVR_REDRAW", winmsg.WVR_REDRAW, 0x300, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "richedit", "EM_STREAMIN", richedit.EM_STREAMIN, 0x449, richedit.Parse, richedit.Names)
	checkDesktopConstant(t, "input", "RID_INPUT", input.RID_INPUT, 0x10000003, input.Parse, input.Names)
	checkDesktopConstant(t, "security", "TOKEN_QUERY", security.TOKEN_QUERY, 8, security.Parse, security.Names)
	checkDesktopConstant(t, "globalization", "CP_UTF8", globalization.CP_UTF8, 65001, globalization.Parse, globalization.Names)
	checkDesktopConstant(t, "winhttp", "HTTP_STATUS_OK", winhttp.HTTP_STATUS_OK, 200, winhttp.Parse, winhttp.Names)
	checkDesktopConstant(t, "cryptography", "CRYPTPROTECT_UI_FORBIDDEN", cryptography.CRYPTPROTECT_UI_FORBIDDEN, 1, cryptography.Parse, cryptography.Names)
	checkDesktopConstant(t, "shell", "THBN_CLICKED", shell.THBN_CLICKED, 0x1800, shell.Parse, shell.Names)

	if security.SE_DEBUG_NAME != "SeDebugPrivilege" || cryptography.BCRYPT_SHA256_ALGORITHM != "SHA256" {
		t.Fatal("incorrect generated security or algorithm name")
	}
}

func TestResourceVersionAndNotificationConstants(t *testing.T) {
	t.Parallel()

	checkDesktopConstant(t, "winmsg", "CBN_SELCHANGE", winmsg.CBN_SELCHANGE, 1, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "FVIRTKEY", winmsg.FVIRTKEY, 1, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "FSHIFT", winmsg.FSHIFT, 4, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "FCONTROL", winmsg.FCONTROL, 8, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "FALT", winmsg.FALT, 16, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "winmsg", "DC_HASDEFID", winmsg.DC_HASDEFID, 0x534B, winmsg.Parse, winmsg.Names)
	checkDesktopConstant(t, "shell", "NIN_KEYSELECT", shell.NIN_KEYSELECT, 0x401, shell.Parse, shell.Names)
	checkDesktopConstant(t, "shell", "NINF_KEY", shell.NINF_KEY, 1, shell.Parse, shell.Names)
	checkDesktopConstant(t, "resource", "RT_MANIFEST", resource.RT_MANIFEST, 24, resource.Parse, resource.Names)
	checkDesktopConstant(t, "resource", "RT_VERSION", resource.RT_VERSION, 16, resource.Parse, resource.Names)
	checkDesktopConstant(t, "resource", "RT_ICON", resource.RT_ICON, 3, resource.Parse, resource.Names)
	checkDesktopConstant(t, "resource", "RT_GROUP_ICON", resource.RT_GROUP_ICON, 14, resource.Parse, resource.Names)
	checkDesktopConstant(t, "resource", "RT_RCDATA", resource.RT_RCDATA, 10, resource.Parse, resource.Names)
	checkDesktopConstant(t, "resource", "RT_STRING", resource.RT_STRING, 6, resource.Parse, resource.Names)
	checkDesktopConstant(t, "versioninfo", "VS_FFI_SIGNATURE", versioninfo.VS_FFI_SIGNATURE, 0xFEEF04BD, versioninfo.Parse, versioninfo.Names)
	checkDesktopConstant(t, "versioninfo", "VS_FFI_STRUCVERSION", versioninfo.VS_FFI_STRUCVERSION, 0x10000, versioninfo.Parse, versioninfo.Names)
	checkDesktopConstant(t, "versioninfo", "VS_FFI_FILEFLAGSMASK", versioninfo.VS_FFI_FILEFLAGSMASK, 0x3F, versioninfo.Parse, versioninfo.Names)
	checkDesktopConstant(t, "versioninfo", "VOS_NT_WINDOWS32", versioninfo.VOS_NT_WINDOWS32, 0x40004, versioninfo.Parse, versioninfo.Names)
	checkDesktopConstant(t, "versioninfo", "VFT_APP", versioninfo.VFT_APP, 1, versioninfo.Parse, versioninfo.Names)
	checkDesktopConstant(t, "pe", "IMAGE_DIRECTORY_ENTRY_SECURITY", pe.IMAGE_DIRECTORY_ENTRY_SECURITY, 4, pe.Parse, pe.Names)

	family := catalog.Family{
		Package:   "winmsg",
		Namespace: "Windows.Win32.UI.WindowsAndMessaging",
		Name:      "ACCEL_VIRT_FLAGS",
	}
	flags, ok := catalog.FormatFlags(family, uint64(winmsg.FVIRTKEY|winmsg.FCONTROL))
	if !ok || !strings.Contains(flags, "FVIRTKEY") || !strings.Contains(flags, "FCONTROL") {
		t.Fatalf("accelerator flags = %q, %v", flags, ok)
	}

	if resource.RT_MANIFEST.Uintptr() != 24 || resource.ID(0xFFFF).Uintptr() != 0xFFFF {
		t.Fatal("incorrect integer resource pointer representation")
	}

	if pe.IMAGE_ORDINAL_FLAG64 != 0x8000000000000000 {
		t.Fatal("64-bit image ordinal flag was truncated")
	}

	name, ok := resource.Name(resource.RT_ICON)
	if !ok || name != "RT_ICON" {
		t.Fatalf("resource type canonical name = %q, %v", name, ok)
	}
}

func TestDesktopFamilyBoundaries(t *testing.T) {
	t.Parallel()

	if _, ok := clipboard.Parse("CF_SCREENFONTS"); ok {
		t.Fatal("font-dialog flags must not be classified as clipboard formats")
	}

	if _, ok := dialogs.Parse("CF_UNICODETEXT"); ok {
		t.Fatal("clipboard formats must not be classified as font-dialog flags")
	}

	if _, ok := controls.Parse("CB_ADDSTRING"); ok {
		t.Fatal("basic control messages must retain their winmsg package assignment")
	}
}
