# 阶段一：需求分析 实施手册

**执行角色**：需求工程师
**Skill 文件**：`.claude/skills/requirements-engineer.md`
**目标分支**：`stage/requirements`

---

## 前置条件

- [ ] 项目立项已批准，业务目标明确
- [ ] 相关干系人已确认
- [ ] Claude Code 已安装并加载 CLAUDE.md

---

## Step 1：启动 Claude Code 并激活角色

```bash
# 1. 进入项目目录
cd feature_orbit

# 2. 启动 Claude Code
claude

# 3. 激活需求工程师 Skill（自然语言触发）
> 我需要开始需求分析阶段，请以需求工程师身份工作
```

Claude Code 自动加载 `.claude/skills/requirements-engineer.md`，进入需求工程师角色。

---

## Step 2：创建需求阶段分支

```bash
# Claude Code 执行（或手动执行）
git checkout develop
git pull origin develop
git checkout -b stage/requirements
```

---

## Step 3：需求收集（与 Claude Code 对话）

向 Claude Code 描述业务背景，它会引导你完成需求收集：

```
> 我们要开发一个功能，背景是：[描述业务背景]
  目标用户是：[描述用户群体]
  核心要解决的问题是：[描述痛点]
```

Claude Code 会自动：
- 识别用户角色（Persona）
- 区分功能性 / 非功能性需求
- 询问遗漏的边界条件
- 标注需求优先级 P0/P1/P2

---

## Step 4：生成 PRD 文档

```
> 请根据我们讨论的内容，生成 docs/requirements/PRD.md
```

Claude Code 自动创建文件，结构如下：

```
docs/requirements/
├── PRD.md           ← 产品需求文档
└── user-stories.md  ← 用户故事列表
```

**PRD 必须包含**：

| 章节 | 内容要求 |
|------|----------|
| 背景与目标 | 业务背景、成功指标 |
| 用户角色 | 主要/次要用户描述 |
| 功能需求 | P0/P1/P2 分级列表 |
| 非功能需求 | 性能、安全、兼容性指标（需量化）|
| 边界说明 | In Scope / Out of Scope |
| 依赖与风险 | 外部依赖、技术风险 |

---

## Step 5：拆解用户故事

```
> 请将 P0 功能需求拆解为用户故事，写入 docs/requirements/user-stories.md
```

每条用户故事必须包含验收标准（Given/When/Then 格式）：

```markdown
## US-001: 用户登录

**故事**：作为注册用户，我希望能用邮箱和密码登录，以便访问个人功能
**优先级**：P0

**验收标准**：
- Given 用户已注册 When 输入正确邮箱密码 Then 登录成功并跳转首页
- Given 用户已注册 When 输入错误密码 Then 提示错误，不泄露账户是否存在
- Given 用户连续失败5次 When 再次尝试 Then 账户锁定30分钟
```

---

## Step 6：需求评审与提交 PR

```bash
# 自检清单（让 Claude Code 执行）
> 请检查需求文档是否完整，对照自检清单逐项核查

# 提交文档
git add docs/requirements/
git commit -m "docs(requirements): 完成 PRD 和用户故事文档"
git push origin stage/requirements
```

然后在 GitHub 创建 PR：`stage/requirements` → `develop`

**PR 标题**：`docs(requirements): 完成需求分析阶段`

---

## 阶段完成标准检查

```
> 请检查需求分析阶段是否满足完成标准
```

Claude Code 自动逐项核查 `SKILL_DONE_CRITERIA`：

- [ ] PRD 文档通过设计师和技术 Leader Review
- [ ] 所有 P0 用户故事有明确验收标准
- [ ] 非功能性需求已量化
- [ ] PR 已合并到 develop
- [ ] 已通知设计师启动阶段二

---

## 常见问题

**Q：需求不清晰怎么办？**
> 继续和 Claude Code 对话，它会主动追问补充信息，直到需求足够清晰。

**Q：干系人意见不一致？**
> 在 PRD 中标注「待确认」，创建 GitHub Issue 跟踪决策项。

**Q：发现新需求怎么办？**
> P2 以下：记录到 backlog，不纳入当前迭代；P0/P1：更新 PRD 并重新评审。
