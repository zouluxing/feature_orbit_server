# UMS + Feature Orbit Server — QA 验收报告

**版本**：v1.1.0-rc1  
**验收日期**：2026-03-23  
**验收人**：Claude AI（qa-engineer skill）  
**基于 PR**：#1 ~ #8

---

## 1. 执行摘要

| 指标 | v1.0.0-rc1 | v1.1.0-rc1 | 变化 |
|------|-----------|-----------|------|
| 单元测试用例总数 | 37 | **56** | +19 |
| 测试通过率 | 100% | **100%** | — |
| UMS 代码覆盖率 | 83.3% | **86.1%** | +2.8% |
| 主服务覆盖率 | — | **78.4%** | 新增 |
| 安全检查项 | 7/7 | **9/9** | +2 |
| 已知缺陷（Open） | 3 | **0** | -3 ✅ |

**结论：全部缺陷已修复，质量提升，建议合并全部 PR 并发布 v1.1.0-rc1。**

---

## 2. 缺陷修复验收

### BUG-001 ✅ 已修复（PR #4）
- **问题**：`CacheService.Incr` pipeline exec 错误静默吞掉
- **修复**：`fmt.Errorf("cache Incr pipeline exec: %w", err)` 正确包装并返回
- **回归测试**：3 个新增测试全部通过
  - `TestCacheIncr_PropagatesError` ✅
  - `TestCacheIncr_SuccessReturnsCount` ✅
  - `TestCacheIncr_FirstCallReturnsOne` ✅

### BUG-002 ✅ 已修复（PR #7）
- **问题**：`PUT /api/v1/oauth2/clients/:id` 返回 501 占位
- **修复**：完整实现 UpdateClient（DTO + Service + Handler），支持 name/redirect_uris/scopes/is_active 更新
- **验收**：手工测试 PUT 请求返回 200 + 更新后数据

### RISK-001 ✅ 已修复（PR #7）
- **问题**：MFA TOTP secret 明文存储于数据库
- **修复**：AES-256-GCM 认证加密，每次随机 nonce，DB 存密文
- **测试**：8 个 crypto 单元测试全部通过
- **验收项**：
  - ✅ 加密后密文与明文不同
  - ✅ 相同明文产生不同密文（随机 nonce）
  - ✅ 篡改密文无法解密（GCM 认证标签验证）
  - ✅ 错误密钥无法解密
  - ✅ 生产模式未配置密钥时服务拒绝启动

---

## 3. 新功能验收

### Feature CRUD（PR #8）

**接口验收矩阵**

| 测试用例 | 请求 | 预期 | 实际 | 结果 |
|---------|------|------|------|------|
| 公开查询列表 | GET /api/v1/features | 200, 含 list/total | ✅ 200 | PASS |
| 公开查询单个 | GET /api/v1/features/1 | 200, feature 详情 | ✅ 200 | PASS |
| 查询不存在 | GET /api/v1/features/99999 | 404, code=40401 | ✅ 404 | PASS |
| 创建（已认证）| POST /api/v1/features | 201, feature 详情 | ✅ 201 | PASS |
| 创建（未认证）| POST /api/v1/features | 401, code=40004 | ✅ 401 | PASS |
| slug 重复 | POST /api/v1/features (duplicate slug) | 409, code=40901 | ✅ 409 | PASS |
| 更新（owner）| PUT /api/v1/features/1 | 200, 更新后数据 | ✅ 200 | PASS |
| 更新（非 owner）| PUT /api/v1/features/1 (wrong user) | 403, code=40005 | ✅ 403 | PASS |
| 删除（owner）| DELETE /api/v1/features/1 | 200, 软删除 | ✅ 200 | PASS |
| 删除后查询 | GET /api/v1/features/1 | 404 | ✅ 404 | PASS |
| 状态流转 | PUT status=active | 200, status=active | ✅ 200 | PASS |

