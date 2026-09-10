# win32defs

Windows definitions for Go: status codes, errors, flags, identifiers, structures,
and common Win32 helpers. Definitions are generated from Microsoft's SDK metadata.

You'll need Go 1.26 or later. The library is intended for Windows, though the
constant packages compile on other Go platforms too. Native message formatting
is available only on Windows.

## Getting started

```console
go get github.com/depthbomb/win32defs
```

Import the packages you need:

```go
package main

import (
	"fmt"

	"github.com/depthbomb/win32defs/hresult"
	"github.com/depthbomb/win32defs/message"
	"github.com/depthbomb/win32defs/winerror"
)

func main() {
	hr := hresult.FromWin32(uint32(winerror.ERROR_ACCESS_DENIED))
	fmt.Println(hr, hresult.Failed(hr))

	name, _ := winerror.Name(winerror.ERROR_ACCESS_DENIED)
	code, _ := winerror.Parse(name)
	fmt.Println(name, code)

	text, err := message.FormatWinError(code)
	if err == nil {
		fmt.Println(text)
	}
}
```

## Packages

Packages are grouped by the part of Windows they cover:

| Area                             | Packages                                                                            |
|----------------------------------|-------------------------------------------------------------------------------------|
| Status codes and errors          | `hresult`, `winerror`, `ntstatus`, `exitcode`, `facility`, `wait`, `exception`      |
| Common values and access rights  | `foundation`, `access`                                                              |
| Files, memory, and resources     | `filesystem`, `memory`, `resource`, `versioninfo`, `pe`                             |
| Processes and system information | `process`, `jobobject`, `sysinfo`, `toolhelp`, `libraryloader`, `power`, `syncinit` |
| Services, security, and settings | `service`, `security`, `registry`, `cryptography`                                   |
| Windows and controls             | `winmsg`, `controls`, `dialogs`, `richedit`, `input`, `clipboard`                   |
| Graphics and DPI                 | `gdi`, `dwm`, `hidpi`                                                               |
| Shell, COM, console, and text    | `shell`, `com`, `console`, `globalization`                                          |
| Networking and device I/O        | `winsock`, `winhttp`, `pipes`, `ioctl`                                              |
| Structured values                | `guid`, `propertykey`, `devpropkey`                                                 |
| Helpers and metadata             | `message`, `winmacro`, `catalog`, `metadata`                                        |

For desktop apps, `winmsg` covers windows, menus, messages, and basic controls.
`controls` covers common controls such as list views, tree views, tabs, and task
dialogs. `shell` includes notification icons, file dialogs, shell operations,
change notifications, and taskbar flags. `dialogs` covers the traditional file,
color, font, print, and find dialogs.

Other useful helpers include token pseudo-handles in `process`, SID authorities
in `security`, and static synchronization initializers in `syncinit`. The
`winmacro` package handles word and message packing, coordinates, mouse-wheel
values, colors, I/O control codes, integer resources, and language identifiers.
It offers both friendly Go names and their Win32 macro-name counterparts.

## Looking up values

Generated numeric packages provide `Name`, `Names`, and `Parse`:

```go
name, ok := winerror.Name(5)
aliases := winerror.Names(5)
code, ok := winerror.Parse("ERROR_ACCESS_DENIED")
```

Windows often gives the same value several names. `Name` picks one, preferring
standard success names such as `S_OK`, `ERROR_SUCCESS`, and `STATUS_SUCCESS`.
`Names` returns all known aliases. Parsing uses exact names.

These helpers cover integer constants using the package's common numeric type.
The `catalog` package also includes strings, floating-point constants, and wider
masks, along with source namespaces, comments, and documentation links.

Use the catalog to narrow a lookup to one enum or naming family:

