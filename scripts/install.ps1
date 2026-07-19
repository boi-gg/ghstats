param(
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\ghstats"
)

$ErrorActionPreference = "Stop"

[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12 -bor [Net.SecurityProtocolType]::Tls13

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

$ProgressPreference = 'SilentlyContinue'
try {
    Write-Host "Downloading $url"
    Invoke-WebRequest -Uri $url -OutFile $tempFile -UseBasicParsing

    if (-not (Test-Path $InstallDir)) {
        Write-Host "Creating $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $target = Join-Path $InstallDir $BinaryName
    try {
        Move-Item -Path $tempFile -Destination $target -Force
    } catch {
        $staged = Join-Path $InstallDir ("ghstats.exe.new-" + [System.Guid]::NewGuid().ToString())
        Move-Item -Path $tempFile -Destination $staged -Force
        Write-Warning "Could not replace $target because it is in use. A new version was staged at $staged. Stop any running ghstats process and run:`n  Move-Item -Force '$staged' '$target'"
    }
} finally {
    if (Test-Path $tempFile) {
        Remove-Item -Path $tempFile -Force -ErrorAction SilentlyContinue
    }
}

Write-Host "ghstats installed to $target"

$normalizedInstall = $InstallDir.TrimEnd('\')
$onPath = ($env:PATH -split ';' | ForEach-Object { $_.TrimEnd('\') }) -contains $normalizedInstall
if (-not $onPath) {
    Write-Warning "$InstallDir is not on your PATH. Add it via System Properties or run:`n  [Environment]::SetEnvironmentVariable('PATH', `"$InstallDir;`$([Environment]::GetEnvironmentVariable('PATH','User'))`", 'User')"
}

Write-Host "Run 'ghstats --help' from a new PowerShell window to verify."
