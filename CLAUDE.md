# Feature Orbit — Claude Code 主配置

## 项目说明
本项目使用 Claude Code Skills 实现分阶段、分角色的软件开发流程自动化。

---

## Skills 目录

所有角色均以 Skill 形式定义，放置于 `.claude/skills/` 目录下。
Claude Code 在执行任务时会自动检索并加载匹配的 Skill。

```
.claude/
├── skills/
│   ├── requirements-engineer.md  # Skill: 需求工程师
│   ├── designer.md               # Skill: 系统设计师
│   ├── developer.md              # Skill: 开发工程师
│   ├── tester.md                 # Skill: 测试工程师
│   ├── qa-engineer.md            # Skill: QA 工程师
│   └── devops-engineer.md        # Skill: 运维工程师
└── stages/
    ├── 01-requirements.md
    ├── 02-design.md
    ├── 03-development.md
    ├── 04-testing.md
    ├── 05-qa.md
    └── 06-deploy.md
```

---

## 开发阶段总览

| 阶段 | Skill | 触发关键词 | 产出物 |
|------|-------|-----------|--------|
| 1. 需求分析 | `requirements-engineer` | 需求、PRD、用户故事 | PRD、User Stories |
| 2. 系统设计 | `designer` | 架构、设计、API、数据库 | 架构图、API文档 |
| 3. 开发实现 | `developer` | 实现、编码、开发、feature | 源码、单元测试 |
| 4. 测试验证 | `tester` | 测试、用例、Bug | 测试报告 |
| 5. 质量保障 | `qa-engineer` | QA、验收、质量 | QA报告 |
| 6. 部署上线 | `devops-engineer` | 部署、上线、发布、CI/CD | 上线报告 |

---

## 全局规则

1. 严格按阶段顺序执行，上一阶段产出物未完成不得进入下一阶段
2. 每个阶段完成后创建 PR，由下一阶段角色 Review 后方可合并
3. 所有产出物统一存放在 `docs/` 对应子目录
4. 代码提交遵循 Conventional Commits 规范
5. Skills 可组合使用，例如开发阶段同时加载 `developer` + `tester`

---

## 快速启动

```bash
# 进入项目，Claude Code 自动加载 CLAUDE.md
cd feature_orbit && claude

# 启动阶段一
> 我需要开始需求分析阶段

# Claude Code 自动匹配并加载 requirements-engineer skill
# 然后按 skill 中定义的步骤逐一执行
```
