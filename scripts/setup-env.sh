#!/bin/bash

# ================================================
# feature_orbit 环境自动检测与准备脚本
# 版本: 1.0.0
# 用法: bash scripts/setup-env.sh
# ================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 计数器
CHECK_PASS=0
CHECK_FAIL=0
AUTO_FIXED=0

log_pass()  { echo -e "${GREEN}✅ $1${NC}"; ((CHECK_PASS++)); }
log_fail()  { echo -e "${RED}❌ $1${NC}"; ((CHECK_FAIL++)); }
log_warn()  { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_fix()   { echo -e "${BLUE}🔧 $1${NC}"; ((AUTO_FIXED++)); }
log_info()  { echo -e "${BLUE}ℹ️  $1${NC}"; }

echo ""
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}   feature_orbit 环境自动检测与准备${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# ------------------------------------------------
# E01: 检测 Node.js
# ------------------------------------------------
echo -e "${BLUE}[E01] 检测 Node.js...${NC}"
if command -v node &> /dev/null; then
  NODE_VER=$(node -v | sed 's/v//' | cut -d. -f1)
  if [ "$NODE_VER" -ge 18 ]; then
    log_pass "Node.js $(node -v) 已就绪"
  else
    log_warn "Node.js 版本过低（当前: $(node -v)，需要: v18+）"
    log_fix "尝试通过 nvm 升级 Node.js..."
    if command -v nvm &> /dev/null; then
      nvm install 20 && nvm use 20
      log_pass "Node.js 已升级到 $(node -v)"
    else
      log_fail "请手动安装 Node.js v18+: https://nodejs.org"
    fi
  fi
else
  log_warn "Node.js 未安装，尝试自动安装..."
  if command -v brew &> /dev/null; then
    log_fix "使用 Homebrew 安装 Node.js..."
    brew install node
    log_pass "Node.js $(node -v) 安装成功"
  elif command -v apt-get &> /dev/null; then
    log_fix "使用 apt-get 安装 Node.js..."
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
    sudo apt-get install -y nodejs
    log_pass "Node.js $(node -v) 安装成功"
  else
    log_fail "无法自动安装，请手动安装: https://nodejs.org"
    exit 1
  fi
fi

# ------------------------------------------------
# E02: 检测 Git
# ------------------------------------------------
echo -e "${BLUE}[E02] 检测 Git...${NC}"
if command -v git &> /dev/null; then
  log_pass "$(git --version) 已就绪"
else
  log_warn "Git 未安装，尝试自动安装..."
  if command -v brew &> /dev/null; then
    log_fix "使用 Homebrew 安装 Git..."
    brew install git && log_pass "Git 安装成功"
  elif command -v apt-get &> /dev/null; then
    log_fix "使用 apt-get 安装 Git..."
    sudo apt-get install -y git && log_pass "Git 安装成功"
  else
    log_fail "无法自动安装 Git，请手动安装: https://git-scm.com"
    exit 1
  fi
fi

# ------------------------------------------------
# E03: 检测 Git 用户配置
# ------------------------------------------------
echo -e "${BLUE}[E03] 检测 Git 用户配置...${NC}"
GIT_NAME=$(git config --global user.name 2>/dev/null || echo "")
GIT_EMAIL=$(git config --global user.email 2>/dev/null || echo "")

if [ -z "$GIT_NAME" ]; then
  log_warn "Git user.name 未配置"
  read -p "  👤 请输入你的 Git 用户名: " INPUT_NAME
  git config --global user.name "$INPUT_NAME"
  log_fix "Git user.name 已设置为: $INPUT_NAME"
else
  log_pass "Git user.name: $GIT_NAME"
fi

if [ -z "$GIT_EMAIL" ]; then
  log_warn "Git user.email 未配置"
  read -p "  📧 请输入你的 Git 邮箱: " INPUT_EMAIL
  git config --global user.email "$INPUT_EMAIL"
  log_fix "Git user.email 已设置为: $INPUT_EMAIL"
else
  log_pass "Git user.email: $GIT_EMAIL"
fi

# ------------------------------------------------
# E04: 检测 Claude Code
# ------------------------------------------------
echo -e "${BLUE}[E04] 检测 Claude Code...${NC}"
if command -v claude &> /dev/null; then
  log_pass "Claude Code 已就绪"
else
  log_warn "Claude Code 未安装，自动安装中..."
  log_fix "执行: npm install -g @anthropic-ai/claude-code"
  npm install -g @anthropic-ai/claude-code
  if command -v claude &> /dev/null; then
    log_pass "Claude Code 安装成功"
  else
    log_fail "Claude Code 安装失败，请手动执行: npm install -g @anthropic-ai/claude-code"
    exit 1
  fi
fi

# ------------------------------------------------
# E05: 检测仓库克隆状态
# ------------------------------------------------
echo -e "${BLUE}[E05] 检测仓库状态...${NC}"
if git rev-parse --git-dir &> /dev/null; then
  REMOTE_URL=$(git remote get-url origin 2>/dev/null || echo "无远端")
  log_pass "Git 仓库已就绪: $REMOTE_URL"
else
  log_warn "当前目录不是 Git 仓库，自动克隆..."
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  PARENT_DIR="$(dirname "$SCRIPT_DIR")"
  cd "$PARENT_DIR"
  log_fix "克隆 feature_orbit 仓库..."
  git clone https://github.com/zouluxing/feature_orbit.git
  cd feature_orbit
  log_pass "仓库克隆成功"
fi

# ------------------------------------------------
# E06: 检测当前分支
# ------------------------------------------------
echo -e "${BLUE}[E06] 检测当前分支...${NC}"
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" = "develop" ]; then
  log_pass "当前分支: develop"
else
  log_warn "当前分支为 '$CURRENT_BRANCH'，自动切换到 develop..."
  if git show-ref --verify --quiet refs/heads/develop; then
    git checkout develop
  else
    git checkout -b develop origin/develop 2>/dev/null || git checkout develop
  fi
  log_fix "已切换到 develop 分支"
  log_pass "当前分支: develop"
fi

# ------------------------------------------------
# E07: 同步远端代码
# ------------------------------------------------
echo -e "${BLUE}[E07] 同步远端代码...${NC}"
git fetch origin --quiet
LOCAL_SHA=$(git rev-parse HEAD)
REMOTE_SHA=$(git rev-parse origin/develop 2>/dev/null || echo "")

if [ -n "$REMOTE_SHA" ] && [ "$LOCAL_SHA" != "$REMOTE_SHA" ]; then
  log_warn "本地代码落后于远端，自动拉取中..."
  git pull origin develop --quiet
  log_fix "代码已同步到最新 ($(git rev-parse --short HEAD))"
else
  log_pass "本地代码已是最新 ($(git rev-parse --short HEAD))"
fi

# ------------------------------------------------
# E08: 检测 Skills 文件完整性
# ------------------------------------------------
echo -e "${BLUE}[E08] 检测 Skills 文件...${NC}"
SKILLS_DIR=".agents/skills"
REQUIRED_SKILLS=("env-setup.md" "requirements-engineer.md" "designer.md" "developer.md" "tester.md" "qa-engineer.md" "devops-engineer.md")
MISSING_SKILLS=[]

for skill in "${REQUIRED_SKILLS[@]}"; do
  if [ ! -f "$SKILLS_DIR/$skill" ]; then
    MISSING_SKILLS+=("$skill")
  fi
done

if [ ${#MISSING_SKILLS[@]} -gt 0 ]; then
  log_warn "发现 ${#MISSING_SKILLS[@]} 个缺失的 Skill 文件: ${MISSING_SKILLS[*]}"
  log_fix "重新拉取最新代码以恢复 Skills..."
  git pull origin develop --quiet
  log_pass "Skills 文件已恢复"
else
  log_pass "所有 Skills 文件完整 (${#REQUIRED_SKILLS[@]}/${#REQUIRED_SKILLS[@]})"
fi

# ------------------------------------------------
# E09: 检测文档目录结构
# ------------------------------------------------
echo -e "${BLUE}[E09] 检测文档目录结构...${NC}"
REQUIRED_DIRS=("docs/requirements" "docs/design" "docs/testing" "docs/qa" "docs/deploy" "docs/playbook")
CREATED_DIRS=0

for dir in "${REQUIRED_DIRS[@]}"; do
  if [ ! -d "$dir" ]; then
    mkdir -p "$dir"
    touch "$dir/.gitkeep"
    ((CREATED_DIRS++))
  fi
done

if [ $CREATED_DIRS -gt 0 ]; then
  log_fix "已自动创建 $CREATED_DIRS 个文档目录"
fi
log_pass "文档目录结构完整"

# ------------------------------------------------
# E10: 检测项目依赖
# ------------------------------------------------
echo -e "${BLUE}[E10] 检测项目依赖...${NC}"
DEP_FOUND=false

if [ -f "pubspec.yaml" ]; then
  DEP_FOUND=true
  if command -v flutter &> /dev/null; then
    log_fix "安装 Flutter 依赖..."
    flutter pub get --quiet && log_pass "Flutter 依赖已就绪"
  else
    log_fail "Flutter SDK 未安装: https://flutter.dev/docs/get-started/install"
  fi
fi

if [ -f "go.mod" ]; then
  DEP_FOUND=true
  if command -v go &> /dev/null; then
    log_fix "安装 Go 依赖..."
    go mod download && log_pass "Go 依赖已就绪"
  else
    log_fail "Go 未安装: https://golang.org/dl/"
  fi
fi

if [ -f "package.json" ]; then
  DEP_FOUND=true
  if [ ! -d "node_modules" ]; then
    log_fix "安装 Node.js 依赖..."
    npm install --quiet && log_pass "Node.js 依赖已就绪"
  else
    log_pass "Node.js 依赖已就绪"
  fi
fi

if [ "$DEP_FOUND" = false ]; then
  log_pass "无项目依赖文件（初始化阶段正常）"
fi

# ------------------------------------------------
# 输出检查报告
# ------------------------------------------------
echo ""
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}              环境检查报告${NC}"
echo -e "${BLUE}================================================${NC}"
echo -e "  通过项: ${GREEN}$CHECK_PASS${NC}"
echo -e "  失败项: ${RED}$CHECK_FAIL${NC}"
echo -e "  自动修复: ${YELLOW}$AUTO_FIXED${NC}"
echo -e "${BLUE}------------------------------------------------${NC}"
echo -e "  Node.js:    $(node -v)"
echo -e "  Git:        $(git --version | cut -d' ' -f3)"
echo -e "  用户:        $(git config --global user.name)"
echo -e "  分支:        $(git branch --show-current)"
echo -e "  代码版本:    $(git rev-parse --short HEAD)"
echo -e "${BLUE}================================================${NC}"

if [ $CHECK_FAIL -eq 0 ]; then
  echo -e "${GREEN}"
  echo "  🚀 环境准备就绪！可以开始工作了。"
  echo -e "${NC}"
  echo -e "${BLUE}================================================${NC}"
  echo ""
  echo "  📋 可用的工作阶段："
  echo "     1. 需求分析  →  说: '开始需求分析'"
  echo "     2. 系统设计  →  说: '开始系统设计'"
  echo "     3. 开发实现  →  说: '开始开发'"
  echo "     4. 测试验证  →  说: '开始测试'"
  echo "     5. 质量保障  →  说: '开始QA验收'"
  echo "     6. 部署上线  →  说: '开始部署'"
  echo ""
  exit 0
else
  echo -e "${RED}"
  echo "  ⚠️  有 $CHECK_FAIL 个检查项未通过，请根据提示手动修复后重新运行。"
  echo -e "${NC}"
  exit 1
fi
