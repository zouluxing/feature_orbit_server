# Skill: 环境自动检测与准备 (Environment Setup)

## SKILL_DESCRIPTION
在任何阶段任务开始前自动触发，检测开发环境是否就绪，若未就绪则自动修复，直到环境完全准备好后才允许进入后续阶段。
触发关键词：环境检查、环境准备、开始工作、启动项目、初始化环境、env check、setup

> ⚠️ 此 Skill 为所有阶段的强制前置 Skill，每次开始新阶段前必须先执行并通过所有检查项。

---

## 运行平台

> 本项目主要在 **Windows** 环境下开发，使用 **Go 语言** 构建 SaaS 系统。
> 推荐使用 PowerShell 执行启动脚本：
> ```powershell
> PowerShell -ExecutionPolicy Bypass -File scripts/setup-env.ps1
> ```

---

## 检查项总览

### 基础环境（所有平台）

| 编号 | 检查项 | 要求 | 自动修复 |
|------|--------|------|----------|
| E01 | Git 安装 | Git 2.x+ | ✅ 引导安装 |
| E02 | Git 用户配置 | user.name / user.email | ✅ 交互配置 |
| E03 | Node.js 版本 | v18+（Claude Code 依赖） | ✅ 引导安装 |
| E04 | Claude Code 安装 | 最新版 | ✅ 自动安装 |
| E05 | 仓库克隆状态 | 已克隆到本地 | ✅ 自动克隆 |
| E06 | 当前分支 | 必须在 develop | ✅ 自动切换 |
| E07 | 本地代码同步 | 与远端 develop 一致 | ✅ 自动拉取 |
| E08 | .agents/skills 目录 | Skills 文件完整 | ✅ 自动拉取 |
| E09 | docs 目录结构 | 文档目录完整 | ✅ 自动创建 |

### Go 开发环境

| 编号 | 检查项 | 要求 | 自动修复 |
|------|--------|------|----------|
| G01 | Go 安装与版本 | Go 1.21+ | ✅ 引导安装 |
| G02 | GOPATH / GOROOT | 环境变量配置正确 | ✅ 自动检测报告 |
| G03 | Go 模块初始化 | go.mod 存在 | ✅ 自动初始化 |
| G04 | Go 项目依赖 | go mod download | ✅ 自动下载 |
| G05 | Go 代码规范工具 | gofmt / golint / golangci-lint | ✅ 自动安装 |
| G06 | Go 测试工具 | go test 可用 | ✅ 自动验证 |
| G07 | 热重载工具 | air（开发期热重载） | ✅ 自动安装 |

### SaaS 基础服务环境

| 编号 | 检查项 | 要求 | 自动修复 |
|------|--------|------|----------|
| S01 | Docker Desktop | 已安装并运行 | ✅ 引导安装 |
| S02 | Docker Compose | v2.x+ | ✅ 引导安装 |
| S03 | 数据库服务 | PostgreSQL（Docker 容器） | ✅ 自动启动 |
| S04 | 缓存服务 | Redis（Docker 容器） | ✅ 自动启动 |
| S05 | 环境变量文件 | .env 文件存在且关键变量完整 | ✅ 从模板生成 |
| S06 | 数据库连通性 | 能够连接到 PostgreSQL | ✅ 自动重试 |
| S07 | 数据库迁移 | migrate 工具 + 迁移脚本 | ✅ 自动执行 |

### Windows 专项检查

| 编号 | 检查项 | 要求 | 自动修复 |
|------|--------|------|----------|
| W01 | PowerShell 版本 | PowerShell 5.1+ 或 PowerShell Core 7+ | ✅ 引导升级 |
| W02 | Windows Terminal | 推荐安装 | ⚠️ 提示安装 |
| W03 | Make 工具 | make 命令可用（通过 Scoop/Choco） | ✅ 自动安装 |
| W04 | 行尾符配置 | Git autocrlf 设置正确 | ✅ 自动配置 |
| W05 | 防火墙/端口 | 8080、5432、6379 端口可用 | ✅ 检测报告 |

---

## SKILL_STEPS

### Step 1 — 基础环境检查（E01~E09）

