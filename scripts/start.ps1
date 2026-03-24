# ================================================================
# start.ps1 - feature_orbit_server 一键启动脚本 (Windows)
# 用法: PowerShell -ExecutionPolicy Bypass -File scripts/start.ps1
# 带删库重建: PowerShell -ExecutionPolicy Bypass -File scripts/start.ps1 -Clean
# ================================================================
param(
    [switch]$Clean  # 启动前先清除数据卷（用于处理数据库结构冲突）
)

$ErrorActionPreference = "Continue"
$ROOT    = Split-Path $PSScriptRoot -Parent
$UMS_DIR = Join-Path $ROOT "ums"

function Write-Step { param($n, $msg) Write-Host "`n[$n] $msg" -ForegroundColor Cyan }
function Write-OK   { param($msg) Write-Host "  OK  $msg" -ForegroundColor Green }
function Write-Fail { param($msg) Write-Host "  FAIL $msg" -ForegroundColor Red; Set-Location $ROOT; exit 1 }
function Write-Info { param($msg) Write-Host "  ...  $msg" -ForegroundColor Gray }
function Write-Warn { param($msg) Write-Host "  WARN $msg" -ForegroundColor Yellow }

Write-Host ""
Write-Host "========================================" -ForegroundColor Blue
Write-Host "   Feature Orbit Server - 启动"          -ForegroundColor Blue
Write-Host "========================================" -ForegroundColor Blue

# -- 前置检查 ---------------------------------------------------------------
Write-Step 0 "前置检查"

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Fail "Docker 未安装或未在 PATH 中，请安装 Docker Desktop"
}
docker info 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) { Write-Fail "Docker Desktop 未运行，请先启动它" }
Write-OK "Docker Desktop 已就绪"

if (-not (Test-Path $UMS_DIR)) { Write-Fail "找不到 ums/ 目录，请在项目根目录运行此脚本" }
Write-OK "项目结构检查通过"

# -- Step 1: 清除旧数据卷（-Clean 模式）-------------------------------------------
if ($Clean) {
    Write-Step 1 "清除旧数据卷（-Clean 模式）"
    Write-Info "正在删除主服务数据卷..."
    Set-Location $ROOT;    docker compose down -v 2>&1 | Out-Null
    Write-Info "正在删除 UMS 数据卷..."
    Set-Location $UMS_DIR; docker compose down -v 2>&1 | Out-Null
    Write-OK "数据卷已全部清除，将全新初始化"
} else {
    Write-Step 1 "检查数据卷状态"
    # 检查是否存在旧 volume，如果存在提示用户
    $appVolume = docker volume ls --filter "name=feature_orbit" --format "{{.Name}}" 2>$null
    $umsVolume = docker volume ls --filter "name=ums"            --format "{{.Name}}" 2>$null
    if ($appVolume -or $umsVolume) {
        Write-OK "数据卷已存在，将使用现有数据重启动"
        Write-Info "如遇到数据库结构冲突错误，请运行: scripts/start.ps1 -Clean"
    } else {
        Write-OK "无旧数据卷，将全新初始化"
    }
}

# -- Step 2: 检查 go.sum -------------------------------------------------------
Write-Step 2 "检查 go.sum 依赖锁定文件"

$go = Get-Command go -ErrorAction SilentlyContinue
if ($go) {
    $umsGoSum  = Join-Path $UMS_DIR "go.sum"
    $rootGoSum = Join-Path $ROOT "go.sum"
    if (-not (Test-Path $umsGoSum)) {
        Write-Info "ums/go.sum 不存在，执行 go mod tidy..."
        Push-Location $UMS_DIR; go mod tidy; Pop-Location
        if (Test-Path $umsGoSum) { Write-OK "ums/go.sum 已生成" }
    } else { Write-OK "ums/go.sum 已存在" }

    if (-not (Test-Path $rootGoSum)) {
        Write-Info "根目录 go.sum 不存在，执行 go mod tidy..."
        Push-Location $ROOT; go mod tidy; Pop-Location
        if (Test-Path $rootGoSum) { Write-OK "根目录 go.sum 已生成" }
    } else { Write-OK "根目录 go.sum 已存在" }
} else {
    Write-Info "Go 未安装，跳过 go.sum 检查（Docker 构建时处理）"
}

# -- Step 3: RSA 密钥 -------------------------------------------------------
Write-Step 3 "检查 UMS RSA 密钥"

$PRIV   = Join-Path $UMS_DIR "configs\ums_rsa_private.pem"
$PUB    = Join-Path $UMS_DIR "configs\ums_rsa_public.pem"
$CFGDIR = Join-Path $UMS_DIR "configs"

