$ErrorActionPreference = "Stop"
$repo = "oSethoum/gvm"
$installDir = if ($env:GVM_INSTALL_DIR) { $env:GVM_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "gvm\bin" }
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$tag = $release.tag_name
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$asset = $release.assets | Where-Object { $_.name -match "windows_${arch}\.zip" } | Select-Object -First 1
if (-not $asset) { throw "No asset found for windows/$arch" }

$tmp = Join-Path $env:TEMP "gvm.zip"
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tmp
Expand-Archive -Path $tmp -DestinationPath $installDir -Force
Remove-Item $tmp

# Refresh current session PATH
if (-not ($env:PATH -split ';' | Where-Object { $_ -eq $installDir })) {
    $env:PATH = $env:PATH + ";$installDir"
}
if ($env:GOBIN -and -not ($env:PATH -split ';' | Where-Object { $_ -eq $env:GOBIN })) {
    $env:PATH = $env:PATH + ";$env:GOBIN"
}
if (-not ($env:PATH -split ';' | Where-Object { $_ -eq '%GOBIN%' })) {
    $env:PATH = $env:PATH + ";%GOBIN%"
}

# Persist to user PATH
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if (-not $userPath) { $userPath = "" }
if (-not ($userPath -split ';' | Where-Object { $_ -eq $installDir })) {
    $newPath = if ($userPath) { "$userPath;$installDir" } else { $installDir }
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    $userPath = $newPath
}
if ($env:GOBIN -and -not ($userPath -split ';' | Where-Object { $_ -eq $env:GOBIN })) {
    $newPath = if ($userPath) { "$userPath;$env:GOBIN" } else { $env:GOBIN }
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    $userPath = $newPath
}
if (-not ($userPath -split ';' | Where-Object { $_ -eq '%GOBIN%' })) {
    $newPath = if ($userPath) { "$userPath;%GOBIN%" } else { "%GOBIN%" }
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    $userPath = $newPath
}

Write-Host "Installed gvm $tag to $installDir\gvm.exe"
Write-Host "gvm is ready in this terminal. For other terminals, restart or run: `$env:PATH = '$installDir;' + `$env:PATH"
