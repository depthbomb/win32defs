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

Generated files are named `zz_generated.go` and must not be edited manually.
