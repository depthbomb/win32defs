# win32defs

`win32defs` provides generated Go definitions for Windows status codes, errors,
flags, identifiers, and value-oriented macros. The constant packages compile on
every Go platform; native message formatting is enabled only on Windows.

The minimum supported Go release is Go 1.26.

## Packages

Core status packages:

- `hresult`
- `winerror`
- `ntstatus`
- `exitcode`
- `facility`
- `wait`
- `exception`

Flag and API-domain packages:

- `access`
- `foundation`
- `filesystem`
- `memory`
- `service`
- `registry`
- `com`
- `console`
- `process`
- `jobobject`
- `power`
- `sysinfo`
- `toolhelp`
- `libraryloader`
- `winsock`
- `winmsg`
- `shell`
- `controls`
- `dialogs`
- `gdi`
- `dwm`
- `hidpi`
- `clipboard`
- `pipes`
- `richedit`
- `input`
- `globalization`
- `winhttp`
- `cryptography`
- `ioctl`
- `security`
- `syncinit`

Generated structured values and helpers:

- `guid`
- `propertykey`
- `devpropkey`
- `message`
- `winmacro`
- `metadata`
- `catalog`

## Examples

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

Each generated numeric package supports exact-name parsing and reverse lookup:

```go
name, ok := winerror.Name(5)
aliases := winerror.Names(5)
code, ok := winerror.Parse("ERROR_ACCESS_DENIED")
```

The `Names` function is important because Windows intentionally assigns
multiple symbolic names to some values.

Canonical names prefer the conventional success symbols (`S_OK`,
`ERROR_SUCCESS`, and `STATUS_SUCCESS`) and then package-specific naming
families. `Names` always returns every known alias.

The HRESULT and NTSTATUS packages also expose exact-name counterparts of the
standard classification and field-extraction macros. The `process` package
includes metadata-defined token pseudo-handle functions, `security` provides
SID identifier authorities, and `syncinit` provides pointer-sized static
synchronization initializers.

Desktop and application support is grouped by API domain:

| Package | Coverage |
| --- | --- |
| `winmsg` | Window operations, menus, hit testing, dialog results, and basic button, combo box, edit, static, and list box controls. |
| `controls` | Common controls, including list views, tree views, tabs, toolbars, progress bars, task dialogs, class names, notifications, and message bases such as `LVM_FIRST`. |
| `dialogs` | Traditional file, color, font, print, and find dialogs. |
| `shell` | Notification icons, file and folder dialogs, shell execution, file operations, shell item attributes and display names, and known-folder flags. |
| `gdi` | Drawing, text formatting, fonts, bitmaps, monitors, and redraw flags. |
| `dwm` | Desktop Window Manager attributes and policies, including immersive dark mode. |
| `hidpi` | DPI awareness, hosting, and scaling enums. |
| `clipboard` | Standard clipboard formats, including `CF_UNICODETEXT`. |
| `pipes` | Named-pipe access, modes, and operation flags. |
| `memory` | Memory and mapping flags plus global, local, and heap allocation flags. |
| `com` | COM flags plus drag/drop effects and variant type identifiers. |
| `richedit` | Rich Edit messages, stream formats, character formatting, and notifications. |
| `input` | Raw input, pointer input, and touch constants. |
| `security` | Token access masks, privileges, security enums, and SID authorities. |
| `globalization` | Code pages, locales, character classification, and text conversion. |
| `winhttp` | HTTP status codes, request flags, authentication, and WinHTTP options. |
| `cryptography` | Cryptographic algorithm names, certificate values, and data-protection flags. |

Family matching accounts for metadata namespaces: clipboard `CF_*` values
remain separate from font-dialog `CF_*` flags, and static-control `SS_*`
styles are included from the system-services namespace. Existing package
assignments and canonical names are preserved when expanding these families.

The `winmsg` package also includes mouse-state and activation/resize values
and the reflected-message base `OCM__BASE`. The `shell` package includes
change notifications, enumeration options, and taskbar button/progress flags.

`FACILITY_NT_BIT` is a 32-bit HRESULT mask, not a facility identifier. Its
previously truncated value has been corrected to `uint32(0x10000000)`; it is
available as `facility.FACILITY_NT_BIT` and through the catalog, but is excluded
from the 16-bit facility-ID lookup functions.

DPI context constants can be passed to syscall bindings with their
pointer-sized representation on every Windows architecture:

```go
context := hidpi.DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2.Uintptr()
```

For lookups scoped to one enum or named prefix, use the opt-in catalog:

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

`Families` discovers available families, and `FamilyDefinitions` enumerates
their members. Integer lookup values accept decimal and Go integer-literal
notation; other values use the catalog's literal representation. Each
definition exposes its family, source scalar kind, and whether its enum is
marked with the metadata Flags attribute. `FormatFlags` rejects ordinary
enums and loose prefixes, and preserves unknown bits as a hexadecimal remainder.
Family membership tables are generated alongside the definition table.
Enumeration needs no runtime index allocation; alias and flag caches are
built on demand for individual families. Returned alias slices are independent
copies and can be modified by the caller.

