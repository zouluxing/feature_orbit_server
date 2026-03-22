# ================================================
# feature_orbit 环境自动检测与准备脚本 (Windows)
# 版本: 2.0.0
# 平台: Windows + Go SaaS 开发
# 用法: PowerShell -ExecutionPolicy Bypass -File scripts/setup-env.ps1
# ================================================

$ErrorActionPreference = "Continue"
$CHECK_PASS = 0
$CHECK_FAIL = 0
$AUTO_FIXED = 0

function Log-Pass  { param($msg) Write-Host "✅ $msg" -ForegroundColor Green;  $script:CHECK_PASS++ }
function Log-Fail  { param($msg) Write-Host "❌ $msg" -ForegroundColor Red;    $script:CHECK_FAIL++ }
function Log-Warn  { param($msg) Write-Host "⚠️  $msg" -ForegroundColor Yellow }
function Log-Fix   { param($msg) Write-Host "🔧 $msg" -ForegroundColor Cyan;   $script:AUTO_FIXED++ }
function Log-Info  { param($msg) Write-Host "ℹ️  $msg" -ForegroundColor Blue }
function Log-Title { param($msg) Write-Host "`n--- $msg ---" -ForegroundColor Magenta }

# 刷新当前会话的环境变量（安装后立即生效）
function Refresh-EnvPath {
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" +
                [System.Environment]::GetEnvironmentVariable("Path", "User")
}

Write-Host ""
Write-Host "================================================" -ForegroundColor Blue
Write-Host "   feature_orbit 环境自动检测与准备 v2.0" -ForegroundColor Blue
Write-Host "   平台: Windows | 语言: Go | 类型: SaaS" -ForegroundColor Blue
Write-Host "================================================" -ForegroundColor Blue
Write-Host ""

# ================================================================
# 一、基础环境检查 E01~E09
# ================================================================
Log-Title "基础环境检查 (E01~E09)"

# E01: Git
Write-Host "[E01] 检测 Git..." -ForegroundColor Blue
if (Get-Command git -ErrorAction SilentlyContinue) {
    Log-Pass "$(git --version) 已就绪"
} else {
    Log-Warn "Git 未安装，尝试通过 winget 安装..."
    if (Get-Command winget -ErrorAction SilentlyContinue) {
        Log-Fix "执行: winget install --id Git.Git -e --source winget"
        winget install --id Git.Git -e --source winget --silent
        Refresh-EnvPath
        if (Get-Command git -ErrorAction SilentlyContinue) { Log-Pass "Git 安装成功" }
        else { Log-Fail "Git 安装失败，请手动安装: https://git-scm.com/download/win" }
    } else {
        Log-Fail "请手动安装 Git: https://git-scm.com/download/win"
    }
}

# E02: Git 用户配置
Write-Host "[E02] 检测 Git 用户配置..." -ForegroundColor Blue
$gitName  = git config --global user.name  2>$null
$gitEmail = git config --global user.email 2>$null
if (-not $gitName) {
    Log-Warn "Git user.name 未配置"
    $inputName = Read-Host "  👤 请输入你的 Git 用户名"
    git config --global user.name $inputName
    Log-Fix "Git user.name 已设置为: $inputName"
    $gitName = $inputName
} else { Log-Pass "Git user.name: $gitName" }

if (-not $gitEmail) {
    Log-Warn "Git user.email 未配置"
    $inputEmail = Read-Host "  📧 请输入你的 Git 邮箱"
    git config --global user.email $inputEmail
    Log-Fix "Git user.email 已设置为: $inputEmail"
} else { Log-Pass "Git user.email: $gitEmail" }

