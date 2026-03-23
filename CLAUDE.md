# Feature Orbit Server — Claude Code 主配置

## 项目说明
本项目使用 Claude Code Skills 实现分阶段、分角色的软件开发流程自动化。

---

## ⚡ 强制前置规则：环境检查

> **所有阶段任务开始前，必须先执行环境检查。**
> Claude Code 在接收到任何阶段指令时，自动加载 `.agents/skills/env-setup.md`，
> 完成全部检查项（E01~E09、G01~G07、S01~S07、W01~W05）并输出「环境准备就绪」后，才允许进入后续阶段。

**一键启动命令（每次开始工作前执行）：**

```powershell
# Windows PowerShell（主要开发平台）
PowerShell -ExecutionPolicy Bypass -File scripts/setup-env.ps1

# 或直接启动 Claude Code（自动触发环境检查）
claude
```

---

## 🏗️ UMS 架构决策（ADR — 已实施）

> **重要**：以下决策已在 PR #1~#5 中落地，后续所有 Claude 会话必须遵守，不得推翻。

### ADR-001: JWT 本地验证（无 UMS 回调）
- **决策**：business systems（如 feature_orbit_server）通过 `/.well-known/jwks.json` 获取 RSA 公钥，在本地验证 UMS 颁发的 JWT，**不回调 UMS**
- **实现**：`internal/middleware/ums.go` — `UMSClient` 本地缓存公钥，TTL=1h
- **理由**：零额外延迟；UMS 故障不影响已登录用户；水平扩展无状态

### ADR-002: Token 策略
- access_token TTL = **15分钟**（RS256 签名）
- refresh_token TTL = **7天**，使用后自动轮换（rotation）
- Logout 将 JTI 写入 Redis 黑名单（TTL=access_token TTL + 1min）

### ADR-003: OAuth2 PKCE 强制
- 授权码流程强制使用 S256 code_challenge
- 适用于所有 SPA 和移动端接入场景

### ADR-004: RBAC 实现
- 模型：`users → user_roles → roles → role_permissions → permissions`
- roles 内置：`admin` / `editor` / `viewer`（is_system=true，不可删除）
- 权限粒度：`resource:action`（如 `user:delete`、`oauth2_client:create`）

### ADR-005: 主服务认证委托
- **feature_orbit_server 自身不实现任何认证逻辑**
- 所有认证通过 `internal/middleware.UMSClient.GinMiddleware()` 完成
- Handler 通过 `middleware.ClaimsFrom(c)` 获取用户身份

---

## 📦 服务端口规划

| 服务 | 端口 | 数据库 | 备注 |
|------|------|--------|------|
| feature_orbit_server | **8080** | PostgreSQL:5432 | 主业务服务 |
| UMS | **8081** | PostgreSQL:**5433** | 独立数据库 ums_db |
| UMS Redis | **6380** | — | DB index 1，避免冲突 |

---

## 🔌 UMS 集成快速参考

```go
// 初始化（main.go 启动时执行一次）
umsClient, err := middleware.NewUMSClient("http://ums:8081", time.Hour)

// 注册中间件
router.Use(umsClient.GinMiddleware())           // 验证 JWT
router.Use(umsClient.RequireRole("admin"))       // 角色守卫（可选）

// Handler 中读取用户信息
claims, _ := middleware.ClaimsFrom(c)
fmt.Println(claims.UserUUID, claims.Email, claims.Roles)
```

---

## Skills 目录

所有角色均以 Skill 形式定义，放置于 `.agents/skills/` 目录下。

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
7. **后续任务自动连续执行**，无需等待用户确认（已由用户授权）

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
