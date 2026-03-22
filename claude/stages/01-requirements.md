# 阶段一：需求分析

## 阶段目标
收集、分析、整理项目需求，产出可执行的需求文档，为设计和开发提供准确输入。

## 执行角色
**需求工程师** → 加载角色：`claude/roles/requirements-engineer.md`

## 分支策略
```bash
git checkout develop
git checkout -b stage/requirements
```

## 阶段步骤

```
Step 1 ──► Step 2 ──► Step 3 ──► Step 4 ──► Step 5
需求收集    需求分析    编写PRD    拆解故事    需求评审
```

## 入口条件
- 项目立项已批准
- 业务目标已明确

## 出口条件
- [ ] PRD 文档已完成并通过评审
- [ ] 用户故事已拆解，P0 故事有验收标准
- [ ] `stage/requirements` 分支 PR 已合并到 `develop`

## 下一阶段
出口条件满足后，通知设计师启动阶段二：系统设计。