# E03: Node.js（Claude Code 依赖）
Write-Host "[E03] 检测 Node.js..." -ForegroundColor Blue
try {
    $nodeVer = (node -v 2>$null).TrimStart('v').Split('.')[0]
    if ([int]$nodeVer -ge 18) { Log-Pass "Node.js $(node -v) 已就绪" }
    else {
        Log-Warn "Node.js 版本过低（需要 v18+），尝试更新..."
        if (Get-Command winget -ErrorAction SilentlyContinue) {
            winget install OpenJS.NodeJS.LTS --silent
            Refresh-EnvPath
            Log-Fix "Node.js 已更新，请重启 PowerShell 后重新运行"
        } else { Log-Fail "请访问 https://nodejs.org 手动安装 Node.js v18+" }
    }
} catch {
    Log-Warn "Node.js 未安装，自动安装中..."
    if (Get-Command winget -ErrorAction SilentlyContinue) {
        Log-Fix "执行: winget install OpenJS.NodeJS.LTS"
        winget install OpenJS.NodeJS.LTS --silent
        Refresh-EnvPath
        if (Get-Command node -ErrorAction SilentlyContinue) { Log-Pass "Node.js $(node -v) 安装成功" }
        else { Log-Fail "安装后需重启 PowerShell，请重新运行脚本" }
    } else { Log-Fail "请访问 https://nodejs.org 手动安装" }
}

# E04: Claude Code
Write-Host "[E04] 检测 Claude Code..." -ForegroundColor Blue
if (Get-Command claude -ErrorAction SilentlyContinue) {
    Log-Pass "Claude Code 已就绪"
} else {
    Log-Warn "Claude Code 未安装，自动安装中..."
    Log-Fix "执行: npm install -g @anthropic-ai/claude-code"
    npm install -g @anthropic-ai/claude-code
    Refresh-EnvPath
    if (Get-Command claude -ErrorAction SilentlyContinue) { Log-Pass "Claude Code 安装成功" }
    else { Log-Fail "Claude Code 安装失败，请手动执行: npm install -g @anthropic-ai/claude-code" }
}

# E05: 仓库克隆状态
Write-Host "[E05] 检测仓库状态..." -ForegroundColor Blue
try {
    git rev-parse --git-dir | Out-Null
    $remoteUrl = git remote get-url origin 2>$null
    Log-Pass "Git 仓库已就绪: $remoteUrl"
} catch {
    Log-Warn "当前目录不是 Git 仓库，自动克隆..."
    Set-Location ..
    git clone https://github.com/zouluxing/feature_orbit.git
    Set-Location feature_orbit
    Log-Fix "仓库克隆成功"
}

# E06: 当前分支
Write-Host "[E06] 检测当前分支..." -ForegroundColor Blue
$currentBranch = git branch --show-current 2>$null
if ($currentBranch -eq "develop") {
    Log-Pass "当前分支: develop"
} else {
    Log-Warn "当前分支为 '$currentBranch'，自动切换到 develop..."
    git checkout develop 2>$null
    if ($LASTEXITCODE -ne 0) { git checkout -b develop origin/develop }
    Log-Fix "已切换到 develop 分支"
    Log-Pass "当前分支: develop"
}

# E07: 同步远端代码
Write-Host "[E07] 同步远端代码..." -ForegroundColor Blue
git fetch origin --quiet
$localSha  = git rev-parse HEAD
$remoteSha = git rev-parse origin/develop 2>$null
if ($localSha -ne $remoteSha) {
    Log-Warn "本地代码落后于远端，自动拉取中..."
    git pull origin develop --quiet
    Log-Fix "代码已同步到最新 ($(git rev-parse --short HEAD))"
} else {
    Log-Pass "本地代码已是最新 ($(git rev-parse --short HEAD))"
}

# E08: Skills 完整性
Write-Host "[E08] 检测 Skills 文件..." -ForegroundColor Blue
$requiredSkills = @("env-setup.md","requirements-engineer.md","designer.md","developer.md","tester.md","qa-engineer.md","devops-engineer.md")
$missingSkills  = $requiredSkills | Where-Object { -not (Test-Path ".agents/skills/$_") }
if ($missingSkills.Count -gt 0) {
    Log-Warn "发现 $($missingSkills.Count) 个缺失的 Skill 文件: $($missingSkills -join ', ')"
    Log-Fix "重新拉取最新代码..."
    git pull origin develop --quiet
    Log-Pass "Skills 文件已恢复"
} else {
    Log-Pass "所有 Skills 文件完整 ($($requiredSkills.Count)/$($requiredSkills.Count))"
}

