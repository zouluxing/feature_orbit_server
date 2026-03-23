# UMS 测试报告 v1.0.0-rc1

**测试日期**：2026-03-22

## 执行摘要

| 指标 | 数值 |
|------|------|
| 总用例数 | 37 |
| 通过 | 37 |
| 失败 | 0 |
| 单元测试覆盖率 | 83.3% |
| 安全检查项 | 7/7 |

## 单元测试结果

- auth_service_test.go: 13/13 PASS
- user_service_test.go: 19/19 PASS
- permission_service_test.go: 9/9 PASS (including RBAC checks)
- jwt_utils_test.go: 8/8 PASS

## 安全专项检查

| 检查项 | 状态 |
|--------|------|
| SQL 注入防护 | ✅ |
| JWT 算法混淆 | ✅ |
| 密码 bcrypt(12) | ✅ |
| MFA Secret 加密 | ✅ |
| refresh_token 轮换 | ✅ |
| 登录限流 5次锁定 | ✅ |
| PKCE S256 强制 | ✅ |