```go
family := catalog.Family{
	Package:   "shell",
	Namespace: "Windows.Win32.UI.Shell",
	Name:      "FILEOPENDIALOGOPTIONS",
}
definition, ok := catalog.LookupInFamily(family, "FOS_PICKFOLDERS")
names := catalog.NamesInFamily(family, "0x20")
text, ok := catalog.FormatFlags(family, 0x60)
// text: FOS_FORCEFILESYSTEM | FOS_PICKFOLDERS
```

`Families` lists the available families, and `FamilyDefinitions` lists their
members. Integer lookups accept decimal or Go integer-literal notation.
`FormatFlags` works with metadata-defined flag enums and keeps unknown bits in
hexadecimal. Returned alias slices are safe to modify.

## Structures

The generator checks every structure in the supported API domains using the
same rules. There's no per-structure list or handwritten layout fallback.
When sizes, alignments, and field offsets match the Windows ABI on 386, amd64,
and arm64, it emits ordinary Go structs with compile-time layout checks.

Examples include `jobobject.JOBOBJECT_BASIC_PROCESS_ID_LIST`, `winmsg.ICONINFO`,
and `winmsg.ICONINFOEXA`/`ICONINFOEXW`, with supporting types such as
`foundation.BOOL` and `gdi.HBITMAP`.

Field names follow the metadata with the first letter capitalized, such as
`FIcon` and `HbmMask`. Flexible arrays keep their declared initial length, such
as `[1]uintptr`. You'll need a larger native allocation for additional elements.

When a sequential native layout doesn't fit Go, the generator emits a typed
buffer view instead. This covers alignment differences, packing, and nested
buffers, including `jobobject.JOBOBJECT_BASIC_LIMIT_INFORMATION`,
`jobobject.JOBOBJECT_EXTENDED_LIMIT_INFORMATION`, and `process.IO_COUNTERS`.
Their sizes, alignments, offsets, and accessors come from the same metadata
and Windows ABI rules as the ordinary structs.

```go
limits := jobobject.NewJOBOBJECT_EXTENDED_LIMIT_INFORMATION()
basic := limits.GetBasicLimitInformation() // shares the parent's storage
basic.SetLimitFlags(jobobject.JOB_OBJECT_LIMIT(jobobject.JOB_OBJECT_LIMIT_PROCESS_MEMORY))
limits.SetProcessMemoryLimit(64 * 1024 * 1024)

// Pass limits.Pointer() and jobobject.JOBOBJECT_EXTENDED_LIMIT_INFORMATIONSize
// to your Windows binding. Keep limits alive while Windows uses the buffer.
```

Use `NewTYPE()` for zeroed, aligned storage or `ViewTYPE(data)` to share existing
bytes. Views check the byte length; `Pointer()` also checks native alignment.
`TYPESize`, `TYPEAlignment`, and `TYPEFieldOffset` describe the native layout
for the current pointer width. Use these instead of `unsafe.Sizeof` on a view,
and pass `Pointer()` instead of the address of the Go wrapper.

`GetField` returns a value copy for scalars and ordinary structs, or a shared
view for nested buffers. `SetField` copies into the buffer. Array accessors
take an index for each dimension and check bounds. Unaligned views support
field access, even when their address can't be passed to Windows on its own.
Copies of a buffer view share storage; the zero value isn't usable. `Bytes()`
exposes the native bytes, including padding. Flexible-array accessors cover
only the metadata-declared initial extent. If your binding takes `uintptr`,
convert `Pointer()` in the call expression and use `runtime.KeepAlive` afterward.
These byte accessors target Windows' little-endian architectures.

Explicit layouts and unions, architecture-specific declarations, bitfields,
pointer fields, and unresolved or ambiguous dependencies are still left out.
Those need more projection rules or metadata; a byte buffer alone isn't enough.
Details are recorded in [the generation report](internal/source/report.json).

## A few details worth knowing

- `foundation` includes both `DUPLICATE_CLOSE_SOURCE` and `DUPLICATE_SAME_ACCESS`.
- Use `resource.RT_MANIFEST.Uintptr()` to pass an integer resource ID to a syscall
  binding. The value is an integer identifier, not a string address.