# E09: 文档目录结构
Write-Host "[E09] 检测文档目录结构..." -ForegroundColor Blue
$requiredDirs = @("docs/requirements","docs/design","docs/testing","docs/qa","docs/deploy","docs/playbook")
$createdDirs  = 0
foreach ($dir in $requiredDirs) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
        New-Item -ItemType File -Force -Path "$dir/.gitkeep" | Out-Null
        $createdDirs++
    }
}
if ($createdDirs -gt 0) { Log-Fix "已自动创建 $createdDirs 个文档目录" }
Log-Pass "文档目录结构完整"

# ================================================================
# 二、Go 开发环境 G01~G07
# ================================================================
Log-Title "Go 开发环境检查 (G01~G07)"

# G01: Go 安装与版本
Write-Host "[G01] 检测 Go 安装与版本..." -ForegroundColor Blue
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Log-Warn "Go 未安装，自动安装中..."
    if (Get-Command winget -ErrorAction SilentlyContinue) {
        Log-Fix "执行: winget install GoLang.Go"
        winget install GoLang.Go --silent
        Refresh-EnvPath
    }
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Log-Fail "Go 安装失败，请手动安装: https://golang.org/dl/"
        Log-Info "安装后请重启 PowerShell 并重新运行此脚本"
    } else {
        Log-Pass "Go $(go version) 安装成功"
    }
} else {
    $goVer = go version
    # 解析版本号
    if ($goVer -match 'go(\d+)\.(\d+)') {
        $major = [int]$Matches[1]; $minor = [int]$Matches[2]
        if ($major -gt 1 -or ($major -eq 1 -and $minor -ge 21)) {
            Log-Pass "$goVer 已就绪"
        } else {
            Log-Warn "Go 版本过低（需要 1.21+），当前: $goVer"
            Log-Fix "执行: winget upgrade GoLang.Go"
            winget upgrade GoLang.Go --silent
            Refresh-EnvPath
            Log-Pass "Go 已升级，请重启 PowerShell 后重新运行"
        }
    } else {
        Log-Pass "$goVer 已就绪"
    }
}

# G02: GOPATH / GOPROXY
Write-Host "[G02] 检测 Go 环境变量..." -ForegroundColor Blue
if (Get-Command go -ErrorAction SilentlyContinue) {
    $goRoot  = go env GOROOT
    $goPath  = go env GOPATH
    $goProxy = go env GOPROXY
    Log-Pass "GOROOT: $goRoot"
    Log-Pass "GOPATH: $goPath"
    # 设置国内加速代理
    if ($goProxy -eq "direct" -or $goProxy -eq "" -or $goProxy -notmatch "goproxy.cn") {
        Log-Warn "GOPROXY 未使用国内加速，自动设置..."
        go env -w GOPROXY=https://goproxy.cn,direct
        go env -w GONOSUMCHECK=*
        Log-Fix "GOPROXY 已设置为: https://goproxy.cn,direct"
    } else {
        Log-Pass "GOPROXY: $goProxy"
    }
    # 确保 GOPATH/bin 在 PATH 中
    $goBin = "$(go env GOPATH)\bin"
    if ($env:Path -notlike "*$goBin*") {
        Log-Warn "GOPATH/bin 不在 PATH 中，自动添加..."
        [System.Environment]::SetEnvironmentVariable("Path", $env:Path + ";$goBin", "User")
        $env:Path += ";$goBin"
        Log-Fix "已将 $goBin 添加到 PATH"
    } else {
        Log-Pass "GOPATH/bin 已在 PATH 中"
    }
} else {
    Log-Warn "Go 未就绪，跳过 G02 检查"
}

