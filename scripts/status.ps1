# ================================================================
# status.ps1 - feature_orbit_server 服务状态查看脚本 (Windows)
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

function Get-ContainerStatus {
    param([string]$Name)
    # 先用 docker ps -a 检查容器是否存在，避免 inspect 报错干扰输出
    $exists = docker ps -a --filter "name=^${Name}$" --format "{{.Names}}" 2>$null
    if ([string]::IsNullOrWhiteSpace($exists)) {
        return $null
    }
    # 容器存在，获取状态
    $state  = docker inspect $Name --format "{{.State.Status}}"        2>$null
    $health = docker inspect $Name --format "{{.State.Health.Status}}" 2>$null
    return @{ State = $state.Trim(); Health = $health.Trim() }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Blue
Write-Host "   Feature Orbit Server - 状态"          -ForegroundColor Blue
Write-Host "========================================" -ForegroundColor Blue

# -- 容器状态 ---------------------------------------------------------
Write-Host ""
Write-Host "[ 容器状态 ]" -ForegroundColor Cyan

$containers = @(
    @{ Name = "ums_postgres";          Label = "UMS PostgreSQL  :5433" },
    @{ Name = "ums_redis";             Label = "UMS Redis        :6380" },
    @{ Name = "ums_app";              Label = "UMS App          :8081" },
    @{ Name = "feature_orbit_postgres"; Label = "App PostgreSQL  :5432" },
    @{ Name = "feature_orbit_app";    Label = "App Server       :8080" }
)

foreach ($c in $containers) {
    $info = Get-ContainerStatus -Name $c.Name
    if ($null -eq $info) {
        Write-Host "  --       $($c.Label)  [未创建]" -ForegroundColor DarkGray
        continue
    }
    $state  = $info.State
    $health = $info.Health
    # health 可能为空字符串（容器没配置 healthcheck）
    if ($health -and $health -ne "<no value>") {
        $display = "$state / $health"
    } else {
        $display = $state
    }
    $isOK = ($state -eq "running") -and ($health -eq "" -or $health -eq "<no value>" -or $health -eq "healthy")
    $color = if ($isOK) { "Green" } elseif ($state -eq "running") { "Yellow" } else { "DarkGray" }
    Write-Host "  $display  $($c.Label)" -ForegroundColor $color
}

# -- HTTP 健康检查 --------------------------------------------------------
Write-Host ""
Write-Host "[ HTTP 健康检查 ]" -ForegroundColor Cyan

$umsOK = Check-Http "UMS  http://localhost:8081/health"              "http://localhost:8081/health"
$appOK = Check-Http "App  http://localhost:8080/health"              "http://localhost:8080/health"
Check-Http         "JWKS http://localhost:8081/.well-known/jwks.json" "http://localhost:8081/.well-known/jwks.json" | Out-Null

# -- 端口监听 ---------------------------------------------------------------
Write-Host ""
Write-Host "[ 端口监听 ]" -ForegroundColor Cyan
$ports = @(
    @{ Port = 5432; Label = "App PostgreSQL" },
    @{ Port = 5433; Label = "UMS PostgreSQL" },
    @{ Port = 6379; Label = "App Redis" },
    @{ Port = 6380; Label = "UMS Redis" },
    @{ Port = 8080; Label = "App Server" },
    @{ Port = 8081; Label = "UMS Server" }
)
foreach ($p in $ports) {
    $used = netstat -ano 2>$null | Select-String (":" + $p.Port + "\s") | Where-Object { $_ -match "LISTENING" }
    if ($used) {
        Write-Host "  LISTEN  :$($p.Port)  $($p.Label)" -ForegroundColor Green
    } else {
        Write-Host "  --      :$($p.Port)  $($p.Label)" -ForegroundColor DarkGray
    }
}

# -- 汇总 ---------------------------------------------------------------
Write-Host ""
Write-Host "========================================" -ForegroundColor Blue
if ($umsOK -and $appOK) {
    Write-Host "   状态: 正常 - 所有服务运行中" -ForegroundColor Green
} elseif ($umsOK -or $appOK) {
    Write-Host "   状态: 部分服务未就绪"         -ForegroundColor Yellow
    Write-Host "   重试: PowerShell -ExecutionPolicy Bypass -File scripts/start.ps1" -ForegroundColor Gray
} else {
    Write-Host "   状态: 服务未运行"               -ForegroundColor Red
    Write-Host "   启动: PowerShell -ExecutionPolicy Bypass -File scripts/start.ps1" -ForegroundColor Gray
}
Write-Host "========================================" -ForegroundColor Blue
Write-Host ""
