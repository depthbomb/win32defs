package catalog_test

import (
	"slices"
	"strconv"
	"testing"

	"github.com/depthbomb/win32defs/catalog"
	"github.com/depthbomb/win32defs/clipboard"
	"github.com/depthbomb/win32defs/com"
	"github.com/depthbomb/win32defs/controls"
	"github.com/depthbomb/win32defs/dialogs"
	"github.com/depthbomb/win32defs/dwm"
	"github.com/depthbomb/win32defs/gdi"
	"github.com/depthbomb/win32defs/hidpi"
	"github.com/depthbomb/win32defs/memory"
	"github.com/depthbomb/win32defs/pipes"
	"github.com/depthbomb/win32defs/shell"
	"github.com/depthbomb/win32defs/winmsg"
)

func checkDesktopConstant[T ~int64 | ~uint32 | ~uint64](t *testing.T, packageName, name string, value, want T, parse func(string) (T, bool), names func(T) []string) {
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
