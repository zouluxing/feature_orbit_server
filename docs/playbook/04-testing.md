# 阶段四：测试验证 实施手册

**执行角色**：测试工程师
**Skill 文件**：`.claude/skills/tester.md`
**目标分支**：`stage/testing`
**前置条件**：所有 feature PR 已合并到 develop，测试环境已部署最新代码

---

## Step 1：启动并激活测试工程师角色

```bash
claude

> 开发阶段已完成，请以测试工程师身份开始测试验证阶段
```

Claude Code 自动读取 PRD、用户故事和 API 设计文档。

```bash
git checkout develop && git pull
git checkout -b stage/testing
```

---

## Step 2：制定测试计划

```
> 请根据用户故事和 API 设计，制定测试计划，生成 docs/testing/test-plan.md
```

**测试计划必须包含**：

```markdown
# 测试计划

## 测试范围
- 纳入范围：所有 P0/P1 用户故事
- 排除范围：P2 功能（下个迭代）

## 测试类型
| 类型 | 工具 | 负责人 |
|------|------|--------|
| 功能测试 | 手动 + 自动化 | 测试工程师 |
| 接口测试 | Postman / curl | 测试工程师 |
| 性能测试 | k6 / JMeter | 测试工程师 |
| 回归测试 | 自动化脚本 | 测试工程师 |

## 测试环境
- 测试地址：http://test.xxx.com
- 数据库：测试库（独立数据）

## 通过标准
- P0/P1 Bug：0 个未关闭
- 测试用例通过率：≥ 95%
- 性能指标达标
```

---

## Step 3：设计测试用例

```
> 请为每条 P0 用户故事设计测试用例，包括正向、逆向和边界用例，
  生成 docs/testing/test-cases.md
```

**用例设计原则（Claude Code 自动应用）**：

| 用例类型 | 说明 | 示例 |
|----------|------|------|
| 正向用例 | 正常流程验证 | 正确账密登录 → 成功 |
| 逆向用例 | 错误输入验证 | 错误密码 → 正确报错 |
| 边界用例 | 极限值验证 | 密码长度恰好100字符 |
| 回归用例 | 已有功能不受影响 | 旧接口仍正常工作 |

---

## Step 4：执行功能测试

```
> 请按测试用例逐条执行测试，记录结果
```

Claude Code 辅助执行测试，自动记录结果：

```bash
# 接口测试示例
curl -X POST http://test.xxx.com/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@test.com","password":"wrong"}'

# 期望：{"code": 40001, "message": "账号或密码错误"}
# 实际：[填写实际结果]
```

**发现 Bug 立即创建 GitHub Issue**：

```markdown
## [用户认证] 连续失败后账号未锁定

**严重程度**：P1 严重
**复现概率**：必现

**复现步骤**：
1. 使用错误密码登录 5 次
2. 第 6 次再次尝试

**期望结果**：返回账号锁定提示
**实际结果**：仍然返回密码错误，未锁定

**截图/日志**：[附件]
```

---

## Step 5：接口自动化测试

```
> 请针对所有 API 接口编写自动化测试脚本，验证接口与设计文档一致
```

```bash
# 运行接口自动化测试
make api-test

# 验证覆盖项
# ✅ 所有接口响应格式与 API 设计文档一致
# ✅ 异常参数处理正确
# ✅ 权限控制有效（401/403 场景）
# ✅ 并发请求数据一致性
```

---

## Step 6：性能测试

```
> 请根据 PRD 中的非功能性需求，执行性能测试，生成 docs/testing/performance-report.md
```

```bash
# 使用 k6 执行性能测试
k6 run tests/performance/load-test.js

# 测试场景
# - 正常负载：50 并发，持续 5 分钟
# - 峰值负载：200 并发，持续 1 分钟
# - 压力测试：逐步加压到系统限制
```

**性能指标记录**：

| 指标 | 目标值 | 实测值 | 是否达标 |
|------|--------|--------|----------|
| P99 响应时间 | < 500ms | | |
| QPS | > 1000 | | |
| 错误率 | < 0.1% | | |
| CPU 使用率 | < 70% | | |

---

## Step 7：生成测试报告

```
> 请汇总测试结果，生成完整的测试报告 docs/testing/test-report.md
```

```bash
git add docs/testing/
git commit -m "docs(testing): 完成测试验证，输出测试报告"
git push origin stage/testing
# 创建 PR: stage/testing → develop
```

---

## Bug 跟踪流程

```
发现 Bug → 创建 Issue（含复现步骤）
    ↓
开发工程师认领修复（新建 fix/xxx 分支）
    ↓
修复完成 → PR 合并到 develop
    ↓
测试工程师回归验证 → 关闭 Issue
    ↓
所有 P0/P1 Bug 关闭 → 测试通过
```

---

## 阶段完成标准

- [ ] P0/P1 Bug 全部修复并回归通过
- [ ] 测试用例通过率 ≥ 95%
- [ ] 性能指标达到设计要求
- [ ] 测试报告已提交
- [ ] PR 已合并到 develop
- [ ] 已通知 QA 工程师启动阶段五
