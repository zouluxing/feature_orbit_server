# ================================================================
# stop.ps1 - feature_orbit_server 一键停止脚本 (Windows)
# 用法:
#   仅停止服务（保留数据）: PowerShell -ExecutionPolicy Bypass -File scripts/stop.ps1
#   彻底清除（含数据）: PowerShell -ExecutionPolicy Bypass -File scripts/stop.ps1 -Clean
# ================================================================
param(
    [switch]$Clean  # 传入 -Clean 则同时删除数据卷
)

$ErrorActionPreference = "Continue"
$ROOT    = Split-Path $PSScriptRoot -Parent
$UMS_DIR = Join-Path $ROOT "ums"

function Write-Step { param($n, $msg) Write-Host "`n[$n] $msg" -ForegroundColor Cyan }
function Write-OK   { param($msg) Write-Host "  OK  $msg" -ForegroundColor Green }
function Write-INFO { param($msg) Write-Host "  ...  $msg" -ForegroundColor Gray }

Write-Host ""
Write-Host "========================================" -ForegroundColor Yellow
if ($Clean) {
    Write-Host "   Feature Orbit Server - 彻底清除（含数据）" -ForegroundColor Yellow
} else {
    Write-Host "   Feature Orbit Server - 停止"         -ForegroundColor Yellow
}
Write-Host "========================================" -ForegroundColor Yellow

$downArgs = if ($Clean) { @("compose", "down", "-v") } else { @("compose", "down") }

# -- Step 1: 停止主服务 -----------------------------------------
Write-Step 1 "停止主服务 (feature_orbit_app / feature_orbit_postgres)"
Set-Location $ROOT
& docker @downArgs 2>&1 | Out-Null
Write-OK "主服务已停止$(if ($Clean) {' 并清除数据卷'})"

# -- Step 2: 停止 UMS -----------------------------------------------
Write-Step 2 "停止 UMS 服务 (ums_app / ums_postgres / ums_redis)"
Set-Location $UMS_DIR
& docker @downArgs 2>&1 | Out-Null
Write-OK "UMS 已停止$(if ($Clean) {' 并清除数据卷'})"

# -- 完成 -----------------------------------------------------------
Set-Location $ROOT
Write-Host ""
Write-Host "========================================" -ForegroundColor Yellow
if ($Clean) {
    Write-Host "   所有服务已停止，数据已清除" -ForegroundColor Yellow
    Write-Host "   下次启动会全新初始化数据库" -ForegroundColor Gray
} else {
    Write-Host "   所有服务已停止，数据已持久化保留" -ForegroundColor Yellow
    Write-Host "   彻底清除（含数据）: scripts/stop.ps1 -Clean" -ForegroundColor Gray
}
Write-Host "   重新启动: scripts/start.ps1"              -ForegroundColor Gray
Write-Host "========================================" -ForegroundColor Yellow
Write-Host ""