# G03: go.mod 初始化
Write-Host "[G03] 检测 Go 模块..." -ForegroundColor Blue
if (Get-Command go -ErrorAction SilentlyContinue) {
    if (-not (Test-Path "go.mod")) {
        Log-Warn "go.mod 不存在，自动初始化..."
        go mod init github.com/zouluxing/feature_orbit
        Log-Fix "go.mod 已初始化: github.com/zouluxing/feature_orbit"
    } else {
        $modName = (Get-Content go.mod | Select-String "^module").ToString().Split(" ")[1]
        Log-Pass "Go 模块: $modName"
    }
}

# G04: Go 依赖下载
Write-Host "[G04] 检测 Go 项目依赖..." -ForegroundColor Blue
if ((Get-Command go -ErrorAction SilentlyContinue) -and (Test-Path "go.mod")) {
    Log-Fix "执行 go mod tidy & download..."
    go mod download 2>$null
    go mod tidy 2>$null
    Log-Pass "Go 依赖已就绪"
} else {
    Log-Pass "无 go.mod，跳过依赖检查（初始化阶段正常）"
}

# G05: 代码规范工具
Write-Host "[G05] 检测 Go 代码规范工具..." -ForegroundColor Blue
if (Get-Command go -ErrorAction SilentlyContinue) {
    # gofmt 内置
    Log-Pass "gofmt 已就绪（Go 内置）"
    # golangci-lint
    if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
        Log-Fix "安装 golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        Refresh-EnvPath
        if (Get-Command golangci-lint -ErrorAction SilentlyContinue) { Log-Pass "golangci-lint 安装成功" }
        else { Log-Warn "golangci-lint 安装后需重启 PowerShell" }
    } else {
        Log-Pass "golangci-lint $(golangci-lint --version 2>$null | Select-String 'version' | ForEach-Object { $_.ToString().Trim() }) 已就绪"
    }
    # goimports
    if (-not (Get-Command goimports -ErrorAction SilentlyContinue)) {
        Log-Fix "安装 goimports..."
        go install golang.org/x/tools/cmd/goimports@latest
        Refresh-EnvPath
        Log-Pass "goimports 安装成功"
    } else {
        Log-Pass "goimports 已就绪"
    }
}

# G06: go test 验证
Write-Host "[G06] 验证 go test 工具..." -ForegroundColor Blue
if (Get-Command go -ErrorAction SilentlyContinue) {
    go help test 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) { Log-Pass "go test 可用" }
    else { Log-Fail "go test 异常，请检查 Go 安装" }
}

# G07: air 热重载
Write-Host "[G07] 检测 air 热重载工具..." -ForegroundColor Blue
if (Get-Command go -ErrorAction SilentlyContinue) {
    if (-not (Get-Command air -ErrorAction SilentlyContinue)) {
        Log-Fix "安装 air 热重载工具..."
        go install github.com/cosmtrek/air@latest
        Refresh-EnvPath
        if (Get-Command air -ErrorAction SilentlyContinue) { Log-Pass "air 安装成功" }
        else { Log-Warn "air 安装后需重启 PowerShell" }
    } else {
        Log-Pass "air 热重载工具已就绪"
    }
    # 生成默认配置文件
    if (-not (Test-Path ".air.toml")) {
        Log-Fix "生成 .air.toml 默认配置..."
        if (Get-Command air -ErrorAction SilentlyContinue) {
            air init 2>$null
            Log-Pass ".air.toml 已生成"
        }
    } else {
        Log-Pass ".air.toml 配置文件已存在"
    }
}

# ================================================================
# 三、SaaS 基础服务检查 S01~S07
# ================================================================
Log-Title "SaaS 基础服务检查 (S01~S07)"

