# 阶段六：部署上线 实施手册

**执行角色**：运维工程师
**Skill 文件**：`.claude/skills/devops-engineer.md`
**目标分支**：`main`（从 develop 合并）
**前置条件**：QA 报告结论为「允许上线」，生产环境资源就绪

---

## Step 1：启动并激活运维工程师角色

```bash
claude

> QA 阶段已通过，请以运维工程师身份开始部署上线阶段
```

Claude Code 自动读取 QA 报告和回滚方案。

---

## Step 2：部署前检查清单

```
> 请执行部署前检查，确认所有前置条件已就绪
```

```bash
# Claude Code 逐项执行检查

# 1. 确认 QA 报告通过
cat docs/qa/qa-report.md | grep "上线建议"

# 2. 检查生产环境资源
ssh prod-server 'df -h && free -h && uptime'

# 3. 备份生产数据
make backup-production
echo "备份完成时间: $(date)" >> docs/deploy/rollback-plan.md

# 4. 记录当前生产版本（用于回滚）
current_sha=$(ssh prod-server 'cat /app/current_version')
echo "回滚版本: $current_sha" >> docs/deploy/rollback-plan.md

# 5. 通知干系人（发送上线公告）
echo "预计上线时间：$(date -d '+30 minutes')，影响范围：xxx"
```

**检查矩阵**：

| 检查项 | 状态 |
|--------|------|
| QA 报告「允许上线」| ⬜ |
| 生产数据已备份 | ⬜ |
| 回滚版本已记录 | ⬜ |
| 生产环境磁盘空间 > 30% | ⬜ |
| 第三方服务正常 | ⬜ |
| 干系人已通知 | ⬜ |

---

## Step 3：配置 CI/CD 流水线

```
> 请检查并更新 CI/CD 流水线配置，确保部署流程自动化
```

Claude Code 创建/更新 `.github/workflows/deploy.yml`：

```yaml
name: Deploy to Production

on:
  push:
    branches: [main]  # 合并到 main 自动触发
  workflow_dispatch:  # 支持手动触发

jobs:
  # 阶段1：运行测试
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Tests
        run: make test
      - name: Check Coverage
        run: make coverage

  # 阶段2：构建镜像
  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build & Push Docker Image
        run: |
          docker build -t $REGISTRY/app:${{ github.sha }} .
          docker push $REGISTRY/app:${{ github.sha }}

  # 阶段3：部署到生产
  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment: production  # 需要审批才能部署
    steps:
      - name: Deploy
        run: |
          ssh $PROD_SERVER \
            "docker pull $REGISTRY/app:${{ github.sha }} && \
             docker-compose up -d --no-deps app"
      - name: Smoke Test
        run: make smoke-test ENV=production
```

---

## Step 4：Staging 环境预验证

```bash
# 先在 Staging 环境跑一遍完整部署流程
make deploy ENV=staging
make smoke-test ENV=staging

# 执行数据库迁移（先在 Staging 验证）
make db-migrate ENV=staging

# 确认 Staging 一切正常后再部署生产
```

---

## Step 5：执行生产部署

```
> Staging 验证通过，请执行生产环境部署
```

```bash
# 1. 将 develop 合并到 main
git checkout main
git merge develop
git tag v1.0.0  # 打版本 Tag
git push origin main --tags

# 2. CI/CD 自动触发（或手动）
# GitHub Actions 检测到 main 分支推送，自动执行部署

# 3. 灰度发布（如支持）
# 第一步：10% 流量 → 观察 5 分钟
# 第二步：50% 流量 → 观察 5 分钟
# 第三步：100% 流量 → 全量上线
```

---

## Step 6：上线验证

```
> 请执行上线后冒烟测试，验证生产环境正常
```

```bash
# 自动化冒烟测试
make smoke-test ENV=production

# 手动验证核心功能（5分钟快速验证）
# ✅ 首页正常加载
# ✅ 用户登录/注册
# ✅ 核心业务流程
# ✅ 支付/关键操作（如有）

# 检查错误日志（部署后前5分钟重点监控）
tail -f /var/log/app/error.log

# 检查关键指标
curl http://prod.xxx.com/health
```

**如发现严重问题立即执行回滚**：

```bash
# 回滚命令（< 5 分钟恢复）
make rollback VERSION=$previous_sha ENV=production

# 回滚后验证
make smoke-test ENV=production
```

---

## Step 7：监控配置与收尾

```
> 请配置监控告警，完成上线收尾工作
```

**监控告警配置**：

| 告警项 | 阈值 | 告警方式 |
|--------|------|----------|
| 错误率 | > 1% | 立即通知 + 准备回滚 |
| P99 响应时间 | > 1s | 通知告警 |
| 服务不可用 | uptime < 99.9% | 立即通知 + 自动重启 |
| CPU 使用率 | > 85% | 通知告警 |
| 磁盘使用率 | > 80% | 通知告警 |

**上线报告生成**：

```
> 请生成上线报告 docs/deploy/release-note.md
```

```markdown
# 上线发布说明 v1.0.0

## 发布时间：2026-03-22 15:00
## 发布人：运维工程师

## 发布内容
- feat: 用户认证模块（US-001, US-002）
- feat: 核心业务功能（US-003 ~ US-008）
- perf: 接口性能优化

## 影响范围
- 影响模块：xxx
- 影响用户：全量
- 预计停服时间：0（滚动部署，无停服）

## 验证结果
- 冒烟测试：✅ 全部通过
- 监控状态：✅ 正常

## 回滚方案
- 回滚命令：`make rollback VERSION=<上一版本>`
- 回滚时间：< 5 分钟
- 决策人：xxx
```

```bash
git add .github/workflows/ docs/deploy/
git commit -m "ci: 完成生产环境部署，配置监控告警"
git push origin main
```

---

## 部署后 24 小时观察期

部署完成后持续观察 24 小时：

```
前 1 小时：每 10 分钟检查一次错误日志和关键指标
1~8 小时：每 30 分钟检查一次
8~24 小时：每小时检查一次

如发现异常：立即评估是否需要回滚
24 小时无问题：本次迭代正式完成 🎉
```

---

## 阶段完成标准

- [ ] 生产环境服务正常运行
- [ ] 冒烟测试全部通过
- [ ] 监控告警已配置
- [ ] 上线报告已发布
- [ ] develop 已合并到 main 并打 Tag
- [ ] 24 小时观察期无重大问题
- [ ] **本次迭代正式完成** 🎉

---

## 完成后归档

```bash
# 整理本次迭代所有文档
git tag -a v1.0.0 -m "Release v1.0.0: 完成核心功能开发"
git push origin --tags

# 归档到 docs/archive/v1.0.0/
# 准备下一迭代需求收集
```