The `winmacro` package provides platform-independent helpers for word and
message-parameter packing, signed coordinate and mouse-wheel extraction, color
values, I/O control-code construction and decomposition, integer resources,
and legacy language and locale identifiers. Friendly Go names and exact Win32
macro-name counterparts are both available.

When Microsoft's structured documentation contains an exact, non-empty match,
the generated value also has a Go comment sourced from that text. Missing or
ambiguous documentation is left blank; the generator does not infer prose.
The opt-in `catalog` package exposes that comment alongside the original
metadata namespace, declaring type, normalized value, and official
documentation URL when available. The `metadata` package exposes both source
package versions, URLs, and SHA-256 digests used for the build.

## Autonomous generation

Constants are not maintained in a seed list. The generator:

1. Discovers the NuGet `PackageBaseAddress` endpoint from the official NuGet V3
   service index.
2. Discovers the newest `Microsoft.Windows.SDK.Win32Metadata` and
   `Microsoft.Windows.SDK.Win32Docs` releases.
3. Downloads both packages, verifies their NuGet signatures, and records their
   SHA-256 digests.
4. Extracts `Windows.Win32.winmd` and the structured `apidocs.msgpack` data.
5. Reads ECMA-335 constants, enum members, struct initializers, and GUID
   attributes with `System.Reflection.Metadata`.
6. Attaches documentation only by exact API/field-name matches and normalizes
   upstream markup into Go comments without paraphrasing it.
7. Classifies and emits deterministic Go packages.
8. Quarantines ambiguous definitions in `internal/source/report.json` instead
   of selecting a value silently.

The top-level packages are deliberately curated around broadly useful Win32
domains; they are not a projection of every metadata enum. The generation
report records total, matched, and unclassified metadata rows by namespace so
that this scope is explicit and coverage regressions are visible. Structured
values and constant-returning inline methods are counted separately.

The source artifact is Microsoft's
[`Microsoft.Windows.SDK.Win32Metadata`](https://www.nuget.org/packages/Microsoft.Windows.SDK.Win32Metadata/)
package, together with its
[`Microsoft.Windows.SDK.Win32Docs`](https://www.nuget.org/packages/Microsoft.Windows.SDK.Win32Docs/)
package. Its [projection guidance](https://github.com/microsoft/win32metadata/blob/main/docs/projections.md)
documents both `.winmd` and the rich documentation artifact as language
projection inputs. Package discovery uses the documented
[NuGet V3 package-content API](https://learn.microsoft.com/nuget/api/package-base-address-resource).

Run an explicit upstream update with:

```console
go run ./cmd/win32defs-gen -latest
```

Ordinary reproducible regeneration uses the version and package digest in
`internal/source/source.lock.json`:

```console
go generate ./...
```

Verified source packages are cached by SHA-256 under the operating system's
user cache directory in `win32defs`. Every reuse checks the archive digest;
only archives that passed NuGet signature verification have a verification
record. Override the location with `-cache-dir`.

After an online generation has populated the package cache and restored the
exporter's .NET dependencies, regenerate without network access using:

```console
go run ./cmd/win32defs-gen -offline
```

Offline mode requires the source lock and fails on missing or invalid cache
entries; it cannot be combined with `-latest`. It builds the exporter with
`--no-restore` so the already restored .NET dependencies are reused.

Generation renders and validates all output in a temporary directory before
replacing live files. Publication skips unchanged files and rolls back file
replacements on an I/O failure. New omissions, rejected definitions, or
collisions fail generation with symbol-level diagnostics. Existing documented
collisions remain quarantined. The report records the emitted symbol inventory
for subsequent regression checks. Use `-accept-changes` only after reviewing
an intentional removal or a new quarantined definition.

Integer normalization rejects out-of-range values before any narrowing and
preserves intentional signed-bit reinterpretation. Finite floating-point
constants retain their source `float32` or `float64` type; non-finite values
are reported as unsupported. Numeric package lookup functions cover integer
constants of the package's common type; the catalog also includes strings,
floating-point values, and explicitly wider masks.

Generation requires Go 1.26+, the .NET 10 SDK, and access to NuGet.org. Library
consumers need only Go.

The scheduled GitHub Actions workflow checks for upstream releases every
Monday, regenerates the repository, runs tests and vet, and commits only when
tracked output changes. The workflow also supports manual dispatch for the
initial or an on-demand run.

## Validation

```console
go generate ./...
git diff --exit-code
go test ./...
go vet ./...
dotnet build tools/winmd-exporter/winmd-exporter.csproj --configuration Release
```

CI runs tests on Linux and Windows and compiles Windows 386 and ARM64 test
binaries. It also verifies that offline regeneration produces no changes.

The catalog and generator benchmarks include family lookup, flag formatting,
first-use family cache construction, and scalar normalization:

```console
go test -p 1 ./catalog ./cmd/win32defs-gen -run '^$' -bench . -benchmem -count=5
```

Generated files are named `zz_generated.go` and must not be edited manually.
