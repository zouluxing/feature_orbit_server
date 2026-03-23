# UMS 系统架构设计

## 1. 整体架构

UMS（User Management Service）以独立微服务形式部署，向所有业务系统提供统一的身份认证、用户 CRUD、RBAC 权限模型以及 OAuth 2.0 授权服务器能力。

```
┌─────────────────────────────────────────┐  ┌───────────────────────────┐
│           UMS Service (:8081)           │  │   Business Systems        │
│                                         │  │                           │
│  Auth Service   User Service            │  │  feature_orbit_server     │
│  Permission     OAuth2 Service          │◄─┤  (:8080)  JWT Middleware  │
│       ↓         ┌────半───┐             │  │  (local verify via JWKS)  │
│  PostgreSQL    Redis       JWKS          │  └───────────────────────────┘
└─────────────────────────────────────────┘
```

## 2. 模块划分

| 模块 | 路径 | 职责 |
|------|------|------|
| API Gateway | internal/handler/middleware.go | 认证中间件、限流、CORS |
| Auth Service | internal/service/auth.go | 登录、登出、MFA、Token 签发 |
| User Service | internal/service/user.go | 用户 CRUD、密码管理 |
| Permission Service | internal/service/permission.go | RBAC 角色与权限绑定 |
| OAuth2 Service | internal/service/oauth2.go | 授权码流程、Token 端点 |
| JWKS Endpoint | internal/handler/auth.go | RSA 公钥分发 |
| Go SDK | pkg/umsclient/client.go | 业务系统一行接入 |

## 3. 技术选型

| 层次 | 技术 | 理由 |
|------|------|------|
| HTTP 框架 | Gin v1.9+ | 高性能、中间件生态成熟 |
| ORM | GORM v2 | 类型安全、迁移友好 |
| JWT | golang-jwt/jwt v5 | RS256 支持，标准兼容 |
| 密码哈希 | bcrypt (cost=12) | 业界标准 |
| 缓存 | Redis 7 | Token 黑名单、会话存储 |
| 数据库 | PostgreSQL 16 | JSONB、事务完整性 |
| TOTP | pquerna/otp | RFC 6238 标准实现 |
| MFA 加密 | AES-256-GCM | 生产密文存储 |

## 4. 部署架构

| 服务 | 端口 | 说明 |
|------|------|------|
| feature_orbit_server | 8080 | 主业务服务 |
| UMS | 8081 | 独立认证服务 |
| PostgreSQL (ums_db) | 5433 | 独立数据库 |
| Redis | 6380 | DB index 1 |

## 5. ADR 关键设计决策

**ADR-001: JWT 本地验证**—业务系统通过 JWKS 缓存 RSA 公钥本地验证 access_token，不回调 UMS。

**ADR-002: Token 策略**—access_token 15min + refresh_token 7d 自动轮换。

**ADR-003: PKCE 强制**—OAuth2 授权码流程强制 S256 code_challenge。

**ADR-004: RBAC**—users → user_roles → roles → role_permissions → permissions。

**ADR-005: 主服务认证委托**—feature_orbit_server 自身不实现任何认证逻辑。
