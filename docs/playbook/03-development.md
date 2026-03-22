# 阶段三：开发实现 实施手册

**执行角色**：开发工程师
**Skill 文件**：`.claude/skills/developer.md`
**目标分支**：`feature/<功能名>`（多个）→ 合并到 `develop`
**前置条件**：阶段二 PR 已合并到 develop

---

## Step 1：启动并激活开发工程师角色

```bash
claude

> 设计阶段已完成，请以开发工程师身份开始功能开发
```

Claude Code 自动读取设计文档，进入开发模式。

---

## Step 2：按用户故事拆分开发任务

```
> 请根据用户故事列表，规划开发任务顺序，并告诉我先从哪个 feature 开始
```

Claude Code 会给出任务拆分建议，例如：

```
建议开发顺序（按依赖关系）：
1. feature/user-auth      ← 基础，其他功能依赖
2. feature/user-profile   ← 依赖 auth
3. feature/core-business  ← 依赖 profile
4. feature/notification   ← 独立模块，可并行
```

---

## Step 3：创建 Feature 分支开发

每个功能独立分支，独立 PR：

```bash
# 从 develop 创建功能分支
git checkout develop && git pull
git checkout -b feature/user-auth

# 告诉 Claude Code 开始实现
> 请实现用户认证模块，参考 docs/design/api-design.md 中的认证接口设计
```

Claude Code 会逐步：
1. 创建项目目录结构
2. 实现核心逻辑
3. 编写单元测试
4. 检查代码规范

---

## Step 4：代码开发规范

Claude Code 在开发过程中自动遵守以下规范：

**提交规范（Conventional Commits）**：
```bash
git commit -m "feat(auth): 实现 JWT 登录接口"
git commit -m "feat(auth): 添加登录失败次数限制"
git commit -m "test(auth): 添加登录接口单元测试"
```

**代码质量要求**：
```bash
# 每次提交前自动执行
make lint      # 代码规范检查
make test      # 单元测试
make coverage  # 覆盖率检查（需 ≥ 80%）
```

**目录结构规范**：
```
src/
├── cmd/          # 入口文件
├── internal/     # 内部模块（不对外暴露）
│   ├── handler/  # HTTP 处理层
│   ├── service/  # 业务逻辑层
│   ├── repo/     # 数据访问层
│   └── model/    # 数据模型
├── pkg/          # 可复用公共包
tests/
├── unit/         # 单元测试
└── integration/  # 集成测试
```

---

## Step 5：单元测试编写

```
> 请为刚才实现的登录模块编写完整的单元测试，覆盖正常流程、错误输入和边界情况
```

**测试覆盖要求**：

| 场景 | 示例 |
|------|------|
| 正常流程 | 正确账号密码 → 返回 Token |
| 错误输入 | 错误密码 → 返回 40001 |
| 边界条件 | 连续失败5次 → 账号锁定 |
| 异常情况 | 数据库超时 → 返回 500 |

```bash
# 运行测试，确保通过
make test
make coverage  # 确保 ≥ 80%
```

---

## Step 6：集成测试

```
> 请运行集成测试，验证各模块间的交互是否正常
```

```bash
make integration-test
```

集成测试验证：
- API 接口与设计文档一致
- 数据库操作正确
- 模块间交互正常
- 权限控制有效

---

## Step 7：提交 PR

```bash
git push origin feature/user-auth
```

在 GitHub 创建 PR：`feature/user-auth` → `develop`

**PR 描述模板**（Claude Code 自动填写）：
```markdown
## 变更说明
实现了用户认证模块，包括：登录、注册、Token 刷新

## 关联 Issue
Closes #1

## 测试说明
- 单元测试覆盖率：87%
- 测试命令：`make test`

## 自检清单
- [x] 代码符合规范
- [x] 单元测试通过
- [x] 集成测试通过
- [x] 文档已更新
```

---

## Step 8：多 Feature 并行开发

当有多个功能需要同时开发时：

```bash
# 每个功能独立分支
git checkout develop && git checkout -b feature/notification

# Claude Code 中切换任务
> 现在切换到通知模块开发，请以开发工程师身份实现消息推送功能
```

**并行开发注意事项**：
- 各 feature 分支互相独立，避免交叉依赖
- 公共代码变更需先合并到 develop，再从 develop 拉取
- 定期 `git rebase develop` 保持分支最新

---

## 阶段完成标准

- [ ] 所有 P0 用户故事已实现
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] CI 流水线全部通过
- [ ] 所有 feature PR 已通过 Code Review 并合并
- [ ] 已通知测试工程师启动阶段四
