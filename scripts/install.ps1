# Install the koru CLI on Windows.
#
#   irm https://koru.js.org/install.ps1 | iex
#
# Set $env:KORU_VERSION = "v0.0.1" to pin a release and $env:KORU_INSTALL_DIR
# to choose the folder (default: %LOCALAPPDATA%\Programs\koru). The archive is
# checked against the SHA-256 published with the release before installing.
# Source: https://github.com/wux4an/koru

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

function Install-Koru {
    $repo = 'wux4an/koru'
    $base = if ($env:KORU_DOWNLOAD_URL) { $env:KORU_DOWNLOAD_URL } else { "https://github.com/$repo/releases" }
    $dir = if ($env:KORU_INSTALL_DIR) { $env:KORU_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\koru' }

    $arch = switch ($env:PROCESSOR_ARCHITECTURE) {
        'AMD64' { 'amd64' }
        'ARM64' { 'arm64' }
        default { throw "koru install: unsupported architecture $($env:PROCESSOR_ARCHITECTURE)" }
    }

    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
    $version = $env:KORU_VERSION
    if (-not $version) {
        # /releases/latest redirects to /releases/tag/vX.Y.Z; no API rate limit.
        $request = [Net.WebRequest]::Create("$base/latest")
        $request.Method = 'HEAD'
        $response = $request.GetResponse()
        $version = $response.ResponseUri.Segments[-1]
        $response.Close()
        if ($version -notmatch '^v\d') { throw "koru install: could not find the latest release at $base/latest" }
    }
    if ($version -notmatch '^v') { $version = "v$version" }

    $name = "koru_$($version.Substring(1))_windows_$arch"
    $url = "$base/download/$version/$name.zip"
    $tmp = Join-Path ([IO.Path]::GetTempPath()) ([IO.Path]::GetRandomFileName())
    New-Item -ItemType Directory -Path $tmp | Out-Null
    try {
        Write-Host "Downloading koru $version for windows/$arch"
        $zip = Join-Path $tmp "$name.zip"
        Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $zip
        $sums = Join-Path $tmp "$name.zip.sha256"
        Invoke-WebRequest -UseBasicParsing -Uri "$url.sha256" -OutFile $sums
        $want = ((Get-Content -Raw $sums) -split '\s+')[0].ToLower()
        $got = (Get-FileHash -Algorithm SHA256 -Path $zip).Hash.ToLower()
        if (-not $want -or $want -ne $got) { throw "koru install: checksum mismatch for $name.zip (expected $want, got $got)" }
        Write-Host "Checksum verified ($got)"

        Expand-Archive -Path $zip -DestinationPath $tmp -Force
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
        Copy-Item -Force (Join-Path $tmp "$name\koru.exe") (Join-Path $dir 'koru.exe')
    } finally {
        Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
    }

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (($userPath -split ';') -notcontains $dir) {
        [Environment]::SetEnvironmentVariable('Path', (@($userPath, $dir) | Where-Object { $_ }) -join ';', 'User')
        $env:Path = "$env:Path;$dir"
        Write-Host "Added $dir to your user PATH (open a new terminal to use it everywhere)."
    }
    Write-Host "Installed $(& (Join-Path $dir 'koru.exe') version) to $dir\koru.exe"
    Write-Host 'Run: koru create my-app'
}

Install-Koru
