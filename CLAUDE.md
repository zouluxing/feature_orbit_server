# Feature Orbit — Claude Code 主配置

## 项目说明
本项目使用 Claude Code Skills 实现分阶段、分角色的软件开发流程自动化。

---

## ⚡ 强制前置规则：环境检查

> **所有阶段任务开始前，必须先执行环境检查。**
> Claude Code 在接收到任何阶段指令时，自动加载 `.agents/skills/env-setup.md`，
> 完成全部检查项（E01~E10）并输出「环境准备就绪」后，才允许进入后续阶段。

**一键启动命令（每次开始工作前执行）：**

```bash
# macOS / Linux
bash scripts/setup-env.sh

# Windows PowerShell
PowerShell -ExecutionPolicy Bypass -File scripts/setup-env.ps1

# 或直接启动 Claude Code（自动触发环境检查）
claude
```

启动后 Claude Code 自动执行：
```
第一步：加载 env-setup Skill → 执行 E01~E10 所有检查
第二步：发现未就绪项 → 自动修复
第三步：所有检查通过 → 输出「环境准备就绪」
第四步：询问用户要开始哪个阶段 → 加载对应 Skill 开始工作
```

---

## Skills 目录

所有角色均以 Skill 形式定义，放置于 `.agents/skills/` 目录下。
Claude Code 在执行任务时会自动检索并加载匹配的 Skill。

```
.agents/
└── skills/
    ├── env-setup.md              # ⚡ 环境检查（所有阶段强制前置）
    ├── requirements-engineer.md  # Skill: 需求工程师
    ├── designer.md               # Skill: 系统设计师
    ├── developer.md              # Skill: 开发工程师
    ├── tester.md                 # Skill: 测试工程师
    ├── qa-engineer.md            # Skill: QA 工程师
    └── devops-engineer.md        # Skill: 运维工程师
```

---

## 开发阶段总览

| 阶段 | Skill | 触发关键词 | 产出物 |
|------|-------|-----------|--------|
| 0. 环境检查 | `env-setup` | 启动/开始工作/任意阶段指令 | 环境就绪报告 |
| 1. 需求分析 | `requirements-engineer` | 需求、PRD、用户故事 | PRD、User Stories |
| 2. 系统设计 | `designer` | 架构、设计、API、数据库 | 架构图、API文档 |
| 3. 开发实现 | `developer` | 实现、编码、开发、feature | 源码、单元测试 |
| 4. 测试验证 | `tester` | 测试、用例、Bug | 测试报告 |
| 5. 质量保障 | `qa-engineer` | QA、验收、质量 | QA报告 |
| 6. 部署上线 | `devops-engineer` | 部署、上线、发布、CI/CD | 上线报告 |

---

## 全局规则

1. **环境检查前置**：任何阶段开始前必须通过 env-setup Skill 的所有检查项
2. **严格按阶段顺序执行**，上一阶段产出物未完成不得进入下一阶段
3. **每个阶段完成后**，创建 PR，由下一阶段角色 Review 后方可合并
4. **所有产出物**统一存放在 `docs/` 对应子目录
5. **代码提交**遵循 Conventional Commits 规范
6. **Skills 可组合使用**，例如开发阶段同时加载 `developer` + `tester`

---

## 快速启动

```bash
# 1. 进入项目目录
cd feature_orbit

# 2. 运行环境检查脚本（自动检测并修复环境）
bash scripts/setup-env.sh        # macOS/Linux
# 或
PowerShell -ExecutionPolicy Bypass -File scripts/setup-env.ps1  # Windows

# 3. 启动 Claude Code
claude

# 4. 告诉 Claude Code 你要做什么
> 开始需求分析阶段
# Claude Code 自动检查环境 → 加载需求工程师 Skill → 开始引导
```

---

## 实施手册

各阶段详细操作步骤请参考：

| 阶段 | 手册文件 |
|------|----------|
| 阶段一：需求分析 | `docs/playbook/01-requirements.md` |
| 阶段二：系统设计 | `docs/playbook/02-design.md` |
| 阶段三：开发实现 | `docs/playbook/03-development.md` |
| 阶段四：测试验证 | `docs/playbook/04-testing.md` |
| 阶段五：质量保障 | `docs/playbook/05-qa.md` |
| 阶段六：部署上线 | `docs/playbook/06-deploy.md` |