if (-not (Test-Path $PRIV) -or -not (Test-Path $PUB)) {
    Write-Info "密钥不存在，正在生成 RSA 2048-bit 密钥对..."
    if (-not (Test-Path $CFGDIR)) { New-Item -ItemType Directory -Path $CFGDIR | Out-Null }
    $openssl = Get-Command openssl -ErrorAction SilentlyContinue
    if ($openssl) {
        & openssl genrsa -out $PRIV 2048 2>&1 | Out-Null
        & openssl rsa -in $PRIV -pubout -out $PUB 2>&1 | Out-Null
        if (Test-Path $PRIV) { Write-OK "RSA 密钥已生成 (openssl)" }
        else { Write-Fail "密钥生成失败，请手动运行: openssl genrsa -out ums/configs/ums_rsa_private.pem 2048" }
    } elseif ($go) {
        $genScript = @"
package main
import ("crypto/rand";"crypto/rsa";"crypto/x509";"encoding/pem";"os")
func main() {
    k,_:=rsa.GenerateKey(rand.Reader,2048)
    pf,_:=os.Create(os.Args[1]);pem.Encode(pf,&pem.Block{Type:"RSA PRIVATE KEY",Bytes:x509.MarshalPKCS1PrivateKey(k)});pf.Close()
    bf,_:=os.Create(os.Args[2]);b,_:=x509.MarshalPKIXPublicKey(&k.PublicKey);pem.Encode(bf,&pem.Block{Type:"PUBLIC KEY",Bytes:b});bf.Close()
}
"@
        $tmp = Join-Path $env:TEMP "genkey.go"
        Set-Content $tmp $genScript
        & go run $tmp $PRIV $PUB
        Remove-Item $tmp -ErrorAction SilentlyContinue
        if (Test-Path $PRIV) { Write-OK "RSA 密钥已生成 (go run)" }
        else { Write-Fail "密钥生成失败" }
    } else {
        Write-Fail "openssl 和 go 均未找到，请安装 Git for Windows"
    }
} else {
    Write-OK "RSA 密钥已存在"
}

# -- Step 4: 启动 UMS ------------------------------------------------------
Write-Step 4 "启动 UMS 服务 (postgres:5433 / redis:6380 / ums:8081)"

Set-Location $UMS_DIR
Write-Info "正在构建并启动 UMS 容器（首次构建需要几分钟）..."
docker compose up -d --build
if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "  UMS 构建失败，常见原因：" -ForegroundColor Yellow
    Write-Host "  1. 依赖下载失败: 检查网络" -ForegroundColor Gray
    Write-Host "  2. 端口占用:     netstat -ano | findstr :5433" -ForegroundColor Gray
    Write-Host "  3. 查看日志:     docker logs ums_app --tail 50" -ForegroundColor Gray
    Set-Location $ROOT; exit 1
}

Write-Info "容器已启动，等待 UMS 就绪..."
$timeout = 90; $elapsed = 0; $ready = $false
while ($elapsed -lt $timeout) {
    try {
        $r = Invoke-WebRequest -Uri "http://localhost:8081/health" -UseBasicParsing -TimeoutSec 3 -ErrorAction Stop
        if ($r.StatusCode -eq 200) { $ready = $true; break }
    } catch {}
    Start-Sleep -Seconds 3; $elapsed += 3
    Write-Info "等待 UMS... ($elapsed/$timeout s)"
}
if (-not $ready) {
    Write-Host ""; Write-Host "  UMS 未在 $timeout 秒内就绪，请查看日志：" -ForegroundColor Yellow
    Write-Host "    docker logs ums_app --tail 50" -ForegroundColor Gray
    Set-Location $ROOT; exit 1
}
Write-OK "UMS 就绪 (http://localhost:8081/health)"

# -- Step 5: 启动主服务 ---------------------------------------------------
Write-Step 5 "启动主服务 (postgres:5432 / app:8080)"

Set-Location $ROOT
Write-Info "正在构建并启动主服务容器..."
docker compose up -d --build
if ($LASTEXITCODE -ne 0) {
    Write-Host ""; Write-Host "  主服务构建失败，请查看日志：" -ForegroundColor Yellow
    Write-Host "    docker logs feature_orbit_app --tail 50" -ForegroundColor Gray
    exit 1
}

Write-Info "容器已启动，等待主服务就绪..."
$timeout = 90; $elapsed = 0; $ready = $false
while ($elapsed -lt $timeout) {
    try {
        $r = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 3 -ErrorAction Stop
        if ($r.StatusCode -eq 200) { $ready = $true; break }
    } catch {}
    Start-Sleep -Seconds 3; $elapsed += 3
    Write-Info "等待主服务... ($elapsed/$timeout s)"
}
if (-not $ready) {
    Write-Host ""; Write-Host "  主服务未在 $timeout 秒内就绪，请查看日志：" -ForegroundColor Yellow
    Write-Host "    docker logs feature_orbit_app --tail 50" -ForegroundColor Gray
    exit 1
}
Write-OK "主服务就绪 (http://localhost:8080/health)"

# -- 完成 -------------------------------------------------------------------
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "   所有服务已就绪"                 -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "  UMS   http://localhost:8081" -ForegroundColor White
Write-Host "  App   http://localhost:8080" -ForegroundColor White
Write-Host ""
Write-Host "  快速验证：" -ForegroundColor Gray
Write-Host "    curl http://localhost:8081/health" -ForegroundColor Gray
Write-Host "    curl http://localhost:8080/health" -ForegroundColor Gray
Write-Host ""
Write-Host "  查看状态: PowerShell -ExecutionPolicy Bypass -File scripts/status.ps1" -ForegroundColor Gray
Write-Host "  停止:     PowerShell -ExecutionPolicy Bypass -File scripts/stop.ps1" -ForegroundColor Gray
Write-Host "  清除重建: PowerShell -ExecutionPolicy Bypass -File scripts/stop.ps1 -Clean" -ForegroundColor Gray
Write-Host ""