```powershell
# [E01] 检测 Git
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Host "❌ [E01] Git 未安装"
    Write-Host "👉 请访问 https://git-scm.com/download/win 下载安装"
    Write-Host "   或使用 winget: winget install --id Git.Git -e --source winget"
    exit 1
} else {
    Write-Host "✅ [E01] $(git --version) 已就绪"
}

# [E02] Git 用户配置
$gitName = git config --global user.name
$gitEmail = git config --global user.email
if (-not $gitName) {
    $gitName = Read-Host "⚠️  [E02] 请输入 Git 用户名"
    git config --global user.name $gitName
}
if (-not $gitEmail) {
    $gitEmail = Read-Host "⚠️  [E02] 请输入 Git 邮箱"
    git config --global user.email $gitEmail
}
Write-Host "✅ [E02] Git 用户: $gitName <$gitEmail>"

# [E03] 检测 Node.js（Claude Code 依赖）
try {
    $nodeVer = (node -v).TrimStart('v').Split('.')[0]
    if ([int]$nodeVer -ge 18) {
        Write-Host "✅ [E03] Node.js $(node -v) 已就绪"
    } else {
        Write-Host "❌ [E03] Node.js 版本过低，请升级到 v18+"
        Start-Process "https://nodejs.org"
        exit 1
    }
} catch {
    Write-Host "❌ [E03] Node.js 未安装，请访问 https://nodejs.org 安装"
    exit 1
}

# [E04] 检测 Claude Code
if (-not (Get-Command claude -ErrorAction SilentlyContinue)) {
    Write-Host "🔧 [E04] 自动安装 Claude Code..."
    npm install -g @anthropic-ai/claude-code
    Write-Host "✅ [E04] Claude Code 安装成功"
} else {
    Write-Host "✅ [E04] Claude Code 已就绪"
}

# [E05] 检测仓库克隆状态
if (-not (git rev-parse --git-dir 2>$null)) {
    Write-Host "🔧 [E05] 自动克隆仓库..."
    Set-Location ..
    git clone https://github.com/zouluxing/feature_orbit_server.git
    Set-Location feature_orbit_server
    Write-Host "✅ [E05] 仓库克隆成功"
} else {
    Write-Host "✅ [E05] Git 仓库已就绪: $(git remote get-url origin)"
}

# [E06~E09] 分支、同步、Skills、文档目录（见 setup-env.ps1 完整脚本）
```

---

### Step 2 — Go 开发环境检查（G01~G07）

```powershell
# [G01] 检测 Go 安装与版本
Write-Host "[G01] 检测 Go 环境..."
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ [G01] Go 未安装"
    Write-Host "👉 自动修复选项："
    Write-Host "   方式一：winget install GoLang.Go"
    Write-Host "   方式二：访问 https://golang.org/dl/ 下载安装"
    if (Get-Command winget -ErrorAction SilentlyContinue) {
        Write-Host "🔧 [G01] 尝试通过 winget 安装 Go..."
        winget install GoLang.Go --silent
        $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
    }
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Host "❌ [G01] Go 安装失败，请手动安装后重新运行"
        exit 1
    }
}
$goVersion = go version
$goVer = ($goVersion -match 'go(\d+)\.(\d+)') | Out-Null; $goMajor = [int]$Matches[1]; $goMinor = [int]$Matches[2]
if ($goMajor -lt 1 -or ($goMajor -eq 1 -and $goMinor -lt 21)) {
    Write-Host "❌ [G01] Go 版本过低（需要 1.21+），当前: $goVersion"
    Write-Host "👉 请更新: winget upgrade GoLang.Go"
    exit 1
}
Write-Host "✅ [G01] $goVersion 已就绪"

# [G02] 检测 GOPATH / GOROOT
Write-Host "[G02] 检测 Go 环境变量..."
$goRoot = go env GOROOT
$goPath = go env GOPATH
$goProxy = go env GOPROXY
Write-Host "✅ [G02] GOROOT: $goRoot"
Write-Host "✅ [G02] GOPATH: $goPath"
if ($goProxy -eq "direct" -or $goProxy -eq "") {
    Write-Host "⚠️  [G02] GOPROXY 未配置，设置国内加速镜像..."
    go env -w GOPROXY=https://goproxy.cn,direct
    go env -w GONOSUMCHECK=*
    Write-Host "✅ [G02] GOPROXY 已设置为: https://goproxy.cn,direct"
} else {
    Write-Host "✅ [G02] GOPROXY: $goProxy"
}

# [G03] 检测 go.mod
Write-Host "[G03] 检测 Go 模块..."
if (-not (Test-Path "go.mod")) {
    Write-Host "⚠️  [G03] go.mod 不存在，自动初始化..."
    go mod init github.com/zouluxing/feature_orbit_server
    Write-Host "✅ [G03] go.mod 已初始化: github.com/zouluxing/feature_orbit_server"
} else {
    $modName = (Get-Content go.mod | Select-String "^module").ToString().Split(" ")[1]
    Write-Host "✅ [G03] Go 模块: $modName"
}

# [G04] 下载 Go 依赖
Write-Host "[G04] 检测 Go 项目依赖..."
go mod download
go mod tidy
Write-Host "✅ [G04] Go 依赖已就绪"

# [G05] 安装代码规范工具
Write-Host "[G05] 检测 Go 代码规范工具..."
Write-Host "✅ [G05] gofmt 已就绪（Go 内置）"
if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
    Write-Host "🔧 [G05] 安装 golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    Write-Host "✅ [G05] golangci-lint 安装成功"
} else {
    Write-Host "✅ [G05] golangci-lint 已就绪"
}

# [G06] 验证 go test
Write-Host "[G06] 验证 Go 测试工具..."
$testResult = go help test 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ [G06] go test 可用"
} else {
    Write-Host "❌ [G06] go test 异常，请检查 Go 安装"
    exit 1
}

# [G07] 安装热重载工具 air
Write-Host "[G07] 检测热重载工具 air..."
if (-not (Get-Command air -ErrorAction SilentlyContinue)) {
    Write-Host "🔧 [G07] 安装 air 热重载工具..."
    go install github.com/cosmtrek/air@latest
    Write-Host "✅ [G07] air 安装成功"
} else {
    Write-Host "✅ [G07] air 热重载工具已就绪"
}
if (-not (Test-Path ".air.toml")) {
    air init
    Write-Host "✅ [G07] .air.toml 配置文件已生成"
}
```