# S01: Docker Desktop
Write-Host "[S01] 检测 Docker Desktop..." -ForegroundColor Blue
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Log-Warn "Docker 未安装"
    Log-Fix "推荐安装: winget install Docker.DockerDesktop"
    Log-Info "或访问: https://www.docker.com/products/docker-desktop"
    Log-Fail "Docker Desktop 未安装（SaaS 开发必需）"
} else {
    try {
        docker info 2>&1 | Out-Null
        if ($LASTEXITCODE -eq 0) { Log-Pass "Docker Desktop 已运行: $(docker --version)" }
        else {
            Log-Fail "Docker 已安装但未运行，请启动 Docker Desktop 后重试"
        }
    } catch {
        Log-Fail "Docker 服务未运行，请手动启动 Docker Desktop"
    }
}

# S02: Docker Compose
Write-Host "[S02] 检测 Docker Compose..." -ForegroundColor Blue
if (Get-Command docker -ErrorAction SilentlyContinue) {
    try {
        $composeVer = docker compose version 2>$null
        if ($LASTEXITCODE -eq 0) { Log-Pass "$composeVer 已就绪" }
        else { Log-Fail "Docker Compose 不可用，请更新 Docker Desktop" }
    } catch {
        Log-Fail "Docker Compose 检测失败"
    }
} else {
    Log-Warn "Docker 未就绪，跳过 S02 检查"
}

# S03~S04: PostgreSQL + Redis 容器
Write-Host "[S03] 检测 PostgreSQL 服务..." -ForegroundColor Blue
if ((Get-Command docker -ErrorAction SilentlyContinue) -and (Test-Path "docker-compose.yml")) {
    $running = docker compose ps --services --filter "status=running" 2>$null
    if ($running -notcontains "postgres") {
        Log-Fix "启动 PostgreSQL 容器..."
        docker compose up -d postgres 2>$null
        Start-Sleep -Seconds 5
        Log-Pass "PostgreSQL 容器已启动"
    } else {
        Log-Pass "PostgreSQL 容器正在运行"
    }

    Write-Host "[S04] 检测 Redis 服务..." -ForegroundColor Blue
    if ($running -notcontains "redis") {
        Log-Fix "启动 Redis 容器..."
        docker compose up -d redis 2>$null
        Log-Pass "Redis 容器已启动"
    } else {
        Log-Pass "Redis 容器正在运行"
    }
} else {
    Log-Pass "[S03] docker-compose.yml 不存在，跳过容器检查（初始化阶段正常）"
    Log-Pass "[S04] 同上"
}

# S05: .env 文件
Write-Host "[S05] 检测环境变量文件..." -ForegroundColor Blue
if (-not (Test-Path ".env")) {
    if (Test-Path ".env.example") {
        Log-Fix "从 .env.example 生成 .env 文件..."
        Copy-Item ".env.example" ".env"
        Log-Pass ".env 文件已生成，请填写关键配置：DB_PASSWORD、JWT_SECRET 等"
    } else {
        Log-Pass ".env 文件不存在（初始化阶段正常，后续需创建）"
    }
} else {
    $envContent  = Get-Content ".env" -ErrorAction SilentlyContinue
    $requiredVars = @("DB_HOST","DB_PORT","DB_NAME","DB_USER","DB_PASSWORD","JWT_SECRET","REDIS_ADDR")
    $missingVars  = $requiredVars | Where-Object { -not ($envContent | Select-String "^$_=.+") }
    if ($missingVars.Count -gt 0) {
        Log-Warn ".env 缺少以下关键变量: $($missingVars -join ', ')"
    } else {
        Log-Pass ".env 环境变量文件完整"
    }
}

# S06: 数据库连通性
Write-Host "[S06] 检测数据库连通性..." -ForegroundColor Blue
if ((Get-Command docker -ErrorAction SilentlyContinue) -and (Test-Path ".env")) {
    $retries = 0; $connected = $false
    while ($retries -lt 5 -and -not $connected) {
        $result = docker exec feature_orbit_postgres pg_isready -U postgres 2>$null
        if ($LASTEXITCODE -eq 0) { $connected = $true }
        else { $retries++; Write-Host "  ⏳ 等待数据库就绪... ($retries/5)"; Start-Sleep -Seconds 3 }
    }
    if ($connected) { Log-Pass "PostgreSQL 连接正常" }
    else { Log-Warn "数据库连接超时，请检查 Docker 容器状态" }
} else {
    Log-Pass "跳过数据库连通性检查（初始化阶段正常）"
}

