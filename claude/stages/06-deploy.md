# 阶段六：部署上线

## 阶段目标
将经过完整验收的代码安全部署到生产环境，配置监控告警，确保服务稳定运行。

## 执行角色
**运维工程师** → 加载角色：`claude/roles/devops-engineer.md`

## 分支策略
```bash
# 将 develop 合并到 main
git checkout main
git merge develop
git tag v<版本号>
git push origin main --tags
```

## 阶段步骤

```
Step 1 ──► Step 2 ──► Step 3 ──► Step 4 ──► Step 5 ──► Step 6
部署准备    CI/CD配置   环境配置    执行部署    上线验证    监控收尾
```

## 入口条件
- 阶段五 QA 报告通过
- 生产环境资源已就绪
- 回滚方案已准备

## 出口条件
- [ ] 生产环境服务正常运行
- [ ] 冒烟测试全部通过
- [ ] 监控告警已配置
- [ ] 上线报告已发布
- [ ] `develop` 已合并到 `main` 并打 Tag

## 迭代完成
本次迭代正式完成，归档文档，准备下一迭代需求收集。
