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
| E09 | 项目目录结构 | 关键目录存在 | ✅ 自动创建 |

### Go 开发环境

| 编号 | 检查项 | 要求 | 自动修复 |
|------|--------|------|----------|
| G01 | Go 安装与版本 | Go 1.21+ | ✅ 自动安装 |
| G02 | GOPATH / GOPROXY | 环境变量配置正确 | ✅ 自动配置 |
| G03 | Go 模块初始化 | go.mod 存在 | ✅ 自动初始化 |
| G04 | Go 项目依赖 | go mod download | ✅ 自动下载 |
| G05 | Go 代码规范工具 | gofmt / golangci-lint | ✅ 自动安装 |
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
| W01 | PowerShell 版本 | PowerShell 5.1+ 或 Core 7+ | ✅ 引导升级 |
| W02 | Windows Terminal | 推荐安装 | ⚠️ 提示安装 |
| W03 | Make 工具 | make 命令可用 | ✅ 自动安装 |
| W04 | 行尾符配置 | Git autocrlf 设置正确 | ✅ 自动配置 |
| W05 | 关键端口 | 8080、5432、6379 端口可用 | ✅ 检测报告 |

---

## SKILL_STEPS

### Step 1 — 基础环境检查（E01~E09）

```powershell
# [E01~E08] 见 scripts/setup-env.ps1

# [E09] 检测并创建项目关键目录结构
$requiredDirs = @(
    "cmd/server",
    "internal/config",
    "internal/handler",
    "internal/service",
    "internal/repository",
    "internal/model",
    "internal/dto",
    "pkg/logger",
    "pkg/response",
    "pkg/errors",
    "api",
    "configs",
    "migrations",
    "tests/unit",
    "tests/integration",
    "tests/fixtures",
    "docs/requirements",
    "docs/design",
    "docs/testing",
    "docs/qa",
    "docs/deploy",
    "docs/playbook"
)
$created = 0
foreach ($dir in $requiredDirs) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
        New-Item -ItemType File -Force -Path "$dir/.gitkeep" | Out-Null
        $created++
    }
}
if ($created -gt 0) { Write-Host "🔧 [E09] 已创建 $created 个项目目录" }
Write-Host "✅ [E09] 项目目录结构完整"
```

---

### Step 2 — Go 开发环境（G01~G07）

```powershell
# [G01] Go 安装与版本检查（见 setup-env.ps1）
# [G02] GOPATH / GOPROXY 配置（见 setup-env.ps1）
# [G03] go.mod 初始化
if (-not (Test-Path "go.mod")) {
    go mod init github.com/zouluxing/feature_orbit_server
    Write-Host "✅ [G03] go.mod 已初始化: github.com/zouluxing/feature_orbit_server"
}
# [G04~G07] 依赖下载、规范工具、测试工具、air（见 setup-env.ps1）
```

---

### Step 3 — SaaS 服务检查（S01~S07）

```powershell
# 使用 Makefile 快捷命令启动服务
make services-up    # 启动 PostgreSQL + Redis
make db-migrate     # 执行数据库迁移
make health         # 检查所有服务健康状态
```

---

### Step 4 — Windows 专项（W01~W05）

```powershell
# 见 scripts/setup-env.ps1 完整实现
```

---

### Step 5 — 输出环境报告

```powershell
Write-Host "================================================"
Write-Host "         feature_orbit_server 环境报告"
Write-Host "================================================"
Write-Host "  Git:        $(git --version)"
Write-Host "  Go:         $(go version)"
Write-Host "  模块名:       github.com/zouluxing/feature_orbit_server"
Write-Host "  Docker:     $(docker --version)"
Write-Host "  分支:        $(git branch --show-current)"
Write-Host "  代码版本:     $(git rev-parse --short HEAD)"
Write-Host "================================================"
Write-Host "  🚀 环境准备就绪！"
Write-Host ""
Write-Host "  💡 常用开发命令（make）："
Write-Host "     make dev          # 热重载启动服务"
Write-Host "     make test         # 运行全部测试"
Write-Host "     make lint         # 代码规范检查"
Write-Host "     make build        # 编译构建"
Write-Host "     make services-up  # 启动 DB/Redis"
Write-Host "================================================"
```

---

## SKILL_OUTPUT

```
✅ E01~E09  基础环境 + 项目目录结构完整
✅ G01~G07  Go 开发环境全部就绪
✅ S01~S07  PostgreSQL / Redis 服务正常
✅ W01~W05  Windows 专项检查通过
🚀 环境准备就绪，可以开始工作
```

## SKILL_DONE_CRITERIA
- 所有检查项均通过（E01~E09、G01~G07、S01~S07、W01~W05）
- 项目目录结构符合 Go SaaS 最佳实践
- Go 模块名为 github.com/zouluxing/feature_orbit_server
- Docker 运行中，PostgreSQL 和 Redis 容器正常
- .env 文件存在且关键变量完整
- 输出「环境准备就绪」后自动进入用户指定阶段