# S07: golang-migrate 迁移工具
Write-Host "[S07] 检测数据库迁移工具..." -ForegroundColor Blue
if (Get-Command go -ErrorAction SilentlyContinue) {
    if (-not (Get-Command migrate -ErrorAction SilentlyContinue)) {
        Log-Fix "安装 golang-migrate..."
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        Refresh-EnvPath
        if (Get-Command migrate -ErrorAction SilentlyContinue) { Log-Pass "golang-migrate 安装成功" }
        else { Log-Warn "migrate 安装后需重启 PowerShell" }
    } else {
        Log-Pass "golang-migrate 工具已就绪"
    }
    # 执行迁移
    if ((Test-Path "migrations") -and (Get-Command migrate -ErrorAction SilentlyContinue) -and (Test-Path ".env")) {
        $dbUrl = (Get-Content ".env" | Select-String "^DATABASE_URL=(.*)").Matches.Groups[1].Value
        if ($dbUrl) {
            Log-Fix "执行数据库迁移..."
            migrate -path migrations -database $dbUrl up 2>$null
            Log-Pass "数据库迁移完成"
        } else {
            Log-Warn "DATABASE_URL 未配置，跳过自动迁移"
        }
    }
}

# ================================================================
# 四、Windows 专项检查 W01~W05
# ================================================================
Log-Title "Windows 专项检查 (W01~W05)"

# W01: PowerShell 版本
Write-Host "[W01] 检测 PowerShell 版本..." -ForegroundColor Blue
$psVer = $PSVersionTable.PSVersion
if ($psVer.Major -ge 7) {
    Log-Pass "PowerShell $psVer (Core) 已就绪"
} elseif ($psVer.Major -ge 5) {
    Log-Pass "PowerShell $psVer 已就绪（推荐升级到 PS7: winget install Microsoft.PowerShell）"
} else {
    Log-Fail "PowerShell 版本过低，请升级到 5.1+"
}

# W02: Windows Terminal
Write-Host "[W02] 检测 Windows Terminal..." -ForegroundColor Blue
if (Get-Command wt -ErrorAction SilentlyContinue) {
    Log-Pass "Windows Terminal 已安装"
} else {
    Log-Warn "推荐安装 Windows Terminal: winget install Microsoft.WindowsTerminal"
}

# W03: make 工具
Write-Host "[W03] 检测 make 工具..." -ForegroundColor Blue
if (-not (Get-Command make -ErrorAction SilentlyContinue)) {
    Log-Warn "make 未安装，自动安装中..."
    if (Get-Command scoop -ErrorAction SilentlyContinue) {
        Log-Fix "通过 Scoop 安装 make..."; scoop install make; Refresh-EnvPath
    } elseif (Get-Command choco -ErrorAction SilentlyContinue) {
        Log-Fix "通过 Chocolatey 安装 make..."; choco install make -y; Refresh-EnvPath
    } elseif (Get-Command winget -ErrorAction SilentlyContinue) {
        Log-Fix "通过 winget 安装 make..."; winget install GnuWin32.Make --silent; Refresh-EnvPath
    } else {
        Log-Warn "无法自动安装 make，推荐先安装 Scoop: irm get.scoop.sh | iex"
    }
    if (Get-Command make -ErrorAction SilentlyContinue) { Log-Pass "make 已就绪" }
    else { Log-Warn "make 安装后可能需重启 PowerShell" }
} else {
    Log-Pass "make 已就绪"
}

