param(
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\ghstats"
)

$ErrorActionPreference = "Stop"

$Repo = "boi-gg/ghstats"
$DownloadBase = "https://github.com/$Repo/releases/latest/download"
$BinaryName = "ghstats.exe"

function Get-Architecture {
    switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
        "Arm64" { return "arm64" }
        "X64" { return "amd64" }
        default { throw "Unsupported architecture $_" }
    }
}

$arch = Get-Architecture
$assetName = "ghstats-windows-$arch.exe"
$url = "$DownloadBase/$assetName"

$tempFile = Join-Path $env:TEMP ("$assetName-" + [System.Guid]::NewGuid().ToString())

Write-Host "Downloading $url"
$ProgressPreference = 'SilentlyContinue'
Invoke-WebRequest -Uri $url -OutFile $tempFile -UseBasicParsing

if (-not (Test-Path $InstallDir)) {
    Write-Host "Creating $InstallDir"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$target = Join-Path $InstallDir $BinaryName
Move-Item -Path $tempFile -Destination $target -Force

Write-Host "ghstats installed to $target"

$normalizedInstall = $InstallDir.TrimEnd('\')
$onPath = ($env:PATH -split ';' | ForEach-Object { $_.TrimEnd('\') }) -contains $normalizedInstall
if (-not $onPath) {
    Write-Warning "$InstallDir is not on your PATH. Add it via System Properties or run:`n  [Environment]::SetEnvironmentVariable('PATH', `"$InstallDir;`$([Environment]::GetEnvironmentVariable('PATH','User'))`", 'User')"
}

Write-Host "Run 'ghstats --help' from a new PowerShell window to verify."
