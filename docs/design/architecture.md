# UMS 系统架构设计

## 1. 整体架构

UMS（User Management Service）以独立微服务形式部署，向所有业务系统提供统一的身份认证、用户 CRUD、RBAC 权限模型以及 OAuth 2.0 授权服务器能力。

```
┌─────────────────────────────────────────┐  ┌───────────────────────────┐
│           UMS Service (:8081)           │  │   Business Systems        │
│                                         │  │                           │
│  API Gateway / Rate Limiter             │  │  feature_orbit_server     │
│       ↓              ↓                  │  │  (:8080)                  │
│  Auth Service   User Service            │  │  ┌─────────────────────┐  │
│  ↓                   ↓                  │◄─┤  │  JWT Middleware      │  │
│  Permission     OAuth2 Service          │  │  │  (local verify)      │  │
│  Service             ↓                  │  │  └─────────────────────┘  │
│       ↓         ┌────┴────┐             │  │                           │
│  PostgreSQL    Redis       JWKS          │  │  Third-party Apps         │
│  (ums_db)    (cache)    Endpoint        │  │  (OAuth2 clients)         │
└─────────────────────────────────────────┘  └───────────────────────────┘
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

## 4. 部署架构

- UMS 端口：8081（独立，与业务服务隔离）
- PostgreSQL：5433（独立 ums_db 数据库）
- Redis：6380（独立，DB index 1）
- 业务服务通过 Docker network 访问 UMS

## 5. 关键设计决策（ADR）

**ADR-001: JWT 本地验证**  
业务系统通过 JWKS 端点缓存 RSA 公钥，本地验证 access_token，不回调 UMS，延迟为 0。仅 token 吊销场景需查 Redis 黑名单。

**ADR-002: access_token TTL = 15 分钟**  
短 TTL 限制泄露影响窗口；refresh_token 7 天，使用后自动轮换（rotation），防重放攻击。

**ADR-003: PKCE 支持**  
OAuth2 授权码流程强制支持 S256 PKCE，防止授权码截获攻击，适用于 SPA 和移动端。
