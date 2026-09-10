package main

import "testing"

func TestDuplicateHandleFamily(t *testing.T) {
	t.Parallel()

	spec := specByName("foundation")
	item := metadataConstant{
		Namespace:     "Windows.Win32.Foundation",
		DeclaringType: "DUPLICATE_HANDLE_OPTIONS",
		Name:          "DUPLICATE_FUTURE_OPTION",
		Enum:          true,
		Flags:         true,
		Kind:          "uint32",
		Value:         "16",
	}
	if !spec.Match(item) {
		t.Fatal("the entire metadata enum must be selected without a member allowlist")
	}

	constant, err := normalizeConstant(item, spec)
	if err != nil || constant.Expression != "16" || constant.Family != item.DeclaringType || !constant.Flags {
		t.Fatalf("enum normalization = %+v, %v", constant, err)
	}

	item.Namespace = "Windows.Win32.Other"
	if spec.Match(item) {
		t.Fatal("duplicate-handle enum classification crossed namespace boundaries")
	}

	item.Namespace = "Windows.Win32.Foundation"
	item.Enum = false
	if spec.Match(item) {
		t.Fatal("loose constants were classified as enum members")
	}
}

func TestDesktopNamespaceBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		packageName string
		name        string
		namespace   string
		want        bool
	}{
		{
			packageName: "versioninfo",
			name:        "VS_FFI_SIGNATURE",
			namespace:   "Windows.Win32.Storage.FileSystem",
			want:        true,
		},
		{
			packageName: "versioninfo",
			name:        "VS_ALLOW_LATIN",
			namespace:   "Windows.Win32.Globalization",
		},
		{
			packageName: "pe",
			name:        "IMAGE_DIRECTORY_ENTRY_SECURITY",
			namespace:   "Windows.Win32.System.Diagnostics.Debug",
			want:        true,
		},
		{
			packageName: "pe",
			name:        "IMAGE_ICON",
			namespace:   "Windows.Win32.UI.WindowsAndMessaging",
		},
		{
			packageName: "resource",
			name:        "RT_RCDATA",
			namespace:   "Windows.Win32.Media.KernelStreaming",
			want:        true,
		},
		{
			packageName: "resource",
			name:        "RT_OTHER",
			namespace:   "Windows.Win32.Media.KernelStreaming",
		},
		{
			packageName: "winmsg",
			name:        "CBN_SELCHANGE",
			namespace:   "Windows.Win32.UI.WindowsAndMessaging",
			want:        true,
		},
		{
			packageName: "clipboard",
			name:        "CF_UNICODETEXT",
			namespace:   "Windows.Win32.System.Ole",
			want:        true,
		},
		{
			packageName: "clipboard",
			name:        "CF_SCREENFONTS",
			namespace:   "Windows.Win32.UI.Controls.Dialogs",
		},
		{
			packageName: "clipboard",
			name:        "CF_ACCEPT",
			namespace:   "Windows.Win32.Networking.WinSock",
		},
		{
			packageName: "winmsg",
			name:        "SS_LEFT",
			namespace:   "Windows.Win32.System.SystemServices",
			want:        true,
		},
		{
			packageName: "winmsg",
			name:        "SS_OTHER",
			namespace:   "Windows.Win32.Devices.Display",
		},
		{
			packageName: "winmsg",
			name:        "EM_SETSEL",
			namespace:   "Windows.Win32.UI.Controls",
			want:        true,
		},
		{
			packageName: "controls",
			name:        "LVM_FIRST",
			namespace:   "Windows.Win32.UI.Controls",
			want:        true,
		},
		{
			packageName: "controls",
			name:        "LVM_FIRST",
			namespace:   "Windows.Win32.Other",
		},
		{
			packageName: "controls",
			name:        "EM_SETSEL",
			namespace:   "Windows.Win32.UI.Controls",
		},
		{
			packageName: "shell",
			name:        "SFGAO_FOLDER",
			namespace:   "Windows.Win32.System.SystemServices",
			want:        true,
		},
		{
			packageName: "memory",
			name:        "GMEM_FIXED",
			namespace:   "Windows.Win32.System.WindowsProgramming",
			want:        true,
		},
		{
			packageName: "memory",
			name:        "LMEM_FIXED",
			namespace:   "Windows.Win32.System.SystemServices",
			want:        true,
		},
		{
			packageName: "memory",
			name:        "HEAP_OTHER",
			namespace:   "Windows.Win32.Other",
		},
		{
			packageName: "pipes",
			name:        "PIPE_ACCESS_DUPLEX",
			namespace:   "Windows.Win32.Storage.FileSystem",
			want:        true,
		},
		{
			packageName: "pipes",
			name:        "PIPE_OTHER",
			namespace:   "Windows.Win32.Devices.Usb",
		},
	}

	for _, test := range tests {
		t.Run(test.packageName+"/"+test.namespace+"/"+test.name, func(t *testing.T) {
			t.Parallel()

			item := metadataConstant{
				Name:      test.name,
				Namespace: test.namespace,
			}
			got := specByName(test.packageName).Match(item)
			if got != test.want {
				t.Fatalf("Match(%#v) = %v, want %v", item, got, test.want)
			}
		})
	}
}