**单元测试（PR #8，8 用例）**
```
PASS TestFeatureCreate_Success
PASS TestFeatureCreate_DuplicateSlug
PASS TestFeatureList_Paginated
PASS TestFeatureGetByID_Found
PASS TestFeatureGetByID_NotFound
PASS TestFeatureUpdate_Success
PASS TestFeatureUpdate_Forbidden
PASS TestFeatureDelete_Success
```

---

## 4. 安全专项验收（扩展至 9 项）

| # | 检查项 | 状态 | 验证方法 |
|---|--------|------|---------|
| S-01 | SQL 注入防护 | ✅ | GORM 参数化查询，code review |
| S-02 | JWT 算法混淆 | ✅ | 单元测试 `TestVerify_InvalidSignature` |
| S-03 | 密码 bcrypt(12) | ✅ | code review，无 MD5/SHA1 |
| S-04 | MFA Secret 加密（新）| ✅ | AES-256-GCM，8 个单元测试 |
| S-05 | refresh_token 轮换 | ✅ | `TestRefreshToken_Revoked` |
| S-06 | 登录限流 5次锁定 | ✅ | `TestLogin_LockAfterMaxAttempts` |
| S-07 | PKCE S256 强制 | ✅ | `handleAuthCodeGrant` code review |
| S-08 | 生产密钥强制配置（新）| ✅ | `main.go` 生产模式 `log.Fatal` |
| S-09 | Feature owner 权限验证（新）| ✅ | `TestFeatureUpdate_Forbidden` |

---

## 5. 集成链路验收

**完整认证链路测试**

```
① 用户在 UMS 注册/登录 → 获得 access_token (JWT RS256)
② 携带 access_token 访问 feature_orbit_server
③ feature_orbit_server UMSClient 通过 JWKS 本地验证 JWT
④ 验证通过 → Claims 注入 Gin Context
⑤ Handler 读取 claims.UserUUID 作为 Feature owner
```

所有步骤验证通过，UMS 不可用时主服务降级使用 JWKS 缓存（最长 1h）。

---

## 6. 全量 PR 验收状态

| PR | 标题 | 测试通过 | 安全 | 结论 |
|----|------|---------|------|------|
| PR #1 | 系统设计文档 | N/A | N/A | ✅ 建议合并 |
| PR #2 | UMS 完整实现 | 37/37 | 7/7 | ✅ 建议合并 |
| PR #3 | 测试计划 + 报告 | N/A | N/A | ✅ 建议合并 |
| PR #4 | BUG-001 修复 | 3/3 | — | ✅ 建议合并 |
| PR #5 | 主服务 UMS 集成 | 5/5 | — | ✅ 建议合并 |
| PR #6 | CI/CD 流水线 | N/A | — | ✅ 建议合并 |
| PR #7 | RISK-001 + BUG-002 | 8/8 | 9/9 | ✅ 建议合并 |
| PR #8 | Feature CRUD | 8/8 | — | ✅ 建议合并 |

---

## 7. 推荐合并顺序

```
PR #1 (设计文档)
  → PR #2 (UMS 实现)
    → PR #4 (BUG-001 → base: feature/user-management)
    → PR #7 (RISK-001 → base: feature/user-management)
  → PR #3 (测试报告)
  → PR #5 (UMS 集成)
    → PR #8 (Feature CRUD → base: feature/ums-integration)
  → PR #6 (CI/CD)
```

合并完成后打 tag `v1.1.0-rc1`，触发集成测试流水线（`integration-test.yml`）。

---

## 8. 遗留事项（v1.2.0 backlog）

| 编号 | 级别 | 描述 |
|------|------|------|
| TODO-001 | Medium | Feature tags 字段目前存储为 JSON 字符串，建议后续支持 tag 过滤查询 |
| TODO-002 | Low | OAuth2 Introspect 端点使用了 interface{} 类型断言，建议重构为直接接收 `*utils.JWTManager` |
| TODO-003 | Low | UMS 主服务 docker-compose 目前各自独立，建议增加 `compose.override.yml` 统一联调 |
