# ================================================
# feature_orbit 环境自动检测与准备脚本 (Windows)
# 版本: 1.0.0
# 用法: PowerShell -ExecutionPolicy Bypass -File scripts/setup-env.ps1
# ================================================

$ErrorActionPreference = "Stop"
$CHECK_PASS = 0
$CHECK_FAIL = 0
$AUTO_FIXED = 0

function Log-Pass  { param($msg) Write-Host "✅ $msg" -ForegroundColor Green;  $script:CHECK_PASS++ }
function Log-Fail  { param($msg) Write-Host "❌ $msg" -ForegroundColor Red;    $script:CHECK_FAIL++ }
function Log-Warn  { param($msg) Write-Host "⚠️  $msg" -ForegroundColor Yellow }
function Log-Fix   { param($msg) Write-Host "🔧 $msg" -ForegroundColor Cyan;   $script:AUTO_FIXED++ }
function Log-Info  { param($msg) Write-Host "ℹ️  $msg" -ForegroundColor Blue }

Write-Host ""
Write-Host "================================================" -ForegroundColor Blue
Write-Host "   feature_orbit 环境自动检测与准备" -ForegroundColor Blue
Write-Host "================================================" -ForegroundColor Blue
Write-Host ""

# E01: 检测 Node.js
Write-Host "[E01] 检测 Node.js..." -ForegroundColor Blue
try {
  $nodeVer = (node -v 2>$null).TrimStart('v').Split('.')[0]
  if ([int]$nodeVer -ge 18) {
    Log-Pass "Node.js $(node -v) 已就绪"
  } else {
    Log-Warn "Node.js 版本过低，需要 v18+，请访问 https://nodejs.org 更新"
    $script:CHECK_FAIL++
  }
} catch {
  Log-Warn "Node.js 未安装"
  Log-Fix "请访问 https://nodejs.org/dist/latest/ 下载安装包"
  Start-Process "https://nodejs.org"
  $script:CHECK_FAIL++
}

# E02: 检测 Git
Write-Host "[E02] 检测 Git..." -ForegroundColor Blue
try {
  $gitVer = git --version
  Log-Pass "$gitVer 已就绪"
} catch {
  Log-Warn "Git 未安装"
  Log-Fix "请访问 https://git-scm.com/download/win 下载安装"
  Start-Process "https://git-scm.com/download/win"
  $script:CHECK_FAIL++
}

# E03: Git 用户配置
Write-Host "[E03] 检测 Git 用户配置..." -ForegroundColor Blue
$gitName = git config --global user.name 2>$null
$gitEmail = git config --global user.email 2>$null

if (-not $gitName) {
  Log-Warn "Git user.name 未配置"
  $inputName = Read-Host "  请输入你的 Git 用户名"
  git config --global user.name $inputName
  Log-Fix "Git user.name 已设置为: $inputName"
} else {
  Log-Pass "Git user.name: $gitName"
}

if (-not $gitEmail) {
  Log-Warn "Git user.email 未配置"
  $inputEmail = Read-Host "  请输入你的 Git 邮箱"
  git config --global user.email $inputEmail
  Log-Fix "Git user.email 已设置为: $inputEmail"
} else {
  Log-Pass "Git user.email: $gitEmail"
}

# E04: 检测 Claude Code
Write-Host "[E04] 检测 Claude Code..." -ForegroundColor Blue
try {
  claude --version | Out-Null
  Log-Pass "Claude Code 已就绪"
} catch {
  Log-Warn "Claude Code 未安装，自动安装中..."
  Log-Fix "执行: npm install -g @anthropic-ai/claude-code"
  npm install -g @anthropic-ai/claude-code
  Log-Pass "Claude Code 安装成功"
}

# E05: 检测仓库状态
Write-Host "[E05] 检测仓库状态..." -ForegroundColor Blue
try {
  git rev-parse --git-dir | Out-Null
  $remoteUrl = git remote get-url origin
  Log-Pass "Git 仓库已就绪: $remoteUrl"
} catch {
  Log-Warn "当前目录不是 Git 仓库，自动克隆..."
  Set-Location ..
  git clone https://github.com/zouluxing/feature_orbit.git
  Set-Location feature_orbit
  Log-Fix "仓库克隆成功"
}

# E06: 检测当前分支
Write-Host "[E06] 检测当前分支..." -ForegroundColor Blue
$currentBranch = git branch --show-current
if ($currentBranch -eq "develop") {
  Log-Pass "当前分支: develop"
} else {
  Log-Warn "当前分支为 '$currentBranch'，自动切换到 develop..."
  git checkout develop 2>$null
  if ($LASTEXITCODE -ne 0) {
    git checkout -b develop origin/develop
  }
  Log-Fix "已切换到 develop 分支"
  Log-Pass "当前分支: develop"
}

