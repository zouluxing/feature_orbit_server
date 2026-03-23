# UMS UI 原型说明

## 核心页面

1. 登录页 /login
2. OAuth2 授权页 /oauth2/authorize
3. 用户管理后台 /admin/users
4. 个人设置 /settings（含 MFA 开关）

## 页面跳转逻辑

```
未登录 → 访问保护页面 → 重定向 /login → 登录成功 → 跳回

OAuth2 三方应用 → /oauth2/authorize?client_id=... → 登录 → 授权确认 → 302 回调
```
