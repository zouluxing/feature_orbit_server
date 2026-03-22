# Skill: 环境自动检测与准备 (Environment Setup)

## SKILL_DESCRIPTION
在任何阶段任务开始前自动触发，检测开发环境是否就绪，若未就绪则自动修复，直到环境完全准备好后才允许进入后续阶段。
触发关键词：环境检查、环境准备、开始工作、启动项目、初始化环境、env check、setup

> ⚠️ 此 Skill 为所有阶段的强制前置 Skill，每次开始新阶段前必须先执行并通过所有检查项。

---

## 检查项总览

| 编号 | 检查项 | 说明 | 自动修复 |
|------|--------|------|----------|
| E01 | Node.js 版本 | 需要 v18+ | ✅ 引导安装 |
| E02 | Git 安装 | 需要 Git 2.x+ | ✅ 引导安装 |
| E03 | Git 用户配置 | user.name / user.email | ✅ 自动配置 |
| E04 | Claude Code 安装 | 需要最新版 | ✅ 自动安装 |
| E05 | 仓库克隆状态 | 是否已克隆到本地 | ✅ 自动克隆 |
| E06 | 当前分支 | 必须在 develop 分支 | ✅ 自动切换 |
| E07 | 本地代码同步 | 与远端 develop 保持一致 | ✅ 自动拉取 |
| E08 | .agents/skills 目录 | Skills 文件是否完整 | ✅ 自动拉取 |
| E09 | docs 目录结构 | 文档目录是否完整 | ✅ 自动创建 |
| E10 | 项目依赖安装 | package.json 依赖 / go mod 等 | ✅ 自动安装 |

---

## SKILL_STEPS

### Step 1 — 检测 Node.js

```bash
# 检测 Node.js 是否安装且版本 >= 18
node_version=$(node -v 2>/dev/null | sed 's/v//' | cut -d. -f1)

if [ -z "$node_version" ]; then
  echo "❌ [E01] Node.js 未安装"
  echo "👉 自动修复：请访问 https://nodejs.org 下载 LTS 版本安装"
  echo "   或使用 nvm 安装："
  echo "   curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash"
  echo "   nvm install 20 && nvm use 20"
  # Windows 用户提示
  echo "   Windows 用户：https://nodejs.org/dist/latest/ 下载 .msi 安装包"
  exit 1
elif [ "$node_version" -lt 18 ]; then
  echo "❌ [E01] Node.js 版本过低（当前: v${node_version}，需要: v18+）"
  echo "👉 自动修复：nvm install 20 && nvm use 20"
  exit 1
else
  echo "✅ [E01] Node.js v$(node -v) 已就绪"
fi
```

---

### Step 2 — 检测 Git

```bash
# 检测 Git 是否安装
if ! command -v git &> /dev/null; then
  echo "❌ [E02] Git 未安装"
  echo "👉 自动修复："
  echo "   macOS:   brew install git"
  echo "   Ubuntu:  sudo apt-get install git"
  echo "   Windows: https://git-scm.com/download/win"
  exit 1
else
  echo "✅ [E02] Git $(git --version) 已就绪"
fi

# 检测 Git 用户配置
git_name=$(git config --global user.name)
git_email=$(git config --global user.email)

if [ -z "$git_name" ]; then
  echo "⚠️  [E03] Git user.name 未配置，自动修复中..."
  read -p "请输入你的 Git 用户名: " input_name
  git config --global user.name "$input_name"
  echo "✅ [E03] Git user.name 已设置为: $input_name"
else
  echo "✅ [E03] Git user.name: $git_name"
fi

if [ -z "$git_email" ]; then
  echo "⚠️  [E03] Git user.email 未配置，自动修复中..."
  read -p "请输入你的 Git 邮箱: " input_email
  git config --global user.email "$input_email"
  echo "✅ [E03] Git user.email 已设置为: $input_email"
else
  echo "✅ [E03] Git user.email: $git_email"
fi
```

---

### Step 3 — 检测 Claude Code

```bash
# 检测 Claude Code 是否安装
if ! command -v claude &> /dev/null; then
  echo "⚠️  [E04] Claude Code 未安装，自动安装中..."
  npm install -g @anthropic-ai/claude-code
  if command -v claude &> /dev/null; then
    echo "✅ [E04] Claude Code 安装成功: $(claude --version)"
  else
    echo "❌ [E04] Claude Code 安装失败，请手动执行: npm install -g @anthropic-ai/claude-code"
    exit 1
  fi
else
  echo "✅ [E04] Claude Code $(claude --version) 已就绪"
fi
```

---

### Step 4 — 检测仓库状态

```bash
# 检测是否在 git 仓库中
if ! git rev-parse --git-dir &> /dev/null; then
  echo "⚠️  [E05] 当前目录不是 Git 仓库，自动克隆中..."
  cd ..
  git clone https://github.com/zouluxing/feature_orbit.git
  cd feature_orbit
  echo "✅ [E05] 仓库克隆成功"
else
  echo "✅ [E05] Git 仓库已就绪: $(git remote get-url origin)"
fi

# 检测当前分支
current_branch=$(git branch --show-current)
if [ "$current_branch" != "develop" ]; then
  echo "⚠️  [E06] 当前分支为 '$current_branch'，自动切换到 develop..."
  git checkout develop 2>/dev/null || git checkout -b develop origin/develop
  echo "✅ [E06] 已切换到 develop 分支"
else
  echo "✅ [E06] 当前分支: develop"
fi

# 同步远端最新代码
echo "🔄 [E07] 同步远端最新代码..."
git fetch origin
local_sha=$(git rev-parse HEAD)
remote_sha=$(git rev-parse origin/develop)

if [ "$local_sha" != "$remote_sha" ]; then
  echo "⚠️  [E07] 本地代码落后于远端，自动拉取中..."
  git pull origin develop
  echo "✅ [E07] 代码已同步到最新"
else
  echo "✅ [E07] 本地代码已是最新"
fi
```

