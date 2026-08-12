# 跨平台构建脚本：生成 linux / windows 两种版本的可执行文件。
# 用法: .\scripts\build.ps1          (构建当前平台 + linux + windows)
#       .\scripts\build.ps1 -Platforms linux,windows
#       .\scripts\build.ps1 -Platforms windows
# 产物输出到 backend\dist\ 目录。

param(
    [string[]]$Platforms = @('linux', 'windows')
)

$ErrorActionPreference = 'Stop'
$BackendDir = Split-Path -Parent $PSScriptRoot
$DistDir = Join-Path $BackendDir 'dist'

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "未找到 go 命令，请先安装 Go 工具链"
}

Write-Host "Backend dir: $BackendDir"
Write-Host "Dist dir:    $DistDir"

$env:CGO_ENABLED = '0'

foreach ($platform in $Platforms) {
    switch ($platform) {
        'linux' {
            $env:GOOS = 'linux'
            $env:GOARCH = 'amd64'
            $output = Join-Path $DistDir 'miaoverse-linux-amd64'
        }
        'windows' {
            $env:GOOS = 'windows'
            $env:GOARCH = 'amd64'
            $output = Join-Path $DistDir 'miaoverse-windows-amd64.exe'
        }
        default {
            throw "不支持的平台: $platform (仅支持 linux / windows)"
        }
    }
    Write-Host "==> building $platform/amd64 -> $output"
    Push-Location $BackendDir
    try {
        go build -trimpath -ldflags "-s -w" -o $output .
        if ($LASTEXITCODE -ne 0) {
            throw "构建 $platform 失败，退出码 $LASTEXITCODE"
        }
    }
    finally {
        Pop-Location
    }
}

Write-Host "构建完成，产物位于: $DistDir"
