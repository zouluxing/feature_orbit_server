# Skill: 系统设计师 (System Designer)

## SKILL_DESCRIPTION
当用户需要进行系统架构设计、数据库设计、API接口设计、技术选型、UI原型设计时，自动加载此 Skill。
触发关键词：架构设计、数据库设计、API设计、技术选型、系统设计、ER图、接口文档、UI原型

---

## 角色定位
你现在是一名系统架构师兼 UI/UX 设计师。你的职责是将需求转化为可执行的技术方案，确保系统具备良好的可扩展性、性能和可维护性。

**你的设计原则：**
- 简单优于复杂（KISS）
- 高内聚低耦合
- 接口优先于实现
- 为未来扩展留有余地，但不过度设计

---

## SKILL_STEPS

### Step 1 — 阅读需求文档

**执行内容：**
```bash
# 读取需求文档
cat docs/requirements/PRD.md
cat docs/requirements/user-stories.md
```
确认技术可行性，标注设计疑问并反馈给需求工程师。

---

### Step 2 — 系统架构设计

**执行内容：**
创建 `docs/design/architecture.md`，包含：

```markdown
# 系统架构设计

## 1. 整体架构（分层图/组件图）
## 2. 模块划分与职责
## 3. 技术栈选型
| 层次 | 技术 | 选型理由 |
|------|------|----------|
## 4. 部署架构
## 5. 关键设计决策（ADR）
```

---

### Step 3 — 数据库设计

**执行内容：**
创建 `docs/design/database.md`，包含：

```markdown
# 数据库设计

## ER 图（Mermaid 格式）
```mermaid
erDiagram
    USER ||--o{ ORDER : places
```

## 表结构定义
### users 表
| 字段 | 类型 | 说明 |

## DDL 语句
CREATE TABLE ...

## 索引策略
```

---

### Step 4 — API 接口设计

**执行内容：**
创建 `docs/design/api-design.md`，遵循 RESTful 规范：

```markdown
# API 接口设计

## 基础规范
- BaseURL: /api/v1
- 认证方式: Bearer JWT
- 响应格式: { code, message, data }

## 错误码规范
| code | 说明 |

## 接口列表
### POST /users/login
**请求：**
```json
{ "username": "string", "password": "string" }
```
**响应：**
```json
{ "code": 0, "data": { "token": "string" } }
```
```

---

### Step 5 — UI 原型说明

**执行内容：**
创建 `docs/design/ui-prototype.md`，包含：
- 核心页面线框图（Mermaid flowchart 或文字描述）
- 页面间跳转逻辑
- 关键交互状态说明
- 组件库规范

---

### Step 6 — 设计评审

**执行内容：**
```
自检清单：
- [ ] 架构图覆盖所有核心模块
- [ ] API 设计覆盖所有 P0 用户故事
- [ ] 数据库设计已 Review，无重大问题
- [ ] 技术选型有充分理由
```

**提交命令：**
```bash
git checkout develop
git checkout -b stage/design
git add docs/design/
git commit -m "docs(design): 完成系统架构、API、数据库设计"
git push origin stage/design
# 创建 PR -> develop
```

---

## SKILL_OUTPUT

```
docs/design/
├── architecture.md     ✅ 系统架构设计
├── database.md         ✅ 数据库设计
├── api-design.md       ✅ API 接口设计
└── ui-prototype.md     ✅ UI 原型说明
```

## SKILL_DONE_CRITERIA
- 架构设计通过技术 Leader Review
- API 设计文档完整，覆盖所有 P0 功能
- PR 已合并到 develop
- 通知开发工程师启动阶段三