---

### Step 5 — 检测 Skills 完整性

```bash
# 检测 .agents/skills 目录及所有 Skill 文件
SKILLS_DIR=".agents/skills"
REQUIRED_SKILLS=(
  "env-setup.md"
  "requirements-engineer.md"
  "designer.md"
  "developer.md"
  "tester.md"
  "qa-engineer.md"
  "devops-engineer.md"
)

if [ ! -d "$SKILLS_DIR" ]; then
  echo "⚠️  [E08] .agents/skills 目录不存在，重新拉取中..."
  git pull origin develop
fi

missing_skills=[]
for skill in "${REQUIRED_SKILLS[@]}"; do
  if [ ! -f "$SKILLS_DIR/$skill" ]; then
    missing_skills+=("$skill")
    echo "❌ [E08] 缺少 Skill 文件: $skill"
  fi
done

if [ ${#missing_skills[@]} -gt 0 ]; then
  echo "⚠️  [E08] 发现缺失的 Skill 文件，尝试重新拉取..."
  git pull origin develop --force
  echo "✅ [E08] Skills 已重新拉取"
else
  echo "✅ [E08] 所有 Skill 文件完整 (${#REQUIRED_SKILLS[@]}/${#REQUIRED_SKILLS[@]})"
fi
```

---

### Step 6 — 检测文档目录结构

```bash
# 检测并创建必要的文档目录
REQUIRED_DIRS=(
  "docs/requirements"
  "docs/design"
  "docs/testing"
  "docs/qa"
  "docs/deploy"
  "docs/playbook"
)

for dir in "${REQUIRED_DIRS[@]}"; do
  if [ ! -d "$dir" ]; then
    echo "⚠️  [E09] 目录不存在，自动创建: $dir"
    mkdir -p "$dir"
    touch "$dir/.gitkeep"
    echo "✅ [E09] 已创建: $dir"
  fi
done
echo "✅ [E09] 文档目录结构完整"
```

---

### Step 7 — 检测项目依赖

```bash
# 根据项目类型自动检测并安装依赖

# Flutter/Dart 项目
if [ -f "pubspec.yaml" ]; then
  echo "🔍 检测到 Flutter 项目"
  if ! command -v flutter &> /dev/null; then
    echo "❌ [E10] Flutter SDK 未安装"
    echo "👉 请访问 https://flutter.dev/docs/get-started/install 安装 Flutter"
    exit 1
  fi
  echo "🔄 [E10] 安装 Flutter 依赖..."
  flutter pub get
  echo "✅ [E10] Flutter 依赖安装完成"
fi

# Go 项目
if [ -f "go.mod" ]; then
  echo "🔍 检测到 Go 项目"
  if ! command -v go &> /dev/null; then
    echo "❌ [E10] Go 未安装，请访问 https://golang.org/dl/ 安装"
    exit 1
  fi
  echo "🔄 [E10] 安装 Go 依赖..."
  go mod download
  echo "✅ [E10] Go 依赖安装完成"
fi

# Node.js 项目
if [ -f "package.json" ]; then
  echo "🔍 检测到 Node.js 项目"
  if [ ! -d "node_modules" ]; then
    echo "🔄 [E10] 安装 Node.js 依赖..."
    npm install
    echo "✅ [E10] Node.js 依赖安装完成"
  else
    echo "✅ [E10] Node.js 依赖已就绪"
  fi
fi

# 无依赖文件
if [ ! -f "pubspec.yaml" ] && [ ! -f "go.mod" ] && [ ! -f "package.json" ]; then
  echo "✅ [E10] 无需安装项目依赖（项目初始化阶段）"
fi
```

---

### Step 8 — 输出环境报告

```bash
echo ""
echo "================================================"
echo "         环境检查报告"
echo "================================================"
echo "✅ E01  Node.js:        $(node -v)"
echo "✅ E02  Git:            $(git --version)"
echo "✅ E03  Git 用户:       $(git config --global user.name)"
echo "✅ E04  Claude Code:    $(claude --version 2>/dev/null || echo '已安装')"
echo "✅ E05  仓库:           $(git remote get-url origin)"
echo "✅ E06  当前分支:       $(git branch --show-current)"
echo "✅ E07  代码同步:       最新"
echo "✅ E08  Skills:         完整"
echo "✅ E09  文档目录:       完整"
echo "✅ E10  项目依赖:       已就绪"
echo "================================================"
echo "🚀 环境准备就绪！可以开始工作了。"
echo "================================================"
echo ""
echo "📋 下一步：告诉我你要开始哪个阶段？"
echo "   例如：'开始需求分析' 或 '继续开发阶段'"
```

---

## SKILL_OUTPUT

```
环境检查报告（全部通过后输出）：
✅ E01 ~ E10 所有检查项通过
🚀 环境准备就绪，可以开始工作
```

## SKILL_DONE_CRITERIA
- 所有 E01~E10 检查项均通过
- 当前位于 develop 分支且代码为最新
- 输出「环境准备就绪」提示
- 自动进入用户指定的阶段任务
