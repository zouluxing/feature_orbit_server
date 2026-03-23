# UMS 测试计划

## 1. 测试范围

| 层级 | 类型 | 工具 | 目标覆盖率 |
|------|------|------|-----------|
| Service 层 | 单元测试 | testify/mock | ≥ 80% |
| Handler 层 | 集成测试（httptest） | gin + httptest | 核心路径全覆盖 |
| API 层 | E2E 冒烟测试 | curl / httpie | P0 用户故事 |
| 安全专项 | 手工 + 自动化 | OWASP 检查单 | 100% 检查项 |

## 2. 测试分层说明

### 2.1 单元测试（Unit Tests）

- **目标**：每个 Service 方法在 Mock 依赖下独立验证
- **位置**：`ums/tests/unit/service/`
- **命名**：`Test<ServiceName>_<Scenario>`
- **执行**：`make test-unit`

### 2.2 集成测试（Integration Tests）

- **目标**：Handler → Service → Repository 端到端调用路径
- **依赖**：真实 PostgreSQL + Redis（Docker Compose）
- **执行**：`make services-up && make test`

### 2.3 E2E 冒烟测试

- **目标**：覆盖所有 P0 API 端点
- **工具**：脚本 `scripts/smoke_test.sh`
- **触发**：每次 PR 合并到 develop 后

## 3. 测试用例矩阵

### 认证模块

| ID | 测试点 | 前置条件 | 输入 | 预期结果 |
|----|--------|----------|------|----------|
| T-001 | 注册成功 | 邮箱未注册 | 合法 email/username/password | 201，返回 user_uuid |
| T-002 | 注册-邮箱重复 | 邮箱已存在 | 同邮箱 | 409，code=40007 |
| T-003 | 注册-密码过短 | — | password 长度 < 8 | 400 |
| T-004 | 登录成功 | 用户已注册、激活 | 正确凭证 | 200，返回 access_token + refresh_token |
| T-005 | 登录-密码错误 | 用户存在 | 错误密码 | 401，code=40001 |
| T-006 | 登录-账号锁定（5次失败） | 4次失败后 | 错误密码 | 401，code=40002 |
| T-007 | 登录-账号已禁用 | status=2 | 正确凭证 | 403，code=40005 |
| T-008 | Refresh Token 轮换 | 持有有效 refresh_token | 合法请求 | 200，新 access_token |
| T-009 | Refresh Token 已用（重放） | refresh_token 已用过 | 同一 refresh_token | 401，code=40004 |
| T-010 | Logout 黑名单 | 已登录 | 合法 Bearer | 200；黑名单后旧 token 401 |

### MFA 模块

| ID | 测试点 | 预期结果 |
|----|--------|----------|
| T-011 | 开启 MFA | 200，返回 secret + QR URI |
| T-012 | 重复开启 MFA | 400，code=40016 |
| T-013 | 登录时 MFA 必填 | 401，code=40015 |
| T-014 | 登录时 MFA 验证码错误 | 401，code=40006 |
| T-015 | 关闭 MFA（正确验证码）| 200 |

### 用户管理模块

| ID | 测试点 | 预期结果 |
|----|--------|----------|
| T-016 | 管理员创建用户 | 201 |
| T-017 | 非管理员创建用户 | 403 |
| T-018 | 分页列表查询 | 200，含 total/page/list |
| T-019 | 软删除用户 | 200；再查 404 |
| T-020 | 禁用/启用用户 | 200 |

### RBAC 模块

| ID | 测试点 | 预期结果 |
|----|--------|----------|
| T-021 | 创建自定义角色 | 201 |
| T-022 | 删除系统角色 | 403，code=40012 |
| T-023 | 给角色绑定权限 | 200 |
| T-024 | 给用户分配角色 | 200 |
| T-025 | 权限检查通过 | bool=true |
| T-026 | 权限检查拒绝 | bool=false |

### OAuth2 模块

| ID | 测试点 | 预期结果 |
|----|--------|----------|
| T-027 | 注册 OAuth2 Client | 201，返回 client_secret（一次性）|
| T-028 | 授权码流程（带 PKCE S256）| 302 redirect with code |
| T-029 | code 换 token | 200，access_token + refresh_token |
| T-030 | PKCE code_verifier 错误 | 400，code=40017 |
| T-031 | refresh_token 换新 token | 200 |
| T-032 | token introspect（有效）| active=true |
| T-033 | token introspect（无效）| active=false |

### JWKS / SDK 模块

| ID | 测试点 | 预期结果 |
|----|--------|----------|
| T-034 | GET /.well-known/jwks.json | 200，含 kty/n/e/kid |
| T-035 | umsclient 本地验证合法 token | claims 正确 |
| T-036 | umsclient 验证篡改 token | error |
| T-037 | umsclient JWKS 缓存 TTL 过期后自动刷新 | 正常验证 |

## 4. 安全专项检查

- [ ] SQL 注入：所有用户输入经 GORM 参数化查询
- [ ] XSS：JSON 响应无 HTML 输出
- [ ] JWT 算法混淆：拒绝 alg=none 及 HS256 请求
- [ ] 敏感字段：password_hash / mfa_secret 不出现在任何响应中
- [ ] refresh_token 存储：仅存 SHA256 哈希，原文不落库
- [ ] 登录限流：5次失败锁定账号 + Redis 窗口限流
- [ ] PKCE：授权码流程强制 S256，拒绝 plain

## 5. 性能基准目标

| 接口 | P99 延迟 | TPS（单节点）|
|------|---------|-------------|
| POST /auth/login | < 200ms | ≥ 200 |
| GET /users/me | < 50ms | ≥ 1000 |
| GET /.well-known/jwks.json | < 10ms | ≥ 5000 |

## 6. 执行计划

```
Day 1: 单元测试执行 + 覆盖率报告
Day 2: 集成测试环境搭建 + Handler 层测试
Day 3: E2E 冒烟测试 + 安全专项
Day 4: 性能基准测试
Day 5: 测试报告撰写 + 缺陷修复验证
```
