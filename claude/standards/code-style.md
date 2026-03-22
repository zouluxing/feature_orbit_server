# 代码规范

## 提交信息规范（Conventional Commits）

```
<type>(<scope>): <subject>

[optional body]
[optional footer]
```

### Type 类型
| Type | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `style` | 代码格式（不影响逻辑）|
| `refactor` | 代码重构 |
| `test` | 测试相关 |
| `chore` | 构建/工具链/依赖更新 |
| `perf` | 性能优化 |
| `ci` | CI/CD 相关 |

### 示例
```
feat(auth): 新增 JWT 登录功能
fix(api): 修复用户列表分页错误
docs(readme): 更新部署文档
```

---

## 分支命名规范

| 分支类型 | 命名格式 | 示例 |
|----------|----------|------|
| 功能开发 | `feature/<功能名>` | `feature/user-auth` |
| Bug 修复 | `fix/<问题描述>` | `fix/login-redirect` |
| 阶段分支 | `stage/<阶段名>` | `stage/requirements` |
| 热修复 | `hotfix/<问题>` | `hotfix/payment-error` |
| 发布 | `release/<版本>` | `release/v1.0.0` |

---

## PR 规范

```markdown
## 变更说明
<!-- 简述本次改动 -->

## 关联 Issue
Closes #<issue编号>

## 改动类型
- [ ] 新功能
- [ ] Bug 修复
- [ ] 重构
- [ ] 文档

## 测试方法
<!-- 如何验证本次改动 -->

## 截图（如有）
```