---

### Step 3 — SaaS 基础服务检查（S01~S07）

```powershell
# [S01] 检测 Docker Desktop
Write-Host "[S01] 检测 Docker Desktop..."
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "❌ [S01] Docker 未安装"
    Write-Host "👉 请访问 https://www.docker.com/products/docker-desktop 安装 Docker Desktop"
    Write-Host "   或使用 winget: winget install Docker.DockerDesktop"
    exit 1
}
try {
    docker info | Out-Null
    Write-Host "✅ [S01] Docker Desktop 已运行: $(docker --version)"
} catch {
    Write-Host "❌ [S01] Docker 已安装但未运行，请启动 Docker Desktop"
    exit 1
}

# [S02] 检测 Docker Compose
Write-Host "[S02] 检测 Docker Compose..."
try {
    $composeVer = docker compose version
    Write-Host "✅ [S02] $composeVer 已就绪"
} catch {
    Write-Host "❌ [S02] Docker Compose 不可用，请更新 Docker Desktop"
    exit 1
}

# [S03~S04] 启动基础服务（PostgreSQL + Redis）
Write-Host "[S03] 检测 PostgreSQL 服务..."
if (Test-Path "docker-compose.yml") {
    $pgRunning = docker compose ps --services --filter "status=running" 2>$null
    if ($pgRunning -notcontains "postgres") {
        Write-Host "🔧 [S03] 启动 PostgreSQL 容器..."
        docker compose up -d postgres
        Start-Sleep -Seconds 5
        Write-Host "✅ [S03] PostgreSQL 容器已启动"
    } else {
        Write-Host "✅ [S03] PostgreSQL 容器正在运行"
    }
    Write-Host "[S04] 检测 Redis 服务..."
    if ($pgRunning -notcontains "redis") {
        Write-Host "🔧 [S04] 启动 Redis 容器..."
        docker compose up -d redis
        Write-Host "✅ [S04] Redis 容器已启动"
    } else {
        Write-Host "✅ [S04] Redis 容器正在运行"
    }
} else {
    Write-Host "⚠️  [S03] docker-compose.yml 不存在，跳过服务检查（初始化阶段正常）"
}

# [S05] 检测 .env 文件
Write-Host "[S05] 检测环境变量文件..."
if (-not (Test-Path ".env")) {
    if (Test-Path ".env.example") {
        Write-Host "🔧 [S05] 从 .env.example 生成 .env 文件..."
        Copy-Item ".env.example" ".env"
        Write-Host "✅ [S05] .env 文件已生成，请检查并填写关键配置项"
        Write-Host "   关键配置项：DB_HOST, DB_PASSWORD, JWT_SECRET, REDIS_ADDR"
    } else {
        Write-Host "⚠️  [S05] .env 和 .env.example 均不存在（初始化阶段正常）"
    }
} else {
    $envContent = Get-Content ".env"
    $requiredVars = @("DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "JWT_SECRET", "REDIS_ADDR")
    $missingVars = @()
    foreach ($var in $requiredVars) {
        if (-not ($envContent | Select-String "^$var=")) {
            $missingVars += $var
        }
    }
    if ($missingVars.Count -gt 0) {
        Write-Host "⚠️  [S05] .env 缺少以下关键变量: $($missingVars -join ', ')"
    } else {
        Write-Host "✅ [S05] .env 环境变量文件完整"
    }
}

# [S06] 检测数据库连通性
Write-Host "[S06] 检测数据库连通性..."
if ((Test-Path ".env") -and (Get-Command docker -ErrorAction SilentlyContinue)) {
    $retries = 0; $maxRetries = 5; $connected = $false
    while ($retries -lt $maxRetries -and -not $connected) {
        $result = docker exec feature_orbit_server_postgres pg_isready -U postgres 2>$null
        if ($LASTEXITCODE -eq 0) {
            $connected = $true
        } else {
            $retries++
            Write-Host "⚠️  [S06] 数据库连接中... ($retries/$maxRetries)"
            Start-Sleep -Seconds 3
        }
    }
    if ($connected) {
        Write-Host "✅ [S06] PostgreSQL 连接正常"
    } else {
        Write-Host "⚠️  [S06] 数据库连接超时，请检查 Docker 容器状态"
    }
} else {
    Write-Host "⚠️  [S06] 跳过数据库连通性检查（初始化阶段正常）"
}

# [S07] 检测数据库迁移工具
Write-Host "[S07] 检测数据库迁移工具..."
if (-not (Get-Command migrate -ErrorAction SilentlyContinue)) {
    Write-Host "🔧 [S07] 安装 golang-migrate 工具..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    if (Get-Command migrate -ErrorAction SilentlyContinue) {
        Write-Host "✅ [S07] migrate 工具安装成功"
    } else {
        Write-Host "⚠️  [S07] migrate 安装后需重启 PowerShell，请重新运行脚本"
    }
} else {
    Write-Host "✅ [S07] golang-migrate 工具已就绪"
}
if ((Test-Path "migrations") -and (Get-Command migrate -ErrorAction SilentlyContinue)) {
    $dbUrl = (Get-Content ".env" | Select-String "^DATABASE_URL=(.*)").Matches.Groups[1].Value
    if ($dbUrl) {
        migrate -path migrations -database $dbUrl up
        Write-Host "✅ [S07] 数据库迁移完成"
    } else {
        Write-Host "⚠️  [S07] DATABASE_URL 未配置，跳过迁移"
    }
}
```

