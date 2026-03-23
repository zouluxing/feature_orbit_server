# Feature Orbit Server — Claude Code 主配置

## 项目说明
本项目使用 Claude Code Skills 实现分阶段、分角色的软件开发流程自动化。

---

## ⚡ 强制前置规则：环境检查

> **所有阶段任务开始前，必须先执行环境检查。**

```powershell
PowerShell -ExecutionPolicy Bypass -File scripts/setup-env.ps1
```

---

## 🏗️ UMS 架构决策（ADR — 已在 develop 落地，不得推翻）

### ADR-001: JWT 本地验证（零 UMS 回调）
业务系统通过 `/.well-known/jwks.json` 缓存 RSA 公钥，本地验证 JWT。
实现：`internal/middleware/ums.go` — UMSClient，JWKS 缓存 TTL=1h，降级服务旧缓存。

### ADR-002: Token 策略
- access_token TTL = **15分钟**（RS256）
- refresh_token TTL = **7天**，使用后自动轮换（rotation）
- Logout 写 JTI 至 Redis 黑名单

### ADR-003: OAuth2 PKCE 强制
授权码流程强制 S256 code_challenge。

### ADR-004: RBAC
`users → user_roles → roles → role_permissions → permissions`
内置角色：admin / editor / viewer（is_system=true）

### ADR-005: 主服务认证委托
**feature_orbit_server 自身不实现任何认证逻辑。**
```go
umsClient, _ := middleware.NewUMSClient(cfg.UMSBaseURL, cfg.UMSCacheTTL)
router.Use(umsClient.GinMiddleware())
claims, _ := middleware.ClaimsFrom(c)
```

### ADR-006: MFA Secret AES-256-GCM 加密（RISK-001 修复）
生产环境强制配置 `UMS_MFA_ENCRYPTION_KEY`（64 hex chars）。
未配置时生产模式启动失败，开发模式降级明文并打 WARN。
```bash
export UMS_MFA_ENCRYPTION_KEY=$(openssl rand -hex 32)
```

---

## 📦 服务端口规划

| 服务 | 端口 | 数据库 |
|------|------|--------|
| feature_orbit_server | **8080** | PostgreSQL:5432 |
| UMS | **8081** | PostgreSQL:**5433** (ums_db) |
| UMS Redis | **6380** | DB index 1 |

---

## 📋 当前 develop 分支内容清单（v1.1.0-rc1）

### UMS 子服务 (`ums/`)
- `cmd/server/main.go` — 入口，DI，MFA 加密器初始化
- `internal/config/` — Viper 配置加载
- `internal/model/` — User / Role / Permission / OAuth2 模型
- `internal/repository/` — UserRepo / RoleRepo / OAuth2Repo
- `internal/service/` — Auth / User / Permission / OAuth2（含 MFA AES-GCM）
- `internal/handler/` — Auth / User / Permission / OAuth2 Handler
- `pkg/crypto/aes.go` — AES-256-GCM 加密工具
- `pkg/utils/jwt.go` — RS256 JWT 签发/验证/JWKS
- `pkg/umsclient/` — Go SDK（业务系统一行接入）
- `migrations/` — PostgreSQL DDL + seed

### 主服务 (`./`)
- `cmd/server/main.go` — 入口，接入 UMS 中间件，Feature CRUD
- `internal/middleware/ums.go` — UMS JWT 中间件
- `internal/model/feature.go` — Feature 模型
- `internal/repository/feature.go` — FeatureRepository
- `internal/service/feature.go` — FeatureService（slug唯一、owner权限）
- `internal/handler/feature.go` — 5个 Feature HTTP Handler
- `internal/dto/feature.go` — DTO
- `migrations/001_create_features.up.sql` — Feature 表 DDL

### CI/CD & 文档
- `.github/workflows/ums-ci.yml` — 覆盖率门禁 ≥80%
- `.github/workflows/app-ci.yml` — 主服务 CI
- `.github/workflows/integration-test.yml` — 每日集成测试
- `docs/design/` — 系统设计文档
- `docs/testing/` — 测试计划、测试报告、QA 验收报告
- `docs/deployment/runbook.md` — 运维手册

---

## Skills 目录

```
.agents/skills/
├── env-setup.md
├── requirements-engineer.md
├── designer.md
├── developer.md
├── tester.md
├── qa-engineer.md
└── devops-engineer.md
```

## 全局规则

1. 环境检查前置
2. 严格按阶段顺序执行
3. 每个阶段完成后创建 PR
4. 代码提交遵循 Conventional Commits 规范
5. 后续任务自动连续执行，无需等待用户确认
6. **所有 ADR-001~006 已落地，后续会话继承，不得推翻**
