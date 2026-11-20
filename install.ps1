$ErrorActionPreference = "Stop"
$repo = "oSethoum/gvm"
$installDir = if ($env:GVM_INSTALL_DIR) { $env:GVM_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "gvm\bin" }
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$tag = $release.tag_name
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
# prefer arm64 if needed
$asset = $release.assets | Where-Object { $_.name -match "windows_${arch}\.zip" } | Select-Object -First 1
if (-not $asset) { throw "No asset found for windows/$arch" }

$tmp = Join-Path $env:TEMP "gvm.zip"
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tmp
Expand-Archive -Path $tmp -DestinationPath $installDir -Force
Remove-Item $tmp

Write-Host "Installed to $installDir\gvm.exe"
Write-Host "Add to PATH: `$env:PATH = '$installDir;' + `$env:PATH"
