$ErrorActionPreference = 'Stop'

$repository = Split-Path -Parent $PSScriptRoot
$vswhere = Join-Path ${env:ProgramFiles(x86)} 'Microsoft Visual Studio/Installer/vswhere.exe'
$installation = & $vswhere -latest -products '*' -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath
if (!$installation) {
    throw 'Windows ABI verification requires the MSVC x86 and x64 build tools.'
}

$compilerVersion = Get-ChildItem -LiteralPath (Join-Path $installation 'VC/Tools/MSVC') -Directory | Sort-Object { [version]$_.Name } -Descending | Select-Object -First 1
$sdkRoot = Join-Path ${env:ProgramFiles(x86)} 'Windows Kits/10/Include'
$sdkVersion = Get-ChildItem -LiteralPath $sdkRoot -Directory | Where-Object { Test-Path -LiteralPath (Join-Path $_.FullName 'um/windows.h') } | Sort-Object { [version]$_.Name } -Descending | Select-Object -First 1
if (!$compilerVersion -or !$sdkVersion) {
    throw 'Windows ABI verification requires an installed Windows SDK and MSVC compiler.'
}

$previousInclude = $env:INCLUDE
$previousCompiler = $env:WIN32DEFS_ABI_CC
$previousArchitecture = $env:WIN32DEFS_ABI_ARCH
Push-Location -LiteralPath $repository
try {
    $env:INCLUDE = @(
        (Join-Path $compilerVersion.FullName 'include')
        (Join-Path $sdkVersion.FullName 'ucrt')
        (Join-Path $sdkVersion.FullName 'shared')
        (Join-Path $sdkVersion.FullName 'um')
    ) -join ';'

    foreach ($target in @(
        @{
            Native = 'x64'
            Go     = 'amd64'
        }
        @{
            Native = 'x86'
            Go     = '386'
        }
    )) {
        $env:WIN32DEFS_ABI_CC = Join-Path $compilerVersion.FullName "bin/Hostx64/$($target.Native)/cl.exe"
        $env:WIN32DEFS_ABI_ARCH = $target.Go
        go test ./cmd/win32defs-gen -run '^TestWindowsSDKStructures$' -count=1 -v
        if ($LASTEXITCODE -ne 0) {
            throw "Windows $($target.Go) SDK ABI verification failed."
        }
    }
} finally {
    $env:INCLUDE = $previousInclude
    $env:WIN32DEFS_ABI_CC = $previousCompiler
    $env:WIN32DEFS_ABI_ARCH = $previousArchitecture
    Pop-Location
}
