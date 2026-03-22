# 阶段二：系统设计

## 阶段目标
将需求转化为可执行的技术方案，产出系统架构、数据库设计、API 设计和 UI 原型。

## 执行角色
**设计师** → 加载角色：`claude/roles/designer.md`

## 分支策略
```bash
git checkout develop
git checkout -b stage/design
```

## 阶段步骤

```
Step 1 ──► Step 2 ──► Step 3 ──► Step 4 ──► Step 5 ──► Step 6
阅读需求    架构设计    数据库设计  API设计    UI原型     设计评审
```

## 入口条件
- 阶段一 PRD 文档已通过评审
- 需求基线已锁定

## 出口条件
- [ ] 系统架构设计文档完成
- [ ] 数据库 ER 图和 DDL 完成
- [ ] API 设计文档覆盖所有 P0 功能
- [ ] `stage/design` 分支 PR 已合并到 `develop`

## 下一阶段
出口条件满足后，通知开发工程师启动阶段三：开发实现。
