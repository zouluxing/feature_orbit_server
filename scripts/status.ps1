# ================================================================
# status.ps1 — feature_orbit_server 服务状态查看脚本 (Windows)
# 用法: PowerShell -ExecutionPolicy Bypass -File scripts/status.ps1
# ================================================================

$ErrorActionPreference = "Continue"
$ROOT = Split-Path $PSScriptRoot -Parent
$UMS_DIR = Join-Path $ROOT "ums"

function Check-Http {
    param($label, $url)
    try {
        $r = Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
        if ($r.StatusCode -eq 200) {
            Write-Host "  OK   $label" -ForegroundColor Green
            return $true
        }
    } catch {}
    Write-Host "  DOWN $label" -ForegroundColor Red
    return $false
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Blue
Write-Host "   Feature Orbit Server — 状态"          -ForegroundColor Blue
Write-Host "========================================" -ForegroundColor Blue

# ── 容器状态 ──────────────────────────────────────────────────
Write-Host "`n[ 容器状态 ]" -ForegroundColor Cyan

$containers = @(
    @{Name="ums_postgres";     Label="UMS PostgreSQL  :5433"},
    @{Name="ums_redis";        Label="UMS Redis        :6380"},
    @{Name="ums_app";          Label="UMS App          :8081"},
    @{Name="feature_orbit_postgres"; Label="App PostgreSQL  :5432"},
    @{Name="feature_orbit_app"; Label="App Server       :8080"}
)

foreach ($c in $containers) {
    $info = docker inspect $c.Name --format '{{.State.Status}} {{.State.Health.Status}}' 2>$null
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($info)) {
        Write-Host "  --   $($c.Label)  [未运行]" -ForegroundColor DarkGray
    } else {
        $parts  = $info.Trim().Split(' ')
        $state  = $parts[0]
        $health = if ($parts.Count -gt 1) { $parts[1] } else { "" }
        $display = if ($health -and $health -ne '<no\ value>') { "$state / $health" } else { $state }
        $color = if ($state -eq "running" -and ($health -eq "" -or $health -eq "healthy")) { "Green" } else { "Yellow" }
        Write-Host "  $display  $($c.Label)" -ForegroundColor $color
    }
}

# ── HTTP 健康检查 ──────────────────────────────────────────────
Write-Host "`n[ HTTP 健康检查 ]" -ForegroundColor Cyan

$umsOK = Check-Http "UMS  http://localhost:8081/health" "http://localhost:8081/health"
$appOK = Check-Http "App  http://localhost:8080/health" "http://localhost:8080/health"
Check-Http "JWKS http://localhost:8081/.well-known/jwks.json" "http://localhost:8081/.well-known/jwks.json" | Out-Null

# ── 端口占用 ───────────────────────────────────────────────────
Write-Host "`n[ 端口监听 ]" -ForegroundColor Cyan
$ports = @(5432, 5433, 6379, 6380, 8080, 8081)
foreach ($port in $ports) {
    $used = netstat -ano 2>$null | Select-String ":$port\s" | Where-Object { $_ -match "LISTENING" }
    if ($used) {
        Write-Host "  LISTEN  :$port" -ForegroundColor Green
    } else {
        Write-Host "  --      :$port" -ForegroundColor DarkGray
    }
}

# ── 汇总 ──────────────────────────────────────────────────────
Write-Host ""
Write-Host "========================================" -ForegroundColor Blue
if ($umsOK -and $appOK) {
    Write-Host "   状态: 正常 — 所有服务运行中"     -ForegroundColor Green
} elseif ($umsOK -or $appOK) {
    Write-Host "   状态: 部分服务未就绪"             -ForegroundColor Yellow
} else {
    Write-Host "   状态: 服务未运行"                 -ForegroundColor Red
    Write-Host "   启动: PowerShell -ExecutionPolicy Bypass -File scripts/start.ps1" -ForegroundColor Gray
}
Write-Host "========================================" -ForegroundColor Blue
Write-Host ""