---

### Step 4 — Windows 专项检查（W01~W05）

```powershell
# [W01] 检测 PowerShell 版本
$psVer = $PSVersionTable.PSVersion
if ($psVer.Major -ge 7) {
    Write-Host "✅ [W01] PowerShell $psVer (Core) 已就绪"
} elseif ($psVer.Major -ge 5) {
    Write-Host "✅ [W01] PowerShell $psVer 已就绪（推荐升级到 PS7: winget install Microsoft.PowerShell）"
} else {
    Write-Host "❌ [W01] PowerShell 版本过低，请升级"
    exit 1
}

# [W02] 检测 Windows Terminal（推荐）
if (Get-Command wt -ErrorAction SilentlyContinue) {
    Write-Host "✅ [W02] Windows Terminal 已安装"
} else {
    Write-Host "⚠️  [W02] 建议安装 Windows Terminal: winget install Microsoft.WindowsTerminal"
}

# [W03] 检测 make 工具
if (-not (Get-Command make -ErrorAction SilentlyContinue)) {
    if (Get-Command scoop -ErrorAction SilentlyContinue) {
        scoop install make
    } elseif (Get-Command choco -ErrorAction SilentlyContinue) {
        choco install make -y
    } elseif (Get-Command winget -ErrorAction SilentlyContinue) {
        winget install GnuWin32.Make
    } else {
        Write-Host "⚠️  [W03] 无法自动安装 make，推荐安装 Scoop: irm get.scoop.sh | iex"
    }
    Write-Host "✅ [W03] make 已就绪"
} else {
    Write-Host "✅ [W03] make 已就绪"
}

# [W04] Git 行尾符配置
$autoCrlf = git config --global core.autocrlf
if ($autoCrlf -ne "true") {
    git config --global core.autocrlf true
    git config --global core.safecrlf warn
    Write-Host "✅ [W04] Git 行尾符已配置 (autocrlf=true)"
} else {
    Write-Host "✅ [W04] Git 行尾符配置正确 (autocrlf=true)"
}

# [W05] 检测关键端口占用
$ports = @{ 8080 = "Go 应用服务"; 5432 = "PostgreSQL"; 6379 = "Redis"; 9090 = "Prometheus（可选）" }
$portConflict = $false
foreach ($port in $ports.Keys) {
    $inUse = netstat -ano | Select-String ":$port " | Where-Object { $_ -match "LISTENING" }
    if ($inUse) {
        Write-Host "⚠️  [W05] 端口 $port ($($ports[$port])) 已被占用"
        $portConflict = $true
    } else {
        Write-Host "✅ [W05] 端口 $port ($($ports[$port])) 可用"
    }
}
if ($portConflict) { Write-Host "⚠️  [W05] 部分端口被占用，请检查" }
```