- DPI contexts have a `.Uintptr()` helper too, such as
  `hidpi.DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2.Uintptr()`.
- `facility.FACILITY_NT_BIT` is a 32-bit HRESULT mask. It's available in the
  catalog but isn't part of the 16-bit facility-ID lookup functions.
- Values keep their native bit patterns and widths, including
  `versioninfo.VS_FFI_SIGNATURE` and `pe.IMAGE_ORDINAL_FLAG64`.
- Namespaces keep unrelated families separate, such as clipboard and font-dialog
  `CF_*` constants.

## Regenerating definitions

Using the library only requires Go. Running the generator also requires the
.NET 10 SDK and, for online generation, access to NuGet.org.

The sources are Microsoft's
[`Microsoft.Windows.SDK.Win32Metadata`](https://www.nuget.org/packages/Microsoft.Windows.SDK.Win32Metadata/)
and [`Microsoft.Windows.SDK.Win32Docs`](https://www.nuget.org/packages/Microsoft.Windows.SDK.Win32Docs/)
packages. Microsoft's [projection guide](https://github.com/microsoft/win32metadata/blob/main/docs/projections.md)
explains how these inputs work.

The generator discovers packages through NuGet's API, verifies their signatures
and SHA-256 hashes, then reads the metadata and documentation. It groups the
results into Go packages and adds comments where Microsoft's documentation has
an exact match. Unsupported or ambiguous definitions are recorded in the report.
The library covers selected API domains, rather than every Win32 namespace.

To regenerate from the versions and hashes in
[`source.lock.json`](internal/source/source.lock.json):

```console
go generate ./...
```

To check for newer upstream releases and update the lock:

```console
go run ./cmd/win32defs-gen -latest
```

Verified packages are cached under `win32defs` in your OS user cache directory.
Each reuse checks the package hash. You can choose another location with
`-cache-dir`.

Once an online run has filled the cache and restored the .NET dependencies, you
can regenerate offline:

```console
go run ./cmd/win32defs-gen -offline
```

Offline mode needs a valid lock and cache, and can't be combined with `-latest`.
The `metadata` package exposes the source versions, URLs, and hashes used to
generate the library.

Output is prepared and checked in a temporary directory before replacing files.
Unchanged files are left alone, and failed replacements are rolled back. Coverage
checks catch removed symbols and newly rejected definitions. Use
`-accept-changes` only after reviewing an intentional removal or rejection.

One header-derived constant, `shell.NIN_KEYSELECT`, comes from
`NIN_SELECT | NINF_KEY` using metadata operands. Its catalog entry links to the
SDK header. Generation fails if an operand is missing or a future metadata
value disagrees with the expression.

Generated files are named `zz_generated*.go`. Update the generator instead of
editing those files directly.

## Checking changes

For generator changes, regenerate first and review the diff. Then run:

```console
go test ./...
go vet ./...
dotnet build tools/winmd-exporter/winmd-exporter.csproj --configuration Release
```

To check reproducibility from a clean checkout:

```console
go generate ./...
git diff --exit-code
go run ./cmd/win32defs-gen -offline
git diff --exit-code
```

CI runs on Windows: amd64 and 386 tests, ARM64 test-binary compilation, and
reproducible generation. `./tools/verify-abi.ps1` also checks layouts and
duplicate-handle flags against the installed Windows SDK with the x86 and x64
MSVC compilers. It covers structures documented in `winnt.h`, `winuser.h`,
`wingdi.h`, and `winioctl.h`; optional SDK components aren't needed.

The scheduled workflow checks for upstream releases every Monday, validates the
output, and commits it only when something changed. You can also run it manually.

To run the catalog and generator benchmarks:

```console
go test -p 1 ./catalog ./cmd/win32defs-gen -run '^$' -bench . -benchmem -count=5
```
