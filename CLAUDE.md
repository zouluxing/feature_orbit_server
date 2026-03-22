# Feature Orbit - Claude Code 开发流程配置

## 项目简介
本项目使用 Claude Code 实现分阶段、分角色的软件开发流程自动化。

---

## 开发阶段与角色总览

| 阶段 | 角色 | 分支命名 | 产出物 |
|------|------|----------|--------|
| 1. 需求分析 | 需求工程师 | `stage/requirements` | PRD 文档、用户故事 |
| 2. 系统设计 | 设计师 | `stage/design` | 架构图、API 设计、UI 原型 |
| 3. 开发实现 | 开发工程师 | `feature/*` | 源代码、单元测试 |
| 4. 测试验证 | 测试工程师 | `stage/testing` | 测试报告、Bug 清单 |
| 5. 质量保障 | QA 工程师 | `stage/qa` | QA 报告、验收结论 |
| 6. 部署上线 | 运维工程师 | `stage/deploy` | 部署脚本、上线报告 |

---

## 全局规则

1. **严格按阶段顺序执行**，上一阶段产出物未完成不得进入下一阶段
2. **每个阶段完成后**，需提交 PR 并由下一阶段角色 Review 后方可合并
3. **所有产出物**统一存放在 `docs/` 对应子目录下
4. **代码提交**遵循 Conventional Commits 规范
5. **角色切换**通过 `/role <角色名>` 指令触发

---

## 快速开始

```bash
# 启动需求分析阶段
/role requirements-engineer

# 查看当前阶段状态
/stage status

# 进入下一阶段
/stage next
```

---

## 角色配置文件路径

- 需求工程师：`claude/roles/requirements-engineer.md`
- 设计师：`claude/roles/designer.md`
- 开发工程师：`claude/roles/developer.md`
- 测试工程师：`claude/roles/tester.md`
- QA 工程师：`claude/roles/qa-engineer.md`
- 运维工程师：`claude/roles/devops-engineer.md`

## 阶段流程文件路径

- 需求阶段：`claude/stages/01-requirements.md`
- 设计阶段：`claude/stages/02-design.md`
- 开发阶段：`claude/stages/03-development.md`
- 测试阶段：`claude/stages/04-testing.md`
- QA 阶段：`claude/stages/05-qa.md`
- 部署阶段：`claude/stages/06-deploy.md`
