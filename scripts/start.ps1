# ================================================================
# start.ps1 — feature_orbit_server 一键启动脚本 (Windows)
# 用法: PowerShell -ExecutionPolicy Bypass -File scripts/start.ps1
# ================================================================

$ErrorActionPreference = "Stop"
$ROOT = Split-Path $PSScriptRoot -Parent
$UMS_DIR = Join-Path $ROOT "ums"

function Write-Step { param($n, $msg) Write-Host "`n[$n] $msg" -ForegroundColor Cyan }
function Write-OK   { param($msg) Write-Host "  OK  $msg" -ForegroundColor Green }
function Write-FAIL { param($msg) Write-Host "  FAIL $msg" -ForegroundColor Red; exit 1 }
function Write-INFO { param($msg) Write-Host "  ...  $msg" -ForegroundColor Gray }

Write-Host ""
Write-Host "========================================" -ForegroundColor Blue
Write-Host "   Feature Orbit Server — 启动"           -ForegroundColor Blue
Write-Host "========================================" -ForegroundColor Blue

# ── 前置检查 ──────────────────────────────────────────────────
Write-Step 0 "前置检查"

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-FAIL "Docker 未安装或未启动，请先启动 Docker Desktop"
}
docker info 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) { Write-FAIL "Docker Desktop 未运行，请先启动它" }
Write-OK "Docker Desktop 已就绪"

if (-not (Test-Path $UMS_DIR)) { Write-FAIL "找不到 ums/ 目录，请确认在项目根目录执行" }
Write-OK "项目结构完整"

# ── Step 1: RSA 密钥 ──────────────────────────────────────────
Write-Step 1 "检查 UMS RSA 密钥"

$PRIV = Join-Path $UMS_DIR "configs\ums_rsa_private.pem"
$PUB  = Join-Path $UMS_DIR "configs\ums_rsa_public.pem"
$CFGDIR = Join-Path $UMS_DIR "configs"

if (-not (Test-Path $PRIV) -or -not (Test-Path $PUB)) {
    Write-INFO "密钥不存在，正在生成 RSA 2048-bit 密钥对..."
    if (-not (Test-Path $CFGDIR)) { New-Item -ItemType Directory -Path $CFGDIR | Out-Null }

    # 优先用 openssl（Git for Windows 自带）
    $openssl = Get-Command openssl -ErrorAction SilentlyContinue
    if ($openssl) {
        & openssl genrsa -out $PRIV 2048 2>&1 | Out-Null
        & openssl rsa -in $PRIV -pubout -out $PUB 2>&1 | Out-Null
        if (Test-Path $PRIV) { Write-OK "RSA 密钥已生成 (openssl)" }
        else { Write-FAIL "openssl 生成密钥失败，请手动执行: openssl genrsa -out ums/configs/ums_rsa_private.pem 2048" }
    } else {
        # 降级：用 Go 程序生成
        $go = Get-Command go -ErrorAction SilentlyContinue
        if ($go) {
            Write-INFO "openssl 未找到，尝试用 Go 生成密钥..."
            $genScript = @"
package main
import ("crypto/rand";"crypto/rsa";"crypto/x509";"encoding/pem";"os")
func main() {
    k,_:=rsa.GenerateKey(rand.Reader,2048)
    pf,_:=os.Create(os.Args[1])
    pem.Encode(pf,&pem.Block{Type:"RSA PRIVATE KEY",Bytes:x509.MarshalPKCS1PrivateKey(k)})
    pf.Close()
    bf,_:=os.Create(os.Args[2])
    b,_:=x509.MarshalPKIXPublicKey(&k.PublicKey)
    pem.Encode(bf,&pem.Block{Type:"PUBLIC KEY",Bytes:b})
    bf.Close()
}
"@
            $tmp = Join-Path $env:TEMP "genkey.go"
            Set-Content $tmp $genScript
            & go run $tmp $PRIV $PUB
            Remove-Item $tmp -ErrorAction SilentlyContinue
            if (Test-Path $PRIV) { Write-OK "RSA 密钥已生成 (go run)" }
            else { Write-FAIL "密钥生成失败，请手动运行: cd ums && make keys-gen" }
        } else {
            Write-FAIL "openssl 和 go 均未找到，无法生成密钥。请安装 Git for Windows 或 Go 后重试"
        }
    }
} else {
    Write-OK "RSA 密钥已存在"
}

# ── Step 2: 启动 UMS（postgres + redis + ums_app）─────────────
Write-Step 2 "启动 UMS 服务 (postgres:5433 / redis:6380 / ums:8081)"

Set-Location $UMS_DIR
docker compose up -d --build 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) { Write-FAIL "UMS docker compose up 失败，运行 'docker logs ums_app' 查看详情" }
Write-INFO "容器已启动，等待健康检查..."

# 轮询 UMS 健康端点，最多 60 秒
$timeout = 60; $elapsed = 0; $ready = $false
while ($elapsed -lt $timeout) {
    try {
        $r = Invoke-WebRequest -Uri "http://localhost:8081/health" -UseBasicParsing -TimeoutSec 3 -ErrorAction Stop
        if ($r.StatusCode -eq 200) { $ready = $true; break }
    } catch {}
    Start-Sleep -Seconds 3; $elapsed += 3
    Write-INFO "等待 UMS 就绪... ($elapsed/$timeout s)"
}
if (-not $ready) { Write-FAIL "UMS 60 秒内未就绪，请运行 'docker logs ums_app --tail 50' 排查" }
Write-OK "UMS 健康检查通过 (http://localhost:8081/health)"

# ── Step 3: 启动主服务（postgres + app）──────────────────────
Write-Step 3 "启动主服务 (postgres:5432 / app:8080)"

Set-Location $ROOT
docker compose up -d --build 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) { Write-FAIL "主服务 docker compose up 失败，运行 'docker logs feature_orbit_app' 查看详情" }
Write-INFO "容器已启动，等待健康检查..."

$timeout = 60; $elapsed = 0; $ready = $false
while ($elapsed -lt $timeout) {
    try {
        $r = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 3 -ErrorAction Stop
        if ($r.StatusCode -eq 200) { $ready = $true; break }
    } catch {}
    Start-Sleep -Seconds 3; $elapsed += 3
    Write-INFO "等待主服务就绪... ($elapsed/$timeout s)"
}
if (-not $ready) { Write-FAIL "主服务 60 秒内未就绪，请运行 'docker logs feature_orbit_app --tail 50' 排查" }
Write-OK "主服务健康检查通过 (http://localhost:8080/health)"

# ── 完成 ──────────────────────────────────────────────────────
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "   所有服务已就绪"                        -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "  UMS  认证服务   http://localhost:8081" -ForegroundColor White
Write-Host "  App  主服务     http://localhost:8080" -ForegroundColor White
Write-Host ""
Write-Host "  快速验证:" -ForegroundColor Gray
Write-Host "    curl http://localhost:8081/.well-known/jwks.json" -ForegroundColor Gray
Write-Host "    curl http://localhost:8080/api/v1/features" -ForegroundColor Gray
Write-Host ""
Write-Host "  停止服务: PowerShell -ExecutionPolicy Bypass -File scripts/stop.ps1" -ForegroundColor Gray
Write-Host "  查看状态: PowerShell -ExecutionPolicy Bypass -File scripts/status.ps1" -ForegroundColor Gray
Write-Host ""
