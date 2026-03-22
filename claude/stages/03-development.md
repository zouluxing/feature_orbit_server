# 阶段三：开发实现

## 阶段目标
依据设计文档实现系统功能，编写完整的单元测试，确保代码质量。

## 执行角色
**开发工程师** → 加载角色：`claude/roles/developer.md`

## 分支策略
```bash
# 每个功能模块单独建分支
git checkout develop
git checkout -b feature/<功能名称>

# 完成后合并回 develop
git checkout develop
git merge feature/<功能名称>
```

## 阶段步骤

```
Step 1 ──► Step 2 ──► Step 3 ──► Step 4 ──► Step 5 ──► Step 6
开发准备    功能实现    单元测试    集成测试    代码自检    提交PR
```

## 入口条件
- 阶段二设计文档已通过评审
- 开发环境已就绪

## 出口条件
- [ ] 所有 P0 用户故事已实现
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] CI 流水线全部通过
- [ ] PR 已通过 Code Review 并合并

## 下一阶段
出口条件满足后，通知测试工程师启动阶段四：测试验证。
