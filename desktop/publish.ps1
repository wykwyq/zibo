param(
    [string]$OutputDirectory = (Join-Path $PSScriptRoot '..\desktop-dist\BililiveGo-Desktop')
)

$ErrorActionPreference = 'Stop'
$repositoryRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$outputPath = [System.IO.Path]::GetFullPath($OutputDirectory)

Push-Location $repositoryRoot
try {
    make build-web build
    dotnet publish (Join-Path $PSScriptRoot 'BililiveGoDesktop.csproj') `
        -c Release -r win-x64 --self-contained true `
        -p:PublishSingleFile=true `
        -p:IncludeNativeLibrariesForSelfExtract=true `
        -p:EnableCompressionInSingleFile=true `
        -o $outputPath
}
finally {
    Pop-Location
}

New-Item -ItemType Directory -Force -Path (Join-Path $outputPath 'backend'), (Join-Path $outputPath 'Videos') | Out-Null
Copy-Item -LiteralPath (Join-Path $repositoryRoot 'bin\bililive-windows-amd64.exe') `
    -Destination (Join-Path $outputPath 'backend\bililive-go.exe') -Force
Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'favicon.ico') -Destination $outputPath -Force

$targetConfig = Join-Path $outputPath 'config.yml'
if (-not (Test-Path -LiteralPath $targetConfig)) {
    $config = Get-Content -LiteralPath (Join-Path $repositoryRoot 'config.yml') -Raw
    $config = $config -replace 'out_put_path: \./', 'out_put_path: ./Videos'
    $config = $config -replace 'bind: :8080', 'bind: 127.0.0.1:28080'
    $config = $config -replace 'convert_to_mp4: false', 'convert_to_mp4: true'
    $config = $config -replace 'is_listening: true', 'is_listening: false'
    $config = $config -replace 'auto_check: true', 'auto_check: false'
    $config = $config -replace 'auto_download: true', 'auto_download: false'
    Set-Content -LiteralPath $targetConfig -Value $config -Encoding utf8NoBOM
}

Write-Host "桌面版已生成：$outputPath"