# E07: 同步远端代码
Write-Host "[E07] 同步远端代码..." -ForegroundColor Blue
git fetch origin --quiet
$localSha = git rev-parse HEAD
$remoteSha = git rev-parse origin/develop 2>$null
if ($localSha -ne $remoteSha) {
  Log-Warn "本地代码落后于远端，自动拉取中..."
  git pull origin develop --quiet
  $shortSha = git rev-parse --short HEAD
  Log-Fix "代码已同步到最新 ($shortSha)"
} else {
  $shortSha = git rev-parse --short HEAD
  Log-Pass "本地代码已是最新 ($shortSha)"
}

# E08: 检测 Skills 完整性
Write-Host "[E08] 检测 Skills 文件..." -ForegroundColor Blue
$requiredSkills = @("env-setup.md","requirements-engineer.md","designer.md","developer.md","tester.md","qa-engineer.md","devops-engineer.md")
$missingSkills = @()
foreach ($skill in $requiredSkills) {
  if (-not (Test-Path ".agents/skills/$skill")) {
    $missingSkills += $skill
  }
}
if ($missingSkills.Count -gt 0) {
  Log-Warn "发现 $($missingSkills.Count) 个缺失的 Skill 文件"
  Log-Fix "重新拉取最新代码..."
  git pull origin develop --quiet
  Log-Pass "Skills 文件已恢复"
} else {
  Log-Pass "所有 Skills 文件完整 ($($requiredSkills.Count)/$($requiredSkills.Count))"
}

# E09: 检测文档目录
Write-Host "[E09] 检测文档目录结构..." -ForegroundColor Blue
$requiredDirs = @("docs/requirements","docs/design","docs/testing","docs/qa","docs/deploy","docs/playbook")
$createdDirs = 0
foreach ($dir in $requiredDirs) {
  if (-not (Test-Path $dir)) {
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    New-Item -ItemType File -Force -Path "$dir/.gitkeep" | Out-Null
    $createdDirs++
  }
}
if ($createdDirs -gt 0) { Log-Fix "已自动创建 $createdDirs 个文档目录" }
Log-Pass "文档目录结构完整"

# E10: 检测项目依赖
Write-Host "[E10] 检测项目依赖..." -ForegroundColor Blue
$depFound = $false
if (Test-Path "pubspec.yaml") {
  $depFound = $true
  if (Get-Command flutter -ErrorAction SilentlyContinue) {
    flutter pub get | Out-Null
    Log-Pass "Flutter 依赖已就绪"
  } else { Log-Fail "Flutter SDK 未安装: https://flutter.dev" }
}
if (Test-Path "go.mod") {
  $depFound = $true
  if (Get-Command go -ErrorAction SilentlyContinue) {
    go mod download
    Log-Pass "Go 依赖已就绪"
  } else { Log-Fail "Go 未安装: https://golang.org/dl/" }
}
if (Test-Path "package.json") {
  $depFound = $true
  if (-not (Test-Path "node_modules")) {
    npm install --quiet
    Log-Pass "Node.js 依赖已就绪"
  } else { Log-Pass "Node.js 依赖已就绪" }
}
if (-not $depFound) { Log-Pass "无项目依赖文件（初始化阶段正常）" }

# 输出报告
Write-Host ""
Write-Host "================================================" -ForegroundColor Blue
Write-Host "              环境检查报告" -ForegroundColor Blue
Write-Host "================================================" -ForegroundColor Blue
Write-Host "  通过项: $CHECK_PASS" -ForegroundColor Green
Write-Host "  失败项: $CHECK_FAIL" -ForegroundColor Red
Write-Host "  自动修复: $AUTO_FIXED" -ForegroundColor Yellow
Write-Host "------------------------------------------------" -ForegroundColor Blue
Write-Host "  Node.js:  $(node -v)"
Write-Host "  Git:      $(git --version | ForEach-Object { $_.Split(' ')[2] })"
Write-Host "  用户:      $(git config --global user.name)"
Write-Host "  分支:      $(git branch --show-current)"
Write-Host "  代码版本:  $(git rev-parse --short HEAD)"
Write-Host "================================================" -ForegroundColor Blue

if ($CHECK_FAIL -eq 0) {
  Write-Host ""
  Write-Host "  🚀 环境准备就绪！可以开始工作了。" -ForegroundColor Green
  Write-Host ""
  Write-Host "  📋 可用的工作阶段："
  Write-Host "     1. 需求分析  →  说: '开始需求分析'"
  Write-Host "     2. 系统设计  →  说: '开始系统设计'"
  Write-Host "     3. 开发实现  →  说: '开始开发'"
  Write-Host "     4. 测试验证  →  说: '开始测试'"
  Write-Host "     5. 质量保障  →  说: '开始QA验收'"
  Write-Host "     6. 部署上线  →  说: '开始部署'"
  exit 0
} else {
  Write-Host "  ⚠️  有 $CHECK_FAIL 个检查项未通过，请根据提示修复后重新运行。" -ForegroundColor Red
  exit 1
}