# W04: Git 行尾符配置
Write-Host "[W04] 检测 Git 行尾符配置..." -ForegroundColor Blue
$autoCrlf = git config --global core.autocrlf 2>$null
if ($autoCrlf -ne "true") {
    Log-Fix "设置 Git core.autocrlf = true（Windows 标准）..."
    git config --global core.autocrlf true
    git config --global core.safecrlf warn
    Log-Pass "Git 行尾符已配置 (autocrlf=true, safecrlf=warn)"
} else {
    Log-Pass "Git 行尾符配置正确 (autocrlf=true)"
}

# W05: 关键端口检测
Write-Host "[W05] 检测关键端口可用性..." -ForegroundColor Blue
$ports = @{ 8080="Go 应用服务"; 5432="PostgreSQL"; 6379="Redis"; 9090="Prometheus(可选)" }
$portConflict = $false
foreach ($port in $ports.Keys) {
    $inUse = netstat -ano 2>$null | Select-String ":$port\s" | Where-Object { $_ -match "LISTENING" }
    if ($inUse) {
        Log-Warn "端口 $port ($($ports[$port])) 已被占用，可能影响服务启动"
        $portConflict = $true
    } else {
        Log-Pass "端口 $port ($($ports[$port])) 可用"
    }
}
if ($portConflict) { Log-Warn "部分端口冲突，请检查占用进程（netstat -ano | findstr :端口号）" }

# ================================================================
# 输出汇总报告
# ================================================================
Write-Host ""
Write-Host "================================================" -ForegroundColor Blue
Write-Host "              环境检查汇总报告" -ForegroundColor Blue
Write-Host "================================================" -ForegroundColor Blue
Write-Host "  通过项:   $CHECK_PASS" -ForegroundColor Green
Write-Host "  失败项:   $CHECK_FAIL" -ForegroundColor Red
Write-Host "  自动修复: $AUTO_FIXED" -ForegroundColor Yellow
Write-Host "------------------------------------------------" -ForegroundColor Blue
if (Get-Command git    -ErrorAction SilentlyContinue) { Write-Host "  Git:         $(git --version)" }
if (Get-Command node   -ErrorAction SilentlyContinue) { Write-Host "  Node.js:     $(node -v)" }
if (Get-Command go     -ErrorAction SilentlyContinue) { Write-Host "  Go:          $(go version)" }
if (Get-Command docker -ErrorAction SilentlyContinue) { Write-Host "  Docker:      $(docker --version)" }
Write-Host "  PowerShell:  $($PSVersionTable.PSVersion)"
try { Write-Host "  分支:        $(git branch --show-current)" } catch {}
try { Write-Host "  代码版本:    $(git rev-parse --short HEAD)" } catch {}
Write-Host "================================================" -ForegroundColor Blue

if ($CHECK_FAIL -eq 0) {
    Write-Host ""
    Write-Host "  🚀 环境准备就绪！可以开始工作了。" -ForegroundColor Green
    Write-Host ""
    Write-Host "  💡 启动开发服务器："
    Write-Host "     air                    # Go 热重载启动"
    Write-Host "     docker compose up -d  # 启动全部服务"
    Write-Host ""
    Write-Host "  📋 可用的工作阶段："
    Write-Host "     1. 需求分析  →  说: '开始需求分析'"
    Write-Host "     2. 系统设计  →  说: '开始系统设计'"
    Write-Host "     3. 开发实现  →  说: '开始开发'"
    Write-Host "     4. 测试验证  →  说: '开始测试'"
    Write-Host "     5. 质量保障  →  说: '开始QA验收'"
    Write-Host "     6. 部署上线  →  说: '开始部署'"
    Write-Host "================================================" -ForegroundColor Blue
    exit 0
} else {
    Write-Host ""
    Write-Host "  ⚠️  有 $CHECK_FAIL 个检查项未通过，请根据提示修复后重新运行。" -ForegroundColor Red
    Write-Host "  重新运行: PowerShell -ExecutionPolicy Bypass -File scripts/setup-env.ps1" -ForegroundColor Yellow
    exit 1
}
