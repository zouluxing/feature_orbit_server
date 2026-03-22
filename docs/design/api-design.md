# UMS API 接口设计

## 基础规范

| 项目 | 规范 |
|------|------|
| Base URL | https://ums.yourdomain.com/api/v1 |
| 认证方式 | Authorization: Bearer <access_token>（JWT RS256）|
| 响应格式 | `{ "code": 0, "message": "ok", "data": {} }` |
| 分页格式 | `{ "list": [], "total": 100, "page": 1, "page_size": 20 }` |
| Token TTL | access_token: 15 分钟 \| refresh_token: 7 天 |

## 错误码

| code | HTTP | 说明 |
|------|------|------|
| 0 | 200 | 成功 |
| 40001 | 401 | 账号或密码错误 |
| 40002 | 401 | 账号已锁定 |
| 40003 | 401 | Token 已过期 |
| 40004 | 401 | Token 无效或已吊销 |
| 40005 | 403 | 权限不足 |
| 40006 | 400 | MFA 验证码错误 |
| 40007 | 409 | 邮箱或用户名已存在 |
| 40008 | 404 | 用户不存在 |
| 40009 | 400 | OAuth2 client_id 无效 |
| 40010 | 400 | redirect_uri 不在白名单 |
| 50001 | 500 | 服务内部错误 |

## 认证接口

### POST /auth/register
```json
请求: { "email": "string", "username": "string", "password": "string" }
响应: { "code": 0, "data": { "user_uuid": "uuid", "email": "string" } }
```

### POST /auth/login
```json
请求: { "email": "string", "password": "string", "mfa_code": "string(可选)" }
响应: { "code": 0, "data": { "access_token": "JWT", "refresh_token": "string", "expires_in": 900, "token_type": "Bearer" } }
```

### POST /auth/token/refresh
```json
请求: { "refresh_token": "string" }
响应: { "code": 0, "data": { "access_token": "JWT", "expires_in": 900 } }
```

### POST /auth/logout
Header: Authorization: Bearer <token>，将 JTI 写入 Redis 黑名单。

### GET /.well-known/jwks.json
返回 RSA 公钥 JWK Set，无需鉴权。业务系统启动时缓存用于本地验证。

## 用户管理接口（需 admin 角色）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /users | 分页查询，支持 email/status 过滤 |
| GET | /users/:uuid | 查询单个用户详情 |
| POST | /users | 创建用户 |
| PUT | /users/:uuid | 更新用户信息 |
| PATCH | /users/:uuid/status | 启用/禁用/解锁 |
| DELETE | /users/:uuid | 软删除 |
| PUT | /users/:uuid/password | 重置密码 |
| POST | /users/:uuid/roles | 分配角色 |
| DELETE | /users/:uuid/roles/:role_id | 移除角色 |
| GET | /users/:uuid/permissions | 查询展开后的权限集 |

## 当前用户接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /users/me | 获取当前用户信息 |
| PUT | /users/me | 更新 profile |
| PUT | /users/me/password | 修改密码 |
| POST | /users/me/mfa/enable | 开启 MFA，返回 TOTP URI |
| POST | /users/me/mfa/disable | 关闭 MFA |

## OAuth 2.0 接口（RFC 6749）

### GET /oauth2/authorize
参数：client_id, response_type=code, redirect_uri, scope, state, code_challenge(PKCE)

### POST /oauth2/token
授权码换 Token：`grant_type=authorization_code&code=...&redirect_uri=...&client_id=...&client_secret=...`  
刷新 Token：`grant_type=refresh_token&refresh_token=...&client_id=...&client_secret=...`

### POST /oauth2/introspect
实时查询 token 有效性（不支持本地验证的系统使用）

### OAuth Client 管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /oauth2/clients | 查询我的应用列表 |
| POST | /oauth2/clients | 注册新应用（返回 secret 仅显示一次）|
| PUT | /oauth2/clients/:id | 更新应用信息 |
| DELETE | /oauth2/clients/:id | 删除应用 |
| POST | /oauth2/clients/:id/secret/rotate | 轮换 client_secret |

## 业务系统快速接入（Go SDK）

```go
// Step 1: 初始化（应用启动时执行一次）
client, err := umsclient.New("https://ums.yourdomain.com", umsclient.Options{})

// Step 2: 注册 Gin 中间件
router.Use(client.GinMiddleware())

// Step 3: 可选 - 角色守卫
router.GET("/admin", client.RequireRole("admin"), adminHandler)

// Step 4: 在 Handler 中读取用户信息
claims, _ := umsclient.ClaimsFrom(c)
fmt.Println(claims.UserUUID, claims.Roles)
```
