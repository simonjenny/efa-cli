# Installs the latest efa-cli release binary for Windows.
# Usage: irm https://raw.githubusercontent.com/simonjenny/efa-cli/main/install.ps1 | iex
$ErrorActionPreference = "Stop"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$repo = "simonjenny/efa-cli"
$installDir = Join-Path $env:LOCALAPPDATA "efa"

$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$tag = $release.tag_name
if (-not $tag) {
    Write-Error "could not determine the latest release tag"
    exit 1
}

$asset = "efa-windows-amd64.exe"
$url = "https://github.com/$repo/releases/download/$tag/$asset"

New-Item -ItemType Directory -Force -Path $installDir | Out-Null
$exePath = Join-Path $installDir "efa.exe"

Write-Host "Downloading $asset ($tag)..."
Invoke-WebRequest -Uri $url -OutFile $exePath

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not ($userPath -split ";" -contains $installDir)) {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Added $installDir to your PATH. Restart your terminal for it to take effect."
}

Write-Host "Installed $(& $exePath --version) to $exePath"