---

### Step 5 — 输出完整环境报告

```powershell
Write-Host ""
Write-Host "================================================"
Write-Host "           环境检查完整报告"
Write-Host "================================================"
Write-Host "  [基础环境]"
Write-Host "  Git:             $(git --version)"
Write-Host "  Node.js:         $(node -v)"
Write-Host "  Claude Code:     已就绪"
Write-Host "  仓库:            feature_orbit_server"
Write-Host "  分支:            $(git branch --show-current)"
Write-Host "  代码版本:        $(git rev-parse --short HEAD)"
Write-Host ""
Write-Host "  [Go 开发环境]"
Write-Host "  Go:              $(go version)"
Write-Host "  模块名:          github.com/zouluxing/feature_orbit_server"
Write-Host "  GOPATH:          $(go env GOPATH)"
Write-Host "  GOPROXY:         $(go env GOPROXY)"
Write-Host "  golangci-lint:   已就绪"
Write-Host "  air（热重载）:   已就绪"
Write-Host "  migrate:         已就绪"
Write-Host ""
Write-Host "  [SaaS 服务]"
Write-Host "  Docker:          $(docker --version)"
Write-Host "  PostgreSQL:      容器运行中 (feature_orbit_server_postgres)"
Write-Host "  Redis:           容器运行中"
Write-Host "  .env 配置:       完整"
Write-Host ""
Write-Host "  [Windows 环境]"
Write-Host "  PowerShell:      $($PSVersionTable.PSVersion)"
Write-Host "  make:            已就绪"
Write-Host "  Git autocrlf:    true"
Write-Host "  关键端口:        可用"
Write-Host "================================================"
Write-Host ""
Write-Host "  🚀 环境准备就绪！可以开始工作了。"
Write-Host ""
Write-Host "  📋 可用的工作阶段："
Write-Host "     1. 需求分析  →  说: '开始需求分析'"
Write-Host "     2. 系统设计  →  说: '开始系统设计'"
Write-Host "     3. 开发实现  →  说: '开始开发'"
Write-Host "     4. 测试验证  →  说: '开始测试'"
Write-Host "     5. 质量保障  →  说: '开始QA验收'"
Write-Host "     6. 部署上线  →  说: '开始部署'"
Write-Host "================================================"
```

---

## SKILL_OUTPUT

```
环境检查报告（全部通过后输出）：
✅ E01~E09  基础环境检查全部通过
✅ G01~G07  Go 开发环境检查全部通过
✅ S01~S07  SaaS 服务检查全部通过
✅ W01~W05  Windows 专项检查全部通过
🚀 环境准备就绪，可以开始工作
```

## SKILL_DONE_CRITERIA
- 所有检查项均通过（E01~E09、G01~G07、S01~S07、W01~W05）
- Go 版本 >= 1.21，模块名为 github.com/zouluxing/feature_orbit_server
- Docker 运行中，PostgreSQL 和 Redis 容器正常
- .env 文件存在且关键变量完整
- 当前位于 develop 分支且代码为最新
- 输出「环境准备就绪」提示后自动进入用户指定阶段
