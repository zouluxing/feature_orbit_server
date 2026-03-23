# ================================================================
# stop.ps1 — feature_orbit_server 一键停止脚本 (Windows)
# 用法: PowerShell -ExecutionPolicy Bypass -File scripts/stop.ps1
# ================================================================

$ErrorActionPreference = "Continue"
$ROOT = Split-Path $PSScriptRoot -Parent
$UMS_DIR = Join-Path $ROOT "ums"

function Write-Step { param($n, $msg) Write-Host "`n[$n] $msg" -ForegroundColor Cyan }
function Write-OK   { param($msg) Write-Host "  OK  $msg" -ForegroundColor Green }
function Write-INFO { param($msg) Write-Host "  ...  $msg" -ForegroundColor Gray }

Write-Host ""
Write-Host "========================================" -ForegroundColor Yellow
Write-Host "   Feature Orbit Server — 停止"          -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Yellow

# ── Step 1: 停止主服务 ────────────────────────────────────────
Write-Step 1 "停止主服务 (feature_orbit_app / feature_orbit_postgres)"
Set-Location $ROOT
docker compose down 2>&1 | Out-Null
Write-OK "主服务已停止"

# ── Step 2: 停止 UMS ──────────────────────────────────────────
Write-Step 2 "停止 UMS 服务 (ums_app / ums_postgres / ums_redis)"
Set-Location $UMS_DIR
docker compose down 2>&1 | Out-Null
Write-OK "UMS 已停止"

# ── 完成 ──────────────────────────────────────────────────────
Write-Host ""
Write-Host "========================================" -ForegroundColor Yellow
Write-Host "   所有服务已停止"                        -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Yellow
Write-Host ""
Write-Host "  数据已持久化到 Docker Volume，下次启动数据不会丢失。" -ForegroundColor Gray
Write-Host "  重新启动: PowerShell -ExecutionPolicy Bypass -File scripts/start.ps1" -ForegroundColor Gray
Write-Host ""

# 可选：彻底清理（含数据卷），默认注释
# 如需清除所有数据重新初始化，取消下方注释后执行
# Write-Host "  彻底清理（含数据）..."
# Set-Location $ROOT;    docker compose down -v 2>&1 | Out-Null
# Set-Location $UMS_DIR; docker compose down -v 2>&1 | Out-Null
# Write-OK "数据卷已清除"
