# UMS API 接口设计

## 基础规范

| 项目 | 规范 |
|------|------|
| Base URL | https://ums.yourdomain.com/api/v1 |
| 认证方式 | Authorization: Bearer <access_token>（JWT RS256）|
| 响应格式 | `{ "code": 0, "message": "ok", "data": {} }` |
| Token TTL | access_token: 15 分钟 / refresh_token: 7 天 |

## 错误码

| code | HTTP | 说明 |
|------|------|------|
| 0 | 200 | 成功 |
| 40001 | 401 | 账号或密码错误 |
| 40004 | 401 | Token 无效或已吊销 |
| 40005 | 403 | 权限不足 |
| 40007 | 409 | 邮箱或用户名已存在 |
| 50001 | 500 | 服务内部错误 |

## 认证接口

- POST /auth/register
- POST /auth/login
- POST /auth/token/refresh
- POST /auth/logout
- GET /.well-known/jwks.json

## 用户管理（需 admin）

- GET/POST /users
- GET/PUT/DELETE /users/:uuid
- PATCH /users/:uuid/status
- POST/DELETE /users/:uuid/roles

## OAuth2.0

- GET /oauth2/authorize
- POST /oauth2/token
- POST /oauth2/introspect
- GET/POST/PUT/DELETE /oauth2/clients

## 业务系统快速接入（Go SDK）

```go
client, _ := umsclient.New("https://ums.yourdomain.com", umsclient.Options{})
router.Use(client.GinMiddleware())
router.Use(client.RequireRole("admin"))
claims, _ := umsclient.ClaimsFrom(c)
```
